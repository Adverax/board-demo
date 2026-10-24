// Command board is the board server (COMP-004): it reads PORT (default 8080,
// CON-006), wires the Board and the HTTP API and serves them until SIGINT or
// SIGTERM (ADR-0003).
package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adverax/board-demo/internal/board"
	"github.com/adverax/board-demo/internal/httpapi"
)

const (
	defaultPort       = "8080"
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 10 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 5 * time.Second
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ln, err := net.Listen("tcp", addr(os.Getenv))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := serve(ctx, ln, httpapi.New(board.New())); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// addr is the listen address: PORT from the environment, or 8080 (CON-006).
func addr(getenv func(string) string) string {
	port := getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	return ":" + port
}

// serve runs h on ln until ctx is cancelled, then drains in-flight requests
// for at most shutdownTimeout and force-closes whatever is left.
func serve(ctx context.Context, ln net.Listener, h http.Handler) error {
	srv := &http.Server{
		Handler:           h,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
	errc := make(chan error, 1)
	go func() { errc <- srv.Serve(ln) }()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	err := srv.Shutdown(shutdownCtx)
	if err != nil {
		_ = srv.Close()
		err = fmt.Errorf("shutdown: %w", err)
	}
	if serveErr := <-errc; !errors.Is(serveErr, http.ErrServerClosed) {
		return errors.Join(err, serveErr)
	}
	return err
}
