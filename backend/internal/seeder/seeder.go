// Package seeder fills the database with demo master data and stock.
//
// It runs automatically on boot (SEED_ENABLED=true) and is idempotent: tables
// that already contain rows are left untouched, so restarting the API never
// duplicates data.
package seeder

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Seeder generates the demo dataset.
type Seeder struct {
	pool *pgxpool.Pool
	rng  *rand.Rand
}

func New(pool *pgxpool.Pool) *Seeder {
	// Fixed seed: the generated dataset is identical on every machine.
	return &Seeder{pool: pool, rng: rand.New(rand.NewSource(20261019))}
}

var categoryNames = []string{
	"Consumable",
	"Sparepart",
	"Material Konstruksi",
	"Alat Tulis Kantor",
	"Packaging",
	"Alat Listrik",
	"Alat Keselamatan (K3)",
	"Chemical",
	"Perkakas Tangan",
	"Aset IT",
	"Furniture",
	"Suku Cadang Kendaraan",
}

var zones = []string{"A", "B", "C", "D", "E", "F"}

var locationTypes = []string{"rack", "bin", "pallet", "cold_storage", "staging"}

var units = []string{"pcs", "box", "roll", "batang", "sak", "liter", "lusin", "unit", "set", "meter"}

var itemNouns = []string{
	"Besi Hollow", "Kabel NYM", "Lakban", "Kardus Packing", "Sarung Tangan",
	"Helm Safety", "Cat Tembok", "Paku Beton", "Mur Baut", "Bearing",
	"Oli Mesin", "Filter Udara", "Lampu LED", "Saklar", "Kunci Inggris",
	"Tang Potong", "Obeng", "Masker", "Pipa PVC", "Semen",
	"Pasir Silika", "Keramik", "Triplek", "Papan Kayu", "Engsel Pintu",
	"Gerinda", "Mata Bor", "Kuas", "Terpal", "Selang Air",
	"Baterai", "Kabel Data", "Mouse", "Keyboard", "Monitor",
	"Printer Ribbon", "Kertas HVS", "Tinta Printer", "Pulpen", "Buku Folio",
}

var itemVariants = []string{
	"Galvanis", "Premium", "Standar", "Heavy Duty", "Anti Rust",
	"Industrial", "Ekonomis", "Tahan Panas", "Waterproof", "Original",
}

var itemSizes = []string{
	"4x4", "2 inch", "3/4 inch", "5 mm", "10 mm", "12 mm", "1.5 mm2",
	"2.5 mm2", "A4", "A3", "500 ml", "1 L", "5 kg", "25 kg", "M8",
	"M10", "XL", "L", "M", "60x60",
}

// Run seeds categories, locations, items and stock when they are still empty.
func (s *Seeder) Run(ctx context.Context, itemCount int) error {
	if itemCount <= 0 {
		itemCount = 1000
	}

	start := time.Now()

	categoryIDs, err := s.seedCategories(ctx)
	if err != nil {
		return err
	}

	locationIDs, err := s.seedLocations(ctx)
	if err != nil {
		return err
	}

	itemIDs, seeded, err := s.seedItems(ctx, itemCount, categoryIDs)
	if err != nil {
		return err
	}

	if seeded {
		if err := s.seedStocks(ctx, itemIDs, locationIDs); err != nil {
			return err
		}
	}

	slog.Info("seeder finished",
		"categories", len(categoryIDs),
		"locations", len(locationIDs),
		"items", len(itemIDs),
		"items_inserted", seeded,
		"duration", time.Since(start).String(),
	)
	return nil
}

func (s *Seeder) seedCategories(ctx context.Context) ([]int64, error) {
	existing, err := s.ids(ctx, `SELECT id FROM categories ORDER BY id`)
	if err != nil {
		return nil, err
	}
	if len(existing) > 0 {
		return existing, nil
	}

	rows := make([][]any, 0, len(categoryNames))
	for i, name := range categoryNames {
		status := "active"
		// Keep two inactive categories so the "inactive category" rule is testable.
		if i == len(categoryNames)-1 || i == len(categoryNames)-2 {
			status = "inactive"
		}
		rows = append(rows, []any{name, status})
	}

	if _, err := s.pool.CopyFrom(ctx,
		pgx.Identifier{"categories"},
		[]string{"name", "status"},
		pgx.CopyFromRows(rows),
	); err != nil {
		return nil, fmt.Errorf("seed categories: %w", err)
	}

	return s.ids(ctx, `SELECT id FROM categories WHERE status = 'active' ORDER BY id`)
}

