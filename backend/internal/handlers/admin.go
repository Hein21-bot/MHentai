package handlers

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"mhentai-backend/internal/database"
	"mhentai-backend/internal/models"
	"mhentai-backend/internal/repository"
	"mhentai-backend/internal/scraper"
	"mhentai-backend/internal/storage"
	"mhentai-backend/internal/telegram"
)

var slugRe = regexp.MustCompile(`[^a-z0-9-]+`)

var orphanCleanupRunning = false

var (
	dedupRunning     = false
	dedupLastDeleted = -1 // -1 = never run; >= 0 after last run
)

// AdminDedupStatus GET /api/admin/chapters/duplicates/status
func AdminDedupStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"running": dedupRunning, "last_deleted": dedupLastDeleted})
}

// ImportJob tracks the progress of a background chapter import.
type ImportJob struct {
	Running  bool      `json:"running"`
	Total    int       `json:"total"`
	Done     int       `json:"done"`
	Saved    int       `json:"saved"`
	Skipped  int       `json:"skipped"`
	Failed   []float64 `json:"failed"`
	SeriesID string    `json:"series_id"`
	Title    string    `json:"title"`
}

var (
	importJobsMu sync.RWMutex
	importJobs   = map[string]*ImportJob{}
)

// AdminImportStatus GET /api/admin/import/status?job_id=X
func AdminImportStatus(c *gin.Context) {
	jobID := c.Query("job_id")
	importJobsMu.RLock()
	job, ok := importJobs[jobID]
	var resp ImportJob
	if ok {
		resp = *job
	}
	importJobsMu.RUnlock()
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// AdminOrphanCleanupStatus GET /api/admin/chapters/orphaned/status
func AdminOrphanCleanupStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"running": orphanCleanupRunning})
}

func makeSlug(title string) string {
	s := strings.ToLower(title)
	s = slugRe.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// AdminGetStats GET /api/admin/stats
func AdminGetStats(c *gin.Context) {
	totalSeries, _ := repository.CountSeries(c.Request.Context())
	totalChapters, _ := repository.CountChapters(c.Request.Context())
	totalImages, totalViews := repository.SumImagesAndViews(c.Request.Context())

	c.JSON(http.StatusOK, gin.H{
		"total_series":   totalSeries,
		"total_chapters": totalChapters,
		"total_images":   totalImages,
		"total_views":    totalViews,
	})
}

// AdminGetRecentSeries GET /api/admin/recent
func AdminGetRecentSeries(c *gin.Context) {
	result, err := repository.ListSeriesPage(c.Request.Context(), repository.ListSeriesParams{
		SortBy: "created_at",
		Limit:  5,
	}, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result.Items})
}

// AdminListSeries GET /api/admin/series?page=1&limit=15&search=
func AdminListSeries(c *gin.Context) {
	page := 1
	limit := 15
	fmt.Sscanf(c.DefaultQuery("page", "1"), "%d", &page)
	fmt.Sscanf(c.DefaultQuery("limit", "15"), "%d", &limit)
	search := c.Query("search")

	result, err := repository.ListSeriesPage(c.Request.Context(), repository.ListSeriesParams{
		SortBy:  "updated_at",
		Limit:   limit,
		Search:  search,
	}, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	totalPages := 0
	if limit > 0 {
		totalPages = (result.Total + limit - 1) / limit
	}
	c.JSON(http.StatusOK, gin.H{
		"data":        result.Items,
		"total":       result.Total,
		"page":        page,
		"total_pages": totalPages,
	})
}

// AdminUpdateSeries PUT /api/admin/series/:id
func AdminUpdateSeries(c *gin.Context) {
	id := c.Param("id")
	var fields map[string]interface{}
	if err := c.ShouldBindJSON(&fields); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	delete(fields, "id")
	delete(fields, "#type")

	if err := repository.UpdateSeriesFields(c.Request.Context(), id, fields); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If language changed, cascade to all chapters of this series
	if lang, ok := fields["language"].(string); ok && lang != "" {
		_ = repository.UpdateChaptersLanguage(c.Request.Context(), id, lang)
	}

	s, _ := repository.GetSeriesByID(c.Request.Context(), id)
	c.JSON(http.StatusOK, s)
}

// AdminCreateSeries POST /api/admin/series
func AdminCreateSeries(c *gin.Context) {
	var body struct {
		Title       string `json:"title"`
		Language    string `json:"language"`
		Status      string `json:"status"`
		Author      string `json:"author"`
		Genres      string `json:"genres"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}
	if body.Language == "" {
		body.Language = "en"
	}
	if body.Status == "" {
		body.Status = "ongoing"
	}
	now := time.Now().UTC().Format(time.RFC3339)
	slug := makeSlug(body.Title)
	s := &models.Series{
		ID:          uuid.NewString(),
		Slug:        slug,
		Title:       body.Title,
		Language:    body.Language,
		Status:      body.Status,
		Author:      body.Author,
		Genres:      body.Genres,
		Description: body.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repository.PutSeries(context.Background(), s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}

// AdminCreateChapter POST /api/admin/chapters
func AdminCreateChapter(c *gin.Context) {
	var body struct {
		SeriesID string  `json:"series_id"`
		Title    string  `json:"title"`
		Number   float64 `json:"number"`
		Language string  `json:"language"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.SeriesID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "series_id is required"})
		return
	}
	series, err := repository.GetSeriesByID(context.Background(), body.SeriesID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "series not found"})
		return
	}
	if body.Language == "" {
		body.Language = series.Language
	}
	now := time.Now().UTC().Format(time.RFC3339)
	slug := fmt.Sprintf("%s-chapter-%v", series.Slug, body.Number)
	ch := &models.Chapter{
		ID:        uuid.NewString(),
		SeriesID:  body.SeriesID,
		Slug:      slug,
		Title:     body.Title,
		Number:    body.Number,
		Language:  body.Language,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repository.PutChapter(context.Background(), ch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = repository.UpdateSeriesFields(context.Background(), body.SeriesID, map[string]interface{}{
		"chapter_count": series.ChapterCount + 1,
		"updated_at":    now,
	})
	c.JSON(http.StatusOK, ch)
}

// AdminDeleteOrphanedChapters DELETE /api/admin/chapters/orphaned
// Starts cleanup in the background (returns immediately) to avoid DynamoDB throttle.
// Scans 5 items per page with 2s delay — safe for 5 RCU provisioned tables.
func AdminDeleteOrphanedChapters(c *gin.Context) {
	if orphanCleanupRunning {
		c.JSON(http.StatusConflict, gin.H{"error": "Cleanup is already running, please wait."})
		return
	}
	orphanCleanupRunning = true
	ctx := context.Background()
	go func() {
		defer func() { orphanCleanupRunning = false }()
		seriesCache := map[string]bool{}
		si := &dynamodb.ScanInput{
			TableName:            aws.String(database.TableChapters),
			ProjectionExpression: aws.String("id, series_id"),
			Limit:                aws.Int32(5),
		}
		for {
			out, err := database.Client.Scan(ctx, si)
			if err != nil {
				return
			}
			for _, item := range out.Items {
				var ch models.Chapter
				if err := attributevalue.UnmarshalMap(item, &ch); err != nil {
					continue
				}
				if ch.SeriesID == "" {
					_ = repository.DeleteChapter(ctx, ch.ID)
					continue
				}
				exists, ok := seriesCache[ch.SeriesID]
				if !ok {
					_, serr := repository.GetSeriesByID(ctx, ch.SeriesID)
					exists = serr == nil
					seriesCache[ch.SeriesID] = exists
				}
				if !exists {
					_ = repository.DeleteChapter(ctx, ch.ID)
				}
			}
			if out.LastEvaluatedKey == nil {
				break
			}
			si.ExclusiveStartKey = out.LastEvaluatedKey
			time.Sleep(2 * time.Second)
		}
	}()
	c.JSON(http.StatusAccepted, gin.H{"message": "Orphan cleanup started in background. It will finish in a few minutes."})
}

// AdminDeduplicateChapters DELETE /api/admin/chapters/duplicates?series_id=X
// Starts dedup in the background (returns 202 immediately) to avoid DynamoDB throttle.
// Poll GET /api/admin/chapters/duplicates/status for completion.
func AdminDeduplicateChapters(c *gin.Context) {
	seriesID := c.Query("series_id")
	if seriesID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "series_id required"})
		return
	}
	if dedupRunning {
		c.JSON(http.StatusConflict, gin.H{"error": "Dedup is already running, please wait."})
		return
	}
	dedupRunning = true
	dedupLastDeleted = -1
	ctx := context.Background()
	go func() {
		defer func() { dedupRunning = false }()

		chapters, err := repository.ListAllChapters(ctx, seriesID)
		if err != nil {
			return
		}

		byNumber := map[float64][]*models.Chapter{}
		for _, ch := range chapters {
			byNumber[ch.Number] = append(byNumber[ch.Number], ch)
		}

		deleted := 0
		for _, group := range byNumber {
			if len(group) <= 1 {
				continue
			}
			best := group[0]
			for _, ch := range group[1:] {
				if len(ch.Images) > len(best.Images) ||
					(len(ch.Images) == len(best.Images) && ch.CreatedAt > best.CreatedAt) {
					best = ch
				}
			}
			for _, ch := range group {
				if ch.ID == best.ID {
					continue
				}
				_ = repository.DeleteChapter(ctx, ch.ID)
				deleted++
				time.Sleep(200 * time.Millisecond)
			}
		}

		remaining, _ := repository.ListAllChapters(ctx, seriesID)
		_ = repository.UpdateSeriesFields(ctx, seriesID, map[string]interface{}{
			"chapter_count": len(remaining),
		})
		dedupLastDeleted = deleted
	}()

	c.JSON(http.StatusAccepted, gin.H{"message": "Dedup started in background."})
}

