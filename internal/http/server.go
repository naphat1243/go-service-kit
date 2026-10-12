// Package http wires the Echo server: middleware, error handling, and routes.
package http

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"github.com/naphat1243/go-service-kit/internal/config"
	"github.com/naphat1243/go-service-kit/internal/pkg/apperror"
)

type Route interface {
	Register(e *echo.Echo)
}

func NewServer(cfg config.Config, log *slog.Logger, routes ...Route) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.HTTPErrorHandler = apperror.HTTPErrorHandler(log)

	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: corsOrigins(cfg),
	}))
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
		rate.Limit(cfg.RateLimitPerSec),
	)))
	e.Use(requestLogger(log))

	for _, r := range routes {
		r.Register(e)
	}
	return e
}

// corsOrigins is deliberately restrictive outside local development:
// configure real origins before shipping. Never ship "*" to production.
func corsOrigins(cfg config.Config) []string {
	if cfg.IsLocal() {
		return []string{"http://localhost:3000"}
	}
	return []string{}
}

func requestLogger(log *slog.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true, LogURI: true, LogMethod: true, LogLatency: true, LogRequestID: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Info("request",
				"method", v.Method,
				"uri", v.URI,
				"status", v.Status,
				"latency_ms", v.Latency.Milliseconds(),
				"request_id", v.RequestID,
			)
			return nil
		},
	})
}

// Start runs the server until ctx is cancelled, then shuts down gracefully.
func Start(ctx context.Context, e *echo.Echo, cfg config.Config, log *slog.Logger) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- e.Start(fmt.Sprintf(":%d", cfg.HTTPPort))
	}()
	log.Info("server started", "port", cfg.HTTPPort, "env", cfg.Env)

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		log.Info("shutting down")
		return e.Shutdown(shutdownCtx)
	}
}
