package handlers

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"mhentai-backend/internal/models"
	"mhentai-backend/internal/repository"
)

// ListSeries GET /api/series
func ListSeries(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "24"))
	if page < 1 {
		page = 1
	}

	result, err := repository.ListSeriesPage(c.Request.Context(), repository.ListSeriesParams{
		Status:   c.Query("status"),
		SortBy:   c.DefaultQuery("sort", "updated_at"),
		Search:   c.Query("q"),
		Language: c.Query("lang"),
		Limit:    limit,
	}, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  result.Items,
		"total": result.Total,
		"page":  page,
		"limit": limit,
	})
}

// GetSeries GET /api/series/:id
func GetSeries(c *gin.Context) {
	identifier := c.Param("id")
	s, err := repository.GetSeriesBySlug(c.Request.Context(), identifier)
	if err == repository.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "series not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	chapters, _ := repository.ListChaptersBySeries(c.Request.Context(), s.ID)
	s.Chapters = nil
	for _, ch := range chapters {
		s.Chapters = append(s.Chapters, *ch)
	}

	// Increment view count async (best-effort)
	go repository.IncrementSeriesViewCount(c.Request.Context(), s.ID) //nolint

	c.JSON(http.StatusOK, s)
}

// GetSeriesLatestChapters GET /api/series/:id/latest-chapters
func GetSeriesLatestChapters(c *gin.Context) {
	seriesID := c.Param("id")
	chapters, err := repository.ListChaptersBySeries(c.Request.Context(), seriesID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// ListChaptersBySeries returns ASC — take last 3
	n := len(chapters)
	if n > 3 {
		chapters = chapters[n-3:]
	}
	// Reverse so newest first
	for i, j := 0, len(chapters)-1; i < j; i, j = i+1, j-1 {
		chapters[i], chapters[j] = chapters[j], chapters[i]
	}
	c.JSON(http.StatusOK, gin.H{"data": chapters})
}

// GetChapter GET /api/chapters/:slug
func GetChapter(c *gin.Context) {
	ch, err := repository.GetChapterBySlug(c.Request.Context(), c.Param("slug"))
	if err == repository.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "chapter not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Load parent series
	if ch.SeriesID != "" {
		series, _ := repository.GetSeriesByID(c.Request.Context(), ch.SeriesID)
		ch.Series = series
	}

	prev, next := repository.GetAdjacentChapters(c.Request.Context(), ch.SeriesID, ch.Number)

	go repository.IncrementChapterViewCount(c.Request.Context(), ch.ID) //nolint

	c.JSON(http.StatusOK, gin.H{
		"chapter":      ch,
		"prev_chapter": prev,
		"next_chapter": next,
	})
}

// GetLatestUpdates GET /api/latest
func GetLatestUpdates(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if limit < 1 || limit > 50 {
		limit = 12
	}
	if page < 1 {
		page = 1
	}
	lang := c.Query("lang")

	// Fetch enough chapters to cover several pages of unique series
	chapters, err := repository.LatestChapters(c.Request.Context(), 100, lang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Group chapters by series — no DB calls yet, just organize
	const maxPerSeries = 3
	seriesOrder := []string{}
	seriesChapters := map[string][]*models.Chapter{}
	for _, ch := range chapters {
		if ch.SeriesID == "" {
			continue
		}
		if _, seen := seriesChapters[ch.SeriesID]; !seen {
			seriesOrder = append(seriesOrder, ch.SeriesID)
			seriesChapters[ch.SeriesID] = nil
		}
		if len(seriesChapters[ch.SeriesID]) < maxPerSeries {
			seriesChapters[ch.SeriesID] = append(seriesChapters[ch.SeriesID], ch)
		}
	}

	total := len(seriesOrder)
	offset := (page - 1) * limit
	if offset >= total {
		c.JSON(http.StatusOK, gin.H{"data": []*models.Chapter{}, "total": total, "page": page})
		return
	}
	end := offset + limit
	if end > total {
		end = total
	}

	// Only fetch series info for this page's series (max `limit` GetItem calls)
	var result []*models.Chapter
	for _, sid := range seriesOrder[offset:end] {
		s, _ := repository.GetSeriesByID(c.Request.Context(), sid)
		for _, ch := range seriesChapters[sid] {
			ch.Series = s
			result = append(result, ch)
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": result, "total": total, "page": page})
}

// refererMap maps image CDN hostnames to the Referer header they require.
var refererMap = map[string]string{
	"img.myanhwa.xyz":         "https://adult.manhwamyanmar.com/",
	"img.manhwamyanmar.com":   "https://adult.manhwamyanmar.com/",
	"img.hentai20.io":         "https://hentai20.io/",
	"img.hentai1.io":          "https://hentai20.io/",
	"s1.manhwa18.net":         "https://manhwa18.net/",
}

var proxyClient = &http.Client{Timeout: 30 * time.Second}

// ProxyImage GET /api/proxy/img?url=<encoded image url>
// Fetches an external image server-side (with correct Referer) and streams it to the client.
func ProxyImage(c *gin.Context) {
	rawURL := c.Query("url")
	if rawURL == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		c.Status(http.StatusBadRequest)
		return
	}

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		c.Status(http.StatusBadGateway)
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	// Set correct Referer based on CDN hostname
	if ref, ok := refererMap[strings.ToLower(parsed.Hostname())]; ok {
		req.Header.Set("Referer", ref)
	}

	resp, err := proxyClient.Do(req)
	if err != nil {
		c.Status(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/jpeg"
	}
	c.Header("Cache-Control", "public, max-age=86400")
	c.DataFromReader(resp.StatusCode, resp.ContentLength, ct, resp.Body, nil)
}
