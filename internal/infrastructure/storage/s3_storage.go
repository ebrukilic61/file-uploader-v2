package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	client     *s3.Client
	bucketName string
	region     string
}

func NewS3Storage(bucketName, region string) (*S3Storage, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("AWS config yüklenemedi: %w", err)
	}
	return &S3Storage{
		client:     s3.NewFromConfig(cfg),
		bucketName: bucketName,
		region:     region,
	}, nil
}

func (s *S3Storage) Upload(file multipart.File, metadata map[string]string) (string, error) {
	key := metadata["filename"]
	if folder, ok := metadata["folder"]; ok {
		key = folder + "/" + key
	}

	_, err := s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:   aws.String(s.bucketName),
		Key:      aws.String(key),
		Body:     file,
		Metadata: metadata,
	})
	if err != nil {
		return "", fmt.Errorf("S3 upload hatası: %w", err)
	}

	// URL artık region alanından okunuyor
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucketName, s.region, key), nil
}

func (s *S3Storage) Download(fileID string) (multipart.File, error) {
	resp, err := s.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fileID),
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the object into a buffer
	tmpFile, err := os.CreateTemp("", "s3download-*")
	if err != nil {
		return nil, fmt.Errorf("geçici dosya oluşturulamadı: %w", err)
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("S3 dosyası kopyalanamadı: %w", err)
	}

	// Seek to beginning for reading
	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("dosya başına alınamadı: %w", err)
	}

	return tmpFile, nil
}

func (s *S3Storage) Delete(fileID string) error {
	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fileID),
	})
	return err
}
