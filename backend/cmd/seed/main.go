// cmd/seed/main.go
// Creates DynamoDB tables and inserts sample series + chapters.
// Usage: go run ./cmd/seed
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"mhentai-backend/internal/database"
	"mhentai-backend/internal/models"
)

func main() {
	if err := godotenv.Load(); err == nil {
		log.Println("Loaded .env")
	}
	database.Init()

	ctx := context.Background()

	createTables(ctx)
	seedData(ctx)

	log.Println("Done.")
}

// ─── Table Creation ────────────────────────────────────────────────────────

func createTables(ctx context.Context) {
	createSeriesTable(ctx)
	createChaptersTable(ctx)
}

func createSeriesTable(ctx context.Context) {
	name := database.TableSeries
	_, err := database.Client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(name),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("slug"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("slug-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("slug"), KeyType: types.KeyTypeHash},
				},
				Projection:            &types.Projection{ProjectionType: types.ProjectionTypeAll},
				ProvisionedThroughput: defaultThroughput(),
			},
		},
		ProvisionedThroughput: defaultThroughput(),
	})
	if err != nil {
		if isAlreadyExists(err) {
			log.Printf("Table %q already exists — skipping", name)
		} else {
			log.Fatalf("createSeriesTable: %v", err)
		}
		return
	}
	waitForTable(ctx, name)
	log.Printf("✓ Created table %q", name)
}

