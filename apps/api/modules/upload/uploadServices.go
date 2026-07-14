package upload

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lil-Strudel/glassact-studios/apps/api/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type UploadResult struct {
	URL         string `json:"url"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	Key         string `json:"key"`
	UploadedAt  string `json:"uploaded_at"`
}

func UploadFileToS3(
	ctx context.Context,
	s3Client *s3.Client,
	cfg *config.Config,
	file io.Reader,
	filename string,
	size int64,
	contentType string,
	uploadPath string,
) (*UploadResult, error) {
	if s3Client == nil {
		return nil, fmt.Errorf("S3 client not initialized")
	}

	filename = filepath.Base(filename)
	filename = strings.ReplaceAll(filename, "..", "")
	filename = strings.TrimSpace(filename)

	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".bin"
	}

	newFilename := uuid.New().String() + ext

	if uploadPath == "" {
		uploadPath = "default"
	}

	key := fmt.Sprintf("file/%s/%s", uploadPath, newFilename)

	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(cfg.S3.Bucket),
		Key:           aws.String(key),
		Body:          file,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to S3: %w", err)
	}

	relativeURL := fmt.Sprintf("/file/%s/%s", uploadPath, newFilename)

	result := &UploadResult{
		URL:         relativeURL,
		Filename:    filename,
		Size:        size,
		ContentType: contentType,
		Key:         key,
		UploadedAt:  time.Now().Format(time.RFC3339),
	}

	return result, nil
}

// GetFileFromS3 reads an object's bytes. `key` is the S3 key (the svg_url with
// its leading slash stripped, e.g. "file/catalog-items/<uuid>.svg").
func GetFileFromS3(
	ctx context.Context,
	s3Client *s3.Client,
	cfg *config.Config,
	key string,
) ([]byte, error) {
	if s3Client == nil {
		return nil, fmt.Errorf("S3 client not initialized")
	}

	out, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(cfg.S3.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer out.Body.Close()

	data, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read S3 object body: %w", err)
	}

	return data, nil
}

func GenerateSignedURL(
	ctx context.Context,
	s3Client *s3.Client,
	cfg *config.Config,
	key string,
	expirationDuration time.Duration,
) (string, error) {
	if s3Client == nil {
		return "", fmt.Errorf("S3 client not initialized")
	}

	if key == "" {
		return "", fmt.Errorf("S3 key cannot be empty")
	}

	if expirationDuration == 0 {
		expirationDuration = 15 * time.Minute
	}

	getObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(cfg.S3.Bucket),
		Key:    aws.String(key),
	}

	presignClient := s3.NewPresignClient(s3Client)
	presignedURL, err := presignClient.PresignGetObject(ctx, getObjectInput,
		func(opts *s3.PresignOptions) {
			opts.Expires = expirationDuration
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.URL, nil
}

// GenerateSignedDownloadURL is like GenerateSignedURL but forces the browser to
// download the object as an attachment named downloadFilename, via S3's
// response-content-disposition override. Use this when serving a stored file
// under a friendly, human-readable name distinct from its opaque S3 key.
func GenerateSignedDownloadURL(
	ctx context.Context,
	s3Client *s3.Client,
	cfg *config.Config,
	key string,
	downloadFilename string,
	expirationDuration time.Duration,
) (string, error) {
	if s3Client == nil {
		return "", fmt.Errorf("S3 client not initialized")
	}

	if key == "" {
		return "", fmt.Errorf("S3 key cannot be empty")
	}

	if expirationDuration == 0 {
		expirationDuration = 15 * time.Minute
	}

	contentDisposition := fmt.Sprintf("attachment; filename=%q", downloadFilename)

	getObjectInput := &s3.GetObjectInput{
		Bucket:                     aws.String(cfg.S3.Bucket),
		Key:                        aws.String(key),
		ResponseContentDisposition: aws.String(contentDisposition),
	}

	presignClient := s3.NewPresignClient(s3Client)
	presignedURL, err := presignClient.PresignGetObject(ctx, getObjectInput,
		func(opts *s3.PresignOptions) {
			opts.Expires = expirationDuration
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned download URL: %w", err)
	}

	return presignedURL.URL, nil
}
