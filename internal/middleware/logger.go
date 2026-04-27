package middleware

import (
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func Logger(logger *zap.Logger) echo.MiddlewareFunc {
	return echomiddleware.RequestLoggerWithConfig(echomiddleware.RequestLoggerConfig{
		LogURI:     true,
		LogStatus:  true,
		LogMethod:  true,
		LogLatency: true,
		LogError:   true,
		LogValuesFunc: func(c echo.Context, v echomiddleware.RequestLoggerValues) error {
			fields := []zap.Field{
				zap.String("method", v.Method),
				zap.String("uri", v.URI),
				zap.Int("status", v.Status),
				zap.Duration("latency", v.Latency),
			}
			if v.Error != nil {
				fields = append(fields, zap.Error(v.Error))
				logger.Error("request", fields...)
			} else {
				logger.Info("request", fields...)
			}
			return nil
		},
	})
}