func createChaptersTable(ctx context.Context) {
	name := database.TableChapters
	_, err := database.Client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(name),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("slug"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("series_id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("number"), AttributeType: types.ScalarAttributeTypeN},
			{AttributeName: aws.String("#type"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("created_at"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("slug-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("slug"), KeyType: types.KeyTypeHash},
				},
				Projection:            &types.Projection{ProjectionType: types.ProjectionTypeAll},
				ProvisionedThroughput: defaultThroughput(),
			},
			{
				IndexName: aws.String("series_id-number-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("series_id"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("number"), KeyType: types.KeyTypeRange},
				},
				Projection:            &types.Projection{ProjectionType: types.ProjectionTypeAll},
				ProvisionedThroughput: defaultThroughput(),
			},
			{
				IndexName: aws.String("type-created_at-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("#type"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("created_at"), KeyType: types.KeyTypeRange},
				},
				Projection:            &types.Projection{ProjectionType: types.ProjectionTypeAll},
				ProvisionedThroughput: defaultThroughput(),
			},
		},
		ProvisionedThroughput: defaultThroughput(),
	})
	if err != nil {
		if isAlreadyExists(err) {
			log.Printf("Table %q already exists — skipping", name)
		} else {
			log.Fatalf("createChaptersTable: %v", err)
		}
		return
	}
	waitForTable(ctx, name)
	log.Printf("✓ Created table %q", name)
}

// ─── Seed Data ─────────────────────────────────────────────────────────────

var sampleSeries = []models.Series{
	{
		Slug:         "the-beginning-after-the-end",
		Title:        "The Beginning After the End",
		CoverURL:     "https://via.placeholder.com/300x420/1a1a2e/7c3aed?text=TBATE",
		Description:  "King Grey has unrivaled strength, wealth, and prestige in a world governed by martial ability. However, solitude lingers closely behind those with great power.",
		Status:       "ongoing",
		Author:       "TurtleMe",
		Genres:       "Action, Adventure, Fantasy, Romance",
		ViewCount:    12540,
		ChapterCount: 3,
		SourceURL:    "https://example.com/manga/the-beginning-after-the-end",
	},
	{
		Slug:         "solo-leveling",
		Title:        "Solo Leveling",
		CoverURL:     "https://via.placeholder.com/300x420/1a1a2e/dc2626?text=SL",
		Description:  "In this world where Hunters with various superpowers battle monsters from invading dungeons, Sung Jinwoo is the weakest of all the Hunters.",
		Status:       "completed",
		Author:       "Chugong",
		Genres:       "Action, Adventure, Fantasy",
		ViewCount:    98700,
		ChapterCount: 3,
		SourceURL:    "https://example.com/manga/solo-leveling",
	},
	{
		Slug:         "omniscient-readers-viewpoint",
		Title:        "Omniscient Reader's Viewpoint",
		CoverURL:     "https://via.placeholder.com/300x420/1a1a2e/0ea5e9?text=ORV",
		Description:  "Only I know the end of this world. One day our MC finds himself stuck in the world of his favorite web novel.",
		Status:       "ongoing",
		Author:       "sing N song",
		Genres:       "Action, Adventure, Drama, Fantasy",
		ViewCount:    34200,
		ChapterCount: 3,
		SourceURL:    "https://example.com/manga/omniscient-readers-viewpoint",
	},
}

// sample placeholder images (16:9 ratio panels)
var placeholderImages = [][]string{
	{
		"https://via.placeholder.com/800x1200/111827/ffffff?text=Page+1",
		"https://via.placeholder.com/800x1200/111827/ffffff?text=Page+2",
		"https://via.placeholder.com/800x1200/111827/ffffff?text=Page+3",
		"https://via.placeholder.com/800x1200/111827/ffffff?text=Page+4",
		"https://via.placeholder.com/800x1200/111827/ffffff?text=Page+5",
	},
	{
		"https://via.placeholder.com/800x1200/1f2937/ffffff?text=Page+1",
		"https://via.placeholder.com/800x1200/1f2937/ffffff?text=Page+2",
		"https://via.placeholder.com/800x1200/1f2937/ffffff?text=Page+3",
		"https://via.placeholder.com/800x1200/1f2937/ffffff?text=Page+4",
	},
	{
		"https://via.placeholder.com/800x1200/0f172a/ffffff?text=Page+1",
		"https://via.placeholder.com/800x1200/0f172a/ffffff?text=Page+2",
		"https://via.placeholder.com/800x1200/0f172a/ffffff?text=Page+3",
		"https://via.placeholder.com/800x1200/0f172a/ffffff?text=Page+4",
		"https://via.placeholder.com/800x1200/0f172a/ffffff?text=Page+5",
		"https://via.placeholder.com/800x1200/0f172a/ffffff?text=Page+6",
	},
}

var chapterTitles = [][]string{
	{"The King's Death", "A New World", "First Steps"},
	{"The Weakest Hunter", "The System Awakens", "Rise of the Shadow Monarch"},
	{"The Novel Ends", "A New Story Begins", "The Reader's Choice"},
}

func seedData(ctx context.Context) {
	now := time.Now().UTC()

	for si, tmpl := range sampleSeries {
		// Check if slug already exists
		existing, _ := querySlug(ctx, database.TableSeries, tmpl.Slug)
		if existing {
			log.Printf("Series %q already exists — skipping", tmpl.Slug)
			continue
		}

		s := tmpl
		s.ID = uuid.NewString()
		s.Type = "SERIES"
		s.CreatedAt = now.Add(-time.Duration(si*24) * time.Hour).Format(time.RFC3339)
		s.UpdatedAt = now.Add(-time.Duration(si) * time.Hour).Format(time.RFC3339)

		if err := putItem(ctx, database.TableSeries, s); err != nil {
			log.Printf("  ✗ series %q: %v", s.Title, err)
			continue
		}
		log.Printf("✓ Series: %q (id=%s)", s.Title, s.ID)

		titles := chapterTitles[si]
		images := placeholderImages[si]
		for ci, title := range titles {
			num := float64(ci + 1)
			chSlug := fmt.Sprintf("%s-chapter-%d", s.Slug, ci+1)

			ch := models.Chapter{
				ID:        uuid.NewString(),
				Type:      "CHAPTER",
				SeriesID:  s.ID,
				Slug:      chSlug,
				Title:     title,
				Number:    num,
				Images:    images,
				ViewCount: int64((ci + 1) * 300),
				CreatedAt: now.Add(-time.Duration((len(titles)-ci)*2) * time.Hour).Format(time.RFC3339),
				UpdatedAt: now.Add(-time.Duration(len(titles)-ci) * time.Hour).Format(time.RFC3339),
			}
			if err := putItem(ctx, database.TableChapters, ch); err != nil {
				log.Printf("  ✗ chapter %q: %v", title, err)
				continue
			}
			log.Printf("  ✓ Chapter %.0f: %q", num, title)
		}
	}
}

// ─── Helpers ───────────────────────────────────────────────────────────────

func putItem(ctx context.Context, table string, v interface{}) error {
	item, err := attributevalue.MarshalMap(v)
	if err != nil {
		return err
	}
	_, err = database.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item:      item,
	})
	return err
}

func querySlug(ctx context.Context, table, slug string) (bool, error) {
	out, err := database.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(table),
		IndexName:              aws.String("slug-index"),
		KeyConditionExpression: aws.String("slug = :slug"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":slug": &types.AttributeValueMemberS{Value: slug},
		},
		Select: types.SelectCount,
	})
	if err != nil {
		return false, err
	}
	return out.Count > 0, nil
}

func waitForTable(ctx context.Context, name string) {
	log.Printf("Waiting for table %q to become active...", name)
	waiter := dynamodb.NewTableExistsWaiter(database.Client)
	if err := waiter.Wait(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(name),
	}, 2*time.Minute); err != nil {
		log.Fatalf("waitForTable %q: %v", name, err)
	}
}

func defaultThroughput() *types.ProvisionedThroughput {
	return &types.ProvisionedThroughput{
		ReadCapacityUnits:  aws.Int64(5),
		WriteCapacityUnits: aws.Int64(5),
	}
}

func isAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*types.ResourceInUseException)
	return ok
}
