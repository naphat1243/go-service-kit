// Package handler contains the HTTP layer: request parsing, response shaping,
// and nothing else. Business rules live in the service layer.
package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// Pinger is anything that can confirm a dependency is reachable (pgxpool.Pool).
type Pinger interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	db Pinger
}

func NewHealthHandler(db Pinger) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Register(e *echo.Echo) {
	e.GET("/healthz", h.live)
	e.GET("/readyz", h.ready)
}

// live reports process liveness only — it must stay dependency-free so
// orchestrators never restart a healthy process because a dependency blinked.
func (h *HealthHandler) live(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HealthHandler) ready(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "degraded", "database": "unreachable"})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
