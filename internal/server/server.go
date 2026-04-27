package server

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/jaejin1/dottie-api/internal/config"
	internaldb "github.com/jaejin1/dottie-api/internal/db"
	"github.com/jaejin1/dottie-api/internal/handler"
	"github.com/jaejin1/dottie-api/internal/middleware"
	"github.com/jaejin1/dottie-api/internal/pkg/firebase"
	"github.com/jaejin1/dottie-api/internal/pkg/kakao"
	"github.com/jaejin1/dottie-api/internal/pkg/mapbox"
	"github.com/jaejin1/dottie-api/internal/pkg/storage"
	"github.com/jaejin1/dottie-api/internal/repository"
	"github.com/jaejin1/dottie-api/internal/service"
)

type Server struct {
	echo   *echo.Echo
	cfg    *config.Config
	logger *zap.Logger
}

func New(cfg *config.Config, logger *zap.Logger, fbClient *firebase.Client, dbPool *pgxpool.Pool) *Server {
	e := echo.New()
	e.HideBanner = true

	// Global middleware
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.RequestID())
	e.Use(middleware.Logger(logger))
	e.Use(middleware.CORS(cfg.CORSOrigins))
	e.Use(middleware.RateLimit())
	e.Use(echomiddleware.BodyLimit("10M"))

	// Custom error handler
	e.HTTPErrorHandler = errorHandler

	// Routes
	healthHandler := handler.NewHealthHandler()
	e.GET("/health", healthHandler.Health)
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"service": "dottie-api", "version": "v1"})
	})

	// Wire dependencies
	queries := internaldb.New(dbPool)
	userRepo := repository.NewUserRepository(queries)
	dayLogRepo := repository.NewDayLogRepository(queries)
	dotRepo := repository.NewDotRepository(queries)

	kakaoClient := kakao.NewClient()

	var mapboxClient mapbox.GeocodingClient
	if cfg.MapboxAccessToken != "" {
		mapboxClient = mapbox.NewClient(cfg.MapboxAccessToken)
	}

	var storageClient storage.StorageClient
	if cfg.R2AccountID != "" {
		storageClient = storage.NewR2Client(cfg.R2AccountID, cfg.R2AccessKeyID, cfg.R2SecretAccessKey, cfg.R2BucketName, cfg.R2PublicURL)
	}

	authSvc := service.NewAuthService(userRepo, kakaoClient, fbClient)
	userSvc := service.NewUserService(userRepo)
	dayLogSvc := service.NewDayLogService(dayLogRepo, dotRepo, userRepo)
	dotSvc := service.NewDotService(dotRepo, dayLogRepo, userRepo, mapboxClient)
	mediaSvc := service.NewMediaService(storageClient, userRepo)

	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userSvc)
	recordingHandler := handler.NewRecordingHandler(dayLogSvc)
	dotHandler := handler.NewDotHandler(dotSvc)
	dayLogHandler := handler.NewDayLogHandler(dayLogSvc)
	mediaHandler := handler.NewMediaHandler(mediaSvc)

	v1 := e.Group("/v1")

	// Auth routes — 인증 미들웨어 없음
	auth := v1.Group("/auth")
	auth.POST("/login", authHandler.Login)

	// 인증 미들웨어 팩토리
	authMW := func() []echo.MiddlewareFunc {
		if fbClient != nil {
			return []echo.MiddlewareFunc{middleware.Auth(fbClient)}
		}
		logger.Warn("firebase not configured — routes are unprotected")
		return nil
	}

	// User routes
	users := v1.Group("/users", authMW()...)
	users.GET("/me", userHandler.GetMe)
	users.PUT("/me", userHandler.UpdateMe)
	users.PUT("/me/character", userHandler.UpdateCharacter)

	// Recording routes
	recordings := v1.Group("/recordings", authMW()...)
	recordings.POST("/start", recordingHandler.Start)
	recordings.POST("/end", recordingHandler.End)

	return &Server{echo: e, cfg: cfg, logger: logger}
}

func (s *Server) Start() error {
	s.logger.Info("starting server", zap.String("port", s.cfg.Port))
	return s.echo.Start(":" + s.cfg.Port)
}

func errorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	code := http.StatusInternalServerError
	var body any

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if msg, ok := he.Message.(map[string]any); ok {
			body = msg
		} else {
			body = map[string]any{
				"error": map[string]string{
					"code":    http.StatusText(code),
					"message": he.Error(),
				},
			}
		}
	} else {
		body = map[string]any{
			"error": map[string]string{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "서버 오류가 발생했습니다",
			},
		}
	}

	_ = c.JSON(code, body)
}
