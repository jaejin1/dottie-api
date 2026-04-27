package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/jaejin1/dottie-api/internal/db"
)

type DotRepository interface {
	Create(ctx context.Context, params db.CreateDotParams) (db.Dot, error)
	GetByDayLog(ctx context.Context, dayLogID pgtype.UUID) ([]db.Dot, error)
	GetByUserAndDate(ctx context.Context, params db.GetDotsByUserAndDateParams) ([]db.Dot, error)
	UpdatePhotoURL(ctx context.Context, params db.UpdateDotPhotoURLParams) (db.Dot, error)
}

type dotRepository struct {
	q db.Querier
}

func NewDotRepository(q db.Querier) DotRepository {
	return &dotRepository{q: q}
}

func (r *dotRepository) Create(ctx context.Context, params db.CreateDotParams) (db.Dot, error) {
	return r.q.CreateDot(ctx, params)
}

func (r *dotRepository) GetByDayLog(ctx context.Context, dayLogID pgtype.UUID) ([]db.Dot, error) {
	return r.q.GetDotsByDayLog(ctx, dayLogID)
}

func (r *dotRepository) GetByUserAndDate(ctx context.Context, params db.GetDotsByUserAndDateParams) ([]db.Dot, error) {
	return r.q.GetDotsByUserAndDate(ctx, params)
}

func (r *dotRepository) UpdatePhotoURL(ctx context.Context, params db.UpdateDotPhotoURLParams) (db.Dot, error) {
	return r.q.UpdateDotPhotoURL(ctx, params)
}