// AdminDeleteEmptySeries DELETE /api/admin/series/empty
// Deletes all series that are garbage: 0 chapters, empty title, or invalid slug ("/", "").
func AdminDeleteEmptySeries(c *gin.Context) {
	all, err := repository.ListSeriesPage(c.Request.Context(), repository.ListSeriesParams{Limit: 500}, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	deleted := 0
	for _, s := range all.Items {
		isGarbage := s.ChapterCount == 0 ||
			s.Title == "" || s.Title == "/" ||
			s.Slug == "" || s.Slug == "/"
		if isGarbage {
			_ = repository.DeleteChaptersBySeries(c.Request.Context(), s.ID)
			if err := repository.DeleteSeries(c.Request.Context(), s.ID); err == nil {
				deleted++
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"deleted": deleted})
}

// AdminDeleteSeries DELETE /api/admin/series/:id
func AdminDeleteSeries(c *gin.Context) {
	id := c.Param("id")
	// Delete all chapters first (cascade)
	_ = repository.DeleteChaptersBySeries(c.Request.Context(), id)
	if err := repository.DeleteSeries(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// AdminListChapters GET /api/admin/chapters?series_id=X&cursor=&limit=20
func AdminListChapters(c *gin.Context) {
	seriesID := c.Query("series_id")
	if seriesID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "series_id required"})
		return
	}
	cursor := c.Query("cursor")
	limit := int32(20)

	page, err := repository.ListChaptersPage(c.Request.Context(), seriesID, limit, cursor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type chapterRow struct {
		models.Chapter
		ImageCount int `json:"image_count"`
	}
	rows := make([]chapterRow, 0, len(page.Chapters))
	for _, ch := range page.Chapters {
		rows = append(rows, chapterRow{Chapter: *ch, ImageCount: len(ch.Images)})
	}
	c.JSON(http.StatusOK, gin.H{
		"data":        rows,
		"next_cursor": page.NextCursor,
	})
}

// AdminDeleteChapter DELETE /api/admin/chapters/:id
func AdminDeleteChapter(c *gin.Context) {
	if err := repository.DeleteChapter(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// AdminPresignUpload POST /api/admin/upload/presign
// Returns a presigned PUT URL for uploading directly to R2.
func AdminPresignUpload(c *gin.Context) {
	if storage.R2 == nil || !storage.R2.Enabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "R2 storage not configured"})
		return
	}
	var req struct {
		Key         string `json:"key" binding:"required"`
		ContentType string `json:"content_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.ContentType == "" {
		req.ContentType = "image/jpeg"
	}
	url, err := storage.R2.PresignPut(c.Request.Context(), req.Key, req.ContentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"url":        url,
		"public_url": storage.R2.PublicURL(req.Key),
	})
}

// ImportRequest is the body for POST /api/admin/import
type ImportRequest struct {
	URL              string               `json:"url" binding:"required"`
	SelectedChapters []ImportChapterInput `json:"selected_chapters"` // chapter data from preview, used directly
	SelectedSlugs    []string             `json:"selected_slugs"`    // fallback: filter re-scraped list by slug
	Force            bool                 `json:"force"`             // overwrite existing chapters instead of skipping
}

type ImportChapterInput struct {
	Slug   string  `json:"slug"`
	Title  string  `json:"title"`
	Number float64 `json:"number"`
	URL    string  `json:"url"`
}

// AdminImport POST /api/admin/import
func AdminImport(c *gin.Context) {
	var req ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	info, err := scraper.ScrapeSeries(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("scrape failed: %v", err)})
		return
	}
	if info.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not extract series title"})
		return
	}

	// Try to reuse an existing series: first by source URL, then by slug
	// (covers the case where the series was manually created before importing)
	series, err := repository.GetSeriesBySourceURL(c.Request.Context(), req.URL)
	if err != nil && err != repository.ErrNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("finding existing series: %v", err)})
		return
	}

	if series == nil {
		baseSlug := info.Slug
		if baseSlug == "" {
			baseSlug = makeSlug(info.Title)
		}
		if existing, err := repository.GetSeriesBySlug(c.Request.Context(), baseSlug); err == nil && existing != nil {
			series = existing
		}
	}

	newlyCreated := false
	if series == nil {
		baseSlug := info.Slug
		if baseSlug == "" {
			baseSlug = makeSlug(info.Title)
		}
		if baseSlug == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "could not derive a slug from the series title"})
			return
		}
		slug := baseSlug
		for i := 1; ; i++ {
			exists, _ := repository.SlugExists(c.Request.Context(), slug)
			if !exists {
				break
			}
			slug = fmt.Sprintf("%s-%d", baseSlug, i)
		}

		coverURL := info.CoverURL
		if storage.R2 != nil && storage.R2.Enabled() && coverURL != "" {
			key := storage.CoverKey(slug, coverURL)
			if uploaded, err := storage.R2.UploadFromURL(c.Request.Context(), coverURL, key, "https://hentai20.io/"); err == nil {
				coverURL = uploaded
			}
		}

		series = &models.Series{
			ID:          uuid.NewString(),
			Slug:        slug,
			Title:       info.Title,
			CoverURL:    coverURL,
			Description: info.Description,
			Status:      info.Status,
			Author:      info.Author,
			Genres:      strings.Join(info.Genres, ", "),
			Language:    "en",
			SourceURL:   req.URL,
		}
		if err := repository.PutSeries(c.Request.Context(), series); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("saving series: %v", err)})
			return
		}
		newlyCreated = true
	} else {
		// Existing series found; update metadata if changed.
		// Preserve user-uploaded cover: only set cover_url if none exists yet.
		metaFields := map[string]interface{}{
			"title":       info.Title,
			"description": info.Description,
			"status":      info.Status,
			"author":      info.Author,
			"genres":      strings.Join(info.Genres, ", "),
		}
		if series.CoverURL == "" && info.CoverURL != "" {
			metaFields["cover_url"] = info.CoverURL
		}
		_ = repository.UpdateSeriesFields(c.Request.Context(), series.ID, metaFields)
	}

	// Build chapter list to import.
	// If the frontend sent selected_chapters (full data from preview), use those directly.
	// This avoids slug mismatch between the preview scrape and the import re-scrape.
	type chapterEntry struct {
		Slug  string
		Title string
		Num   float64
		URL   string
	}
	var toImport []chapterEntry
	if len(req.SelectedChapters) > 0 {
		for _, c := range req.SelectedChapters {
			toImport = append(toImport, chapterEntry{c.Slug, c.Title, c.Number, c.URL})
		}
	} else {
		selectedSlugs := make(map[string]bool, len(req.SelectedSlugs))
		for _, s := range req.SelectedSlugs {
			selectedSlugs[s] = true
		}
		seriesSlug := series.Slug
		for _, ch := range info.Chapters {
			chSlug := ch.Slug
			if chSlug == "" {
				chSlug = fmt.Sprintf("%s-chapter-%.0f", seriesSlug, ch.Number)
			}
			if len(selectedSlugs) > 0 && !selectedSlugs[chSlug] {
				continue
			}
			toImport = append(toImport, chapterEntry{chSlug, ch.Title, ch.Number, ch.URL})
		}
	}

	if len(toImport) == 0 {
		if newlyCreated {
			_ = repository.DeleteSeries(c.Request.Context(), series.ID)
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "series": nil, "chapters_saved": 0, "chapters_skipped": 0, "chapters_failed": []float64{}})
		return
	}

	jobID := uuid.NewString()
	job := &ImportJob{
		Running:  true,
		Total:    len(toImport),
		Failed:   []float64{},
		SeriesID: series.ID,
		Title:    series.Title,
	}
	importJobsMu.Lock()
	importJobs[jobID] = job
	importJobsMu.Unlock()

	bgCtx := context.Background()
	seriesSlug := series.Slug
	seriesChapterCount := series.ChapterCount
	force := req.Force
	isNew := newlyCreated
	go func() {
		defer func() {
			importJobsMu.Lock()
			job.Running = false
			importJobsMu.Unlock()
		}()
		for _, ch := range toImport {
			chSlug := ch.Slug
			if chSlug == "" {
				chSlug = fmt.Sprintf("%s-chapter-%.0f", seriesSlug, ch.Num)
			}
			existingID, _ := repository.GetChapterIDBySlug(bgCtx, chSlug)
			if existingID != "" {
				if !force {
					importJobsMu.Lock()
					job.Skipped++
					job.Done++
					importJobsMu.Unlock()
					continue
				}
				_ = repository.DeleteChapter(bgCtx, existingID)
			}
			chapter := &models.Chapter{
				ID:        uuid.NewString(),
				SeriesID:  series.ID,
				Slug:      chSlug,
				Title:     ch.Title,
				Number:    ch.Num,
				Language:  "en",
				SourceURL: ch.URL,
			}
			if ch.URL != "" {
				chapter.Images = scrapeAndOptionallyProxy(bgCtx, ch.URL, seriesSlug, chSlug, true)
				time.Sleep(500 * time.Millisecond)
			}
			importJobsMu.Lock()
			if err := repository.PutChapter(bgCtx, chapter); err != nil {
				job.Failed = append(job.Failed, ch.Num)
			} else {
				job.Saved++
			}
			job.Done++
			importJobsMu.Unlock()
		}
		importJobsMu.RLock()
		saved := job.Saved
		importJobsMu.RUnlock()
		_ = repository.UpdateSeriesFields(bgCtx, series.ID, map[string]interface{}{
			"chapter_count": seriesChapterCount + saved,
		})
		if isNew {
			telegram.NotifyNewSeries(telegram.SeriesInfo{
				Title:        series.Title,
				Description:  series.Description,
				CoverURL:     series.CoverURL,
				ChapterCount: seriesChapterCount + saved,
				Status:       series.Status,
				Genres:       series.Genres,
				Slug:         series.Slug,
			})
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"job_id": jobID,
		"series": series,
		"total":  len(toImport),
	})
}

// AdminImportChapterImages POST /api/admin/import/chapter-images
func AdminImportChapterImages(c *gin.Context) {
	var req struct {
		ChapterID  string `json:"chapter_id" binding:"required"`
		ChapterURL string `json:"chapter_url" binding:"required"`
		ProxyToR2  bool   `json:"proxy_to_r2"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ch, err := repository.GetChapterByID(c.Request.Context(), req.ChapterID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chapter not found"})
		return
	}

	seriesSlug := ""
	if ch.SeriesID != "" {
		s, _ := repository.GetSeriesByID(c.Request.Context(), ch.SeriesID)
		if s != nil {
			seriesSlug = s.Slug
		}
	}

	var images []string
	if strings.Contains(req.ChapterURL, "mangaboost.com") {
		images, _ = scraper.ScrapeMangaBoostChapterImages(req.ChapterURL)
		if len(images) == 0 {
			images = scrapeAndOptionallyProxy(c.Request.Context(), req.ChapterURL, seriesSlug, ch.Slug, req.ProxyToR2)
		}
	} else {
		images = scrapeAndOptionallyProxy(c.Request.Context(), req.ChapterURL, seriesSlug, ch.Slug, req.ProxyToR2)
	}
	if err := repository.UpdateChapterImages(c.Request.Context(), req.ChapterID, images, req.ChapterURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "images_count": len(images)})
}

// AdminScrapeChapterList POST /api/admin/import/chapters
func AdminScrapeChapterList(c *gin.Context) {
	var req struct {
		SeriesID  string `json:"series_id" binding:"required"`
		SourceURL string `json:"source_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	series, err := repository.GetSeriesByID(c.Request.Context(), req.SeriesID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "series not found"})
		return
	}

	srcURL := req.SourceURL
	if srcURL == "" {
		srcURL = series.SourceURL
	}

	info, err := scraper.ScrapeSeries(srcURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	saved := 0
	for _, ch := range info.Chapters {
		chSlug := ch.Slug
		if chSlug == "" {
			chSlug = fmt.Sprintf("%s-chapter-%.0f", series.Slug, ch.Number)
		}
		exists, _ := repository.ChapterSlugExists(c.Request.Context(), chSlug)
		if exists {
			continue
		}
		_ = repository.PutChapter(c.Request.Context(), &models.Chapter{
			ID:       uuid.NewString(),
			SeriesID: series.ID,
			Slug:     chSlug,
			Title:    ch.Title,
			Number:   ch.Number,
		})
		saved++
	}

	_ = repository.UpdateSeriesFields(c.Request.Context(), series.ID, map[string]interface{}{
		"chapter_count": series.ChapterCount + saved,
	})

	c.JSON(http.StatusOK, gin.H{"success": true, "chapters_added": saved})
}

// AdminPreviewURL POST /api/admin/import/preview-url
// Scrapes a URL and returns series info + chapter list without saving anything.
// Used by the import page to show a preview before the user confirms.
func AdminPreviewURL(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	info, err := scraper.ScrapeSeries(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("scrape failed: %v", err)})
		return
	}
	if info.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not extract series title from that URL"})
		return
	}

	type chapterPreview struct {
		Title  string  `json:"title"`
		Number float64 `json:"number"`
		Slug   string  `json:"slug"`
		URL    string  `json:"url"`
	}

	chapters := make([]chapterPreview, 0, len(info.Chapters))
	for _, ch := range info.Chapters {
		slug := ch.Slug
		if slug == "" {
			slug = fmt.Sprintf("%s-chapter-%.0f", makeSlug(info.Title), ch.Number)
		}
		chapters = append(chapters, chapterPreview{
			Title:  ch.Title,
			Number: ch.Number,
			Slug:   slug,
			URL:    ch.URL,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"title":       info.Title,
		"cover_url":   info.CoverURL,
		"description": info.Description,
		"status":      info.Status,
		"author":      info.Author,
		"genres":      strings.Join(info.Genres, ", "),
		"chapters":    chapters,
	})
}

// AdminPreviewChapters POST /api/admin/import/preview
// Scrapes chapter list from a source URL without saving anything.
// Returns each chapter with an "exists" flag showing if it's already imported.
func AdminPreviewChapters(c *gin.Context) {
	var req struct {
		SeriesID  string `json:"series_id" binding:"required"`
		SourceURL string `json:"source_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	series, err := repository.GetSeriesByID(c.Request.Context(), req.SeriesID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "series not found"})
		return
	}

	srcURL := req.SourceURL
	if srcURL == "" {
		srcURL = series.SourceURL
	}
	if srcURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no source_url provided and series has no source URL"})
		return
	}

	info, err := scraper.ScrapeSeries(srcURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	type previewChapter struct {
		Title  string  `json:"title"`
		Number float64 `json:"number"`
		Slug   string  `json:"slug"`
		URL    string  `json:"url"`
		Exists bool    `json:"exists"` // already imported
	}

	result := make([]previewChapter, 0, len(info.Chapters))
	for _, ch := range info.Chapters {
		chSlug := ch.Slug
		if chSlug == "" {
			chSlug = fmt.Sprintf("%s-chapter-%.0f", series.Slug, ch.Number)
		}
		exists, _ := repository.ChapterSlugExists(c.Request.Context(), chSlug)
		result = append(result, previewChapter{
			Title:  ch.Title,
			Number: ch.Number,
			Slug:   chSlug,
			URL:    ch.URL,
			Exists: exists,
		})
	}

	c.JSON(http.StatusOK, gin.H{"chapters": result, "series": series})
}

// AdminImportSelectedChapters POST /api/admin/import/selected
// Imports only the chapters whose slugs are in the "slugs" list.
func AdminImportSelectedChapters(c *gin.Context) {
	var req struct {
		SeriesID     string   `json:"series_id" binding:"required"`
		SourceURL    string   `json:"source_url"`
		Slugs        []string `json:"slugs" binding:"required"` // slugs to import
		ScrapeImages bool     `json:"scrape_images"`
		ProxyToR2    bool     `json:"proxy_to_r2"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	series, err := repository.GetSeriesByID(c.Request.Context(), req.SeriesID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "series not found"})
		return
	}

	srcURL := req.SourceURL
	if srcURL == "" {
		srcURL = series.SourceURL
	}

	info, err := scraper.ScrapeSeries(srcURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build a set of slugs the user selected
	wantSlug := make(map[string]bool, len(req.Slugs))
	for _, s := range req.Slugs {
		wantSlug[s] = true
	}

	saved := 0
	for _, ch := range info.Chapters {
		chSlug := ch.Slug
		if chSlug == "" {
			chSlug = fmt.Sprintf("%s-chapter-%.0f", series.Slug, ch.Number)
		}
		if !wantSlug[chSlug] {
			continue
		}
		exists, _ := repository.ChapterSlugExists(c.Request.Context(), chSlug)
		if exists {
			continue
		}

		chapter := &models.Chapter{
			ID:        uuid.NewString(),
			SeriesID:  series.ID,
			Slug:      chSlug,
			Title:     ch.Title,
			Number:    ch.Number,
			SourceURL: ch.URL,
		}
		if req.ScrapeImages && ch.URL != "" {
			chapter.Images = scrapeAndOptionallyProxy(c.Request.Context(), ch.URL, series.Slug, chSlug, req.ProxyToR2)
			time.Sleep(500 * time.Millisecond)
		}
		if err := repository.PutChapter(c.Request.Context(), chapter); err != nil {
			continue
		}
		saved++
	}

	_ = repository.UpdateSeriesFields(c.Request.Context(), series.ID, map[string]interface{}{
		"chapter_count": series.ChapterCount + saved,
	})

	c.JSON(http.StatusOK, gin.H{"success": true, "chapters_saved": saved})
}

// scrapeAndOptionallyProxy scrapes image URLs and optionally uploads them to R2.
func scrapeAndOptionallyProxy(ctx context.Context, chapterURL, seriesSlug, chapterSlug string, proxyToR2 bool) []string {
	rawImages, err := scraper.ScrapeChapterImages(chapterURL)
	if err != nil {
		return nil
	}

	if !proxyToR2 || storage.R2 == nil || !storage.R2.Enabled() {
		return rawImages
	}

	proxied := make([]string, 0, len(rawImages))
	for i, imgURL := range rawImages {
		key := storage.ImageKey(seriesSlug, chapterSlug, i, imgURL)
		if uploaded, err := storage.R2.UploadFromURL(ctx, imgURL, key, "https://hentai20.io/"); err == nil {
			proxied = append(proxied, uploaded)
		} else {
			proxied = append(proxied, imgURL) // fall back to original URL on error
		}
	}
	return proxied
}

// AdminPreviewMangaBoost POST /api/admin/import/mangaboost/preview
// Scrapes a mangaboost.com manga URL and returns series + chapter list without saving.
func AdminPreviewMangaBoost(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	info, err := scraper.ScrapeMangaBoostSeries(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("scrape failed: %v", err)})
		return
	}
	if info.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not extract series title from that URL"})
		return
	}

	type chapterPreview struct {
		Title  string  `json:"title"`
		Number float64 `json:"number"`
		Slug   string  `json:"slug"`
		URL    string  `json:"url"`
	}

	chapters := make([]chapterPreview, 0, len(info.Chapters))
	for _, ch := range info.Chapters {
		slug := ch.Slug
		if slug == "" {
			slug = fmt.Sprintf("%s-chapter-%.0f", makeSlug(info.Title), ch.Number)
		}
		chapters = append(chapters, chapterPreview{
			Title:  ch.Title,
			Number: ch.Number,
			Slug:   slug,
			URL:    ch.URL,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"title":       info.Title,
		"cover_url":   info.CoverURL,
		"description": info.Description,
		"status":      info.Status,
		"author":      info.Author,
		"genres":      strings.Join(info.Genres, ", "),
		"chapters":    chapters,
	})
}

// AdminImportMangaBoost POST /api/admin/import/mangaboost
// Imports a mangaboost.com series (optionally only selected chapters).
func AdminImportMangaBoost(c *gin.Context) {
	var req struct {
		URL              string               `json:"url" binding:"required"`
		SelectedChapters []ImportChapterInput `json:"selected_chapters"`
		SelectedSlugs    []string             `json:"selected_slugs"`
		Force            bool                 `json:"force"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	info, err := scraper.ScrapeMangaBoostSeries(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("scrape failed: %v", err)})
		return
	}
	if info.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not extract series title"})
		return
	}

	// Reuse existing series: first try source URL, then fall back to slug match
	// (covers the case where the series was manually created before importing)
	series, err := repository.GetSeriesBySourceURL(c.Request.Context(), req.URL)
	if err != nil && err != repository.ErrNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("finding existing series: %v", err)})
		return
	}

	if series == nil {
		baseSlug := info.Slug
		if baseSlug == "" {
			baseSlug = makeSlug(info.Title)
		}
		// Try to find an existing series with the same slug before creating a new one
		if existing, err := repository.GetSeriesBySlug(c.Request.Context(), baseSlug); err == nil && existing != nil {
			series = existing
		}
	}

	mbNewlyCreated := false
	if series == nil {
		baseSlug := info.Slug
		if baseSlug == "" {
			baseSlug = makeSlug(info.Title)
		}
		slug := baseSlug
		for i := 1; ; i++ {
			exists, _ := repository.SlugExists(c.Request.Context(), slug)
			if !exists {
				break
			}
			slug = fmt.Sprintf("%s-%d", baseSlug, i)
		}

		coverURL := info.CoverURL
		if storage.R2 != nil && storage.R2.Enabled() && coverURL != "" {
			key := storage.CoverKey(slug, coverURL)
			if uploaded, err := storage.R2.UploadFromURL(c.Request.Context(), coverURL, key, "https://mangaboost.com/"); err == nil {
				coverURL = uploaded
			}
		}

		series = &models.Series{
			ID:          uuid.NewString(),
			Slug:        slug,
			Title:       info.Title,
			CoverURL:    coverURL,
			Description: info.Description,
			Status:      info.Status,
			Author:      info.Author,
			Genres:      strings.Join(info.Genres, ", "),
			Language:    "my",
			SourceURL:   req.URL,
		}
		if err := repository.PutSeries(c.Request.Context(), series); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("saving series: %v", err)})
			return
		}
		mbNewlyCreated = true
	} else {
		mbMetaFields := map[string]interface{}{
			"title":       info.Title,
			"description": info.Description,
			"status":      info.Status,
			"author":      info.Author,
			"genres":      strings.Join(info.Genres, ", "),
		}
		if series.CoverURL == "" && info.CoverURL != "" {
			mbMetaFields["cover_url"] = info.CoverURL
		}
		_ = repository.UpdateSeriesFields(c.Request.Context(), series.ID, mbMetaFields)
	}

	// Build chapter list: prefer selected_chapters from preview (avoids slug mismatch),
	// fall back to filtering the freshly scraped list by selected_slugs.
	type mbEntry struct {
		Slug  string
		Title string
		Num   float64
		URL   string
	}
	var mbToImport []mbEntry
	if len(req.SelectedChapters) > 0 {
		for _, c := range req.SelectedChapters {
			mbToImport = append(mbToImport, mbEntry{c.Slug, c.Title, c.Number, c.URL})
		}
	} else {
		selectedSlugs := make(map[string]bool, len(req.SelectedSlugs))
		for _, s := range req.SelectedSlugs {
			selectedSlugs[s] = true
		}
		for _, ch := range info.Chapters {
			chSlug := ch.Slug
			if chSlug == "" {
				chSlug = fmt.Sprintf("%s-chapter-%.0f", series.Slug, ch.Number)
			}
			if len(selectedSlugs) > 0 && !selectedSlugs[chSlug] {
				continue
			}
			mbToImport = append(mbToImport, mbEntry{chSlug, ch.Title, ch.Number, ch.URL})
		}
	}

	if len(mbToImport) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": true, "series": nil, "chapters_saved": 0, "chapters_skipped": 0, "chapters_failed": []float64{}})
		return
	}

	jobID := uuid.NewString()
	job := &ImportJob{
		Running:  true,
		Total:    len(mbToImport),
		Failed:   []float64{},
		SeriesID: series.ID,
		Title:    series.Title,
	}
	importJobsMu.Lock()
	importJobs[jobID] = job
	importJobsMu.Unlock()

	bgCtx := context.Background()
	seriesSlug := series.Slug
	seriesChapterCount := series.ChapterCount
	force := req.Force
	isMbNew := mbNewlyCreated
	go func() {
		defer func() {
			importJobsMu.Lock()
			job.Running = false
			importJobsMu.Unlock()
		}()
		for _, ch := range mbToImport {
			chSlug := ch.Slug
			if chSlug == "" {
				chSlug = fmt.Sprintf("%s-chapter-%.0f", seriesSlug, ch.Num)
			}
			existingID, _ := repository.GetChapterIDBySlug(bgCtx, chSlug)
			if existingID != "" {
				if !force {
					importJobsMu.Lock()
					job.Skipped++
					job.Done++
					importJobsMu.Unlock()
					continue
				}
				_ = repository.DeleteChapter(bgCtx, existingID)
			}
			chapter := &models.Chapter{
				ID:        uuid.NewString(),
				SeriesID:  series.ID,
				Slug:      chSlug,
				Title:     ch.Title,
				Number:    ch.Num,
				Language:  "my",
				SourceURL: ch.URL,
			}
			if ch.URL != "" {
				imgs, imgErr := scraper.ScrapeMangaBoostChapterImages(ch.URL)
				if imgErr == nil && len(imgs) > 0 {
					chapter.Images = imgs
				}
				time.Sleep(500 * time.Millisecond)
			}
			importJobsMu.Lock()
			if err := repository.PutChapter(bgCtx, chapter); err != nil {
				job.Failed = append(job.Failed, ch.Num)
			} else {
				job.Saved++
			}
			job.Done++
			importJobsMu.Unlock()
		}
		importJobsMu.RLock()
		saved := job.Saved
		importJobsMu.RUnlock()
		_ = repository.UpdateSeriesFields(bgCtx, series.ID, map[string]interface{}{
			"chapter_count": seriesChapterCount + saved,
		})
		if isMbNew {
			telegram.NotifyNewSeries(telegram.SeriesInfo{
				Title:        series.Title,
				Description:  series.Description,
				CoverURL:     series.CoverURL,
				ChapterCount: seriesChapterCount + saved,
				Status:       series.Status,
				Genres:       series.Genres,
				Slug:         series.Slug,
			})
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"job_id": jobID,
		"series": series,
		"total":  len(mbToImport),
	})
}

