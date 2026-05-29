package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// R2Store wraps the S3 client pointed at Cloudflare R2.
type R2Store struct {
	client    *s3.Client
	bucket    string
	publicURL string // e.g. https://pub-xxx.r2.dev or https://cdn.yourdomain.com
}

var R2 *R2Store

// InitR2 creates the R2 client using environment variables:
//
//	R2_ACCOUNT_ID        — Cloudflare account ID
//	R2_ACCESS_KEY_ID     — R2 API token access key
//	R2_SECRET_ACCESS_KEY — R2 API token secret key
//	R2_BUCKET            — bucket name (default: mhentai)
//	R2_PUBLIC_URL        — public base URL for serving objects
func InitR2() {
	accountID := os.Getenv("R2_ACCOUNT_ID")
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	bucket := getEnv("R2_BUCKET", "mhentai")
	publicURL := strings.TrimRight(os.Getenv("R2_PUBLIC_URL"), "/")

	if accountID == "" || accessKey == "" || secretKey == "" {
		log.Println("R2 storage not configured (R2_ACCOUNT_ID / R2_ACCESS_KEY_ID / R2_SECRET_ACCESS_KEY missing) — image proxy disabled")
		return
	}

	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("auto"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		log.Printf("R2 config error: %v — image proxy disabled", err)
		return
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	R2 = &R2Store{client: client, bucket: bucket, publicURL: publicURL}
	log.Printf("R2 storage ready (bucket=%s, public_url=%s)", bucket, publicURL)
}

// Enabled returns whether R2 is configured.
func (r *R2Store) Enabled() bool {
	return r != nil && r.client != nil
}

// UploadFromURL downloads an image from srcURL and uploads it to R2.
// referer is sent as the Referer header; pass "" to omit it.
// Returns the public URL of the uploaded object.
func (r *R2Store) UploadFromURL(ctx context.Context, srcURL, key, referer string) (string, error) {
	// Download
	req, _ := http.NewRequestWithContext(ctx, "GET", srcURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download status %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	_, err = r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		Body:        resp.Body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("upload to R2: %w", err)
	}

	return r.PublicURL(key), nil
}

// UploadReader uploads from an io.Reader directly.
func (r *R2Store) UploadReader(ctx context.Context, key, contentType string, body io.Reader) (string, error) {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("upload to R2: %w", err)
	}
	return r.PublicURL(key), nil
}

// PresignPut returns a presigned PUT URL valid for 1 hour.
func (r *R2Store) PresignPut(ctx context.Context, key, contentType string) (string, error) {
	presigner := s3.NewPresignClient(r.client)
	res, err := presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(time.Hour))
	if err != nil {
		return "", err
	}
	return res.URL, nil
}

// Delete removes an object from R2.
func (r *R2Store) Delete(ctx context.Context, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	return err
}

// PublicURL builds the public URL for an object key.
func (r *R2Store) PublicURL(key string) string {
	if r.publicURL != "" {
		return r.publicURL + "/" + key
	}
	// Fall back to path-style (not ideal for production)
	return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/%s", "auto", r.bucket, key)
}

// ImageKey generates a deterministic R2 key for a chapter image.
// e.g. "chapters/my-series/chapter-1/001.jpg"
func ImageKey(seriesSlug, chapterSlug string, index int, srcURL string) string {
	ext := path.Ext(srcURL)
	if ext == "" || len(ext) > 5 {
		ext = ".jpg"
	}
	// Strip query string from ext
	if i := strings.Index(ext, "?"); i != -1 {
		ext = ext[:i]
	}
	return fmt.Sprintf("chapters/%s/%s/%03d%s", seriesSlug, chapterSlug, index+1, ext)
}

// CoverKey generates a deterministic R2 key for a series cover.
func CoverKey(seriesSlug, srcURL string) string {
	ext := path.Ext(srcURL)
	if ext == "" || len(ext) > 5 {
		ext = ".jpg"
	}
	if i := strings.Index(ext, "?"); i != -1 {
		ext = ext[:i]
	}
	return fmt.Sprintf("covers/%s/cover%s", seriesSlug, ext)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
