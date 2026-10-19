package logger

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Mutation types recorded in the stock audit log.
const (
	MutationReceive   = "RECEIVE"
	MutationPlacement = "PLACEMENT"
)

// Mutation is one line of the stock transaction log.
//
// Business rule: "setiap mutasi stok (receive) harus mencatat log transaksi
// minimal di level aplikasi (console/file)" - every entry is therefore written
// to stdout (structured) and appended as JSON to STOCK_LOG_FILE.
type Mutation struct {
	RequestID    string    `json:"request_id,omitempty"`
	Type         string    `json:"type"`
	ItemID       int64     `json:"item_id"`
	SKU          string    `json:"sku"`
	LocationID   int64     `json:"location_id"`
	LocationCode string    `json:"location_code"`
	QtyBefore    int64     `json:"qty_before"`
	QtyDelta     int64     `json:"qty_delta"`
	QtyAfter     int64     `json:"qty_after"`
	Reference    string    `json:"reference,omitempty"`
	Note         string    `json:"note,omitempty"`
	At           time.Time `json:"at"`
}

// StockLogger appends stock mutations to a file and mirrors them on the console.
type StockLogger struct {
	mu   sync.Mutex
	file *os.File
	path string
}

// NewStockLogger opens (and creates) the append only mutation log file. When the
// file cannot be opened the logger still works in console-only mode.
func NewStockLogger(path string) (*StockLogger, error) {
	if path == "" {
		return &StockLogger{}, nil
	}

	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return &StockLogger{}, fmt.Errorf("create stock log dir: %w", err)
		}
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return &StockLogger{}, fmt.Errorf("open stock log file: %w", err)
	}

	return &StockLogger{file: file, path: path}, nil
}

// Log records one mutation on the console and in the audit file.
func (s *StockLogger) Log(m Mutation) {
	if m.At.IsZero() {
		m.At = time.Now().UTC()
	}
	if m.Type == "" {
		m.Type = MutationReceive
	}

	slog.Info("stock mutation",
		"type", m.Type,
		"request_id", m.RequestID,
		"item_id", m.ItemID,
		"sku", m.SKU,
		"location_id", m.LocationID,
		"location_code", m.LocationCode,
		"qty_before", m.QtyBefore,
		"qty_delta", m.QtyDelta,
		"qty_after", m.QtyAfter,
		"reference", m.Reference,
		"at", m.At.Format(time.RFC3339Nano),
	)

	if s.file == nil {
		return
	}

	line, err := json.Marshal(m)
	if err != nil {
		slog.Error("marshal stock mutation", "error", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.file.Write(append(line, '\n')); err != nil {
		slog.Error("write stock mutation log", "error", err, "path", s.path)
	}
}

// Close releases the audit file handle.
func (s *StockLogger) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file == nil {
		return nil
	}
	err := s.file.Close()
	s.file = nil
	return err
}
