package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lanefedov/metrics/internal/handler"
	"github.com/lanefedov/metrics/internal/server"
	"github.com/lanefedov/metrics/internal/service"
	"github.com/lanefedov/metrics/internal/storage"
)

func main() {
	cfg, err := loadServerConfig(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var store storage.MetricsStorage
	var fileStore *storage.MemStorage
	var databasePing handler.PingFunc
	switch cfg.storageMode {
	case storageModeDatabase:
		if err := runDatabaseMigrations(cfg.databaseDSN); err != nil {
			log.Fatal(err)
		}

		pool, err := pgxpool.New(context.Background(), cfg.databaseDSN)
		if err != nil {
			log.Fatal(err)
		}
		defer pool.Close()

		databasePing = pool.Ping
		store = storage.NewPostgresStorage(pool)
	case storageModeFile:
		fileStore = storage.NewMemStorage()
		if cfg.restore {
			if err := fileStore.LoadFromFile(cfg.fileStoragePath); err != nil {
				log.Fatal(err)
			}
		}
		store = fileStore
	default:
		store = storage.NewMemStorage()
	}

	metricsService := service.NewMetricsService(store)
	if fileStore != nil {
		if cfg.storeInterval == 0 {
			metricsService = service.NewMetricsServiceWithAfterUpdate(store, func() error {
				return fileStore.SaveToFile(cfg.fileStoragePath)
			})
		} else {
			startPeriodicSave(ctx, fileStore, cfg.fileStoragePath, cfg.storeInterval)
		}
	}

	h := server.NewHandler(metricsService, cfg.key, databasePing)
	httpServer := &http.Server{
		Addr:              cfg.address,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("shutdown server: %v", err)
		}
	}()

	log.Printf("listening on %s", cfg.address)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

	if fileStore == nil {
		return
	}
	if err := fileStore.SaveToFile(cfg.fileStoragePath); err != nil {
		log.Printf("save metrics: %v", err)
	}
}

func startPeriodicSave(ctx context.Context, store *storage.MemStorage, path string, interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := store.SaveToFile(path); err != nil {
					log.Printf("save metrics: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
