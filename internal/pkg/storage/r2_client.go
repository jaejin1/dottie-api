package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type StorageClient interface {
	GenerateUploadURL(ctx context.Context, key, contentType string) (string, error)
	GetPublicURL(key string) string
}

type R2Client struct {
	presigner *s3.PresignClient
	bucket    string
	publicURL string
}

func NewR2Client(accountID, accessKeyID, secretAccessKey, bucket, publicURL string) *R2Client {
	cfg := aws.Config{
		Region: "auto",
		Credentials: credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID))
	})

	return &R2Client{
		presigner: s3.NewPresignClient(client),
		bucket:    bucket,
		publicURL: publicURL,
	}
}

func (r *R2Client) GenerateUploadURL(ctx context.Context, key, contentType string) (string, error) {
	result, err := r.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(1*time.Hour))
	if err != nil {
		return "", err
	}
	return result.URL, nil
}

func (r *R2Client) GetPublicURL(key string) string {
	return fmt.Sprintf("%s/%s", r.publicURL, key)
}
