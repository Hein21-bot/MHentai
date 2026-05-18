package handlers

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"mhentai-backend/internal/models"
	"mhentai-backend/internal/repository"
	"mhentai-backend/internal/scraper"
	"mhentai-backend/internal/storage"
)

var slugRe = regexp.MustCompile(`[^a-z0-9-]+`)

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

// AdminListSeries GET /api/admin/series
func AdminListSeries(c *gin.Context) {
	result, err := repository.ListSeriesPage(c.Request.Context(), repository.ListSeriesParams{
		SortBy: "updated_at",
		Limit:  200,
	}, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result.Items})
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

// AdminListChapters GET /api/admin/chapters?series_id=X
func AdminListChapters(c *gin.Context) {
	chapters, err := repository.ListAllChapters(c.Request.Context(), c.Query("series_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Attach image count
	type chapterRow struct {
		models.Chapter
		ImageCount int `json:"image_count"`
	}
	rows := make([]chapterRow, 0, len(chapters))
	for _, ch := range chapters {
		rows = append(rows, chapterRow{Chapter: *ch, ImageCount: len(ch.Images)})
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
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
	URL           string   `json:"url" binding:"required"`
	SelectedSlugs []string `json:"selected_slugs"` // if set, only import these chapters
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

	if series == nil {
		// Unique slug
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

		// Always upload cover to R2
		coverURL := info.CoverURL
		if storage.R2 != nil && storage.R2.Enabled() && coverURL != "" {
			key := storage.CoverKey(slug, coverURL)
			if uploaded, err := storage.R2.UploadFromURL(c.Request.Context(), coverURL, key); err == nil {
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
	} else {
		// Existing series found; update metadata if changed.
		_ = repository.UpdateSeriesFields(c.Request.Context(), series.ID, map[string]interface{}{
			"title":       info.Title,
			"cover_url":   info.CoverURL,
			"description": info.Description,
			"status":      info.Status,
			"author":      info.Author,
			"genres":      strings.Join(info.Genres, ", "),
		})
	}

	// Build selected slugs set (empty = import all)
	selectedSlugs := make(map[string]bool, len(req.SelectedSlugs))
	for _, s := range req.SelectedSlugs {
		selectedSlugs[s] = true
	}

	savedChapters, skippedChapters := 0, 0
	seriesSlug := series.Slug
	for _, ch := range info.Chapters {
		chSlug := ch.Slug
		if chSlug == "" {
			chSlug = fmt.Sprintf("%s-chapter-%.0f", seriesSlug, ch.Number)
		}
		// Skip if not in selection (when selection is provided)
		if len(selectedSlugs) > 0 && !selectedSlugs[chSlug] {
			skippedChapters++
			continue
		}
		// If slug exists but under a different series, move it to this series
		if existing, _ := repository.GetChapterBySlug(c.Request.Context(), chSlug); existing != nil {
			if existing.SeriesID == series.ID {
				skippedChapters++
				continue
			}
			// Orphaned chapter — reassign to current series
			_ = repository.UpdateSeriesFields(c.Request.Context(), existing.ID, map[string]interface{}{
				"series_id": series.ID,
				"language":  "en",
			})
			savedChapters++
			continue
		}

		chapter := &models.Chapter{
			ID:        uuid.NewString(),
			SeriesID:  series.ID,
			Slug:      chSlug,
			Title:     ch.Title,
			Number:    ch.Number,
			Language:  "en",
			SourceURL: ch.URL,
		}

		// Always scrape images and upload to R2
		if ch.URL != "" {
			chapter.Images = scrapeAndOptionallyProxy(c.Request.Context(), ch.URL, seriesSlug, chSlug, true)
			time.Sleep(500 * time.Millisecond) // be polite between chapter fetches
		}

		if err := repository.PutChapter(c.Request.Context(), chapter); err != nil {
			continue
		}
		savedChapters++
	}

	// Update chapter count relative to the existing series count.
	_ = repository.UpdateSeriesFields(c.Request.Context(), series.ID, map[string]interface{}{
		"chapter_count": series.ChapterCount + savedChapters,
	})
	series.ChapterCount += savedChapters

	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"series":           series,
		"chapters_saved":   savedChapters,
		"chapters_skipped": skippedChapters,
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
		if uploaded, err := storage.R2.UploadFromURL(ctx, imgURL, key); err == nil {
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
		URL           string   `json:"url" binding:"required"`
		SelectedSlugs []string `json:"selected_slugs"`
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
			if uploaded, err := storage.R2.UploadFromURL(c.Request.Context(), coverURL, key); err == nil {
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
	} else {
		_ = repository.UpdateSeriesFields(c.Request.Context(), series.ID, map[string]interface{}{
			"title":       info.Title,
			"cover_url":   info.CoverURL,
			"description": info.Description,
			"status":      info.Status,
			"author":      info.Author,
			"genres":      strings.Join(info.Genres, ", "),
		})
	}

	selectedSlugs := make(map[string]bool, len(req.SelectedSlugs))
	for _, s := range req.SelectedSlugs {
		selectedSlugs[s] = true
	}

	savedChapters, skippedChapters := 0, 0
	seriesSlug := series.Slug
	for _, ch := range info.Chapters {
		chSlug := ch.Slug
		if chSlug == "" {
			chSlug = fmt.Sprintf("%s-chapter-%.0f", seriesSlug, ch.Number)
		}
		if len(selectedSlugs) > 0 && !selectedSlugs[chSlug] {
			skippedChapters++
			continue
		}
		// If slug exists but under a different series, move it to this series
		if existing, _ := repository.GetChapterBySlug(c.Request.Context(), chSlug); existing != nil {
			if existing.SeriesID == series.ID {
				skippedChapters++
				continue
			}
			// Orphaned chapter — reassign to current series
			_ = repository.UpdateSeriesFields(c.Request.Context(), existing.ID, map[string]interface{}{
				"series_id": series.ID,
				"language":  "my",
			})
			savedChapters++
			continue
		}

		chapter := &models.Chapter{
			ID:        uuid.NewString(),
			SeriesID:  series.ID,
			Slug:      chSlug,
			Title:     ch.Title,
			Number:    ch.Number,
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

		if err := repository.PutChapter(c.Request.Context(), chapter); err != nil {
			continue
		}
		savedChapters++
	}

	_ = repository.UpdateSeriesFields(c.Request.Context(), series.ID, map[string]interface{}{
		"chapter_count": series.ChapterCount + savedChapters,
	})
	series.ChapterCount += savedChapters

	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"series":           series,
		"chapters_saved":   savedChapters,
		"chapters_skipped": skippedChapters,
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
