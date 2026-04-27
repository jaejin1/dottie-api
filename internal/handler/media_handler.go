package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/jaejin1/dottie-api/internal/model/dto"
	"github.com/jaejin1/dottie-api/internal/service"
)

type MediaHandler struct {
	mediaService service.MediaService
}

func NewMediaHandler(mediaService service.MediaService) *MediaHandler {
	return &MediaHandler{mediaService: mediaService}
}

func (h *MediaHandler) Upload(c echo.Context) error {
	uid := c.Get("firebase_uid").(string)

	var req dto.MediaUploadRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "BAD_REQUEST", "message": "잘못된 요청 형식입니다"},
		})
	}
	if req.ContentType == "" {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "BAD_REQUEST", "message": "content_type은 필수입니다"},
		})
	}

	result, err := h.mediaService.GenerateUploadURL(c.Request().Context(), uid, &req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}
