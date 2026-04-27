package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().UTC(),
	})
}
