package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func CORS(origins string) echo.MiddlewareFunc {
	allowedOrigins := strings.Split(origins, ",")
	if len(allowedOrigins) == 0 || origins == "" {
		allowedOrigins = []string{"*"}
	}

	return echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
	})
}
