package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// S3StorageService implements StorageService using AWS S3
type S3StorageService struct {
	client *s3.Client
	bucket string
	region string
}

// S3Config holds configuration for S3 storage
type S3Config struct {
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
}

// NewS3StorageService creates a new S3 storage service
func NewS3StorageService(ctx context.Context, cfg S3Config) (*S3StorageService, error) {
	// Create AWS config with credentials
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(awsCfg)

	return &S3StorageService{
		client: client,
		bucket: cfg.Bucket,
		region: cfg.Region,
	}, nil
}

// Upload uploads content to S3 with retry logic
func (s *S3StorageService) Upload(ctx context.Context, userID string, documentID string, filename string, content io.Reader, contentType string) (string, error) {
	// Generate storage key with user-specific prefix
	storageKey := fmt.Sprintf("%s/%s", userID, documentID)

	// Prepare upload input
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(storageKey),
		Body:        content,
		ContentType: aws.String(contentType),
		Metadata: map[string]string{
			"filename":    filename,
			"document-id": documentID,
			"user-id":     userID,
		},
	}

	// Implement retry logic (3 attempts with exponential backoff)
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		_, err := s.client.PutObject(ctx, input)
		if err == nil {
			return storageKey, nil
		}

		lastErr = err

		// Don't retry on last attempt
		if attempt < 3 {
			// Exponential backoff: 1s, 2s
			backoff := time.Duration(attempt) * time.Second
			time.Sleep(backoff)
		}
	}

	return "", fmt.Errorf("failed to upload to S3 after 3 attempts: %w", lastErr)
}

// Download retrieves content from S3
func (s *S3StorageService) Download(ctx context.Context, storageKey string) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(storageKey),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}

	return result.Body, nil
}

// Delete removes content from S3
func (s *S3StorageService) Delete(ctx context.Context, storageKey string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(storageKey),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}

// GetURL returns a presigned URL for accessing the content
func (s *S3StorageService) GetURL(ctx context.Context, storageKey string, expirationMinutes int) (string, error) {
	// Create presign client
	presignClient := s3.NewPresignClient(s.client)

	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(storageKey),
	}

	// Generate presigned URL
	result, err := presignClient.PresignGetObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expirationMinutes) * time.Minute
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return result.URL, nil
}

// GenerateDocumentID generates a unique document ID
func GenerateDocumentID() string {
	return uuid.New().String()
}
