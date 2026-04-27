package service

import (
	"context"
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"github.com/jaejin1/dottie-api/internal/db"
	"github.com/jaejin1/dottie-api/internal/model/dto"
	"github.com/jaejin1/dottie-api/internal/repository"
)

type DayLogService interface {
	StartRecording(ctx context.Context, firebaseUID, date string) (*dto.DayLogResponse, error)
	EndRecording(ctx context.Context, firebaseUID, dayLogID string) (*dto.DayLogResponse, error)
	ListDayLogs(ctx context.Context, firebaseUID string, limit, offset int32) ([]dto.DayLogResponse, error)
	GetDayLogWithDots(ctx context.Context, firebaseUID, dayLogID string) (*dto.DayLogDetailResponse, error)
}

type dayLogService struct {
	dayLogRepo repository.DayLogRepository
	dotRepo    repository.DotRepository
	userRepo   repository.UserRepository
}

func NewDayLogService(dayLogRepo repository.DayLogRepository, dotRepo repository.DotRepository, userRepo repository.UserRepository) DayLogService {
	return &dayLogService{dayLogRepo: dayLogRepo, dotRepo: dotRepo, userRepo: userRepo}
}

func (s *dayLogService) StartRecording(ctx context.Context, firebaseUID, date string) (*dto.DayLogResponse, error) {
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

	pgDate := pgtype.Date{Time: t, Valid: true}
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}

	dayLog, err := s.dayLogRepo.Create(ctx, db.CreateDayLogParams{
		UserID:    user.ID,
		Date:      pgDate,
		StartedAt: now,
	})
	if err != nil {
		// unique constraint 위반 — (user_id, date) 중복
		if isUniqueViolation(err) {
			return nil, echo.NewHTTPError(http.StatusConflict, map[string]any{
				"error": map[string]string{"code": "ALREADY_RECORDING", "message": "해당 날짜의 기록이 이미 존재합니다"},
			})
		}
		return nil, err
	}

	resp := toDayLogResponse(dayLog)
	return &resp, nil
}

func (s *dayLogService) EndRecording(ctx context.Context, firebaseUID, dayLogID string) (*dto.DayLogResponse, error) {
	user, err := s.userRepo.GetByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, err
	}

	var pgID pgtype.UUID
	if err := pgID.Scan(dayLogID); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "INVALID_ID", "message": "유효하지 않은 ID입니다"},
		})
	}

	dayLog, err := s.dayLogRepo.GetByID(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, echo.NewHTTPError(http.StatusNotFound, map[string]any{
				"error": map[string]string{"code": "DAYLOG_NOT_FOUND", "message": "기록을 찾을 수 없습니다"},
			})
		}
		return nil, err
	}

	if dayLog.UserID != user.ID {
		return nil, echo.NewHTTPError(http.StatusForbidden, map[string]any{
			"error": map[string]string{"code": "FORBIDDEN", "message": "권한이 없습니다"},
		})
	}

	if !dayLog.IsRecording {
		return nil, echo.NewHTTPError(http.StatusConflict, map[string]any{
			"error": map[string]string{"code": "ALREADY_ENDED", "message": "이미 종료된 기록입니다"},
		})
	}

	dots, err := s.dotRepo.GetByDayLog(ctx, dayLog.ID)
	if err != nil {
		return nil, err
	}

	distM, placeCount, durationSec := calcStats(dots)

	updated, err := s.dayLogRepo.End(ctx, db.EndDayLogParams{
		ID:               dayLog.ID,
		EndedAt:          pgtype.Timestamptz{Time: time.Now(), Valid: true},
		TotalDistanceM:   pgtype.Float8{Float64: distM, Valid: true},
		PlaceCount:       pgtype.Int4{Int32: placeCount, Valid: true},
		TotalDurationSec: pgtype.Int4{Int32: durationSec, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	resp := toDayLogResponse(updated)
	return &resp, nil
}

func (s *dayLogService) ListDayLogs(ctx context.Context, firebaseUID string, limit, offset int32) ([]dto.DayLogResponse, error) {
	user, err := s.userRepo.GetByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, err
	}

	dayLogs, err := s.dayLogRepo.ListByUser(ctx, db.ListDayLogsByUserParams{
		UserID: user.ID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	result := make([]dto.DayLogResponse, len(dayLogs))
	for i, dl := range dayLogs {
		result[i] = toDayLogResponse(dl)
	}
	return result, nil
}

func (s *dayLogService) GetDayLogWithDots(ctx context.Context, firebaseUID, dayLogID string) (*dto.DayLogDetailResponse, error) {
	user, err := s.userRepo.GetByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, err
	}

	var pgID pgtype.UUID
	if err := pgID.Scan(dayLogID); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "INVALID_ID", "message": "유효하지 않은 ID입니다"},
		})
	}

	dayLog, err := s.dayLogRepo.GetByID(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, echo.NewHTTPError(http.StatusNotFound, map[string]any{
				"error": map[string]string{"code": "DAYLOG_NOT_FOUND", "message": "기록을 찾을 수 없습니다"},
			})
		}
		return nil, err
	}

	if dayLog.UserID != user.ID {
		return nil, echo.NewHTTPError(http.StatusForbidden, map[string]any{
			"error": map[string]string{"code": "FORBIDDEN", "message": "권한이 없습니다"},
		})
	}

	dots, err := s.dotRepo.GetByDayLog(ctx, dayLog.ID)
	if err != nil {
		return nil, err
	}

	dotResponses := make([]dto.DotResponse, len(dots))
	for i, d := range dots {
		dotResponses[i] = toDotResponse(d)
	}

	return &dto.DayLogDetailResponse{
		DayLogResponse: toDayLogResponse(dayLog),
		Dots:           dotResponses,
	}, nil
}

