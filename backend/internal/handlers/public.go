package handlers

import (
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
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
		Genre:    c.Query("genre"),
		Letter:   c.Query("letter"),
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
	chapters, err := repository.LatestChaptersBySeries(c.Request.Context(), seriesID, 3)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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
// Orders by series updated_at so bulk chapter imports don't flood the feed.
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

	seriesResult, err := repository.ListSeriesPage(c.Request.Context(), repository.ListSeriesParams{
		SortBy:   "updated_at",
		Limit:    limit,
		Language: lang,
		Status:   c.Query("status"),
	}, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Fetch latest chapters for all series concurrently instead of sequentially.
	ctx := c.Request.Context()
	slots := make([][]*models.Chapter, len(seriesResult.Items))
	var wg sync.WaitGroup
	for i := range seriesResult.Items {
		wg.Add(1)
		go func(idx int, s *models.Series) {
			defer wg.Done()
			chs, _ := repository.LatestChaptersBySeries(ctx, s.ID, 3)
			for _, ch := range chs {
				ch.Series = s
			}
			slots[idx] = chs
		}(i, seriesResult.Items[i])
	}
	wg.Wait()

	result := make([]*models.Chapter, 0, len(seriesResult.Items)*3)
	for _, chs := range slots {
		result = append(result, chs...)
	}

	c.JSON(http.StatusOK, gin.H{"data": result, "total": seriesResult.Total, "page": page})
}

// GetGenres GET /api/genres
// Returns sorted list of distinct genres from the cached series data.
func GetGenres(c *gin.Context) {
	all, err := repository.GetAllSeries(c.Request.Context(), c.Query("lang"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	seen := map[string]struct{}{}
	var genres []string
	for _, s := range all {
		for _, g := range strings.Split(s.Genres, ",") {
			t := strings.TrimSpace(g)
			if t != "" {
				if _, ok := seen[t]; !ok {
					seen[t] = struct{}{}
					genres = append(genres, t)
				}
			}
		}
	}
	sort.Strings(genres)
	c.JSON(http.StatusOK, gin.H{"data": genres})
}

// GetRecommendations GET /api/recommendations
// Returns a random sample of series matching the requested genre, from cache.
func GetRecommendations(c *gin.Context) {
	genre := c.Query("genre")
	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	if limit < 1 || limit > 20 {
		limit = 5
	}

	all, err := repository.GetAllSeries(c.Request.Context(), c.Query("lang"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var filtered []*models.Series
	for _, s := range all {
		if status != "" && s.Status != status {
			continue
		}
		if genre == "" || containsGenre(s.Genres, genre) {
			filtered = append(filtered, s)
		}
	}

	cp := make([]*models.Series, len(filtered))
	copy(cp, filtered)
	rand.Shuffle(len(cp), func(i, j int) { cp[i], cp[j] = cp[j], cp[i] })
	if len(cp) > limit {
		cp = cp[:limit]
	}
	c.JSON(http.StatusOK, gin.H{"data": cp})
}

func containsGenre(genres, target string) bool {
	for _, g := range strings.Split(genres, ",") {
		if strings.TrimSpace(g) == target {
			return true
		}
	}
	return false
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
