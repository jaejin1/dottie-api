package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/jaejin1/dottie-api/internal/db"
	"github.com/jaejin1/dottie-api/internal/model/dto"
	"github.com/jaejin1/dottie-api/internal/pkg/firebase"
	"github.com/jaejin1/dottie-api/internal/pkg/kakao"
	"github.com/jaejin1/dottie-api/internal/repository"
)

type AuthService interface {
	KakaoLogin(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
}

type authService struct {
	userRepo    repository.UserRepository
	kakaoClient kakao.Client
	fbClient    *firebase.Client
}

func NewAuthService(userRepo repository.UserRepository, kakaoClient kakao.Client, fbClient *firebase.Client) AuthService {
	return &authService{
		userRepo:    userRepo,
		kakaoClient: kakaoClient,
		fbClient:    fbClient,
	}
}

func (s *authService) KakaoLogin(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	kakaoUser, err := s.kakaoClient.GetUser(ctx, req.Token)
	if err != nil {
		return nil, fmt.Errorf("kakao token verification failed: %w", err)
	}

	firebaseUID := fmt.Sprintf("kakao:%d", kakaoUser.ID)

	existingUser, err := s.userRepo.GetByFirebaseUID(ctx, firebaseUID)
	isNew := false

	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}

		// 신규 가입
		nickname := req.Nickname
		if nickname == "" {
			nickname = kakaoUser.Nickname
		}

		defaultCharacter, _ := json.Marshal(dto.CharacterConfig{
			Color:      "blue",
			Accessory:  "none",
			Expression: "default",
		})

		var profileImage pgtype.Text
		if kakaoUser.ProfileImage != "" {
			profileImage = pgtype.Text{String: kakaoUser.ProfileImage, Valid: true}
		}

		existingUser, err = s.userRepo.Create(ctx, db.CreateUserParams{
			FirebaseUid:     firebaseUID,
			Provider:        "kakao",
			Nickname:        nickname,
			ProfileImage:    profileImage,
			CharacterConfig: defaultCharacter,
		})
		if err != nil {
			return nil, err
		}
		isNew = true
	}

	customToken, err := s.fbClient.CreateCustomToken(ctx, firebaseUID)
	if err != nil {
		return nil, fmt.Errorf("firebase custom token creation failed: %w", err)
	}

	return &dto.LoginResponse{
		FirebaseCustomToken: customToken,
		User:                toUserResponse(existingUser),
		IsNew:               isNew,
	}, nil
}
