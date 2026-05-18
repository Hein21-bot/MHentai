package repository

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"mhentai-backend/internal/database"
	"mhentai-backend/internal/models"
)

var ErrNotFound = fmt.Errorf("not found")

// chapterLiteProjection is a ProjectionExpression that excludes the images field.
// Used when we only need chapter metadata, not image URLs.
const chapterLiteProjection = "id, series_id, slug, title, #n, #lang, view_count, source_url, created_at, updated_at"

var chapterLiteNames = map[string]string{
	"#n":    "number",
	"#lang": "language",
}

// GetChapterByID fetches a chapter by primary key (includes images).
func GetChapterByID(ctx context.Context, id string) (*models.Chapter, error) {
	out, err := database.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(database.TableChapters),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, err
	}
	if out.Item == nil {
		return nil, ErrNotFound
	}
	var ch models.Chapter
	if err = attributevalue.UnmarshalMap(out.Item, &ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

// GetChapterBySlug fetches a chapter via the slug GSI (includes images for reading).
func GetChapterBySlug(ctx context.Context, slug string) (*models.Chapter, error) {
	out, err := database.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(database.TableChapters),
		IndexName:              aws.String("slug-index"),
		KeyConditionExpression: aws.String("slug = :slug"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":slug": &types.AttributeValueMemberS{Value: slug},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, err
	}
	if len(out.Items) == 0 {
		return nil, ErrNotFound
	}
	var ch models.Chapter
	if err = attributevalue.UnmarshalMap(out.Items[0], &ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

// ListChaptersBySeries returns all chapters for a series, sorted by number ASC.
// Images are excluded — use GetChapterBySlug when images are needed.
func ListChaptersBySeries(ctx context.Context, seriesID string) ([]*models.Chapter, error) {
	out, err := database.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(database.TableChapters),
		IndexName:              aws.String("series_id-number-index"),
		KeyConditionExpression: aws.String("series_id = :sid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":sid": &types.AttributeValueMemberS{Value: seriesID},
		},
		ProjectionExpression:     aws.String(chapterLiteProjection),
		ExpressionAttributeNames: chapterLiteNames,
		ScanIndexForward:         aws.Bool(true), // ascending by number
	})
	if err != nil {
		return nil, err
	}
	var chapters []*models.Chapter
	for _, item := range out.Items {
		var ch models.Chapter
		if err = attributevalue.UnmarshalMap(item, &ch); err == nil {
			chapters = append(chapters, &ch)
		}
	}
	return chapters, nil
}

// LatestChaptersBySeries returns the last n chapters for a series (highest numbers first).
// Uses DynamoDB Limit instead of fetching all chapters — O(n) not O(total).
func LatestChaptersBySeries(ctx context.Context, seriesID string, n int) ([]*models.Chapter, error) {
	out, err := database.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(database.TableChapters),
		IndexName:              aws.String("series_id-number-index"),
		KeyConditionExpression: aws.String("series_id = :sid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":sid": &types.AttributeValueMemberS{Value: seriesID},
		},
		ProjectionExpression:     aws.String(chapterLiteProjection),
		ExpressionAttributeNames: chapterLiteNames,
		ScanIndexForward:         aws.Bool(false), // descending = highest number first
		Limit:                    aws.Int32(int32(n)),
	})
	if err != nil {
		return nil, err
	}
	var chapters []*models.Chapter
	for _, item := range out.Items {
		var ch models.Chapter
		if err = attributevalue.UnmarshalMap(item, &ch); err == nil {
			chapters = append(chapters, &ch)
		}
	}
	return chapters, nil
}

// GetAdjacentChapters returns prev and next chapters using two targeted GSI queries
// instead of fetching the full chapter list.
func GetAdjacentChapters(ctx context.Context, seriesID string, number float64) (prev *models.Chapter, next *models.Chapter) {
	numAV, err := attributevalue.Marshal(number)
	if err != nil {
		return nil, nil
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		out, err := database.Client.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(database.TableChapters),
			IndexName:              aws.String("series_id-number-index"),
			KeyConditionExpression: aws.String("series_id = :sid AND #n < :num"),
			ExpressionAttributeNames: map[string]string{
				"#n":    "number",
				"#lang": "language",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":sid": &types.AttributeValueMemberS{Value: seriesID},
				":num": numAV,
			},
			ProjectionExpression: aws.String(chapterLiteProjection),
			ScanIndexForward:     aws.Bool(false), // descending → closest lower number first
			Limit:                aws.Int32(1),
		})
		if err == nil && len(out.Items) > 0 {
			var ch models.Chapter
			if attributevalue.UnmarshalMap(out.Items[0], &ch) == nil {
				prev = &ch
			}
		}
	}()

	go func() {
		defer wg.Done()
		out, err := database.Client.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(database.TableChapters),
			IndexName:              aws.String("series_id-number-index"),
			KeyConditionExpression: aws.String("series_id = :sid AND #n > :num"),
			ExpressionAttributeNames: map[string]string{
				"#n":    "number",
				"#lang": "language",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":sid": &types.AttributeValueMemberS{Value: seriesID},
				":num": numAV,
			},
			ProjectionExpression: aws.String(chapterLiteProjection),
			ScanIndexForward:     aws.Bool(true), // ascending → closest higher number first
			Limit:                aws.Int32(1),
		})
		if err == nil && len(out.Items) > 0 {
			var ch models.Chapter
			if attributevalue.UnmarshalMap(out.Items[0], &ch) == nil {
				next = &ch
			}
		}
	}()

	wg.Wait()
	return
}

