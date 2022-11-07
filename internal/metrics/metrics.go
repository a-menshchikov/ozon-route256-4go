package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const (
	readHeaderTimeout      = 5 * time.Second
	shutdownContextTimeout = 5 * time.Second
)

type Server struct {
	port uint16

	logger *zap.Logger
}

func NewServer(port uint16, l *zap.Logger) *Server {
	return &Server{
		port:   port,
		logger: l,
	}
}

func (s *Server) Run(ctx context.Context) error {
	router := chi.NewRouter()

	router.Use(
		middleware.NoCache,
	)

	router.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", s.port),
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			s.logger.Error("failed to run metrics server", zap.Error(err))
		}
	}()

	<-ctx.Done()

	s.logger.Info("metrics server shutdown")
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), shutdownContextTimeout)
	defer cancel()

	if err := server.Shutdown(ctxWithTimeout); err != nil {
		return errors.Wrap(err, "cannot shutdown metrics server")
	}

	return nil
}

func ResultLabel(err error) string {
	if err != nil {
		return "error"
	}

	return "ok"
}
