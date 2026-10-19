package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nugroho/siswa/backend/internal/model"
)

// CategoryRepository is the data access layer for the categories table.
type CategoryRepository struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

// List returns categories ordered by name; onlyActive filters inactive ones out.
// item_count is the number of live (non soft deleted) items in the category.
func (r *CategoryRepository) List(ctx context.Context, onlyActive bool) ([]model.Category, error) {
	q := `
		SELECT c.id, c.name, c.status,
		       (SELECT COUNT(*) FROM items i
		         WHERE i.category_id = c.id AND i.deleted_at IS NULL) AS item_count
		FROM categories c`
	if onlyActive {
		q += ` WHERE c.status = 'active'`
	}
	q += ` ORDER BY c.name ASC`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("select categories: %w", err)
	}
	defer rows.Close()

	categories := make([]model.Category, 0, 32)
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Status, &c.ItemCount); err != nil {
			return nil, fmt.Errorf("scan categories: %w", err)
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

// CategoryNameIndex is the unique index guarding the category name.
const CategoryNameIndex = "ux_categories_name"

// Create inserts a new category and fills the generated id.
func (r *CategoryRepository) Create(ctx context.Context, category *model.Category) error {
	const q = `
		INSERT INTO categories (name, status)
		VALUES ($1, $2)
		RETURNING id`

	if err := r.pool.QueryRow(ctx, q, category.Name, category.Status).Scan(&category.ID); err != nil {
		return fmt.Errorf("insert category: %w", err)
	}
	return nil
}

// ExistsName reports whether a category with that name already exists
// (case insensitive).
func (r *CategoryRepository) ExistsName(ctx context.Context, name string) (bool, error) {
	const q = `SELECT EXISTS (SELECT 1 FROM categories WHERE LOWER(name) = LOWER($1))`

	var exists bool
	if err := r.pool.QueryRow(ctx, q, name).Scan(&exists); err != nil {
		return false, fmt.Errorf("check category name: %w", err)
	}
	return exists, nil
}

// GetByID returns one category.
func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (*model.Category, error) {
	const q = `SELECT id, name, status FROM categories WHERE id = $1`

	var c model.Category
	if err := r.pool.QueryRow(ctx, q, id).Scan(&c.ID, &c.Name, &c.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("select category: %w", err)
	}
	return &c, nil
}

// Count returns the number of categories; used by the seeder.
func (r *CategoryRepository) Count(ctx context.Context) (int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM categories`).Scan(&total); err != nil {
		return 0, fmt.Errorf("count categories: %w", err)
	}
	return total, nil
}

// LocationRepository is the data access layer for the locations table.
type LocationRepository struct {
	pool *pgxpool.Pool
}

func NewLocationRepository(pool *pgxpool.Pool) *LocationRepository {
	return &LocationRepository{pool: pool}
}

// List returns every warehouse location ordered by code.
// item_count is the number of distinct live items stored at the location; a slot
// that was reserved with qty 0 still counts, because the item occupies it.
func (r *LocationRepository) List(ctx context.Context) ([]model.Location, error) {
	const q = `
		SELECT l.id, l.code, l.zone, l.type,
		       (SELECT COUNT(*) FROM stocks s
		          JOIN items i ON i.id = s.item_id AND i.deleted_at IS NULL
		         WHERE s.location_id = l.id) AS item_count
		FROM locations l
		ORDER BY l.code ASC`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("select locations: %w", err)
	}
	defer rows.Close()

	locations := make([]model.Location, 0, 64)
	for rows.Next() {
		var l model.Location
		if err := rows.Scan(&l.ID, &l.Code, &l.Zone, &l.Type, &l.ItemCount); err != nil {
			return nil, fmt.Errorf("scan locations: %w", err)
		}
		locations = append(locations, l)
	}
	return locations, rows.Err()
}

// LocationCodeIndex is the unique index guarding the location code.
const LocationCodeIndex = "ux_locations_code"

// Create inserts a new warehouse location and fills the generated id.
func (r *LocationRepository) Create(ctx context.Context, location *model.Location) error {
	const q = `
		INSERT INTO locations (code, zone, type)
		VALUES ($1, $2, $3)
		RETURNING id`

	if err := r.pool.QueryRow(ctx, q, location.Code, location.Zone, location.Type).Scan(&location.ID); err != nil {
		return fmt.Errorf("insert location: %w", err)
	}
	return nil
}

// ExistsCode reports whether a location with that code already exists.
func (r *LocationRepository) ExistsCode(ctx context.Context, code string) (bool, error) {
	const q = `SELECT EXISTS (SELECT 1 FROM locations WHERE UPPER(code) = UPPER($1))`

	var exists bool
	if err := r.pool.QueryRow(ctx, q, code).Scan(&exists); err != nil {
		return false, fmt.Errorf("check location code: %w", err)
	}
	return exists, nil
}

// GetByID returns one location.
func (r *LocationRepository) GetByID(ctx context.Context, id int64) (*model.Location, error) {
	const q = `SELECT id, code, zone, type FROM locations WHERE id = $1`

	var l model.Location
	if err := r.pool.QueryRow(ctx, q, id).Scan(&l.ID, &l.Code, &l.Zone, &l.Type); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("select location: %w", err)
	}
	return &l, nil
}

// Count returns the number of locations; used by the seeder.
func (r *LocationRepository) Count(ctx context.Context) (int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM locations`).Scan(&total); err != nil {
		return 0, fmt.Errorf("count locations: %w", err)
	}
	return total, nil
}
