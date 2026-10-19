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
func (r *CategoryRepository) List(ctx context.Context, onlyActive bool) ([]model.Category, error) {
	q := `SELECT id, name, status FROM categories`
	if onlyActive {
		q += ` WHERE status = 'active'`
	}
	q += ` ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("select categories: %w", err)
	}
	defer rows.Close()

	categories := make([]model.Category, 0, 32)
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Status); err != nil {
			return nil, fmt.Errorf("scan categories: %w", err)
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
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
func (r *LocationRepository) List(ctx context.Context) ([]model.Location, error) {
	const q = `SELECT id, code, zone, type FROM locations ORDER BY code ASC`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("select locations: %w", err)
	}
	defer rows.Close()

	locations := make([]model.Location, 0, 64)
	for rows.Next() {
		var l model.Location
		if err := rows.Scan(&l.ID, &l.Code, &l.Zone, &l.Type); err != nil {
			return nil, fmt.Errorf("scan locations: %w", err)
		}
		locations = append(locations, l)
	}
	return locations, rows.Err()
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
