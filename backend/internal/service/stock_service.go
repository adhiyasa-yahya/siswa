package service

import (
	"context"
	"errors"
	"time"

	"github.com/nugroho/siswa/backend/internal/dto"
	"github.com/nugroho/siswa/backend/internal/errs"
	"github.com/nugroho/siswa/backend/internal/logger"
	"github.com/nugroho/siswa/backend/internal/model"
	"github.com/nugroho/siswa/backend/internal/repository"
)

// StockRepo is the persistence contract required by StockService.
type StockRepo interface {
	ListByItem(ctx context.Context, itemID int64) ([]model.Stock, error)
	ApplyMutation(ctx context.Context, itemID, locationID, delta int64,
		validate func(qtyBefore, delta int64) error) (*repository.MutationResult, error)
}

// LocationReader is the slice of the location repository the services need.
type LocationReader interface {
	GetByID(ctx context.Context, id int64) (*model.Location, error)
	List(ctx context.Context) ([]model.Location, error)
	Create(ctx context.Context, location *model.Location) error
	ExistsCode(ctx context.Context, code string) (bool, error)
}

// MutationLogger records stock transactions (console + file).
type MutationLogger interface {
	Log(m logger.Mutation)
}

// StockService implements the stock use cases.
type StockService struct {
	stocks    StockRepo
	items     ItemRepo
	locations LocationReader
	mutations MutationLogger
}

func NewStockService(stocks StockRepo, items ItemRepo, locations LocationReader, mutations MutationLogger) *StockService {
	return &StockService{stocks: stocks, items: items, locations: locations, mutations: mutations}
}

const timeLayout = "2006-01-02T15:04:05Z07:00"

// Receive books incoming stock for one item into one location.
//
// Rules enforced here, at service level:
//   - qty must be a positive integer;
//   - the item must exist and must not be soft deleted;
//   - the location must exist;
//   - the resulting quantity must never be negative (checked inside the
//     transaction, while the stock row is locked);
//   - every successful mutation is written to the transaction log
//     (console + STOCK_LOG_FILE).
func (s *StockService) Receive(ctx context.Context, req dto.ReceiveStockRequest, requestID string) (*dto.ReceiveStockResponse, error) {
	if req.Qty <= 0 {
		return nil, errs.InvalidQuantity("Quantity to receive must be greater than 0")
	}

	item, err := s.items.GetByID(ctx, req.ItemID, false)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errs.NotFoundMsg("Item does not exist or has been deleted")
		}
		return nil, errs.Database(err)
	}

	location, err := s.locations.GetByID(ctx, req.LocationID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errs.InvalidLocation(req.LocationID)
		}
		return nil, errs.Database(err)
	}

	result, err := s.stocks.ApplyMutation(ctx, req.ItemID, req.LocationID, req.Qty,
		func(qtyBefore, delta int64) error {
			// Service level guard: the stock may never become negative.
			if qtyBefore+delta < 0 {
				return errs.NegativeStock(req.ItemID, req.LocationID, qtyBefore, delta)
			}
			return nil
		})
	if err != nil {
		if appErr, ok := errs.As(err); ok {
			return nil, appErr
		}
		if repository.IsCheckViolation(err, repository.QtyNonNegativeConstraint) {
			return nil, errs.NegativeStock(req.ItemID, req.LocationID, 0, req.Qty)
		}
		if repository.IsForeignKeyViolation(err) {
			return nil, errs.InvalidLocation(req.LocationID)
		}
		return nil, errs.Database(err)
	}

	s.mutations.Log(logger.Mutation{
		RequestID:    requestID,
		Type:         logger.MutationReceive,
		ItemID:       item.ID,
		SKU:          item.SKU,
		LocationID:   location.ID,
		LocationCode: location.Code,
		QtyBefore:    result.QtyBefore,
		QtyDelta:     req.Qty,
		QtyAfter:     result.QtyAfter,
		Reference:    req.Reference,
		Note:         req.Note,
		At:           time.Now().UTC(),
	})

	return &dto.ReceiveStockResponse{
		StockID:      result.StockID,
		ItemID:       item.ID,
		SKU:          item.SKU,
		LocationID:   location.ID,
		LocationCode: location.Code,
		QtyBefore:    result.QtyBefore,
		QtyReceived:  req.Qty,
		QtyAfter:     result.QtyAfter,
		Reference:    req.Reference,
		UpdatedAt:    result.UpdatedAt.UTC().Format(timeLayout),
	}, nil
}

// StockByItem returns the stock of one item across every location.
func (s *StockService) StockByItem(ctx context.Context, itemID int64) (*dto.ItemStockResponse, error) {
	item, err := s.items.GetByID(ctx, itemID, false)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errs.NotFound("Item", itemID)
		}
		return nil, errs.Database(err)
	}

	stocks, err := s.stocks.ListByItem(ctx, itemID)
	if err != nil {
		return nil, errs.Database(err)
	}

	locations := make([]dto.StockPerLoc, 0, len(stocks))
	var total int64
	for _, st := range stocks {
		total += st.Qty
		locations = append(locations, dto.StockPerLoc{
			StockID:      st.ID,
			LocationID:   st.LocationID,
			LocationCode: st.LocationCode,
			Zone:         st.LocationZone,
			Type:         st.LocationType,
			Qty:          st.Qty,
			UpdatedAt:    st.UpdatedAt.UTC().Format(timeLayout),
		})
	}

	return &dto.ItemStockResponse{
		ItemID:    item.ID,
		SKU:       item.SKU,
		Name:      item.Name,
		Unit:      item.Unit,
		TotalQty:  total,
		Locations: locations,
	}, nil
}

// Locations exposes the location master data (used by the receive form).
func (s *StockService) Locations(ctx context.Context) ([]model.Location, error) {
	locations, err := s.locations.List(ctx)
	if err != nil {
		return nil, errs.Database(err)
	}
	return locations, nil
}

// CreateLocation registers a new warehouse location (rack / zone).
// The code must be unique (duplicate -> 409 / code 1012).
func (s *StockService) CreateLocation(ctx context.Context, req dto.CreateLocationRequest) (*model.Location, error) {
	req.Normalize()

	exists, err := s.locations.ExistsCode(ctx, req.Code)
	if err != nil {
		return nil, errs.Database(err)
	}
	if exists {
		return nil, errs.DuplicateLocation(req.Code)
	}

	location := &model.Location{Code: req.Code, Zone: req.Zone, Type: req.Type}

	if err := s.locations.Create(ctx, location); err != nil {
		if repository.IsUniqueViolation(err, repository.LocationCodeIndex) {
			return nil, errs.DuplicateLocation(req.Code)
		}
		if repository.IsCheckViolation(err, "locations_type_check") {
			return nil, errs.Validation("Validation failed",
				errs.FieldError{Field: "type", Reason: "Must be one of: rack, bin, pallet, cold_storage, staging"})
		}
		return nil, errs.Database(err)
	}
	return location, nil
}
