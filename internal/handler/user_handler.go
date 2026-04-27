package handler

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"

	"github.com/jaejin1/dottie-api/internal/model/dto"
	"github.com/jaejin1/dottie-api/internal/service"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetMe(c echo.Context) error {
	uid := c.Get("firebase_uid").(string)

	user, err := h.userService.GetMe(c.Request().Context(), uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, map[string]any{
				"error": map[string]string{"code": "USER_NOT_FOUND", "message": "사용자를 찾을 수 없습니다"},
			})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
			"error": map[string]string{"code": "INTERNAL_SERVER_ERROR", "message": "서버 오류가 발생했습니다"},
		})
	}

	return c.JSON(http.StatusOK, map[string]any{"data": user})
}

func (h *UserHandler) UpdateMe(c echo.Context) error {
	uid := c.Get("firebase_uid").(string)

	var req dto.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "BAD_REQUEST", "message": "잘못된 요청 형식입니다"},
		})
	}

	if len([]rune(req.Nickname)) > 30 {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "VALIDATION_ERROR", "message": "닉네임은 30자 이하여야 합니다"},
		})
	}

	user, err := h.userService.UpdateMe(c.Request().Context(), uid, &req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, map[string]any{
				"error": map[string]string{"code": "USER_NOT_FOUND", "message": "사용자를 찾을 수 없습니다"},
			})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
			"error": map[string]string{"code": "INTERNAL_SERVER_ERROR", "message": "서버 오류가 발생했습니다"},
		})
	}

	return c.JSON(http.StatusOK, map[string]any{"data": user})
}

func (h *UserHandler) UpdateCharacter(c echo.Context) error {
	uid := c.Get("firebase_uid").(string)

	var req dto.UpdateCharacterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "BAD_REQUEST", "message": "잘못된 요청 형식입니다"},
		})
	}

	if req.Color == "" || req.Accessory == "" || req.Expression == "" {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "VALIDATION_ERROR", "message": "color, accessory, expression은 필수입니다"},
		})
	}

	user, err := h.userService.UpdateCharacter(c.Request().Context(), uid, &req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, map[string]any{
				"error": map[string]string{"code": "USER_NOT_FOUND", "message": "사용자를 찾을 수 없습니다"},
			})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
			"error": map[string]string{"code": "INTERNAL_SERVER_ERROR", "message": "서버 오류가 발생했습니다"},
		})
	}

	return c.JSON(http.StatusOK, map[string]any{"data": user})
}
