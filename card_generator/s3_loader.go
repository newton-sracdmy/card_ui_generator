package card_generator

import (
	"bufio"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	_ "golang.org/x/image/webp"
)

const (
	DefaultS3Region = "ap-south-1"

	DefaultS3BucketName = "one-card-035359787041-ap-south-1-an"

	DefaultS3Timeout = 20 * time.Second

	MaxImageDownloadBytes = 20 * 1024 * 1024
)

type S3Config struct {
	BucketName    string
	Region        string
	Timeout       time.Duration
	MaxImageBytes int64
}

type S3Loader struct {
	client     *s3.Client
	bucketName string
	timeout    time.Duration
	maxBytes   int64
}

var (
	cardAssetCache sync.Map

	defaultLoader   *S3Loader
	defaultLoaderMu sync.RWMutex
	initLock        sync.Mutex
)

func getFirstEnv(keys ...string) string {
	for _, k := range keys {
		if val := strings.TrimSpace(os.Getenv(k)); val != "" {
			return val
		}
	}
	return ""
}

func getS3Env(keys ...string) string {
	return getFirstEnv(keys...)
}

func LoadS3ConfigFromEnv() S3Config {
	bucket := getFirstEnv("AWS_S3_BUCKET", "AWS_BUCKET_NAME", "S3_BUCKET_NAME", "AWS_S3_BUCKET_NAME", "s3_backet_name")
	if bucket == "" {
		bucket = DefaultS3BucketName
	}

	region := getFirstEnv("AWS_REGION", "AWS_DEFAULT_REGION")
	if region == "" {
		region = DefaultS3Region
	}

	return S3Config{
		BucketName:    bucket,
		Region:        region,
		Timeout:       DefaultS3Timeout,
		MaxImageBytes: MaxImageDownloadBytes,
	}
}

func NewS3Loader(ctx context.Context, cfg S3Config) (*S3Loader, error) {
	if cfg.BucketName == "" {
		return nil, fmt.Errorf("FAILED_TO_CREATE_S3_LOADER: s3 bucket name is not configured")
	}
	if cfg.Region == "" {
		cfg.Region = DefaultS3Region
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultS3Timeout
	}
	if cfg.MaxImageBytes <= 0 {
		cfg.MaxImageBytes = MaxImageDownloadBytes
	}

	var configOpts []func(*config.LoadOptions) error
	configOpts = append(configOpts, config.WithRegion(cfg.Region))

	accessKey := getFirstEnv("AWS_ACCESS_KEY_ID", "AWS_ACCESS_KEY")
	secretKey := getFirstEnv("AWS_SECRET_ACCESS_KEY", "AWS_SECRET_KEY")
	if accessKey != "" && secretKey != "" {
		configOpts = append(configOpts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		))
	}

	if ctx == nil {
		ctx = context.Background()
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, configOpts...)
	if err != nil {
		return nil, fmt.Errorf("FAILED_TO_LOAD_AWS_CONFIGURATION: %w", err)
	}

	return &S3Loader{
		client:     s3.NewFromConfig(awsCfg),
		bucketName: cfg.BucketName,
		timeout:    cfg.Timeout,
		maxBytes:   cfg.MaxImageBytes,
	}, nil
}

