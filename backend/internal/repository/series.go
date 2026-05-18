package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"mhentai-backend/internal/database"
	"mhentai-backend/internal/models"
)

// GetSeriesByID fetches a series by primary key.
func GetSeriesByID(ctx context.Context, id string) (*models.Series, error) {
	out, err := database.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(database.TableSeries),
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
	var s models.Series
	if err = attributevalue.UnmarshalMap(out.Item, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// GetSeriesBySlug fetches a series via the slug GSI.
func GetSeriesBySlug(ctx context.Context, slug string) (*models.Series, error) {
	out, err := database.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(database.TableSeries),
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
	var s models.Series
	if err = attributevalue.UnmarshalMap(out.Items[0], &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// GetSeriesBySourceURL scans the series table for a matching source URL.
func GetSeriesBySourceURL(ctx context.Context, sourceURL string) (*models.Series, error) {
	input := &dynamodb.ScanInput{
		TableName:        aws.String(database.TableSeries),
		FilterExpression: aws.String("#source_url = :source_url"),
		ExpressionAttributeNames: map[string]string{
			"#source_url": "source_url",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":source_url": &types.AttributeValueMemberS{Value: sourceURL},
		},
		Limit: aws.Int32(1),
	}

	out, err := database.Client.Scan(ctx, input)
	if err != nil {
		return nil, err
	}
	if len(out.Items) == 0 {
		return nil, ErrNotFound
	}
	var s models.Series
	if err = attributevalue.UnmarshalMap(out.Items[0], &s); err != nil {
		return nil, err
	}
	return &s, nil
}

type ListSeriesParams struct {
	Status   string // "" = all, "ongoing", "completed"
	SortBy   string // "updated_at" | "created_at" | "title" | "views"
	Search   string
	Language string // "" = all, "en", "my"
	Limit    int
	LastKey  map[string]types.AttributeValue
}

type ListSeriesResult struct {
	Items   []*models.Series
	LastKey map[string]types.AttributeValue
	Total   int // approximate from scan, 0 when using index queries
}

// ListSeries returns a paginated list of series. Uses a Scan for search/sort flexibility.
func ListSeries(ctx context.Context, p ListSeriesParams) (*ListSeriesResult, error) {
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 24
	}

	// Build filter expression
	filterExpr := ""
	exprNames := map[string]string{}
	exprVals := map[string]types.AttributeValue{}

	if p.Status != "" {
		filterExpr = "#status = :status"
		exprNames["#status"] = "status"
		exprVals[":status"] = &types.AttributeValueMemberS{Value: p.Status}
	}
	if p.Search != "" {
		searchFilter := "contains(#title, :search)"
		exprNames["#title"] = "title"
		exprVals[":search"] = &types.AttributeValueMemberS{Value: p.Search}
		if filterExpr != "" {
			filterExpr += " AND " + searchFilter
		} else {
			filterExpr = searchFilter
		}
	}
	if p.Language != "" {
		langFilter := "#language = :language"
		exprNames["#language"] = "language"
		exprVals[":language"] = &types.AttributeValueMemberS{Value: p.Language}
		if filterExpr != "" {
			filterExpr += " AND " + langFilter
		} else {
			filterExpr = langFilter
		}
	}

	// Scan the table (works for any filter/sort combination)
	input := &dynamodb.ScanInput{
		TableName: aws.String(database.TableSeries),
	}
	if filterExpr != "" {
		input.FilterExpression = aws.String(filterExpr)
		if len(exprNames) > 0 {
			input.ExpressionAttributeNames = exprNames
		}
		input.ExpressionAttributeValues = exprVals
	}
	if p.LastKey != nil {
		input.ExclusiveStartKey = p.LastKey
	}

	// Fetch enough items to sort client-side (DynamoDB Scan doesn't support ORDER BY)
	// For production scale, use GSIs with pre-sorted data. For now, scan all and sort.
	var allItems []*models.Series
	var lastKey map[string]types.AttributeValue

	for {
		out, err := database.Client.Scan(ctx, input)
		if err != nil {
			return nil, err
		}
		for _, item := range out.Items {
			var s models.Series
			if err = attributevalue.UnmarshalMap(item, &s); err == nil {
				allItems = append(allItems, &s)
			}
		}
		lastKey = out.LastEvaluatedKey
		if lastKey == nil {
			break
		}
		// Continue scanning if we haven't exhausted the table
		input.ExclusiveStartKey = lastKey
	}

	// Sort in-memory
	sortSeries(allItems, p.SortBy)

	total := len(allItems)

	// Manual pagination after sort
	start := 0
	if p.LastKey != nil {
		// Simple offset: not ideal but works for small-medium datasets
		// For true cursor pagination with DynamoDB, store offset in cursor
	}
	end := start + p.Limit
	if end > len(allItems) {
		end = len(allItems)
	}

	var nextKey map[string]types.AttributeValue
	if end < len(allItems) {
		// Encode the next page start index as a synthetic last key
		// (not a real DynamoDB key — we handle pagination in the handler via page offset)
		nextKey = map[string]types.AttributeValue{
			"_page_offset": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", end)},
		}
	}

	return &ListSeriesResult{
		Items:   allItems[start:end],
		LastKey: nextKey,
		Total:   total,
	}, nil
}

// ListSeriesPage returns a specific page of the sorted/filtered list.
func ListSeriesPage(ctx context.Context, p ListSeriesParams, page int) (*ListSeriesResult, error) {
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 24
	}
	if page < 1 {
		page = 1
	}

	filterExpr := ""
	exprNames := map[string]string{}
	exprVals := map[string]types.AttributeValue{}

	if p.Status != "" {
		filterExpr = "#status = :status"
		exprNames["#status"] = "status"
		exprVals[":status"] = &types.AttributeValueMemberS{Value: p.Status}
	}
	if p.Search != "" {
		cond := "contains(#title, :search)"
		exprNames["#title"] = "title"
		exprVals[":search"] = &types.AttributeValueMemberS{Value: p.Search}
		if filterExpr != "" {
			filterExpr += " AND " + cond
		} else {
			filterExpr = cond
		}
	}
	if p.Language != "" {
		langCond := "#language = :language"
		exprNames["#language"] = "language"
		exprVals[":language"] = &types.AttributeValueMemberS{Value: p.Language}
		if filterExpr != "" {
			filterExpr += " AND " + langCond
		} else {
			filterExpr = langCond
		}
	}

	input := &dynamodb.ScanInput{
		TableName: aws.String(database.TableSeries),
	}
	if filterExpr != "" {
		input.FilterExpression = aws.String(filterExpr)
		if len(exprNames) > 0 {
			input.ExpressionAttributeNames = exprNames
		}
		input.ExpressionAttributeValues = exprVals
	}

	var allItems []*models.Series
	for {
		out, err := database.Client.Scan(ctx, input)
		if err != nil {
			return nil, err
		}
		for _, item := range out.Items {
			var s models.Series
			if err = attributevalue.UnmarshalMap(item, &s); err == nil {
				allItems = append(allItems, &s)
			}
		}
		if out.LastEvaluatedKey == nil {
			break
		}
		input.ExclusiveStartKey = out.LastEvaluatedKey
	}

	sortSeries(allItems, p.SortBy)
	total := len(allItems)

	offset := (page - 1) * p.Limit
	end := offset + p.Limit
	if offset >= total {
		return &ListSeriesResult{Items: []*models.Series{}, Total: total}, nil
	}
	if end > total {
		end = total
	}

	return &ListSeriesResult{Items: allItems[offset:end], Total: total}, nil
}

// PutSeries creates or replaces a series item.
func PutSeries(ctx context.Context, s *models.Series) error {
	s.Type = "SERIES"
	now := time.Now().UTC().Format(time.RFC3339)
	if s.CreatedAt == "" {
		s.CreatedAt = now
	}
	s.UpdatedAt = now

	item, err := attributevalue.MarshalMap(s)
	if err != nil {
		return err
	}
	_, err = database.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(database.TableSeries),
		Item:      item,
	})
	return err
}

