package repository

import (
	"context"

	"github.com/jaejin1/dottie-api/internal/db"
)

type UserRepository interface {
	GetByFirebaseUID(ctx context.Context, firebaseUID string) (db.User, error)
	Create(ctx context.Context, params db.CreateUserParams) (db.User, error)
	Update(ctx context.Context, params db.UpdateUserParams) (db.User, error)
	UpdateCharacter(ctx context.Context, params db.UpdateCharacterConfigParams) (db.User, error)
}

type userRepository struct {
	q db.Querier
}

func NewUserRepository(q db.Querier) UserRepository {
	return &userRepository{q: q}
}

func (r *userRepository) GetByFirebaseUID(ctx context.Context, firebaseUID string) (db.User, error) {
	return r.q.GetUserByFirebaseUID(ctx, firebaseUID)
}

func (r *userRepository) Create(ctx context.Context, params db.CreateUserParams) (db.User, error) {
	return r.q.CreateUser(ctx, params)
}

func (r *userRepository) Update(ctx context.Context, params db.UpdateUserParams) (db.User, error) {
	return r.q.UpdateUser(ctx, params)
}

func (r *userRepository) UpdateCharacter(ctx context.Context, params db.UpdateCharacterConfigParams) (db.User, error) {
	return r.q.UpdateCharacterConfig(ctx, params)
}
