package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"bookmarkd"
	"bookmarkd/internal/core"
	"bookmarkd/internal/inmem"
	"bookmarkd/internal/server"
	"bookmarkd/internal/sqlite"

	"github.com/go-chi/httplog/v2"
)

// Build version, injected during build.
var (
	version string
	commit  string
)

func main() {
	// Propagate build information to root package.
	bookmarkd.Version = strings.TrimPrefix(version, "")
	bookmarkd.Commit = commit

	ctx := context.Background()
	if err := run(ctx, os.Getenv, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(
	ctx context.Context,
	getenv func(string) string,
	stdout io.Writer,
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	logger := httplog.NewLogger("bookmarkd", httplog.Options{
		// JSON:             true,
		LogLevel:         slog.LevelDebug,
		Concise:          true,
		RequestHeaders:   true,
		MessageFieldName: "msg",
		// TimeFieldFormat: time.RFC850,
		Tags: map[string]string{
			"version": version,
			"env":     "dev",
		},
		QuietDownRoutes: []string{
			"/",
			"/ping",
		},
		QuietDownPeriod: 10 * time.Second,
		Writer:          stdout,
	})

	config, err := core.NewConfig(getenv)
	if err != nil {
		return err
	}

	// Initialize error tracking.
	// if config.Rollbar.Token != "" {
	// 	rollbar.SetToken(config.Rollbar.Token)
	// 	rollbar.SetEnvironment("production")
	// 	rollbar.SetCodeVersion(version)
	// 	rollbar.SetServerRoot("github.com/nickv0/bookmarkd")
	// 	core.ReportError = rollbarReportError
	// 	core.ReportPanic = rollbarReportPanic
	// 	log.Printf("rollbar error tracking enabled")
	// }

	db := sqlite.NewDB(config.DbDsn)
	if err := db.Open(); err != nil {
		return fmt.Errorf("cannot open db: %w", err)
	}

	// inmem
	registrationStore := inmem.NewRegistrationStore()
	eventService := inmem.NewEventService()

	// sqlite
	bookmarkStore := sqlite.NewBookmarkStore(db)
	sessionService := sqlite.NewSessionStore(db)
	userStore := sqlite.NewUserStore(db)

	db.EventService = eventService

	httpServer := server.NewServer(
		logger,
		config,
		registrationStore,
		bookmarkStore,
		eventService,
		sessionService,
		userStore,
	)

	go func() {
		logger.Info("listening", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		// make a new context for the Shutdown (thanks Alessandro Rosetti)
		// shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()
	return nil
}
