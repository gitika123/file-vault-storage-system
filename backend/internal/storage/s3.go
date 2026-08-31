package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Store uses any S3-compatible service (AWS S3, MinIO, or a cloud provider)
// as shared blob storage. Temporary staging remains local and is removed after upload.
type S3Store struct {
	Client   *s3.Client
	Bucket   string
	TempRoot string
}

func NewS3Store(ctx context.Context, bucket, region, endpoint, accessKey, secretKey, tempRoot string) (*S3Store, error) {
	loadOptions := []func(*awsconfig.LoadOptions) error{awsconfig.WithRegion(region)}
	if accessKey != "" || secretKey != "" {
		loadOptions = append(loadOptions, awsconfig.WithCredentialsProvider(aws.NewCredentialsCache(aws.CredentialsProviderFunc(func(context.Context) (aws.Credentials, error) {
			return aws.Credentials{AccessKeyID: accessKey, SecretAccessKey: secretKey}, nil
		}))))
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, err
	}
	if endpoint != "" {
		cfg.BaseEndpoint = aws.String(endpoint)
	}
	return &S3Store{Client: s3.NewFromConfig(cfg, func(o *s3.Options) { o.UsePathStyle = endpoint != "" }), Bucket: bucket, TempRoot: tempRoot}, nil
}

func (s *S3Store) Initialize() error { return os.MkdirAll(filepath.Join(s.TempRoot, ".tmp"), 0o750) }

func (s *S3Store) Stage(ctx context.Context, source io.Reader, maxBytes int64) (TempObject, error) {
	return (LocalStore{Root: s.TempRoot}).Stage(ctx, source, maxBytes)
}
func (s *S3Store) Commit(temp TempObject, key string) error {
	if err := validateKey(key); err != nil {
		return err
	}
	f, err := os.Open(temp.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = s.Client.PutObject(context.Background(), &s3.PutObjectInput{Bucket: aws.String(s.Bucket), Key: aws.String(key), Body: f})
	if err != nil {
		return fmt.Errorf("put blob in object storage: %w", err)
	}
	return os.Remove(temp.Path)
}
func (s *S3Store) Open(key string) (*os.File, error) {
	if err := validateKey(key); err != nil {
		return nil, err
	}
	out, err := s.Client.GetObject(context.Background(), &s3.GetObjectInput{Bucket: aws.String(s.Bucket), Key: aws.String(key)})
	if err != nil {
		return nil, err
	}
	tmp, err := os.CreateTemp(filepath.Join(s.TempRoot, ".tmp"), "download-*")
	if err != nil {
		out.Body.Close()
		return nil, err
	}
	if _, err = io.Copy(tmp, out.Body); err != nil {
		out.Body.Close()
		tmp.Close()
		os.Remove(tmp.Name())
		return nil, err
	}
	out.Body.Close()
	if err = tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return nil, err
	}
	return tmp, nil
}
func (s *S3Store) Delete(key string) error {
	if err := validateKey(key); err != nil {
		return err
	}
	_, err := s.Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{Bucket: aws.String(s.Bucket), Key: aws.String(key)})
	return err
}
func (s *S3Store) Discard(temp TempObject) error { return os.Remove(temp.Path) }
