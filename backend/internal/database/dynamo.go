package database

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var Client *dynamodb.Client

// TableSeries is the DynamoDB table name for series.
var TableSeries = "mhentai_series"

// TableChapters is the DynamoDB table name for chapters.
var TableChapters = "mhentai_chapters"

// Init creates the DynamoDB client using environment variables:
//
//	AWS_REGION           (default: us-east-1)
//	AWS_ACCESS_KEY_ID
//	AWS_SECRET_ACCESS_KEY
//	DYNAMODB_ENDPOINT    (optional, for local DynamoDB / DynamoDB Local)
//	DYNAMODB_TABLE_SERIES   (default: mhentai_series)
//	DYNAMODB_TABLE_CHAPTERS (default: mhentai_chapters)
func Init() {
	region := getEnv("AWS_REGION", "us-east-1")
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")

	if t := os.Getenv("DYNAMODB_TABLE_SERIES"); t != "" {
		TableSeries = t
	}
	if t := os.Getenv("DYNAMODB_TABLE_CHAPTERS"); t != "" {
		TableChapters = t
	}

	var cfg aws.Config
	var err error

	if accessKey != "" && secretKey != "" {
		cfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		)
	} else {
		// Fall back to default credential chain (IAM role, ~/.aws/credentials, etc.)
		cfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(region),
		)
	}
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	opts := []func(*dynamodb.Options){}
	if endpoint != "" {
		opts = append(opts, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	}

	Client = dynamodb.NewFromConfig(cfg, opts...)
	log.Printf("DynamoDB client ready (region=%s, series_table=%s, chapters_table=%s)", region, TableSeries, TableChapters)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
