package graceful

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const shutdownTimeout = 10 * time.Second

// Server is a long-running service that can be started and stopped gracefully.
type Server interface {
	Label() string
	Start() error
	Shutdown(ctx context.Context) error
}

// Context returns a context that is canceled on SIGINT or SIGTERM,
// along with a cancel function for manual cancellation.
func Context() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-sigCh:
			slog.Info("received signal, shutting down", "signal", sig)
			cancel()
		case <-ctx.Done():
		}
		signal.Stop(sigCh)
	}()

	return ctx, cancel
}

// Serve starts all servers and blocks until the context is canceled.
// On cancellation, it shuts down all servers concurrently with a bounded timeout.
func Serve(ctx context.Context, servers ...Server) {
	var wg sync.WaitGroup

	for _, s := range servers {
		wg.Add(1)
		go func(srv Server) {
			defer wg.Done()
			slog.Info("starting server", "server", srv.Label())
			if err := srv.Start(); err != nil {
				slog.Error("server failed", "server", srv.Label(), "error", err)
			}
		}(s)
	}

	<-ctx.Done()

	slog.Info("shutting down servers")

	// The caller's ctx is already canceled. Use a fresh context with a timeout
	// so shutdown handlers actually get time to drain connections and close listeners.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	var shutdownWg sync.WaitGroup
	for _, s := range servers {
		shutdownWg.Add(1)
		go func(srv Server) {
			defer shutdownWg.Done()
			if err := srv.Shutdown(shutdownCtx); err != nil {
				slog.Error("shutdown failed", "server", srv.Label(), "error", err)
			}
		}(s)
	}
	shutdownWg.Wait()

	wg.Wait()
}