// ── helpers ────────────────────────────────────────────────

func calcStats(dots []db.Dot) (distM float64, placeCount int32, durationSec int32) {
	if len(dots) == 0 {
		return 0, 0, 0
	}

	placeSet := make(map[string]struct{})
	for i, d := range dots {
		if d.PlaceName.Valid && d.PlaceName.String != "" {
			placeSet[d.PlaceName.String] = struct{}{}
		}
		if i > 0 {
			distM += haversineMeters(
				dots[i-1].Latitude, dots[i-1].Longitude,
				d.Latitude, d.Longitude,
			)
		}
	}

	placeCount = int32(len(placeSet))

	first := dots[0].Timestamp.Time
	last := dots[len(dots)-1].Timestamp.Time
	durationSec = int32(last.Sub(first).Seconds())

	return distM, placeCount, durationSec
}

func haversineMeters(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000.0
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δφ := (lat2 - lat1) * math.Pi / 180
	Δλ := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) + math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func toDayLogResponse(dl db.DayLog) dto.DayLogResponse {
	r := dto.DayLogResponse{
		ID:          dl.ID.String(),
		Date:        dl.Date.Time.Format("2006-01-02"),
		StartedAt:   dl.StartedAt.Time,
		IsRecording: dl.IsRecording,
	}
	if dl.EndedAt.Valid {
		t := dl.EndedAt.Time
		r.EndedAt = &t
	}
	if dl.TotalDistanceM.Valid {
		v := dl.TotalDistanceM.Float64
		r.TotalDistanceM = &v
	}
	if dl.PlaceCount.Valid {
		v := dl.PlaceCount.Int32
		r.PlaceCount = &v
	}
	if dl.TotalDurationSec.Valid {
		v := dl.TotalDurationSec.Int32
		r.TotalDurationSec = &v
	}
	return r
}

func toDotResponse(d db.Dot) dto.DotResponse {
	r := dto.DotResponse{
		ID:        d.ID.String(),
		DayLogID:  d.DayLogID.String(),
		Latitude:  d.Latitude,
		Longitude: d.Longitude,
		Timestamp: d.Timestamp.Time,
	}
	if d.PlaceName.Valid {
		r.PlaceName = &d.PlaceName.String
	}
	if d.PlaceCategory.Valid {
		r.PlaceCategory = &d.PlaceCategory.String
	}
	if d.PhotoUrl.Valid {
		r.PhotoURL = &d.PhotoUrl.String
	}
	if d.Memo.Valid {
		r.Memo = &d.Memo.String
	}
	if d.Emotion.Valid {
		r.Emotion = &d.Emotion.String
	}
	return r
}

func isUniqueViolation(err error) bool {
	return err != nil && (contains(err.Error(), "unique") || contains(err.Error(), "duplicate"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
