// Command server runs the Bushwhack web service.
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brennanMKE/Bushwhack/internal/server"
	"github.com/brennanMKE/Bushwhack/patterns"
	"github.com/brennanMKE/Bushwhack/web"
)

var (
	version = "dev"
	updated = "" // YYYY-MM-DD of the last commit, for the sitemap
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := flag.String("addr", net.JoinHostPort("127.0.0.1", port), "listen address (default 127.0.0.1:$PORT)")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	lib, err := patterns.Load()
	if err != nil {
		log.Error("patterns", "err", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go lib.ComputeSafeSizes(ctx)
	srv := &http.Server{
		Addr: *addr,
		Handler: server.New(server.Config{
			Static:   web.FS(),
			Patterns: lib,
			Version:  version,
			Updated:  updated,
			Log:      log,
		}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       90 * time.Second,
		MaxHeaderBytes:    64 << 10,
	}

	go func() {
		log.Info("listening", "addr", *addr, "version", version)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server", "err", err)
			os.Exit(1)
		}
	}()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdown); err != nil {
		log.Error("shutdown", "err", err)
	}
}
