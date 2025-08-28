package s3bucket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"email-parser-poc/internal/domain/entities"
	"email-parser-poc/internal/ports/outgoing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Storage struct {
	Client     *s3.Client
	BucketName string
}

func NewS3Storage(bucketName string, localstackEndpoint string) (outgoing.StorageService, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test-key", "test-secret", "test-session",
		)),
		config.WithRegion("us-east-1"),
		config.WithBaseEndpoint(localstackEndpoint),
	)

	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &Storage{
		Client:     client,
		BucketName: bucketName,
	}, nil
}

// Add the required methods to implement the StorageService interface
func (s *Storage) UploadEmails(ctx context.Context, emails *entities.EmailList) (string, error) {
	emailsJSON, err := json.MarshalIndent(emails, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal emails to JSON: %w", err)
	}

	filename := fmt.Sprintf("emails/emails_%s.json", time.Now().Format("20060102_150405"))

	_, err = s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.BucketName),
		Key:         aws.String(filename),
		Body:        bytes.NewReader(emailsJSON),
		ContentType: aws.String("application/json"),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload emails to S3: %w", err)
	}

	return filename, nil
}
