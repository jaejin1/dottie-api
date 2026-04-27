package service

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/jaejin1/dottie-api/internal/db"
	"github.com/jaejin1/dottie-api/internal/model/dto"
	"github.com/jaejin1/dottie-api/internal/repository"
)

type UserService interface {
	GetMe(ctx context.Context, firebaseUID string) (*dto.UserResponse, error)
	UpdateMe(ctx context.Context, firebaseUID string, req *dto.UpdateUserRequest) (*dto.UserResponse, error)
	UpdateCharacter(ctx context.Context, firebaseUID string, req *dto.UpdateCharacterRequest) (*dto.UserResponse, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetMe(ctx context.Context, firebaseUID string) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, err
	}
	resp := toUserResponse(user)
	return &resp, nil
}

func (s *userService) UpdateMe(ctx context.Context, firebaseUID string, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, err
	}

	var profileImage pgtype.Text
	if req.ProfileImage != nil {
		profileImage = pgtype.Text{String: *req.ProfileImage, Valid: true}
	}

	updated, err := s.userRepo.Update(ctx, db.UpdateUserParams{
		ID:           user.ID,
		Nickname:     req.Nickname,
		ProfileImage: profileImage,
	})
	if err != nil {
		return nil, err
	}
	resp := toUserResponse(updated)
	return &resp, nil
}

func (s *userService) UpdateCharacter(ctx context.Context, firebaseUID string, req *dto.UpdateCharacterRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, err
	}

	characterJSON, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	updated, err := s.userRepo.UpdateCharacter(ctx, db.UpdateCharacterConfigParams{
		ID:              user.ID,
		CharacterConfig: characterJSON,
	})
	if err != nil {
		return nil, err
	}
	resp := toUserResponse(updated)
	return &resp, nil
}

// toUserResponse는 auth_service.go와 공유하는 변환 헬퍼
func toUserResponse(u db.User) dto.UserResponse {
	var profileImage *string
	if u.ProfileImage.Valid {
		profileImage = &u.ProfileImage.String
	}

	var character dto.CharacterConfig
	_ = json.Unmarshal(u.CharacterConfig, &character)

	return dto.UserResponse{
		ID:              u.ID.String(),
		Nickname:        u.Nickname,
		ProfileImage:    profileImage,
		CharacterConfig: character,
		Provider:        u.Provider,
		CreatedAt:       u.CreatedAt.Time,
	}
}
