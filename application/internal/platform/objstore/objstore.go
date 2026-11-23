// Package objstore talks to the S3-compatible object storage (RustFS).
//
// Clients upload to the private raw bucket via presigned PUT; the app only
// HEADs / range-GETs raw objects to validate them. Processed files live in the
// media bucket and are read through the CDN: PublicURL builds that URL from a key.
// Bucket names and keys never leave the app; only public URLs do.
package objstore

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketRaw       string
	PublicBaseURL   string
}

// RustFS ignores the region, but SigV4 needs one.
const region = "us-east-1"

var ErrNotFound = errors.New("objstore: object not found")

type Store struct {
	client     *s3.Client
	presign    *s3.PresignClient
	bucketRaw  string
	publicBase string
}

// New builds the client. It performs no I/O.
func New(cfg Config) (*Store, error) {
	if _, err := url.ParseRequestURI(cfg.Endpoint); err != nil {
		return nil, errors.New("objstore: invalid endpoint url")
	}
	base, err := url.ParseRequestURI(cfg.PublicBaseURL)
	if err != nil || base.Host == "" {
		return nil, errors.New("objstore: invalid public base url")
	}

	client := s3.New(s3.Options{
		Region:       region,
		BaseEndpoint: aws.String(cfg.Endpoint),
		UsePathStyle: true,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		// Newer SDKs add CRC checksums by default, which S3-compatible stores
		// and presigned PUTs from browsers do not handle well.
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	})

	return &Store{
		client:     client,
		presign:    s3.NewPresignClient(client),
		bucketRaw:  cfg.BucketRaw,
		publicBase: strings.TrimRight(cfg.PublicBaseURL, "/"),
	}, nil
}

// PresignRawPut returns a URL the client uses to PUT exactly size bytes of
// contentType to key in the raw bucket. The URL is a credential: never log it.
func (s *Store) PresignRawPut(ctx context.Context, key, contentType string, size int64, ttl time.Duration) (string, error) {
	req, err := s.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucketRaw),
		Key:           aws.String(key),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("objstore: presign put: %w", err)
	}
	return req.URL, nil
}

type ObjectInfo struct {
	Size        int64
	ContentType string
}

// HeadRaw returns size and content type of a raw object, or ErrNotFound.
func (s *Store) HeadRaw(ctx context.Context, key string) (ObjectInfo, error) {
	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketRaw),
		Key:    aws.String(key),
	})
	if err != nil {
		return ObjectInfo{}, mapErr("head", err)
	}
	return ObjectInfo{
		Size:        aws.ToInt64(out.ContentLength),
		ContentType: aws.ToString(out.ContentType),
	}, nil
}

// ReadRawRange reads up to length bytes from offset of a raw object
// (used to read image / video headers).
func (s *Store) ReadRawRange(ctx context.Context, key string, offset, length int64) ([]byte, error) {
	if offset < 0 || length <= 0 {
		return nil, fmt.Errorf("objstore: invalid range offset=%d length=%d", offset, length)
	}
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketRaw),
		Key:    aws.String(key),
		Range:  aws.String(fmt.Sprintf("bytes=%d-%d", offset, offset+length-1)),
	})
	if err != nil {
		return nil, mapErr("get range", err)
	}
	defer out.Body.Close()

	b, err := io.ReadAll(io.LimitReader(out.Body, length))
	if err != nil {
		return nil, fmt.Errorf("objstore: read range: %w", err)
	}
	return b, nil
}

// PublicURL returns the CDN URL of a processed object in the media bucket.
func (s *Store) PublicURL(key string) string {
	return s.publicBase + "/" + strings.TrimLeft(key, "/")
}

func mapErr(op string, err error) error {
	var re *awshttp.ResponseError
	if errors.As(err, &re) && re.HTTPStatusCode() == 404 {
		return ErrNotFound
	}
	return fmt.Errorf("objstore: %s: %w", op, err)
}
