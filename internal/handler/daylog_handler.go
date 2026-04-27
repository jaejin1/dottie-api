package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/jaejin1/dottie-api/internal/service"
)

type DayLogHandler struct {
	dayLogService service.DayLogService
}

func NewDayLogHandler(dayLogService service.DayLogService) *DayLogHandler {
	return &DayLogHandler{dayLogService: dayLogService}
}

func (h *DayLogHandler) List(c echo.Context) error {
	uid := c.Get("firebase_uid").(string)

	limit := int32(20)
	offset := int32(0)
	if l := c.QueryParam("limit"); l != "" {
		if v, err := strconv.ParseInt(l, 10, 32); err == nil && v > 0 {
			limit = int32(v)
		}
	}
	if o := c.QueryParam("offset"); o != "" {
		if v, err := strconv.ParseInt(o, 10, 32); err == nil && v >= 0 {
			offset = int32(v)
		}
	}

	result, err := h.dayLogService.ListDayLogs(c.Request().Context(), uid, limit, offset)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{
		"data": result,
		"meta": map[string]int32{"limit": limit, "offset": offset},
	})
}

func (h *DayLogHandler) GetDetail(c echo.Context) error {
	uid := c.Get("firebase_uid").(string)
	id := c.Param("id")

	result, err := h.dayLogService.GetDayLogWithDots(c.Request().Context(), uid, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}