// LatestChapters returns the most recently added chapters across all series.
func LatestChapters(ctx context.Context, limit int, language string) ([]*models.Chapter, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(database.TableChapters),
		IndexName:              aws.String("type-created_at-index"),
		KeyConditionExpression: aws.String("#t = :type"),
		ExpressionAttributeNames: map[string]string{
			"#t":    "#type",
			"#lang": "language",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":type": &types.AttributeValueMemberS{Value: "CHAPTER"},
		},
		ProjectionExpression: aws.String("id, series_id, slug, title, #n, #lang, view_count, created_at, updated_at"),
		ScanIndexForward:     aws.Bool(false), // newest first
	}

	// Reuse the #n name already declared
	input.ExpressionAttributeNames["#n"] = "number"

	if language != "" {
		input.FilterExpression = aws.String("#lang = :language")
		input.ExpressionAttributeValues[":language"] = &types.AttributeValueMemberS{Value: language}
	} else {
		input.Limit = aws.Int32(int32(limit))
	}

	var chapters []*models.Chapter
	for {
		out, err := database.Client.Query(ctx, input)
		if err != nil {
			return nil, err
		}
		for _, item := range out.Items {
			var ch models.Chapter
			if err = attributevalue.UnmarshalMap(item, &ch); err == nil {
				chapters = append(chapters, &ch)
			}
		}
		if len(chapters) >= limit || out.LastEvaluatedKey == nil {
			break
		}
		input.ExclusiveStartKey = out.LastEvaluatedKey
	}

	if len(chapters) > limit {
		chapters = chapters[:limit]
	}
	return chapters, nil
}

// ListAllChapters scans all chapters (admin use). Includes images.
func ListAllChapters(ctx context.Context, seriesID string) ([]*models.Chapter, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(database.TableChapters),
	}
	if seriesID != "" {
		input.FilterExpression = aws.String("series_id = :sid")
		input.ExpressionAttributeValues = map[string]types.AttributeValue{
			":sid": &types.AttributeValueMemberS{Value: seriesID},
		}
	}

	var chapters []*models.Chapter
	for {
		out, err := database.Client.Scan(ctx, input)
		if err != nil {
			return nil, err
		}
		for _, item := range out.Items {
			var ch models.Chapter
			if err = attributevalue.UnmarshalMap(item, &ch); err == nil {
				chapters = append(chapters, &ch)
			}
		}
		if out.LastEvaluatedKey == nil {
			break
		}
		input.ExclusiveStartKey = out.LastEvaluatedKey
	}

	sort.Slice(chapters, func(i, j int) bool {
		return chapters[i].Number < chapters[j].Number
	})
	return chapters, nil
}

// PutChapter creates or replaces a chapter item.
func PutChapter(ctx context.Context, ch *models.Chapter) error {
	ch.Type = "CHAPTER"
	now := time.Now().UTC().Format(time.RFC3339)
	if ch.CreatedAt == "" {
		ch.CreatedAt = now
	}
	ch.UpdatedAt = now

	item, err := attributevalue.MarshalMap(ch)
	if err != nil {
		return err
	}
	_, err = database.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(database.TableChapters),
		Item:      item,
	})
	return err
}

// UpdateChaptersLanguage updates the language field on every chapter belonging to a series.
// Runs up to 25 concurrent UpdateItem calls to avoid N sequential round-trips.
func UpdateChaptersLanguage(ctx context.Context, seriesID, language string) error {
	chapters, err := ListAllChapters(ctx, seriesID)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)

	const maxConcurrent = 25
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	var firstErr error
	var errMu sync.Mutex

	for _, ch := range chapters {
		wg.Add(1)
		sem <- struct{}{}
		go func(id string) {
			defer wg.Done()
			defer func() { <-sem }()
			_, e := database.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
				TableName: aws.String(database.TableChapters),
				Key: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: id},
				},
				UpdateExpression: aws.String("SET #lang = :lang, updated_at = :now"),
				ExpressionAttributeNames: map[string]string{
					"#lang": "language",
				},
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":lang": &types.AttributeValueMemberS{Value: language},
					":now":  &types.AttributeValueMemberS{Value: now},
				},
			})
			if e != nil {
				errMu.Lock()
				if firstErr == nil {
					firstErr = e
				}
				errMu.Unlock()
			}
		}(ch.ID)
	}
	wg.Wait()
	return firstErr
}

