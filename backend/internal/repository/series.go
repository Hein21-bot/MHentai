package repository

import (
	"context"
	"fmt"
	"strings"
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

// BatchGetSeriesByIDs fetches multiple series by primary key in a single BatchGetItem call.
// Returns a map of id → Series. Missing items are absent from the map (not an error).
func BatchGetSeriesByIDs(ctx context.Context, ids []string) (map[string]*models.Series, error) {
	if len(ids) == 0 {
		return map[string]*models.Series{}, nil
	}

	keys := make([]map[string]types.AttributeValue, len(ids))
	for i, id := range ids {
		keys[i] = map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		}
	}

	result := make(map[string]*models.Series, len(ids))
	remaining := keys

	for len(remaining) > 0 {
		out, err := database.Client.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
			RequestItems: map[string]types.KeysAndAttributes{
				database.TableSeries: {Keys: remaining},
			},
		})
		if err != nil {
			return nil, err
		}
		for _, item := range out.Responses[database.TableSeries] {
			var s models.Series
			if attributevalue.UnmarshalMap(item, &s) == nil {
				result[s.ID] = &s
			}
		}
		remaining = nil
		if unprocessed, ok := out.UnprocessedKeys[database.TableSeries]; ok {
			remaining = unprocessed.Keys
		}
	}

	return result, nil
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
// Note: Limit is intentionally omitted — on a Scan, DynamoDB applies Limit before
// FilterExpression, so Limit=1 would evaluate only one item and miss any non-first match.
func GetSeriesBySourceURL(ctx context.Context, sourceURL string) (*models.Series, error) {
	input := &dynamodb.ScanInput{
		TableName:        aws.String(database.TableSeries),
		FilterExpression: aws.String("source_url = :source_url"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":source_url": &types.AttributeValueMemberS{Value: sourceURL},
		},
	}
	for {
		out, err := database.Client.Scan(ctx, input)
		if err != nil {
			return nil, err
		}
		if len(out.Items) > 0 {
			var s models.Series
			if err = attributevalue.UnmarshalMap(out.Items[0], &s); err != nil {
				return nil, err
			}
			return &s, nil
		}
		if out.LastEvaluatedKey == nil {
			break
		}
		input.ExclusiveStartKey = out.LastEvaluatedKey
	}
	return nil, ErrNotFound
}

type ListSeriesParams struct {
	Status   string // "" = all, "ongoing", "completed"
	SortBy   string // "updated_at" | "created_at" | "title" | "views"
	Search   string
	Language string // "" = all, "en", "my"
	Genre    string // "" = all; applied in-memory from cache, not in DynamoDB
	Letter   string // "" = all; "A"-"Z", "0-9", "#" — applied in-memory from cache
	Limit    int
	LastKey  map[string]types.AttributeValue
}

type ListSeriesResult struct {
	Items   []*models.Series
	LastKey map[string]types.AttributeValue
	Total   int
}

// ListSeriesPage returns a specific page of the sorted/filtered list.
// Results are cached for 2 minutes keyed by (status, sort, search, language).
func ListSeriesPage(ctx context.Context, p ListSeriesParams, page int) (*ListSeriesResult, error) {
	if p.Limit <= 0 || p.Limit > 500 {
		p.Limit = 24
	}
	if page < 1 {
		page = 1
	}

	key := listCacheKey(p)

	allItems, cached := cacheGetList(key)
	if !cached {
		var err error
		allItems, err = scanAllSeries(ctx, p)
		if err != nil {
			return nil, err
		}
		cacheSetList(key, allItems)
	}

	// Letter filter applied in-memory so it reuses the shared cache entry.
	if p.Letter != "" {
		var lf []*models.Series
		for _, s := range allItems {
			runes := []rune(s.Title)
			if len(runes) == 0 {
				continue
			}
			first := strings.ToUpper(string(runes[0]))
			var match bool
			switch p.Letter {
			case "#":
				match = (first < "A" || first > "Z") && (first < "0" || first > "9")
			case "0-9":
				match = first >= "0" && first <= "9"
			default:
				match = first == strings.ToUpper(p.Letter)
			}
			if match {
				lf = append(lf, s)
			}
		}
		allItems = lf
	}

	// Genre filter applied in-memory so it reuses the shared cache entry.
	if p.Genre != "" {
		var gf []*models.Series
		for _, s := range allItems {
			for _, g := range strings.Split(s.Genres, ",") {
				if strings.TrimSpace(g) == p.Genre {
					gf = append(gf, s)
					break
				}
			}
		}
		allItems = gf
	}

	// Sort a copy so the cached slice (shared across sort orders) is never mutated.
	sorted := make([]*models.Series, len(allItems))
	copy(sorted, allItems)
	sortSeries(sorted, p.SortBy)

	total := len(sorted)
	offset := (page - 1) * p.Limit
	if offset >= total {
		return &ListSeriesResult{Items: []*models.Series{}, Total: total}, nil
	}
	end := offset + p.Limit
	if end > total {
		end = total
	}

	return &ListSeriesResult{Items: sorted[offset:end], Total: total}, nil
}

// GetAllSeries returns the full unsorted series list for a language from cache (or scans once).
// Used by genre/recommendation endpoints that need all series without pagination.
func GetAllSeries(ctx context.Context, lang string) ([]*models.Series, error) {
	key := listCacheKey(ListSeriesParams{Language: lang})
	items, cached := cacheGetList(key)
	if !cached {
		var err error
		items, err = scanAllSeries(ctx, ListSeriesParams{Language: lang})
		if err != nil {
			return nil, err
		}
		cacheSetList(key, items)
	}
	return items, nil
}

// scanAllSeries does the full DynamoDB scan with the given filters.
func scanAllSeries(ctx context.Context, p ListSeriesParams) ([]*models.Series, error) {
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
		input.ExpressionAttributeNames = exprNames
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
	return allItems, nil
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
	if err == nil {
		InvalidateSeriesCache()
	}
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
	if err == nil {
		InvalidateSeriesCache()
	}
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
	if err == nil {
		InvalidateSeriesCache()
	}
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
	if err == nil {
		InvalidateSeriesCache()
	}
	return err
}

// CountSeries returns the approximate item count via DescribeTable (no RCU cost).
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
