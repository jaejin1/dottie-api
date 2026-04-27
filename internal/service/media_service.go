package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/jaejin1/dottie-api/internal/model/dto"
	"github.com/jaejin1/dottie-api/internal/pkg/storage"
	"github.com/jaejin1/dottie-api/internal/repository"
)

const maxFileSize = 10 * 1024 * 1024 // 10MB

var allowedContentTypes = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/heic": "heic",
}

type MediaService interface {
	GenerateUploadURL(ctx context.Context, firebaseUID string, req *dto.MediaUploadRequest) (*dto.MediaUploadResponse, error)
}

type mediaService struct {
	storage  storage.StorageClient
	userRepo repository.UserRepository
}

func NewMediaService(storageClient storage.StorageClient, userRepo repository.UserRepository) MediaService {
	return &mediaService{storage: storageClient, userRepo: userRepo}
}

func (s *mediaService) GenerateUploadURL(ctx context.Context, firebaseUID string, req *dto.MediaUploadRequest) (*dto.MediaUploadResponse, error) {
	if s.storage == nil {
		return nil, echo.NewHTTPError(http.StatusServiceUnavailable, map[string]any{
			"error": map[string]string{"code": "STORAGE_NOT_CONFIGURED", "message": "스토리지가 설정되지 않았습니다"},
		})
	}

	if req.FileSize > maxFileSize {
		return nil, echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "FILE_TOO_LARGE", "message": "파일 크기가 너무 큽니다 (최대 10MB)"},
		})
	}

	ext, ok := allowedContentTypes[strings.ToLower(req.ContentType)]
	if !ok {
		return nil, echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error": map[string]string{"code": "INVALID_FILE_TYPE", "message": "지원하지 않는 파일 형식입니다 (jpeg, png, heic만 허용)"},
		})
	}

	user, err := s.userRepo.GetByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, err
	}

	date := time.Now().Format("20060102")
	key := fmt.Sprintf("users/%s/dots/%s/%s.%s", user.ID.String(), date, uuid.New().String(), ext)

	uploadURL, err := s.storage.GenerateUploadURL(ctx, key, req.ContentType)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
			"error": map[string]string{"code": "STORAGE_ERROR", "message": "업로드 URL 생성에 실패했습니다"},
		})
	}

	return &dto.MediaUploadResponse{
		UploadURL: uploadURL,
		PublicURL: s.storage.GetPublicURL(key),
		ExpiresIn: 3600,
	}, nil
}
