package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/jaejin1/dottie-api/internal/model/dto"
	"github.com/jaejin1/dottie-api/internal/service"
)

type DotHandler struct {
	dotService service.DotService
}

func NewDotHandler(dotService service.DotService) *DotHandler {
	return &DotHandler{dotService: dotService}
}

func (h *DotHandler) CreateDot(c echo.Context) error {
	uid := c.Get("firebase_uid").(string)

	var req dto.CreateDotRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "BAD_REQUEST", "message": "잘못된 요청 형식입니다"},
		})
	}
	if req.DayLogID == "" || req.Timestamp == "" {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "BAD_REQUEST", "message": "day_log_id, timestamp는 필수입니다"},
		})
	}

	result, err := h.dotService.CreateDot(c.Request().Context(), uid, &req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": result})
}

func (h *DotHandler) CreateDotsBatch(c echo.Context) error {
	uid := c.Get("firebase_uid").(string)

	var req dto.CreateDotsBatchRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "BAD_REQUEST", "message": "잘못된 요청 형식입니다"},
		})
	}
	if req.DayLogID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "BAD_REQUEST", "message": "day_log_id는 필수입니다"},
		})
	}

	result, err := h.dotService.CreateDotsBatch(c.Request().Context(), uid, &req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (h *DotHandler) GetDotsByDate(c echo.Context) error {
	uid := c.Get("firebase_uid").(string)

	date := c.QueryParam("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	result, err := h.dotService.GetDotsByDate(c.Request().Context(), uid, date)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}