func getDefaultS3Loader() (*S3Loader, error) {
	defaultLoaderMu.RLock()
	loader := defaultLoader
	defaultLoaderMu.RUnlock()
	if loader != nil {
		return loader, nil
	}

	initLock.Lock()
	defer initLock.Unlock()

	if defaultLoader != nil {
		return defaultLoader, nil
	}

	cfg := LoadS3ConfigFromEnv()
	newLoader, err := NewS3Loader(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	defaultLoaderMu.Lock()
	defaultLoader = newLoader
	defaultLoaderMu.Unlock()

	return defaultLoader, nil
}

func initS3Client() error {
	_, err := getDefaultS3Loader()
	return err
}

func ExtractObjectKey(keyOrURL string, bucketName string) string {
	val := strings.TrimSpace(keyOrURL)
	if val == "" {
		return ""
	}

	if strings.HasPrefix(val, "s3://") {
		trimmed := strings.TrimPrefix(val, "s3://")
		if idx := strings.Index(trimmed, "?"); idx != -1 {
			trimmed = trimmed[:idx]
		}
		parts := strings.SplitN(trimmed, "/", 2)
		if len(parts) == 2 {
			if unescaped, err := url.PathUnescape(parts[1]); err == nil {
				return unescaped
			}
			return parts[1]
		}
		return trimmed
	}

	if strings.HasPrefix(val, "http://") || strings.HasPrefix(val, "https://") {
		parsed, err := url.Parse(val)
		if err == nil && parsed.Host != "" {
			path := strings.TrimPrefix(parsed.Path, "/")
			if unescaped, err := url.PathUnescape(path); err == nil {
				path = unescaped
			}

			if bucketName != "" && strings.HasPrefix(path, bucketName+"/") {
				path = strings.TrimPrefix(path, bucketName+"/")
			}
			return path
		}
	}

	key := val
	if idx := strings.Index(key, "?"); idx != -1 {
		key = key[:idx]
	}
	key = strings.TrimPrefix(key, "/")

	if unescaped, err := url.PathUnescape(key); err == nil {
		key = unescaped
	}

	if bucketName != "" && strings.HasPrefix(key, bucketName+"/") {
		key = strings.TrimPrefix(key, bucketName+"/")
	}

	return key
}

func extractS3ObjectKey(keyOrURL string) string {
	cfg := LoadS3ConfigFromEnv()
	return ExtractObjectKey(keyOrURL, cfg.BucketName)
}

func isS3Source(value string) bool {
	val := strings.TrimSpace(value)
	if val == "" || strings.HasPrefix(val, "data:") {
		return false
	}

	if strings.HasPrefix(val, "s3://") {
		return true
	}

	if strings.HasPrefix(val, "http://") || strings.HasPrefix(val, "https://") {
		return strings.Contains(val, "amazonaws.com") && !strings.Contains(val, "X-Amz-Signature")
	}

	lower := strings.ToLower(val)
	return strings.Contains(val, "/") ||
		strings.HasSuffix(lower, ".png") ||
		strings.HasSuffix(lower, ".jpg") ||
		strings.HasSuffix(lower, ".jpeg") ||
		strings.HasSuffix(lower, ".webp")
}

func (l *S3Loader) DownloadImage(ctx context.Context, keyOrURL string) (image.Image, error) {
	if l == nil || l.client == nil {
		return nil, fmt.Errorf("FAILED_TO_DOWNLOAD_IMAGE: s3 client is not available")
	}

	key := ExtractObjectKey(keyOrURL, l.bucketName)
	if key == "" {
		return nil, fmt.Errorf("FAILED_TO_DOWNLOAD_IMAGE: invalid or empty s3 object key: %s", keyOrURL)
	}

	reqCtx := ctx
	if reqCtx == nil {
		reqCtx = context.Background()
	}
	if _, hasDeadline := reqCtx.Deadline(); !hasDeadline && l.timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(reqCtx, l.timeout)
		defer cancel()
	}

	result, err := l.client.GetObject(reqCtx, &s3.GetObjectInput{
		Bucket: aws.String(l.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("FAILED_TO_GET_OBJECT (bucket: %s, key: %s): %w", l.bucketName, key, err)
	}
	defer result.Body.Close()

	limitedReader := io.LimitReader(result.Body, l.maxBytes)
	bufferedReader := bufio.NewReader(limitedReader)

	img, _, err := image.Decode(bufferedReader)
	if err != nil {
		return nil, fmt.Errorf("FAILED_TO_DECODE_IMAGE (bucket: %s, key: %s): %w", l.bucketName, key, err)
	}

	return img, nil
}

func downloadImageFromS3(keyOrURL string) (image.Image, error) {
	loader, err := getDefaultS3Loader()
	if err != nil {
		return nil, err
	}
	return loader.DownloadImage(context.Background(), keyOrURL)
}

func DownloadImageFromS3WithContext(ctx context.Context, keyOrURL string) (image.Image, error) {
	loader, err := getDefaultS3Loader()
	if err != nil {
		return nil, err
	}
	return loader.DownloadImage(ctx, keyOrURL)
}
