package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

// RateLimit returns a rate limiter middleware: 100 req/min per IP.
func RateLimit() echo.MiddlewareFunc {
	return echomiddleware.RateLimiterWithConfig(echomiddleware.RateLimiterConfig{
		Store: echomiddleware.NewRateLimiterMemoryStoreWithConfig(
			echomiddleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(100.0 / 60.0), // 100 req/min
				Burst:     20,
				ExpiresIn: 0,
			},
		),
		IdentifierExtractor: func(c echo.Context) (string, error) {
			return c.RealIP(), nil
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return c.JSON(http.StatusTooManyRequests, map[string]any{
				"error": map[string]string{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "요청이 너무 많습니다. 잠시 후 다시 시도해주세요",
				},
			})
		},
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			return c.JSON(http.StatusTooManyRequests, map[string]any{
				"error": map[string]string{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "요청이 너무 많습니다. 잠시 후 다시 시도해주세요",
				},
			})
		},
	})
}