// UpdateSeriesFields does a partial update using UpdateItem.
func UpdateSeriesFields(ctx context.Context, id string, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}
	fields["updated_at"] = time.Now().UTC().Format(time.RFC3339)

	expr := "SET "
	exprNames := map[string]string{}
	exprVals := map[string]types.AttributeValue{}
	i := 0
	for k, v := range fields {
		placeholder := fmt.Sprintf(":v%d", i)
		namePlaceholder := fmt.Sprintf("#f%d", i)
		if i > 0 {
			expr += ", "
		}
		expr += fmt.Sprintf("%s = %s", namePlaceholder, placeholder)
		exprNames[namePlaceholder] = k
		av, err := attributevalue.Marshal(v)
		if err != nil {
			return err
		}
		exprVals[placeholder] = av
		i++
	}

	_, err := database.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(database.TableSeries),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression:          aws.String(expr),
		ExpressionAttributeNames:  exprNames,
		ExpressionAttributeValues: exprVals,
	})
	return err
}

// UpdateSeriesCoverURL updates the cover_url for a series.
func UpdateSeriesCoverURL(ctx context.Context, id, coverURL string) error {
	_, err := database.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(database.TableSeries),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression: aws.String("SET cover_url = :url, updated_at = :now"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":url": &types.AttributeValueMemberS{Value: coverURL},
			":now": &types.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
		},
	})
	return err
}

// IncrementSeriesViewCount atomically increments view_count.
func IncrementSeriesViewCount(ctx context.Context, id string) error {
	_, err := database.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(database.TableSeries),
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

// DeleteSeries removes a series item (chapters are deleted separately).
func DeleteSeries(ctx context.Context, id string) error {
	_, err := database.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(database.TableSeries),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	return err
}

// CountSeries returns the approximate item count by scanning the table.
func CountSeries(ctx context.Context) (int64, error) {
	out, err := database.Client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(database.TableSeries),
	})
	if err != nil {
		return 0, err
	}
	if out.Table.ItemCount == nil {
		return 0, nil
	}
	return *out.Table.ItemCount, nil
}

// SlugExists returns true if the slug is already taken.
func SlugExists(ctx context.Context, slug string) (bool, error) {
	out, err := database.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(database.TableSeries),
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
