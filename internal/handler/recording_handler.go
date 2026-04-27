package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/jaejin1/dottie-api/internal/model/dto"
	"github.com/jaejin1/dottie-api/internal/service"
)

type RecordingHandler struct {
	dayLogService service.DayLogService
}

func NewRecordingHandler(dayLogService service.DayLogService) *RecordingHandler {
	return &RecordingHandler{dayLogService: dayLogService}
}

func (h *RecordingHandler) Start(c echo.Context) error {
	uid := c.Get("firebase_uid").(string)

	var req dto.StartRecordingRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "BAD_REQUEST", "message": "잘못된 요청 형식입니다"},
		})
	}
	if req.Date == "" {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "BAD_REQUEST", "message": "date는 필수입니다"},
		})
	}

	result, err := h.dayLogService.StartRecording(c.Request().Context(), uid, req.Date)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": result})
}

func (h *RecordingHandler) End(c echo.Context) error {
	uid := c.Get("firebase_uid").(string)

	var req dto.EndRecordingRequest
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

	result, err := h.dayLogService.EndRecording(c.Request().Context(), uid, req.DayLogID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}
