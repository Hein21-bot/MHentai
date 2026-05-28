package scraper

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var yotepyaIndexRe = regexp.MustCompile(`/index(\d*)\.html$`)

// ScrapeYotepyaSeries scrapes series info + chapter list from yotepya.com.
// Series URL format: https://yotepya.com/contents/{id}/
func ScrapeYotepyaSeries(seriesURL string) (*SeriesInfo, error) {
	// Normalise: ensure trailing slash
	if !strings.HasSuffix(seriesURL, "/") {
		seriesURL += "/"
	}

	doc, err := fetchYotepyaDoc(seriesURL)
	if err != nil {
		return nil, fmt.Errorf("fetching series page: %w", err)
	}

	info := &SeriesInfo{SourceURL: seriesURL}

	// Title from <title>, strip " - YotePya"
	info.Title = strings.TrimSuffix(
		strings.TrimSpace(doc.Find("title").Text()),
		" - YotePya",
	)
	info.Slug = extractSlugFromURL(seriesURL)
	if info.Slug == "" {
		// fallback: last non-empty path segment
		parts := strings.Split(strings.TrimSuffix(seriesURL, "/"), "/")
		info.Slug = parts[len(parts)-1]
	}

	// Description from <p class="card-text"> after "ဇာတ်လမ်းအကျဉ်း"
	doc.Find(".card-text").Each(func(_ int, s *goquery.Selection) {
		if info.Description == "" {
			t := strings.TrimSpace(s.Text())
			if len(t) > 10 {
				info.Description = t
			}
		}
	})

	// Cover: first data-original image on the series page (chapter thumbnail)
	doc.Find("img[data-original]").Each(func(_ int, s *goquery.Selection) {
		if info.CoverURL == "" {
			if v, ok := s.Attr("data-original"); ok && v != "" {
				info.CoverURL = strings.TrimSpace(v)
			}
		}
	})

	info.Status = "ongoing"

	// Collect chapters from all index pages
	// Pages: index.html (=main), index1.html, index2.html, ...
	chaptersSeen := make(map[int]bool)
	var allChapters []ChapterInfo

	baseURL := seriesURL // e.g. https://yotepya.com/contents/71/

	fetchChaptersFromDoc := func(d *goquery.Document) {
		d.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
			href, _ := s.Attr("href")
			if !strings.Contains(href, "./chapters/") && !strings.Contains(href, "/chapters/") {
				return
			}
			// Extract chapter number from filename e.g. ./chapters/5.html → 5
			numStr := regexp.MustCompile(`/chapters/(\d+)\.html`).FindStringSubmatch(href)
			if len(numStr) < 2 {
				return
			}
			num, err := strconv.Atoi(numStr[1])
			if err != nil || chaptersSeen[num] {
				return
			}
			chaptersSeen[num] = true

			title := strings.TrimSpace(s.Find(".btn").Text())
			if title == "" {
				title = fmt.Sprintf("Chapter %d", num)
			}

			chapterURL := baseURL + "chapters/" + numStr[1] + ".html"
			slug := fmt.Sprintf("%s-chapter-%d", info.Slug, num)

			allChapters = append(allChapters, ChapterInfo{
				Title:  title,
				Slug:   slug,
				Number: float64(num),
				URL:    chapterURL,
			})
		})
	}

	// Chapters from main page
	fetchChaptersFromDoc(doc)

	// Fetch additional index pages until 404
	for i := 1; i <= 50; i++ {
		pageURL := fmt.Sprintf("%sindex%d.html", baseURL, i)
		d, err := fetchYotepyaDoc(pageURL)
		if err != nil {
			break
		}
		before := len(allChapters)
		fetchChaptersFromDoc(d)
		if len(allChapters) == before {
			break // no new chapters found, stop
		}
	}

	// Sort ascending by chapter number (1 → 2 → 3 → ...)
	for i := 0; i < len(allChapters); i++ {
		for j := i + 1; j < len(allChapters); j++ {
			if allChapters[i].Number > allChapters[j].Number {
				allChapters[i], allChapters[j] = allChapters[j], allChapters[i]
			}
		}
	}

	info.Chapters = allChapters
	return info, nil
}

// ScrapeYotepyaChapterImages extracts image URLs from a yotepya.com chapter page.
func ScrapeYotepyaChapterImages(chapterURL string) ([]string, error) {
	doc, err := fetchYotepyaDoc(chapterURL)
	if err != nil {
		return nil, fmt.Errorf("fetching chapter page: %w", err)
	}

	seen := make(map[string]bool)
	var imgs []string

	doc.Find("img[data-original]").Each(func(_ int, s *goquery.Selection) {
		src, _ := s.Attr("data-original")
		if src == "" || strings.HasPrefix(src, "data:") {
			return
		}
		// Skip ads/icon images
		if strings.Contains(src, "ico.png") || strings.Contains(src, "favicon") {
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

func fetchYotepyaDoc(rawURL string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://yotepya.com/")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("page not found (404)")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}
	return goquery.NewDocumentFromReader(resp.Body)
}
