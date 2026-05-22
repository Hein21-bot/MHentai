package repository

import (
	"fmt"
	"sync"
	"time"

	"mhentai-backend/internal/models"
)

const cacheTTL = 2 * time.Minute

type seriesCacheEntry struct {
	items     []*models.Series
	expiresAt time.Time
}

var (
	seriesListCache = map[string]*seriesCacheEntry{}
	cacheMu         sync.RWMutex
)

func listCacheKey(p ListSeriesParams) string {
	// SortBy is intentionally excluded: the cache stores unsorted data and callers
	// sort a copy, so all sort variants share the same underlying scan.
	return fmt.Sprintf("%s|%s|%s", p.Status, p.Search, p.Language)
}

func cacheGetList(key string) ([]*models.Series, bool) {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	e, ok := seriesListCache[key]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.items, true
}

func cacheSetList(key string, items []*models.Series) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	seriesListCache[key] = &seriesCacheEntry{
		items:     items,
		expiresAt: time.Now().Add(cacheTTL),
	}
}

// InvalidateSeriesCache clears all cached series lists.
// Call this after any write to the series table.
func InvalidateSeriesCache() {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	seriesListCache = map[string]*seriesCacheEntry{}
}