// AdminRescrapeSeriesImages POST /api/admin/import/mangaboost/rescrape
// Re-scrapes chapter images for all chapters of a series that have a mangaboost source_url.
func AdminRescrapeSeriesImages(c *gin.Context) {
	var req struct {
		SeriesID string `json:"series_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chapters, err := repository.ListAllChapters(c.Request.Context(), req.SeriesID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated, failed := 0, 0
	for _, ch := range chapters {
		if ch.SourceURL == "" {
			continue
		}
		imgs, err := scraper.ScrapeMangaBoostChapterImages(ch.SourceURL)
		if err != nil || len(imgs) == 0 {
			failed++
			continue
		}
		if err := repository.UpdateChapterImages(c.Request.Context(), ch.ID, imgs, ch.SourceURL); err != nil {
			failed++
			continue
		}
		updated++
		time.Sleep(300 * time.Millisecond)
	}

	c.JSON(http.StatusOK, gin.H{"updated": updated, "failed": failed})
}

// AdminRescrapeSeries POST /api/admin/import/rescrape
// Re-scrapes chapter images for all chapters of a series using the general scraper (works for all sources).
func AdminRescrapeSeries(c *gin.Context) {
	var req struct {
		SeriesID string `json:"series_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	series, err := repository.GetSeriesByID(c.Request.Context(), req.SeriesID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "series not found"})
		return
	}

	chapters, err := repository.ListAllChapters(c.Request.Context(), req.SeriesID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated, failed := 0, 0
	for _, ch := range chapters {
		if ch.SourceURL == "" {
			continue
		}
		imgs := scrapeAndOptionallyProxy(c.Request.Context(), ch.SourceURL, series.Slug, ch.Slug, true)
		if len(imgs) == 0 {
			failed++
			continue
		}
		if err := repository.UpdateChapterImages(c.Request.Context(), ch.ID, imgs, ch.SourceURL); err != nil {
			failed++
			continue
		}
		updated++
		time.Sleep(300 * time.Millisecond)
	}

	c.JSON(http.StatusOK, gin.H{"updated": updated, "failed": failed})
}

// AdminUploadSeriesCover POST /api/admin/upload/series-cover
func AdminUploadSeriesCover(c *gin.Context) {
	seriesID := c.PostForm("series_id")
	if seriesID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "series_id required"})
		return
	}

	file, err := c.FormFile("cover")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cover file required"})
		return
	}

	// Get series to get slug
	s, err := repository.GetSeriesByID(c.Request.Context(), seriesID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("get series: %v", err)})
		return
	}

	// Open file
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "open file"})
		return
	}
	defer f.Close()

	// Upload to R2
	key := storage.CoverKey(s.Slug, file.Filename)
	url, err := storage.R2.UploadReader(c.Request.Context(), key, file.Header.Get("Content-Type"), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("upload to R2: %v", err)})
		return
	}

	// Update series cover_url
	if err := repository.UpdateSeriesCoverURL(c.Request.Context(), seriesID, url); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("update series: %v", err)})
		return
	}

	telegram.NotifyNewSeries(telegram.SeriesInfo{
		Title:       s.Title,
		Description: s.Description,
		CoverURL:    url,
		Status:      s.Status,
		Genres:      s.Genres,
		Slug:        s.Slug,
	})

	c.JSON(http.StatusOK, gin.H{"success": true, "cover_url": url})
}

