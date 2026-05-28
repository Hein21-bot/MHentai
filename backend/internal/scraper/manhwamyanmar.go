package scraper

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// chapterListRe matches the " Chapter N Chapter N ..." suffix in og:description
var chapterListRe = regexp.MustCompile(`\s+Chapter\s+\d+\s+Chapter\s+\d+.*$`)

// ScrapeManhwaMyamarSeries scrapes series info + chapter list from adult.manhwamyanmar.com.
func ScrapeManhwaMyamarSeries(seriesURL string) (*SeriesInfo, error) {
	doc, err := fetchManhwaMyamarDoc(seriesURL)
	if err != nil {
		return nil, fmt.Errorf("fetching series page: %w", err)
	}

	info := &SeriesInfo{SourceURL: seriesURL}
	info.Slug = extractSlugFromURL(seriesURL)

	// Title from og:title, strip " - Manhwa Myanmar" suffix
	doc.Find(`meta[property="og:title"]`).Each(func(_ int, s *goquery.Selection) {
		if v, ok := s.Attr("content"); ok && info.Title == "" {
			t := strings.TrimSpace(v)
			t = strings.TrimSuffix(t, " - Manhwa Myanmar")
			info.Title = strings.TrimSpace(t)
		}
	})

	// Cover from og:image
	doc.Find(`meta[property="og:image"]`).Each(func(_ int, s *goquery.Selection) {
		if v, ok := s.Attr("content"); ok && info.CoverURL == "" {
			info.CoverURL = strings.TrimSpace(v)
		}
	})

	// Parse og:description: "Genre1, Genre2 Status: Ongoing [real description] Chapter 1 Chapter 2..."
	var ogDesc string
	doc.Find(`meta[property="og:description"]`).Each(func(_ int, s *goquery.Selection) {
		if v, ok := s.Attr("content"); ok {
			ogDesc = strings.TrimSpace(v)
		}
	})

	if ogDesc != "" {
		statusIdx := strings.Index(ogDesc, "Status:")
		if statusIdx > 0 {
			// Genres: everything before "Status:"
			for _, g := range strings.Split(ogDesc[:statusIdx], ",") {
				g = strings.TrimSpace(g)
				if g != "" {
					info.Genres = append(info.Genres, g)
				}
			}
			// Status + description: rest after "Status:"
			rest := strings.TrimSpace(ogDesc[statusIdx+7:])
			lv := strings.ToLower(rest)
			if strings.HasPrefix(lv, "completed") {
				info.Status = "completed"
				rest = strings.TrimSpace(rest[len("completed"):])
			} else if strings.HasPrefix(lv, "ongoing") {
				info.Status = "ongoing"
				rest = strings.TrimSpace(rest[len("ongoing"):])
			} else {
				info.Status = "ongoing"
			}
			// Strip trailing chapter list
			rest = chapterListRe.ReplaceAllString(rest, "")
			info.Description = strings.TrimSpace(rest)
		} else {
			info.Description = ogDesc
			info.Status = "ongoing"
		}
	}

	// Chapter links — <a href="https://18.manhwamyanmar.com/...">Chapter N</a>
	seen := make(map[string]bool)
	doc.Find(".entry-content a").Each(func(_ int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok || !strings.Contains(href, "18.manhwamyanmar.com") || seen[href] {
			return
		}
		seen[href] = true
		title := strings.TrimSpace(s.Text())
		if title == "" {
			title = extractSlugFromURL(href)
		}
		num := parseChapterNumber(title)
		chSlug := extractSlugFromURL(href)
		info.Chapters = append(info.Chapters, ChapterInfo{
			Title:  title,
			Slug:   chSlug,
			Number: num,
			URL:    href,
		})
	})

	return info, nil
}

// ScrapeManhwaMyamarChapterImages extracts image URLs from an 18.manhwamyanmar.com chapter page.
func ScrapeManhwaMyamarChapterImages(chapterURL string) ([]string, error) {
	doc, err := fetchManhwaMyamarDoc(chapterURL)
	if err != nil {
		return nil, fmt.Errorf("fetching chapter page: %w", err)
	}

	seen := make(map[string]bool)
	var imgs []string
	doc.Find("img").Each(func(_ int, s *goquery.Selection) {
		// Real URLs are in data-src (lazy load); noscript fallbacks are unparsed by goquery
		src, _ := s.Attr("data-src")
		if src == "" {
			src, _ = s.Attr("src")
		}
		if src == "" || strings.HasPrefix(src, "data:") || strings.HasSuffix(src, ".gif") {
			return
		}
		if !strings.Contains(src, "manhwamm.cloud") {
			return
		}
		if !seen[src] {
			seen[src] = true
			imgs = append(imgs, src)
		}
	})

	if len(imgs) == 0 {
		return nil, fmt.Errorf("no images found on chapter page")
	}
	return imgs, nil
}

func fetchManhwaMyamarDoc(rawURL string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://adult.manhwamyanmar.com/")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 || resp.StatusCode == 503 {
		return nil, fmt.Errorf("blocked by site (status %d)", resp.StatusCode)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}
	return goquery.NewDocumentFromReader(resp.Body)
}
