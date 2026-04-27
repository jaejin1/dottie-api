package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/jaejin1/dottie-api/internal/db"
)

type DayLogRepository interface {
	Create(ctx context.Context, params db.CreateDayLogParams) (db.DayLog, error)
	GetByID(ctx context.Context, id pgtype.UUID) (db.DayLog, error)
	GetByUserAndDate(ctx context.Context, params db.GetDayLogByUserAndDateParams) (db.DayLog, error)
	ListByUser(ctx context.Context, params db.ListDayLogsByUserParams) ([]db.DayLog, error)
	End(ctx context.Context, params db.EndDayLogParams) (db.DayLog, error)
}

type dayLogRepository struct {
	q db.Querier
}

func NewDayLogRepository(q db.Querier) DayLogRepository {
	return &dayLogRepository{q: q}
}

func (r *dayLogRepository) Create(ctx context.Context, params db.CreateDayLogParams) (db.DayLog, error) {
	return r.q.CreateDayLog(ctx, params)
}

func (r *dayLogRepository) GetByID(ctx context.Context, id pgtype.UUID) (db.DayLog, error) {
	return r.q.GetDayLogByID(ctx, id)
}

func (r *dayLogRepository) GetByUserAndDate(ctx context.Context, params db.GetDayLogByUserAndDateParams) (db.DayLog, error) {
	return r.q.GetDayLogByUserAndDate(ctx, params)
}

func (r *dayLogRepository) ListByUser(ctx context.Context, params db.ListDayLogsByUserParams) ([]db.DayLog, error) {
	return r.q.ListDayLogsByUser(ctx, params)
}

func (r *dayLogRepository) End(ctx context.Context, params db.EndDayLogParams) (db.DayLog, error) {
	return r.q.EndDayLog(ctx, params)
}