// UpdateChapterImages replaces the images list for a chapter and optionally saves the source URL.
func UpdateChapterImages(ctx context.Context, id string, images []string, sourceURL string) error {
	av, _ := attributevalue.Marshal(images)
	expr := "SET images = :imgs, updated_at = :now"
	vals := map[string]types.AttributeValue{
		":imgs": av,
		":now":  &types.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
	}
	if sourceURL != "" {
		expr += ", source_url = :src"
		vals[":src"] = &types.AttributeValueMemberS{Value: sourceURL}
	}
	_, err := database.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(database.TableChapters),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression:          aws.String(expr),
		ExpressionAttributeValues: vals,
	})
	return err
}

// IncrementChapterViewCount atomically increments view_count.
func IncrementChapterViewCount(ctx context.Context, id string) error {
	_, err := database.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(database.TableChapters),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression: aws.String("SET view_count = view_count + :one"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":one": &types.AttributeValueMemberN{Value: "1"},
		},
	})
	return err
}

// DeleteChapter removes a chapter item.
func DeleteChapter(ctx context.Context, id string) error {
	_, err := database.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(database.TableChapters),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	return err
}

// DeleteChaptersBySeries deletes all chapters belonging to a series using BatchWriteItem.
// Groups deletes in batches of 25 (DynamoDB limit) instead of N sequential deletes.
func DeleteChaptersBySeries(ctx context.Context, seriesID string) error {
	chapters, err := ListAllChapters(ctx, seriesID)
	if err != nil {
		return err
	}
	if len(chapters) == 0 {
		return nil
	}

	const batchSize = 25
	for i := 0; i < len(chapters); i += batchSize {
		end := i + batchSize
		if end > len(chapters) {
			end = len(chapters)
		}
		requests := make([]types.WriteRequest, end-i)
		for j, ch := range chapters[i:end] {
			requests[j] = types.WriteRequest{
				DeleteRequest: &types.DeleteRequest{
					Key: map[string]types.AttributeValue{
						"id": &types.AttributeValueMemberS{Value: ch.ID},
					},
				},
			}
		}
		if _, err = database.Client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				database.TableChapters: requests,
			},
		}); err != nil {
			return err
		}
	}
	return nil
}

// ChapterSlugExists checks if a chapter slug is already in use.
func ChapterSlugExists(ctx context.Context, slug string) (bool, error) {
	out, err := database.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(database.TableChapters),
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

// CountChapters returns approximate chapter count via DescribeTable (no RCU cost).
func CountChapters(ctx context.Context) (int64, error) {
	out, err := database.Client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(database.TableChapters),
	})
	if err != nil {
		return 0, err
	}
	if out.Table.ItemCount == nil {
		return 0, nil
	}
	return *out.Table.ItemCount, nil
}

// SumImagesAndViews scans all chapters and series to compute total image count and total view count.
func SumImagesAndViews(ctx context.Context) (totalImages int64, totalViews int64) {
	chInput := &dynamodb.ScanInput{
		TableName:            aws.String(database.TableChapters),
		ProjectionExpression: aws.String("images, view_count"),
	}
	for {
		out, err := database.Client.Scan(ctx, chInput)
		if err != nil {
			break
		}
		for _, item := range out.Items {
			var ch models.Chapter
			if attributevalue.UnmarshalMap(item, &ch) == nil {
				totalImages += int64(len(ch.Images))
				totalViews += int64(ch.ViewCount)
			}
		}
		if out.LastEvaluatedKey == nil {
			break
		}
		chInput.ExclusiveStartKey = out.LastEvaluatedKey
	}

	sInput := &dynamodb.ScanInput{
		TableName:            aws.String(database.TableSeries),
		ProjectionExpression: aws.String("view_count"),
	}
	for {
		out, err := database.Client.Scan(ctx, sInput)
		if err != nil {
			break
		}
		for _, item := range out.Items {
			var s models.Series
			if attributevalue.UnmarshalMap(item, &s) == nil {
				totalViews += int64(s.ViewCount)
			}
		}
		if out.LastEvaluatedKey == nil {
			break
		}
		sInput.ExclusiveStartKey = out.LastEvaluatedKey
	}
	return
}
