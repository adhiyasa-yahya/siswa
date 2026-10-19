package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nugroho/siswa/backend/internal/dto"
	"github.com/nugroho/siswa/backend/internal/model"
)

// SKUActiveIndex is the partial unique index guarding SKU uniqueness on live rows.
const SKUActiveIndex = "ux_items_sku_active"

// ItemRepository is the data access layer for the items table.
type ItemRepository struct {
	pool *pgxpool.Pool
}

func NewItemRepository(pool *pgxpool.Pool) *ItemRepository {
	return &ItemRepository{pool: pool}
}

const itemColumns = `
	i.id, i.sku, i.name, i.category_id, c.name AS category_name, i.unit,
	i.created_at, i.updated_at, i.deleted_at,
	COALESCE((SELECT SUM(s.qty) FROM stocks s WHERE s.item_id = i.id), 0) AS total_qty,
	COALESCE((
		SELECT JSON_AGG(
			JSON_BUILD_OBJECT('location_id', l.id, 'code', l.code, 'zone', l.zone, 'qty', s.qty)
			ORDER BY l.code
		)
		FROM stocks s
		JOIN locations l ON l.id = s.location_id
		WHERE s.item_id = i.id
	), '[]') AS locations`

// Placement is the optional initial stock position of a new item.
type Placement struct {
	LocationID int64
	Qty        int64
}

// Create inserts a new item and fills the generated id/timestamps.
//
// When placement is not nil the stock row is created in the same transaction,
// so an item is never left half-placed.
func (r *ItemRepository) Create(ctx context.Context, item *model.Item, placement *Placement) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin create item: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const insertItem = `
		INSERT INTO items (sku, name, category_id, unit)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	if err := tx.QueryRow(ctx, insertItem, item.SKU, item.Name, item.CategoryID, item.Unit).
		Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return fmt.Errorf("insert item: %w", err)
	}

	if placement != nil {
		const insertStock = `
			INSERT INTO stocks (item_id, location_id, qty)
			VALUES ($1, $2, $3)`

		if _, err := tx.Exec(ctx, insertStock, item.ID, placement.LocationID, placement.Qty); err != nil {
			return fmt.Errorf("place item stock: %w", err)
		}
		item.TotalQty = placement.Qty
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit create item: %w", err)
	}
	return nil
}

// GetByID returns a single item. Soft deleted rows are excluded unless
// includeDeleted is true.
func (r *ItemRepository) GetByID(ctx context.Context, id int64, includeDeleted bool) (*model.Item, error) {
	q := `
		SELECT ` + itemColumns + `
		FROM items i
		JOIN categories c ON c.id = i.category_id
		WHERE i.id = $1`
	if !includeDeleted {
		q += ` AND i.deleted_at IS NULL`
	}

	row := r.pool.QueryRow(ctx, q, id)

	item, err := scanItem(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("select item: %w", err)
	}
	return item, nil
}

// List returns one page of items plus the total number of matching rows.
func (r *ItemRepository) List(ctx context.Context, q dto.ListItemQuery) ([]model.Item, int64, error) {
	// The sort column/direction are whitelisted by dto.ListItemQuery.Sanitize().
	where := `
		WHERE i.deleted_at IS NULL
		  AND ($1 = '' OR i.sku ILIKE '%' || $1 || '%' OR i.name ILIKE '%' || $1 || '%')
		  AND ($2::bigint IS NULL OR i.category_id = $2::bigint)`

	var categoryID any
	if q.CategoryID != nil {
		categoryID = *q.CategoryID
	}

	var total int64
	countQ := `SELECT COUNT(*) FROM items i ` + where
	if err := r.pool.QueryRow(ctx, countQ, q.Search, categoryID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count items: %w", err)
	}
	if total == 0 {
		return []model.Item{}, 0, nil
	}

	listQ := fmt.Sprintf(`
		SELECT %s
		FROM items i
		JOIN categories c ON c.id = i.category_id
		%s
		ORDER BY i.%s %s, i.id %s
		LIMIT $3 OFFSET $4`, itemColumns, where, q.SortBy, q.SortDir, q.SortDir)

	rows, err := r.pool.Query(ctx, listQ, q.Search, categoryID, q.Limit, q.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("select items: %w", err)
	}
	defer rows.Close()

	items := make([]model.Item, 0, q.Limit)
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan items: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate items: %w", err)
	}

	return items, total, nil
}

// Update overwrites the mutable fields of a live item.
func (r *ItemRepository) Update(ctx context.Context, item *model.Item) error {
	const q = `
		UPDATE items
		SET sku = $2, name = $3, category_id = $4, unit = $5, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, q, item.ID, item.SKU, item.Name, item.CategoryID, item.Unit).
		Scan(&item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("update item: %w", err)
	}
	return nil
}

// SoftDelete flags the item as deleted without removing the row.
func (r *ItemRepository) SoftDelete(ctx context.Context, id int64) error {
	const q = `
		UPDATE items
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`

	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("soft delete item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ExistsSKU reports whether another live item already uses the SKU.
// excludeID is ignored when zero (create flow).
func (r *ItemRepository) ExistsSKU(ctx context.Context, sku string, excludeID int64) (bool, error) {
	const q = `
		SELECT EXISTS (
			SELECT 1 FROM items
			WHERE sku = $1 AND deleted_at IS NULL AND ($2::bigint IS NULL OR id <> $2::bigint)
		)`

	var exclude any
	if excludeID > 0 {
		exclude = excludeID
	}

	var exists bool
	if err := r.pool.QueryRow(ctx, q, sku, exclude).Scan(&exists); err != nil {
		return false, fmt.Errorf("check sku: %w", err)
	}
	return exists, nil
}

// Count returns the number of rows in items (including soft deleted ones); used
// by the seeder to decide whether seeding is needed.
func (r *ItemRepository) Count(ctx context.Context) (int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM items`).Scan(&total); err != nil {
		return 0, fmt.Errorf("count all items: %w", err)
	}
	return total, nil
}

// rowScanner is satisfied by both pgx.Row and pgx.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanItem(row rowScanner) (*model.Item, error) {
	var item model.Item
	var locations []byte

	err := row.Scan(
		&item.ID, &item.SKU, &item.Name, &item.CategoryID, &item.CategoryName, &item.Unit,
		&item.CreatedAt, &item.UpdatedAt, &item.DeletedAt, &item.TotalQty, &locations,
	)
	if err != nil {
		return nil, err
	}

	// locations is the JSON_AGG array built by itemColumns.
	item.Locations = []model.ItemLocation{}
	if len(locations) > 0 {
		if err := json.Unmarshal(locations, &item.Locations); err != nil {
			return nil, fmt.Errorf("decode item locations: %w", err)
		}
	}

	return &item, nil
}
