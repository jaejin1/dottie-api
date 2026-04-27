package service

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"github.com/jaejin1/dottie-api/internal/db"
	"github.com/jaejin1/dottie-api/internal/model/dto"
	"github.com/jaejin1/dottie-api/internal/pkg/mapbox"
	"github.com/jaejin1/dottie-api/internal/repository"
)

type DotService interface {
	CreateDot(ctx context.Context, firebaseUID string, req *dto.CreateDotRequest) (*dto.CreateDotResponse, error)
	CreateDotsBatch(ctx context.Context, firebaseUID string, req *dto.CreateDotsBatchRequest) (*dto.CreateDotsBatchResponse, error)
	GetDotsByDate(ctx context.Context, firebaseUID, date string) ([]dto.DotResponse, error)
}

type dotService struct {
	dotRepo    repository.DotRepository
	dayLogRepo repository.DayLogRepository
	userRepo   repository.UserRepository
	mapbox     mapbox.GeocodingClient // nil 허용 (미설정 시)
}

func NewDotService(
	dotRepo repository.DotRepository,
	dayLogRepo repository.DayLogRepository,
	userRepo repository.UserRepository,
	mapboxClient mapbox.GeocodingClient,
) DotService {
	return &dotService{
		dotRepo:    dotRepo,
		dayLogRepo: dayLogRepo,
		userRepo:   userRepo,
		mapbox:     mapboxClient,
	}
}

func (s *dotService) CreateDot(ctx context.Context, firebaseUID string, req *dto.CreateDotRequest) (*dto.CreateDotResponse, error) {
	user, err := s.userRepo.GetByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, err
	}

	dayLog, err := s.getDayLogAndCheckOwner(ctx, req.DayLogID, user.ID)
	if err != nil {
		return nil, err
	}

	ts, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "INVALID_TIMESTAMP", "message": "timestamp 형식이 올바르지 않습니다 (RFC3339)"},
		})
	}

	params := db.CreateDotParams{
		DayLogID:  dayLog.ID,
		UserID:    user.ID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Timestamp: pgtype.Timestamptz{Time: ts, Valid: true},
	}

	// optional fields
	if req.PlaceName != nil && *req.PlaceName != "" {
		params.PlaceName = pgtype.Text{String: *req.PlaceName, Valid: true}
	} else if s.mapbox != nil {
		// best-effort 역지오코딩
		if geo, err := s.mapbox.ReverseGeocode(ctx, req.Latitude, req.Longitude); err == nil && geo != nil {
			params.PlaceName = pgtype.Text{String: geo.PlaceName, Valid: true}
			params.PlaceCategory = pgtype.Text{String: geo.Category, Valid: true}
		}
	}
	if req.PlaceCategory != nil {
		params.PlaceCategory = pgtype.Text{String: *req.PlaceCategory, Valid: true}
	}
	if req.Memo != nil {
		params.Memo = pgtype.Text{String: *req.Memo, Valid: true}
	}
	if req.Emotion != nil {
		params.Emotion = pgtype.Text{String: *req.Emotion, Valid: true}
	}

	dot, err := s.dotRepo.Create(ctx, params)
	if err != nil {
		return nil, err
	}

	return &dto.CreateDotResponse{ID: dot.ID.String()}, nil
}

func (s *dotService) CreateDotsBatch(ctx context.Context, firebaseUID string, req *dto.CreateDotsBatchRequest) (*dto.CreateDotsBatchResponse, error) {
	if len(req.Dots) > 50 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "BATCH_LIMIT_EXCEEDED", "message": "한 번에 최대 50개까지 업로드할 수 있습니다"},
		})
	}

	user, err := s.userRepo.GetByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, err
	}

	dayLog, err := s.getDayLogAndCheckOwner(ctx, req.DayLogID, user.ID)
	if err != nil {
		return nil, err
	}

	result := &dto.CreateDotsBatchResponse{
		Synced: make([]dto.BatchSyncedItem, 0, len(req.Dots)),
		Failed: make([]dto.BatchFailedItem, 0),
	}

	for _, item := range req.Dots {
		ts, err := time.Parse(time.RFC3339, item.Timestamp)
		if err != nil {
			result.Failed = append(result.Failed, dto.BatchFailedItem{
				ClientID: item.ClientID,
				Reason:   "invalid timestamp format",
			})
			continue
		}

		params := db.CreateDotParams{
			DayLogID:  dayLog.ID,
			UserID:    user.ID,
			Latitude:  item.Latitude,
			Longitude: item.Longitude,
			Timestamp: pgtype.Timestamptz{Time: ts, Valid: true},
		}
		if item.PlaceName != nil {
			params.PlaceName = pgtype.Text{String: *item.PlaceName, Valid: true}
		}
		if item.PlaceCategory != nil {
			params.PlaceCategory = pgtype.Text{String: *item.PlaceCategory, Valid: true}
		}
		if item.Memo != nil {
			params.Memo = pgtype.Text{String: *item.Memo, Valid: true}
		}
		if item.Emotion != nil {
			params.Emotion = pgtype.Text{String: *item.Emotion, Valid: true}
		}

		dot, err := s.dotRepo.Create(ctx, params)
		if err != nil {
			result.Failed = append(result.Failed, dto.BatchFailedItem{
				ClientID: item.ClientID,
				Reason:   "failed to save",
			})
			continue
		}
		result.Synced = append(result.Synced, dto.BatchSyncedItem{
			ClientID: item.ClientID,
			ServerID: dot.ID.String(),
		})
	}

	return result, nil
}

func (s *dotService) GetDotsByDate(ctx context.Context, firebaseUID, date string) ([]dto.DotResponse, error) {
	user, err := s.userRepo.GetByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, err
	}

	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "INVALID_DATE", "message": "날짜 형식이 올바르지 않습니다 (YYYY-MM-DD)"},
		})
	}

	dots, err := s.dotRepo.GetByUserAndDate(ctx, db.GetDotsByUserAndDateParams{
		UserID: user.ID,
		Date:   pgtype.Date{Time: t, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	result := make([]dto.DotResponse, len(dots))
	for i, d := range dots {
		result[i] = toDotResponse(d)
	}
	return result, nil
}

func (s *dotService) getDayLogAndCheckOwner(ctx context.Context, dayLogID string, userID pgtype.UUID) (db.DayLog, error) {
	var pgID pgtype.UUID
	if err := pgID.Scan(dayLogID); err != nil {
		return db.DayLog{}, echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "INVALID_ID", "message": "유효하지 않은 day_log_id입니다"},
		})
	}

	dayLog, err := s.dayLogRepo.GetByID(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.DayLog{}, echo.NewHTTPError(http.StatusNotFound, map[string]any{
				"error": map[string]string{"code": "DAYLOG_NOT_FOUND", "message": "기록을 찾을 수 없습니다"},
			})
		}
		return db.DayLog{}, err
	}

	if dayLog.UserID != userID {
		return db.DayLog{}, echo.NewHTTPError(http.StatusForbidden, map[string]any{
			"error": map[string]string{"code": "FORBIDDEN", "message": "권한이 없습니다"},
		})
	}
	return dayLog, nil
}
