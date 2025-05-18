package repository

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/o1egl/govatar"
)

type StaticRepository interface {
	GetImages(urls []string) ([][]byte, error)
	UploadImage(fileBytes []byte, filename, contentType string) error
	DeleteImage(user_id int, filename string) error
	GenerateImage(contentType string, ismale bool) ([]byte, error)
}

type StaticRepo struct {
	Client     *minio.Client
	BucketName string
}

func (sr *StaticRepo) UploadImage(fileBytes []byte, filename, contentType string) error {
	ctx := context.Background()

	_, err := sr.Client.PutObject(ctx, sr.BucketName, filename,
		bytes.NewReader(fileBytes),
		int64(len(fileBytes)),
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return fmt.Errorf("failed to upload image to minio: %w", err)
	}
	return nil
}

func (sr *StaticRepo) GetImages(urls []string) ([][]byte, error) {
	var results [][]byte

	for _, objectName := range urls {
		obj, err := sr.Client.GetObject(context.Background(), sr.BucketName, objectName, minio.GetObjectOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get object %s: %w", objectName, err)
		}

		data, err := io.ReadAll(obj)
		if err != nil {
			return nil, fmt.Errorf("failed to read object %s: %w", objectName, err)
		}

		results = append(results, data)
	}

	return results, nil
}

func NewStaticRepo() (*StaticRepo, error) {
	endpoint := "minio:9000"
	accessKeyID := os.Getenv("MINIO_ROOT_USER")
	secretAccessKey := os.Getenv("MINIO_ROOT_PASSWORD")
	useSSL := false

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return &StaticRepo{}, err
	}

	bucketName := "profile-photos"
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, err
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}

	return &StaticRepo{
		Client:     minioClient,
		BucketName: bucketName,
	}, nil
}

func NewStaticRepoCl(client *minio.Client, bucket string) *StaticRepo {
	return &StaticRepo{
		Client:     client,
		BucketName: bucket,
	}
}

func (sr *StaticRepo) DeleteImage(user_id int, filename string) error {
	ctx := context.Background()

	return sr.Client.RemoveObject(ctx, sr.BucketName, filename, minio.RemoveObjectOptions{})
}

func (sr *StaticRepo) GenerateImage(contentType string, ismale bool) ([]byte, error) {
	var img image.Image
	var sex govatar.Gender
	if ismale {
		sex = govatar.MALE
	} else {
		sex = govatar.FEMALE
	}

	img, err := govatar.Generate(sex)
	if err != nil {
		return []byte{}, fmt.Errorf("error generating image: %v", err)
	}

	var buf bytes.Buffer

	if contentType == "image/png" {
		err = png.Encode(&buf, img)
	} else if contentType == "image/jpeg" {
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 75})
	}

	if err != nil {
		return []byte{}, fmt.Errorf("error generating image: %v", err)
	}

	ansBytes := buf.Bytes()
	return ansBytes, nil
}
