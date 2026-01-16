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
) (*UploadResult, error) {
	if s3Client == nil {
		return nil, fmt.Errorf("S3 client not initialized")
	}

	filename = filepath.Base(filename)
	filename = strings.ReplaceAll(filename, "..", "")
	filename = strings.TrimSpace(filename)

	timestamp := time.Now().Unix()
	key := fmt.Sprintf("uploads/%d-%s", timestamp, filename)

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

	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.S3.Bucket, cfg.S3.Region, key)

	result := &UploadResult{
		URL:         url,
		Filename:    filename,
		Size:        size,
		ContentType: contentType,
		Key:         key,
		UploadedAt:  time.Now().Format(time.RFC3339),
	}

	return result, nil
}
