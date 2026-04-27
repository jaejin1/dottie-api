package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/jaejin1/dottie-api/internal/model/dto"
	"github.com/jaejin1/dottie-api/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "BAD_REQUEST", "message": "잘못된 요청 형식입니다"},
		})
	}

	if req.Token == "" {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "BAD_REQUEST", "message": "token은 필수입니다"},
		})
	}

	if req.Provider != "kakao" {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "INVALID_PROVIDER", "message": "지원하지 않는 로그인 방식입니다"},
		})
	}

	result, err := h.authService.KakaoLogin(c.Request().Context(), &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{
			"error": map[string]string{"code": "TOKEN_VERIFICATION_FAILED", "message": "토큰 검증에 실패했습니다"},
		})
	}

	return c.JSON(http.StatusOK, map[string]any{"data": result})
}