// AdminUploadChapterImages POST /api/admin/upload/chapter-images
func AdminUploadChapterImages(c *gin.Context) {
	chapterID := c.PostForm("chapter_id")
	if chapterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chapter_id required"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "multipart form required"})
		return
	}
	files := form.File["images"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "images files required"})
		return
	}

	// Get chapter and series
	ch, err := repository.GetChapterByID(c.Request.Context(), chapterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("get chapter: %v", err)})
		return
	}
	s, err := repository.GetSeriesByID(c.Request.Context(), ch.SeriesID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("get series: %v", err)})
		return
	}

	var urls []string
	for i, file := range files {
		f, err := file.Open()
		if err != nil {
			continue
		}
		key := storage.ImageKey(s.Slug, ch.Slug, i, file.Filename)
		url, err := storage.R2.UploadReader(c.Request.Context(), key, file.Header.Get("Content-Type"), f)
		f.Close()
		if err != nil {
			continue
		}
		urls = append(urls, url)
	}

	// Update chapter images
	if err := repository.UpdateChapterImages(c.Request.Context(), chapterID, urls, ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("update chapter: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "images_count": len(urls), "images": urls})
}

// AdminPreviewManhwaMyanmar POST /api/admin/import/manhwamyanmar/preview
func AdminPreviewManhwaMyanmar(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	info, err := scraper.ScrapeManhwaMyamarSeries(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("scrape failed: %v", err)})
		return
	}
	if info.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not extract series title from that URL"})
		return
	}

	type chapterPreview struct {
		Title  string  `json:"title"`
		Number float64 `json:"number"`
		Slug   string  `json:"slug"`
		URL    string  `json:"url"`
	}
	chapters := make([]chapterPreview, 0, len(info.Chapters))
	for _, ch := range info.Chapters {
		slug := ch.Slug
		if slug == "" {
			slug = fmt.Sprintf("%s-chapter-%.0f", makeSlug(info.Title), ch.Number)
		}
		chapters = append(chapters, chapterPreview{
			Title:  ch.Title,
			Number: ch.Number,
			Slug:   slug,
			URL:    ch.URL,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"title":       info.Title,
		"cover_url":   info.CoverURL,
		"description": info.Description,
		"status":      info.Status,
		"author":      info.Author,
		"genres":      strings.Join(info.Genres, ", "),
		"chapters":    chapters,
	})
}

