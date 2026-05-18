// cmd/wipe/main.go
// Deletes all series and chapters from DynamoDB tables configured in .env.
package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/joho/godotenv"

	"mhentai-backend/internal/database"
	"mhentai-backend/internal/repository"
)

func main() {
	if err := godotenv.Load(); err == nil {
		log.Println("Loaded .env")
	}
	database.Init()

	ctx := context.Background()
	if err := wipeSeries(ctx); err != nil {
		log.Fatalf("failed to wipe series: %v", err)
	}
	if err := wipeChapters(ctx); err != nil {
		log.Fatalf("failed to wipe orphan chapters: %v", err)
	}
	log.Println("Finished wiping all series and chapters.")
}

func wipeSeries(ctx context.Context) error {
	log.Println("Scanning series table...")
	return scanAndDelete(ctx, database.TableSeries, func(id string) error {
		log.Printf("Deleting series %q", id)
		return repository.DeleteSeries(ctx, id)
	})
}

func wipeChapters(ctx context.Context) error {
	log.Println("Scanning chapters table...")
	return scanAndDelete(ctx, database.TableChapters, func(id string) error {
		log.Printf("Deleting chapter %q", id)
		return repository.DeleteChapter(ctx, id)
	})
}

func scanAndDelete(ctx context.Context, table string, deleteFn func(string) error) error {
	input := &dynamodb.ScanInput{TableName: aws.String(table), ProjectionExpression: aws.String("id")}
	for {
		out, err := database.Client.Scan(ctx, input)
		if err != nil {
			return err
		}
		for _, item := range out.Items {
			var key struct{ ID string `dynamodbav:"id"` }
			if err := attributevalue.UnmarshalMap(item, &key); err != nil {
				return err
			}
			if err := deleteFn(key.ID); err != nil {
				return err
			}
		}
		if out.LastEvaluatedKey == nil {
			break
		}
		input.ExclusiveStartKey = out.LastEvaluatedKey
	}
	return nil
}
