// Command api is the entrypoint of the SISWA backend.
//
//	SISWA - Sistem Item Stok Warehouse & Aset
//
// On boot it connects to PostgreSQL, applies the embedded migrations, connects
// to Redis (optional), seeds demo data, then serves the HTTP API on APP_PORT.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nugroho/siswa/backend/internal/cache"
	"github.com/nugroho/siswa/backend/internal/config"
	"github.com/nugroho/siswa/backend/internal/database"
	"github.com/nugroho/siswa/backend/internal/handler"
	"github.com/nugroho/siswa/backend/internal/logger"
	"github.com/nugroho/siswa/backend/internal/logstore"
	"github.com/nugroho/siswa/backend/internal/repository"
	"github.com/nugroho/siswa/backend/internal/router"
	"github.com/nugroho/siswa/backend/internal/seeder"
	"github.com/nugroho/siswa/backend/internal/service"
	"github.com/nugroho/siswa/backend/internal/validation"
)

// version is reported by /health and the root endpoint.
const version = "1.0.0"

func main() {
	if err := run(); err != nil {
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()
	logger.Setup(cfg.AppEnv, cfg.LogLevel)
	validation.Setup()

	slog.Info("starting service",
		"app", cfg.AppName,
		"env", cfg.AppEnv,
		"port", cfg.AppPort,
		"version", version,
	)

	// Cancelled on SIGINT/SIGTERM so every component shuts down cleanly.
	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// --- PostgreSQL ---------------------------------------------------------
	bootCtx, cancelBoot := context.WithTimeout(rootCtx, 90*time.Second)
	defer cancelBoot()

	pool, err := database.Connect(bootCtx, cfg)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	if err := database.Migrate(bootCtx, pool); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	// --- Redis (optional) ---------------------------------------------------
	rdb, err := cache.Connect(bootCtx, cfg)
	if err != nil {
		slog.Warn("redis unavailable, request logs will only be written to postgres", "error", err)
		rdb = nil
	}
	if rdb != nil {
		defer func() { _ = rdb.Close() }()
	}

	// --- Request log store --------------------------------------------------
	store := logstore.New(cfg, rdb, pool)
	store.StartRedisBridge(rootCtx)
	defer store.Close()

	// --- Stock mutation log (console + file) --------------------------------
	stockLogger, err := logger.NewStockLogger(cfg.StockLogFile)
	if err != nil {
		slog.Warn("stock mutation file log disabled", "error", err)
	}
	defer func() { _ = stockLogger.Close() }()

	// --- Seeder -------------------------------------------------------------
	if cfg.SeedEnabled {
		if err := seeder.New(pool).Run(bootCtx, cfg.SeedItemCount); err != nil {
			// A failed seed must not stop the API from serving.
			slog.Error("seeder failed", "error", err)
		}
	} else {
		slog.Info("seeder disabled")
	}

	// --- Wiring -------------------------------------------------------------
	itemRepo := repository.NewItemRepository(pool)
	categoryRepo := repository.NewCategoryRepository(pool)
	locationRepo := repository.NewLocationRepository(pool)
	stockRepo := repository.NewStockRepository(pool)

	itemService := service.NewItemService(itemRepo, categoryRepo, locationRepo, stockLogger)
	stockService := service.NewStockService(stockRepo, itemRepo, locationRepo, stockLogger)

	engine := router.New(cfg, store, router.Handlers{
		Item:   handler.NewItemHandler(itemService),
		Stock:  handler.NewStockHandler(stockService),
		Log:    handler.NewLogHandler(store),
		System: handler.NewSystemHandler(pool, rdb, cfg.AppName, version),
	})

	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
		// No WriteTimeout: the SSE log stream is a long lived response.
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("http server listening",
			"addr", server.Addr,
			"swagger", "http://localhost:"+cfg.AppPort+"/swagger/index.html",
		)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("http server: %w", err)
	case <-rootCtx.Done():
		slog.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	slog.Info("service stopped")
	return nil
}