// AdminImportManhwaMyanmar POST /api/admin/import/manhwamyanmar
func AdminImportManhwaMyanmar(c *gin.Context) {
	var req struct {
		URL              string               `json:"url" binding:"required"`
		SelectedChapters []ImportChapterInput `json:"selected_chapters"`
		SelectedSlugs    []string             `json:"selected_slugs"`
		Force            bool                 `json:"force"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	info, err := scraper.ScrapeManhwaMyamarSeries(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("scrape failed: %v", err)})
		return
	}
	if info.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not extract series title"})
		return
	}

	series, err := repository.GetSeriesBySourceURL(c.Request.Context(), req.URL)
	if err != nil && err != repository.ErrNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("finding existing series: %v", err)})
		return
	}

	if series == nil {
		baseSlug := info.Slug
		if baseSlug == "" {
			baseSlug = makeSlug(info.Title)
		}
		if existing, err := repository.GetSeriesBySlug(c.Request.Context(), baseSlug); err == nil && existing != nil {
			series = existing
		}
	}

	mmNewlyCreated := false
	if series == nil {
		baseSlug := info.Slug
		if baseSlug == "" {
			baseSlug = makeSlug(info.Title)
		}
		slug := baseSlug
		for i := 1; ; i++ {
			exists, _ := repository.SlugExists(c.Request.Context(), slug)
			if !exists {
				break
			}
			slug = fmt.Sprintf("%s-%d", baseSlug, i)
		}

		coverURL := info.CoverURL
		if storage.R2 != nil && storage.R2.Enabled() && coverURL != "" {
			key := storage.CoverKey(slug, coverURL)
			if uploaded, err := storage.R2.UploadFromURL(c.Request.Context(), coverURL, key, "https://adult.manhwamyanmar.com/"); err == nil {
				coverURL = uploaded
			}
		}

		series = &models.Series{
			ID:          uuid.NewString(),
			Slug:        slug,
			Title:       info.Title,
			CoverURL:    coverURL,
			Description: info.Description,
			Status:      info.Status,
			Author:      info.Author,
			Genres:      strings.Join(info.Genres, ", "),
			Language:    "my",
			SourceURL:   req.URL,
		}
		if err := repository.PutSeries(c.Request.Context(), series); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("saving series: %v", err)})
			return
		}
		mmNewlyCreated = true
	} else {
		metaFields := map[string]interface{}{
			"title":       info.Title,
			"description": info.Description,
			"status":      info.Status,
			"author":      info.Author,
			"genres":      strings.Join(info.Genres, ", "),
		}
		if series.CoverURL == "" && info.CoverURL != "" {
			metaFields["cover_url"] = info.CoverURL
		}
		_ = repository.UpdateSeriesFields(c.Request.Context(), series.ID, metaFields)
	}

	type mmEntry struct {
		Slug  string
		Title string
		Num   float64
		URL   string
	}
	var mmToImport []mmEntry
	if len(req.SelectedChapters) > 0 {
		for _, ch := range req.SelectedChapters {
			mmToImport = append(mmToImport, mmEntry{ch.Slug, ch.Title, ch.Number, ch.URL})
		}
	} else {
		selectedSlugs := make(map[string]bool, len(req.SelectedSlugs))
		for _, s := range req.SelectedSlugs {
			selectedSlugs[s] = true
		}
		for _, ch := range info.Chapters {
			chSlug := ch.Slug
			if chSlug == "" {
				chSlug = fmt.Sprintf("%s-chapter-%.0f", series.Slug, ch.Number)
			}
			if len(selectedSlugs) > 0 && !selectedSlugs[chSlug] {
				continue
			}
			mmToImport = append(mmToImport, mmEntry{chSlug, ch.Title, ch.Number, ch.URL})
		}
	}

	if len(mmToImport) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": true, "series": nil, "chapters_saved": 0, "chapters_skipped": 0, "chapters_failed": []float64{}})
		return
	}

	jobID := uuid.NewString()
	job := &ImportJob{
		Running:  true,
		Total:    len(mmToImport),
		Failed:   []float64{},
		SeriesID: series.ID,
		Title:    series.Title,
	}
	importJobsMu.Lock()
	importJobs[jobID] = job
	importJobsMu.Unlock()

	bgCtx := context.Background()
	seriesSlug := series.Slug
	seriesChapterCount := series.ChapterCount
	force := req.Force
	isMmNew := mmNewlyCreated
	go func() {
		defer func() {
			importJobsMu.Lock()
			job.Running = false
			importJobsMu.Unlock()
		}()
		for _, ch := range mmToImport {
			chSlug := ch.Slug
			if chSlug == "" {
				chSlug = fmt.Sprintf("%s-chapter-%.0f", seriesSlug, ch.Num)
			}
			existingID, _ := repository.GetChapterIDBySlug(bgCtx, chSlug)
			if existingID != "" {
				if !force {
					importJobsMu.Lock()
					job.Skipped++
					job.Done++
					importJobsMu.Unlock()
					continue
				}
				_ = repository.DeleteChapter(bgCtx, existingID)
			}
			chapter := &models.Chapter{
				ID:        uuid.NewString(),
				SeriesID:  series.ID,
				Slug:      chSlug,
				Title:     ch.Title,
				Number:    ch.Num,
				Language:  "my",
				SourceURL: ch.URL,
			}
			if ch.URL != "" {
				imgs, imgErr := scraper.ScrapeManhwaMyamarChapterImages(ch.URL)
				if imgErr == nil && len(imgs) > 0 {
					if storage.R2 != nil && storage.R2.Enabled() {
						uploaded := make([]string, 0, len(imgs))
						for i, imgURL := range imgs {
							key := storage.ImageKey(seriesSlug, chSlug, i, imgURL)
							if u, err := storage.R2.UploadFromURL(bgCtx, imgURL, key, "https://adult.manhwamyanmar.com/"); err == nil {
								uploaded = append(uploaded, u)
							} else {
								uploaded = append(uploaded, imgURL)
							}
						}
						chapter.Images = uploaded
					} else {
						chapter.Images = imgs
					}
				}
				time.Sleep(500 * time.Millisecond)
			}
			importJobsMu.Lock()
			if err := repository.PutChapter(bgCtx, chapter); err != nil {
				job.Failed = append(job.Failed, ch.Num)
			} else {
				job.Saved++
			}
			job.Done++
			importJobsMu.Unlock()
		}
		importJobsMu.RLock()
		saved := job.Saved
		importJobsMu.RUnlock()
		_ = repository.UpdateSeriesFields(bgCtx, series.ID, map[string]interface{}{
			"chapter_count": seriesChapterCount + saved,
		})
		if isMmNew {
			telegram.NotifyNewSeries(telegram.SeriesInfo{
				Title:        series.Title,
				Description:  series.Description,
				CoverURL:     series.CoverURL,
				ChapterCount: seriesChapterCount + saved,
				Status:       series.Status,
				Genres:       series.Genres,
				Slug:         series.Slug,
			})
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"job_id": jobID,
		"series": series,
		"total":  len(mmToImport),
	})
}

// AdminPreviewYotepya POST /api/admin/import/yotepya/preview
func AdminPreviewYotepya(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	info, err := scraper.ScrapeYotepyaSeries(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("scrape failed: %v", err)})
		return
	}
	if info.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not extract series title from that URL"})
		return
	}

	type chapterPreview struct {
		Title  string  `json:"title"`
		Number float64 `json:"number"`
		Slug   string  `json:"slug"`
		URL    string  `json:"url"`
	}
	chapters := make([]chapterPreview, 0, len(info.Chapters))
	for _, ch := range info.Chapters {
		slug := ch.Slug
		if slug == "" {
			slug = fmt.Sprintf("%s-chapter-%.0f", makeSlug(info.Title), ch.Number)
		}
		chapters = append(chapters, chapterPreview{
			Title:  ch.Title,
			Number: ch.Number,
			Slug:   slug,
			URL:    ch.URL,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"title":       info.Title,
		"cover_url":   info.CoverURL,
		"description": info.Description,
		"status":      info.Status,
		"author":      info.Author,
		"genres":      strings.Join(info.Genres, ", "),
		"chapters":    chapters,
	})
}

// AdminImportYotepya POST /api/admin/import/yotepya
func AdminImportYotepya(c *gin.Context) {
	var req struct {
		URL              string               `json:"url" binding:"required"`
		SelectedChapters []ImportChapterInput `json:"selected_chapters"`
		SelectedSlugs    []string             `json:"selected_slugs"`
		Force            bool                 `json:"force"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	info, err := scraper.ScrapeYotepyaSeries(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("scrape failed: %v", err)})
		return
	}
	if info.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not extract series title"})
		return
	}

	series, err := repository.GetSeriesBySourceURL(c.Request.Context(), req.URL)
	if err != nil && err != repository.ErrNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("finding existing series: %v", err)})
		return
	}

	if series == nil {
		baseSlug := info.Slug
		if baseSlug == "" {
			baseSlug = makeSlug(info.Title)
		}
		if existing, err := repository.GetSeriesBySlug(c.Request.Context(), baseSlug); err == nil && existing != nil {
			series = existing
		}
	}

	ytNewlyCreated := false
	if series == nil {
		baseSlug := info.Slug
		if baseSlug == "" {
			baseSlug = makeSlug(info.Title)
		}
		slug := baseSlug
		for i := 1; ; i++ {
			exists, _ := repository.SlugExists(c.Request.Context(), slug)
			if !exists {
				break
			}
			slug = fmt.Sprintf("%s-%d", baseSlug, i)
		}

		coverURL := info.CoverURL
		if storage.R2 != nil && storage.R2.Enabled() && coverURL != "" {
			key := storage.CoverKey(slug, coverURL)
			if uploaded, err := storage.R2.UploadFromURL(c.Request.Context(), coverURL, key, "https://yotepya.com/"); err == nil {
				coverURL = uploaded
			}
		}

		series = &models.Series{
			ID:          uuid.NewString(),
			Slug:        slug,
			Title:       info.Title,
			CoverURL:    coverURL,
			Description: info.Description,
			Status:      info.Status,
			Author:      info.Author,
			Genres:      strings.Join(info.Genres, ", "),
			Language:    "my",
			SourceURL:   req.URL,
		}
		if err := repository.PutSeries(c.Request.Context(), series); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("saving series: %v", err)})
			return
		}
		ytNewlyCreated = true
	} else {
		metaFields := map[string]interface{}{
			"title":       info.Title,
			"description": info.Description,
			"status":      info.Status,
			"author":      info.Author,
			"genres":      strings.Join(info.Genres, ", "),
		}
		if series.CoverURL == "" && info.CoverURL != "" {
			metaFields["cover_url"] = info.CoverURL
		}
		_ = repository.UpdateSeriesFields(c.Request.Context(), series.ID, metaFields)
	}

	type ytEntry struct {
		Slug  string
		Title string
		Num   float64
		URL   string
	}
	var ytToImport []ytEntry
	if len(req.SelectedChapters) > 0 {
		for _, ch := range req.SelectedChapters {
			ytToImport = append(ytToImport, ytEntry{ch.Slug, ch.Title, ch.Number, ch.URL})
		}
	} else {
		selectedSlugs := make(map[string]bool, len(req.SelectedSlugs))
		for _, s := range req.SelectedSlugs {
			selectedSlugs[s] = true
		}
		for _, ch := range info.Chapters {
			chSlug := ch.Slug
			if chSlug == "" {
				chSlug = fmt.Sprintf("%s-chapter-%.0f", series.Slug, ch.Number)
			}
			if len(selectedSlugs) > 0 && !selectedSlugs[chSlug] {
				continue
			}
			ytToImport = append(ytToImport, ytEntry{chSlug, ch.Title, ch.Number, ch.URL})
		}
	}

	if len(ytToImport) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": true, "series": nil, "chapters_saved": 0, "chapters_skipped": 0, "chapters_failed": []float64{}})
		return
	}

	jobID := uuid.NewString()
	job := &ImportJob{
		Running:  true,
		Total:    len(ytToImport),
		Failed:   []float64{},
		SeriesID: series.ID,
		Title:    series.Title,
	}
	importJobsMu.Lock()
	importJobs[jobID] = job
	importJobsMu.Unlock()

	bgCtx := context.Background()
	seriesSlug := series.Slug
	seriesChapterCount := series.ChapterCount
	force := req.Force
	isYtNew := ytNewlyCreated
	go func() {
		defer func() {
			importJobsMu.Lock()
			job.Running = false
			importJobsMu.Unlock()
		}()
		for _, ch := range ytToImport {
			chSlug := ch.Slug
			if chSlug == "" {
				chSlug = fmt.Sprintf("%s-chapter-%.0f", seriesSlug, ch.Num)
			}
			existingID, _ := repository.GetChapterIDBySlug(bgCtx, chSlug)
			if existingID != "" {
				if !force {
					importJobsMu.Lock()
					job.Skipped++
					job.Done++
					importJobsMu.Unlock()
					continue
				}
				_ = repository.DeleteChapter(bgCtx, existingID)
			}
			chapter := &models.Chapter{
				ID:        uuid.NewString(),
				SeriesID:  series.ID,
				Slug:      chSlug,
				Title:     ch.Title,
				Number:    ch.Num,
				Language:  "my",
				SourceURL: ch.URL,
			}
			if ch.URL != "" {
				imgs, imgErr := scraper.ScrapeYotepyaChapterImages(ch.URL)
				if imgErr == nil && len(imgs) > 0 {
					if storage.R2 != nil && storage.R2.Enabled() {
						uploaded := make([]string, 0, len(imgs))
						for i, imgURL := range imgs {
							key := storage.ImageKey(seriesSlug, chSlug, i, imgURL)
							if u, err := storage.R2.UploadFromURL(bgCtx, imgURL, key, "https://yotepya.com/"); err == nil {
								uploaded = append(uploaded, u)
							} else {
								uploaded = append(uploaded, imgURL)
							}
						}
						chapter.Images = uploaded
					} else {
						chapter.Images = imgs
					}
				}
				time.Sleep(500 * time.Millisecond)
			}
			importJobsMu.Lock()
			if err := repository.PutChapter(bgCtx, chapter); err != nil {
				job.Failed = append(job.Failed, ch.Num)
			} else {
				job.Saved++
			}
			job.Done++
			importJobsMu.Unlock()
		}
		importJobsMu.RLock()
		saved := job.Saved
		importJobsMu.RUnlock()
		_ = repository.UpdateSeriesFields(bgCtx, series.ID, map[string]interface{}{
			"chapter_count": seriesChapterCount + saved,
		})
		if isYtNew {
			telegram.NotifyNewSeries(telegram.SeriesInfo{
				Title:        series.Title,
				Description:  series.Description,
				CoverURL:     series.CoverURL,
				ChapterCount: seriesChapterCount + saved,
				Status:       series.Status,
				Genres:       series.Genres,
				Slug:         series.Slug,
			})
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"job_id": jobID,
		"series": series,
		"total":  len(ytToImport),
	})
}
