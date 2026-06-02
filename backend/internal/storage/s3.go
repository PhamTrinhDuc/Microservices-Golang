package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type S3StorageProvider struct {
	client     *s3.Client
	bucketName string
	region     string
	customHost string
}

func NewS3StorageProvider(bucket, region, accessKey, secretKey, customEndpoint string) (*S3StorageProvider, error) {
	var cfg aws.Config
	var err error

	if accessKey != "" && secretKey != "" {
		creds := credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
			config.WithCredentialsProvider(creds),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	}

	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if customEndpoint != "" {
			o.BaseEndpoint = aws.String(customEndpoint)
		}
	})

	return &S3StorageProvider{
		client:     client,
		bucketName: bucket,
		region:     region,
		customHost: customEndpoint,
	}, nil
}

func (p *S3StorageProvider) UploadFile(ctx context.Context, filename string, file io.Reader, size int64, contentType string) (string, error) {
	uniqueName := fmt.Sprintf("%s-%s", uuid.New().String(), filename)

	_, err := p.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(p.bucketName),
		Key:         aws.String(uniqueName),
		Body:        file,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	var publicURL string
	if p.customHost != "" {
		publicURL = fmt.Sprintf("%s/%s/%s", strings.TrimSuffix(p.customHost, "/"), p.bucketName, uniqueName)
	} else {
		publicURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", p.bucketName, p.region, uniqueName)
	}
	return publicURL, nil
}

func (p *S3StorageProvider) DeleteFile(ctx context.Context, fileURL string) error {
	parts := strings.Split(fileURL, "/")
	if len(parts) == 0 {
		return fmt.Errorf("invalid S3 file URL")
	}
	key := parts[len(parts)-1]

	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.bucketName),
		Key:    aws.String(key),
	})
	return err
}