func (s *Seeder) seedLocations(ctx context.Context) ([]int64, error) {
	existing, err := s.ids(ctx, `SELECT id FROM locations ORDER BY id`)
	if err != nil {
		return nil, err
	}
	if len(existing) > 0 {
		return existing, nil
	}

	rows := make([][]any, 0, 72)
	for _, zone := range zones {
		for aisle := 1; aisle <= 3; aisle++ {
			for slot := 1; slot <= 4; slot++ {
				locType := locationTypes[(aisle+slot)%len(locationTypes)]
				code := fmt.Sprintf("%s-%02d-%s%02d", zone, aisle, prefixOf(locType), slot)
				rows = append(rows, []any{code, zone, locType})
			}
		}
	}

	if _, err := s.pool.CopyFrom(ctx,
		pgx.Identifier{"locations"},
		[]string{"code", "zone", "type"},
		pgx.CopyFromRows(rows),
	); err != nil {
		return nil, fmt.Errorf("seed locations: %w", err)
	}

	return s.ids(ctx, `SELECT id FROM locations ORDER BY id`)
}

func prefixOf(locType string) string {
	switch locType {
	case "rack":
		return "R"
	case "bin":
		return "B"
	case "pallet":
		return "P"
	case "cold_storage":
		return "C"
	default:
		return "S"
	}
}

// seedItems inserts itemCount items when the table is empty. It returns the ids
// of all items and whether new rows were inserted.
func (s *Seeder) seedItems(ctx context.Context, itemCount int, categoryIDs []int64) ([]int64, bool, error) {
	existing, err := s.ids(ctx, `SELECT id FROM items ORDER BY id`)
	if err != nil {
		return nil, false, err
	}
	if len(existing) > 0 {
		slog.Info("seeder: items table already populated, skipping", "rows", len(existing))
		return existing, false, nil
	}
	if len(categoryIDs) == 0 {
		return nil, false, fmt.Errorf("seed items: no active category available")
	}

	rows := make([][]any, 0, itemCount)
	for i := 1; i <= itemCount; i++ {
		noun := itemNouns[s.rng.Intn(len(itemNouns))]
		variant := itemVariants[s.rng.Intn(len(itemVariants))]
		size := itemSizes[s.rng.Intn(len(itemSizes))]

		sku := fmt.Sprintf("SKU-%05d", i)
		name := fmt.Sprintf("%s %s %s", noun, size, variant)
		categoryID := categoryIDs[s.rng.Intn(len(categoryIDs))]
		unit := units[s.rng.Intn(len(units))]

		rows = append(rows, []any{sku, name, categoryID, unit})
	}

	copied, err := s.pool.CopyFrom(ctx,
		pgx.Identifier{"items"},
		[]string{"sku", "name", "category_id", "unit"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return nil, false, fmt.Errorf("seed items: %w", err)
	}
	slog.Info("seeder: items inserted", "rows", copied)

	ids, err := s.ids(ctx, `SELECT id FROM items ORDER BY id`)
	return ids, true, err
}

// seedStocks spreads every item over one to three locations.
func (s *Seeder) seedStocks(ctx context.Context, itemIDs, locationIDs []int64) error {
	if len(itemIDs) == 0 || len(locationIDs) == 0 {
		return nil
	}

	count, err := s.count(ctx, `SELECT COUNT(*) FROM stocks`)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	rows := make([][]any, 0, len(itemIDs)*2)
	for _, itemID := range itemIDs {
		placements := 1 + s.rng.Intn(3)
		used := make(map[int64]bool, placements)
		for p := 0; p < placements; p++ {
			locationID := locationIDs[s.rng.Intn(len(locationIDs))]
			if used[locationID] {
				continue
			}
			used[locationID] = true
			qty := int64(s.rng.Intn(500))
			rows = append(rows, []any{itemID, locationID, qty})
		}
	}

	copied, err := s.pool.CopyFrom(ctx,
		pgx.Identifier{"stocks"},
		[]string{"item_id", "location_id", "qty"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("seed stocks: %w", err)
	}
	slog.Info("seeder: stocks inserted", "rows", copied)
	return nil
}

func (s *Seeder) ids(ctx context.Context, query string) ([]int64, error) {
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("seeder query ids: %w", err)
	}
	defer rows.Close()

	ids, err := pgx.CollectRows(rows, pgx.RowTo[int64])
	if err != nil {
		return nil, fmt.Errorf("seeder scan ids: %w", err)
	}
	return ids, nil
}

func (s *Seeder) count(ctx context.Context, query string) (int64, error) {
	var total int64
	if err := s.pool.QueryRow(ctx, query).Scan(&total); err != nil {
		return 0, fmt.Errorf("seeder count: %w", err)
	}
	return total, nil
}
