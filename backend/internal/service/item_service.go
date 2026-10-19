// Package service holds the business rules of the SISWA API.
//
// Validation that matters for the business (unique SKU, usable category,
// non-negative stock) lives here, not only in the database, so the rules hold
// no matter which transport calls them.
package service

import (
	"context"
	"errors"

	"github.com/nugroho/siswa/backend/internal/dto"
	"github.com/nugroho/siswa/backend/internal/errs"
	"github.com/nugroho/siswa/backend/internal/model"
	"github.com/nugroho/siswa/backend/internal/repository"
)

// ItemRepo is the persistence contract required by ItemService.
type ItemRepo interface {
	Create(ctx context.Context, item *model.Item) error
	GetByID(ctx context.Context, id int64, includeDeleted bool) (*model.Item, error)
	List(ctx context.Context, q dto.ListItemQuery) ([]model.Item, int64, error)
	Update(ctx context.Context, item *model.Item) error
	SoftDelete(ctx context.Context, id int64) error
	ExistsSKU(ctx context.Context, sku string, excludeID int64) (bool, error)
}

// CategoryReader is the slice of the category repository the services need.
type CategoryReader interface {
	GetByID(ctx context.Context, id int64) (*model.Category, error)
	List(ctx context.Context, onlyActive bool) ([]model.Category, error)
}

// ItemService implements the item use cases.
type ItemService struct {
	items      ItemRepo
	categories CategoryReader
}

func NewItemService(items ItemRepo, categories CategoryReader) *ItemService {
	return &ItemService{items: items, categories: categories}
}

// Create registers a new item.
//
// Rules: the SKU must be unique among live items (duplicate -> 409 / code 1003)
// and the category must exist and be active.
func (s *ItemService) Create(ctx context.Context, req dto.CreateItemRequest) (*model.Item, error) {
	req.Normalize()

	if err := s.assertCategoryUsable(ctx, req.CategoryID); err != nil {
		return nil, err
	}

	exists, err := s.items.ExistsSKU(ctx, req.SKU, 0)
	if err != nil {
		return nil, errs.Database(err)
	}
	if exists {
		return nil, errs.DuplicateSKU(req.SKU)
	}

	item := &model.Item{
		SKU:        req.SKU,
		Name:       req.Name,
		CategoryID: req.CategoryID,
		Unit:       req.Unit,
	}

	if err := s.items.Create(ctx, item); err != nil {
		// A concurrent insert can still win the race; the partial unique index
		// is the final authority.
		if repository.IsUniqueViolation(err, repository.SKUActiveIndex) {
			return nil, errs.DuplicateSKU(req.SKU)
		}
		if repository.IsForeignKeyViolation(err) {
			return nil, errs.InvalidCategory(req.CategoryID, "Category does not exist")
		}
		return nil, errs.Database(err)
	}

	// Re-read so the response carries category_name and total_qty.
	created, err := s.items.GetByID(ctx, item.ID, false)
	if err != nil {
		return item, nil
	}
	return created, nil
}

// List returns a filtered, paginated page of live items.
func (s *ItemService) List(ctx context.Context, q dto.ListItemQuery) ([]model.Item, int64, error) {
	q.Sanitize()

	items, total, err := s.items.List(ctx, q)
	if err != nil {
		return nil, 0, errs.Database(err)
	}
	return items, total, nil
}

// Get returns a single live item.
func (s *ItemService) Get(ctx context.Context, id int64) (*model.Item, error) {
	item, err := s.items.GetByID(ctx, id, false)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errs.NotFound("Item", id)
		}
		return nil, errs.Database(err)
	}
	return item, nil
}

// Update replaces the mutable fields of a live item.
func (s *ItemService) Update(ctx context.Context, id int64, req dto.UpdateItemRequest) (*model.Item, error) {
	req.Normalize()

	if _, err := s.Get(ctx, id); err != nil {
		return nil, err
	}

	if err := s.assertCategoryUsable(ctx, req.CategoryID); err != nil {
		return nil, err
	}

	exists, err := s.items.ExistsSKU(ctx, req.SKU, id)
	if err != nil {
		return nil, errs.Database(err)
	}
	if exists {
		return nil, errs.DuplicateSKU(req.SKU)
	}

	item := &model.Item{
		ID:         id,
		SKU:        req.SKU,
		Name:       req.Name,
		CategoryID: req.CategoryID,
		Unit:       req.Unit,
	}

	if err := s.items.Update(ctx, item); err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return nil, errs.NotFound("Item", id)
		case repository.IsUniqueViolation(err, repository.SKUActiveIndex):
			return nil, errs.DuplicateSKU(req.SKU)
		case repository.IsForeignKeyViolation(err):
			return nil, errs.InvalidCategory(req.CategoryID, "Category does not exist")
		default:
			return nil, errs.Database(err)
		}
	}

	updated, err := s.items.GetByID(ctx, id, false)
	if err != nil {
		return item, nil
	}
	return updated, nil
}

// Delete soft deletes an item: the row stays in the database and only receives
// its deleted_at flag.
func (s *ItemService) Delete(ctx context.Context, id int64) error {
	if err := s.items.SoftDelete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return errs.NotFound("Item", id)
		}
		return errs.Database(err)
	}
	return nil
}

// Categories exposes the category master data (used by the frontend dropdown).
func (s *ItemService) Categories(ctx context.Context, onlyActive bool) ([]model.Category, error) {
	categories, err := s.categories.List(ctx, onlyActive)
	if err != nil {
		return nil, errs.Database(err)
	}
	return categories, nil
}

func (s *ItemService) assertCategoryUsable(ctx context.Context, categoryID int64) error {
	category, err := s.categories.GetByID(ctx, categoryID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return errs.InvalidCategory(categoryID, "Category does not exist")
		}
		return errs.Database(err)
	}
	if category.Status != model.CategoryStatusActive {
		return errs.InvalidCategory(categoryID, "Category is inactive")
	}
	return nil
}
