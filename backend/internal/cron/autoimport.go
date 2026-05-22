package cron

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"mhentai-backend/internal/models"
	"mhentai-backend/internal/repository"
	"mhentai-backend/internal/scraper"
	"mhentai-backend/internal/storage"
)

const autoImportInterval = 24 * time.Hour
const autoImportDelay = 2 * time.Minute // wait after startup before first run

// StartAutoImport runs a background goroutine that imports new chapters every 24 hours
// for all series whose source_url is from adult.manhwamyanmar.com.
func StartAutoImport() {
	go func() {
		log.Printf("[cron] Auto-import scheduled. First run in %v.", autoImportDelay)
		time.Sleep(autoImportDelay)
		for {
			log.Println("[cron] Auto-import started.")
			total := runAutoImport()
			log.Printf("[cron] Auto-import done. %d new chapters added. Next run in %v.", total, autoImportInterval)
			time.Sleep(autoImportInterval)
		}
	}()
}

func runAutoImport() int {
	ctx := context.Background()

	all, err := repository.GetAllSeries(ctx, "")
	if err != nil {
		log.Printf("[cron] Failed to fetch series list: %v", err)
		return 0
	}

	totalAdded := 0
	for _, s := range all {
		if s.SourceURL == "" || !strings.Contains(s.SourceURL, "manhwamyanmar.com") {
			continue
		}
		if s.Status != "ongoing" {
			continue
		}
		added, err := importNewChapters(ctx, s)
		if err != nil {
			log.Printf("[cron] Error on series %q: %v", s.Title, err)
			continue
		}
		if added > 0 {
			log.Printf("[cron] +%d chapters for %q", added, s.Title)
			totalAdded += added
		}
		time.Sleep(2 * time.Second) // be polite to source site
	}
	return totalAdded
}

func importNewChapters(ctx context.Context, series *models.Series) (int, error) {
	info, err := scraper.ScrapeSeries(series.SourceURL)
	if err != nil {
		return 0, fmt.Errorf("scrape: %w", err)
	}

	saved := 0
	for _, ch := range info.Chapters {
		chSlug := ch.Slug
		if chSlug == "" {
			chSlug = fmt.Sprintf("%s-chapter-%.0f", series.Slug, ch.Number)
		}

		// Skip chapters that already exist
		existingID, _ := repository.GetChapterIDBySlug(ctx, chSlug)
		if existingID != "" {
			continue
		}

		// Scrape images for new chapter
		var images []string
		if ch.URL != "" {
			raw, err := scraper.ScrapeChapterImages(ch.URL)
			if err == nil && len(raw) > 0 {
				images = uploadToR2(ctx, raw, series.Slug, chSlug)
			}
			time.Sleep(500 * time.Millisecond)
		}

		chapter := &models.Chapter{
			ID:        uuid.NewString(),
			SeriesID:  series.ID,
			Slug:      chSlug,
			Title:     ch.Title,
			Number:    ch.Number,
			Language:  series.Language,
			SourceURL: ch.URL,
			Images:    images,
		}
		if err := repository.PutChapter(ctx, chapter); err != nil {
			log.Printf("[cron] Failed to save chapter %s: %v", chSlug, err)
			continue
		}
		saved++
	}

	if saved > 0 {
		_ = repository.UpdateSeriesFields(ctx, series.ID, map[string]interface{}{
			"chapter_count": series.ChapterCount + saved,
		})
		repository.InvalidateSeriesCache()
	}
	return saved, nil
}

func uploadToR2(ctx context.Context, images []string, seriesSlug, chapterSlug string) []string {
	if storage.R2 == nil || !storage.R2.Enabled() {
		return images
	}
	result := make([]string, 0, len(images))
	for i, imgURL := range images {
		key := storage.ImageKey(seriesSlug, chapterSlug, i, imgURL)
		if uploaded, err := storage.R2.UploadFromURL(ctx, imgURL, key); err == nil {
			result = append(result, uploaded)
		} else {
			result = append(result, imgURL)
		}
	}
	return result
}
