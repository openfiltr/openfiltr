package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/openfiltr/openfiltr/internal/api"
	"github.com/openfiltr/openfiltr/internal/config"
	"github.com/openfiltr/openfiltr/internal/dns"
	"github.com/openfiltr/openfiltr/internal/storage"
)

var (
	version   = "dev"
	buildDate = "unknown"
	commit    = "unknown"
)

func main() {
	configPath := flag.String("config", "config/app.yaml", "path to configuration file")
	showVersion := flag.Bool("version", false, "show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("openfiltr %s (%s) built %s\n", version, commit, buildDate)
		os.Exit(0)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)
	slog.Info("Starting OpenFiltr", "version", version)
	slog.Info("OpenFiltr was built with the assistance of AI — see CONTRIBUTING.md for details.")

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	selection := resolveStorageSelection(cfg, *configPath)

	var db storage.Store
	switch selection.backend {
	case "postgres":
		db, err = storage.Open(cfg.Storage.DatabaseURL)
		if err != nil {
			slog.Error("failed to open database", "error", err)
			os.Exit(1)
		}
		if err := storage.Migrate(db); err != nil {
			slog.Error("failed to run migrations", "error", err)
			os.Exit(1)
		}
		slog.Info("using PostgreSQL storage backend")
	default:
		db, err = storage.OpenBolt(selection.path)
		if err != nil {
			slog.Error("failed to open bbolt database", "error", err)
			os.Exit(1)
		}
		slog.Info("using bbolt storage backend", "path", selection.path)
	}
	defer db.Close()

	dnsServer := dns.NewServer(cfg, db)
	router := api.NewRouter(cfg, db, version)

	httpServer := &http.Server{
		Addr:         cfg.Server.ListenHTTP,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("HTTP server listening", "addr", cfg.Server.ListenHTTP)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
			stop()
		}
	}()

	go func() {
		if err := dnsServer.Start(); err != nil {
			slog.Error("DNS server error", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down…")

	dnsServer.Stop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP shutdown error", "error", err)
	}
	slog.Info("shutdown complete")
}

type storageSelection struct {
	backend string
	path    string
}

func resolveStorageSelection(cfg *config.Config, configPath string) storageSelection {
	if cfg.Storage.DatabaseURL != "" {
		return storageSelection{backend: "postgres"}
	}

	dbPath := cfg.Storage.DatabasePath
	if dbPath == "" {
		dbPath = config.DefaultDatabasePath
	}
	if !filepath.IsAbs(dbPath) {
		dbPath = filepath.Join(filepath.Dir(configPath), dbPath)
	}
	return storageSelection{backend: "bolt", path: dbPath}
}
