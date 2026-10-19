package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nugroho/siswa/backend/internal/model"
)

// QtyNonNegativeConstraint is the database level guard mirroring the service
// level "stock must never be negative" rule.
const QtyNonNegativeConstraint = "stocks_qty_non_negative"

// StockRepository is the data access layer for the stocks table.
type StockRepository struct {
	pool *pgxpool.Pool
}

func NewStockRepository(pool *pgxpool.Pool) *StockRepository {
	return &StockRepository{pool: pool}
}

// MutationResult describes the effect of a stock mutation.
type MutationResult struct {
	StockID   int64
	QtyBefore int64
	QtyAfter  int64
	UpdatedAt time.Time
}

// ListByItem returns the stock of one item across every location.
func (r *StockRepository) ListByItem(ctx context.Context, itemID int64) ([]model.Stock, error) {
	const q = `
		SELECT s.id, s.item_id, s.location_id, s.qty, s.updated_at,
		       l.code, l.zone, l.type
		FROM stocks s
		JOIN locations l ON l.id = s.location_id
		WHERE s.item_id = $1
		ORDER BY l.code ASC`

	rows, err := r.pool.Query(ctx, q, itemID)
	if err != nil {
		return nil, fmt.Errorf("select stocks: %w", err)
	}
	defer rows.Close()

	stocks := make([]model.Stock, 0, 8)
	for rows.Next() {
		var s model.Stock
		if err := rows.Scan(&s.ID, &s.ItemID, &s.LocationID, &s.Qty, &s.UpdatedAt,
			&s.LocationCode, &s.LocationZone, &s.LocationType); err != nil {
			return nil, fmt.Errorf("scan stocks: %w", err)
		}
		stocks = append(stocks, s)
	}
	return stocks, rows.Err()
}

// ApplyMutation adds delta to the stock of (itemID, locationID) atomically.
//
// The row is created on demand and then locked with SELECT ... FOR UPDATE, so
// concurrent receipts of the same item/location are serialised. validate is
// called with the locked "before" quantity: returning an error aborts the
// transaction, which is how the service enforces the non-negative stock rule
// before the database CHECK constraint would fire.
func (r *StockRepository) ApplyMutation(
	ctx context.Context,
	itemID, locationID, delta int64,
	validate func(qtyBefore, delta int64) error,
) (*MutationResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin stock mutation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Make sure the (item_id, location_id) row exists before locking it.
	const ensure = `
		INSERT INTO stocks (item_id, location_id, qty)
		VALUES ($1, $2, 0)
		ON CONFLICT (item_id, location_id) DO NOTHING`
	if _, err := tx.Exec(ctx, ensure, itemID, locationID); err != nil {
		return nil, fmt.Errorf("ensure stock row: %w", err)
	}

	const lock = `
		SELECT id, qty FROM stocks
		WHERE item_id = $1 AND location_id = $2
		FOR UPDATE`

	var stockID, qtyBefore int64
	if err := tx.QueryRow(ctx, lock, itemID, locationID).Scan(&stockID, &qtyBefore); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("lock stock row: %w", err)
	}

	if validate != nil {
		if err := validate(qtyBefore, delta); err != nil {
			return nil, err
		}
	}

	const update = `
		UPDATE stocks
		SET qty = qty + $2, updated_at = NOW()
		WHERE id = $1
		RETURNING qty, updated_at`

	var result MutationResult
	result.StockID = stockID
	result.QtyBefore = qtyBefore
	if err := tx.QueryRow(ctx, update, stockID, delta).Scan(&result.QtyAfter, &result.UpdatedAt); err != nil {
		return nil, fmt.Errorf("update stock: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit stock mutation: %w", err)
	}
	return &result, nil
}

// Count returns the number of stock rows; used by the seeder.
func (r *StockRepository) Count(ctx context.Context) (int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM stocks`).Scan(&total); err != nil {
		return 0, fmt.Errorf("count stocks: %w", err)
	}
	return total, nil
}
