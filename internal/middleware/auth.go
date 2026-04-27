package middleware

import (
	"strings"

	"github.com/jaejin1/dottie-api/internal/pkg/firebase"
	"github.com/labstack/echo/v4"
)

func Auth(fbClient *firebase.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return echo.NewHTTPError(401, map[string]any{
					"error": map[string]string{
						"code":    "UNAUTHORIZED",
						"message": "인증 토큰이 없습니다",
					},
				})
			}
			idToken := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := fbClient.VerifyIDToken(c.Request().Context(), idToken)
			if err != nil {
				return echo.NewHTTPError(401, map[string]any{
					"error": map[string]string{
						"code":    "INVALID_TOKEN",
						"message": "유효하지 않은 인증 토큰입니다",
					},
				})
			}

			c.Set("firebase_uid", token.UID)
			return next(c)
		}
	}
}
