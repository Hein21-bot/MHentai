package scraper

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type SeriesInfo struct {
	Title       string
	Slug        string
	CoverURL    string
	Description string
	Status      string
	Author      string
	Genres      []string
	Chapters    []ChapterInfo
	SourceURL   string
}

type ChapterInfo struct {
	Title  string
	Slug   string
	Number float64
	URL    string
	Images []string
}

var httpClient = &http.Client{
	Timeout: 60 * time.Second,
}

func fetchDoc(rawURL string) (*goquery.Document, error) {
	return fetchDocWithRetry(rawURL, 3)
}

func fetchDocWithRetry(rawURL string, maxRetries int) (*goquery.Document, error) {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}
		doc, err := doFetch(rawURL)
		if err == nil {
			return doc, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func doFetch(rawURL string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://hentai20.io/")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	// NOTE: Do NOT set Accept-Encoding manually — Go's http.Client handles
	// gzip decompression automatically only when it adds the header itself.
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 || resp.StatusCode == 503 {
		return nil, fmt.Errorf("blocked by site (status %d) — Cloudflare may be active", resp.StatusCode)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}
	return goquery.NewDocumentFromReader(resp.Body)
}

func extractSlugFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	path := strings.Trim(u.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}
	// If path is like /manga/series-name/ return series-name
	if len(parts) >= 2 && parts[0] == "manga" {
		return parts[1]
	}
	// Otherwise return the last path segment
	return parts[len(parts)-1]
}

var chapterNumRe = regexp.MustCompile(`(?i)chapter[- _]+([\d]+(?:\.[\d]+)?)`)

func parseChapterNumber(title string) float64 {
	// Try "Chapter N" pattern
	m := chapterNumRe.FindStringSubmatch(title)
	if len(m) >= 2 {
		n, _ := strconv.ParseFloat(m[1], 64)
		return n
	}
	// Try trailing number
	re2 := regexp.MustCompile(`([\d]+(?:\.[\d]+)?)\s*$`)
	m2 := re2.FindStringSubmatch(strings.TrimSpace(title))
	if len(m2) >= 2 {
		n, _ := strconv.ParseFloat(m2[1], 64)
		return n
	}
	return 0
}

// ScrapeSeries scrapes series info + chapter list from a hentai20.io manga URL
func ScrapeSeries(seriesURL string) (*SeriesInfo, error) {
	doc, err := fetchDoc(seriesURL)
	if err != nil {
		return nil, fmt.Errorf("fetching series page: %w", err)
	}

	info := &SeriesInfo{
		SourceURL: seriesURL,
	}

	// Title — try multiple selectors across different themes
	titleSelectors := []string{
		"h1.entry-title", // hentai20.io
		"h1.post-title",  // manhwamyanmar.com (contains inner <a>)
		"h1.manga-title",
		".manga-info h1",
		".post-title h1",
	}
	for _, sel := range titleSelectors {
		t := strings.TrimSpace(doc.Find(sel).First().Text())
		if t != "" {
			info.Title = t
			break
		}
	}
	if info.Title == "" {
		info.Title = strings.TrimSpace(doc.Find("h1").First().Text())
	}
	if info.Title == "" {
		// Last resort: page title tag
		info.Title = strings.TrimSpace(doc.Find("title").Text())
		for _, sep := range []string{" – ", " | ", " - "} {
			if idx := strings.Index(info.Title, sep); idx != -1 {
				info.Title = info.Title[:idx]
				break
			}
		}
	}
	info.Title = strings.TrimSpace(info.Title)

	// Slug from URL
	info.Slug = extractSlugFromURL(seriesURL)

	// Cover image - hentai20.io uses .thumb img or img.wp-post-image
	coverSelectors := []string{
		".thumb img",
		".summary_image img",
		"img.wp-post-image",
		".manga-cover img",
	}
	for _, sel := range coverSelectors {
		doc.Find(sel).Each(func(i int, s *goquery.Selection) {
			if info.CoverURL != "" {
				return
			}
			for _, attr := range []string{"src", "data-src", "data-lazy-src"} {
				if v, ok := s.Attr(attr); ok && v != "" && !strings.HasPrefix(v, "data:") {
					info.CoverURL = strings.TrimSpace(v)
					return
				}
			}
		})
		if info.CoverURL != "" {
			break
		}
	}

	// Description - hentai20.io uses .entry-content p
	doc.Find(".entry-content p, .summary__content p, .description-summary p, #editdescription p, [itemprop='description'] p, .seriestucontent p").Each(func(i int, s *goquery.Selection) {
		if info.Description == "" {
			t := strings.TrimSpace(s.Text())
			if len(t) > 10 {
				info.Description = t
			}
		}
	})
	if info.Description == "" {
		if metaDesc, ok := doc.Find("meta[name='description']").Attr("content"); ok {
			info.Description = strings.TrimSpace(metaDesc)
		}
	}

	// Status - from infotable rows (td pairs)
	doc.Find(".infotable tr, .post-content_item").Each(func(_ int, s *goquery.Selection) {
		tds := s.Find("td")
		if tds.Length() >= 2 {
			label := strings.ToLower(strings.TrimSpace(tds.Eq(0).Text()))
			val := strings.TrimSpace(tds.Eq(1).Text())
			if strings.Contains(label, "status") {
				lv := strings.ToLower(val)
				if strings.Contains(lv, "ongoing") || strings.Contains(lv, "on going") {
					info.Status = "ongoing"
				} else if strings.Contains(lv, "complet") {
					info.Status = "completed"
				} else {
					info.Status = "ongoing"
				}
			}
			if strings.Contains(label, "author") {
				info.Author = val
			}
		}
	})

	// Fallback status from .post-status area
	if info.Status == "" {
		doc.Find(".post-status .summary-content").Each(func(_ int, s *goquery.Selection) {
			if info.Status == "" {
				lv := strings.ToLower(strings.TrimSpace(s.Text()))
				if strings.Contains(lv, "ongoing") {
					info.Status = "ongoing"
				} else if strings.Contains(lv, "complet") {
					info.Status = "completed"
				}
			}
		})
	}
	if info.Status == "" {
		info.Status = "ongoing"
	}

	// Genres - hentai20.io uses <b>Genres</b>: <a rel="tag">...</a>
	// Also try .genres-content a, .mgen a, a[rel="tag"]
	genresSeen := map[string]bool{}
	addGenre := func(g string) {
		g = strings.TrimSpace(g)
		if g != "" && !genresSeen[g] {
			genresSeen[g] = true
			info.Genres = append(info.Genres, g)
		}
	}

	doc.Find("a[rel='tag']").Each(func(_ int, s *goquery.Selection) {
		addGenre(s.Text())
	})
	// Also check .genres-content a and .mgen a
	if len(info.Genres) == 0 {
		doc.Find(".genres-content a, .genre a, .mgen a, .genre-item, .manga-genres a").Each(func(_ int, s *goquery.Selection) {
			addGenre(s.Text())
		})
	}

	// Chapter list - hentai20.io uses #chapterlist ul li with data-num
	doc.Find("#chapterlist li, .chapterlist li").Each(func(_ int, s *goquery.Selection) {
		a := s.Find("a").First()
		href, _ := a.Attr("href")
		if href == "" {
			return
		}
		// Get chapter title from .chapternum span
		title := strings.TrimSpace(s.Find(".chapternum").Text())
		if title == "" {
			title = strings.TrimSpace(a.Text())
		}

		// Get number from data-num attribute first
		var num float64
		if numStr, ok := s.Attr("data-num"); ok {
			num, _ = strconv.ParseFloat(strings.TrimSpace(numStr), 64)
		}
		if num == 0 {
			num = parseChapterNumber(title)
		}

		chSlug := extractSlugFromURL(href)
		info.Chapters = append(info.Chapters, ChapterInfo{
			Title:  title,
			Slug:   chSlug,
			Number: num,
			URL:    href,
		})
	})

	// Fallback: try older madara theme selectors
	if len(info.Chapters) == 0 {
		doc.Find(".wp-manga-chapter, .chapter-item, li.chapter").Each(func(_ int, s *goquery.Selection) {
			a := s.Find("a").First()
			href, _ := a.Attr("href")
			title := strings.TrimSpace(a.Text())
			if href == "" {
				return
			}
			chSlug := extractSlugFromURL(href)
			num := parseChapterNumber(title)
			info.Chapters = append(info.Chapters, ChapterInfo{
				Title:  title,
				Slug:   chSlug,
				Number: num,
				URL:    href,
			})
		})
	}

	// Fallback: blog/button style — manhwamyanmar.com uses
	// <a href="..." rel="noopener"><button class="btn"> Chapter N </button></a>
	if len(info.Chapters) == 0 {
		doc.Find(".entry-content a[rel='noopener'], .post-content a[rel='noopener'], .entry-content center a, .post-body a[rel='noopener']").Each(func(_ int, s *goquery.Selection) {
			href, _ := s.Attr("href")
			if href == "" {
				return
			}
			title := strings.TrimSpace(s.Text())
			if title == "" {
				title = strings.TrimSpace(s.Find("button").Text())
			}
			num := parseChapterNumber(title)
			if num == 0 {
				return // skip non-chapter links
			}
			chSlug := extractSlugFromURL(href)
			if chSlug == "" {
				chSlug = fmt.Sprintf("%s-chapter-%.0f", info.Slug, num)
			}
			info.Chapters = append(info.Chapters, ChapterInfo{
				Title:  title,
				Slug:   chSlug,
				Number: num,
				URL:    href,
			})
		})
	}

	// Sort chapters ascending by number
	sort.Slice(info.Chapters, func(i, j int) bool {
		return info.Chapters[i].Number < info.Chapters[j].Number
	})

	return info, nil
}

// imgSrcRe extracts src from img tags
var imgSrcRe = regexp.MustCompile(`<img[^>]+src="([^"]+)"`)

// ScrapeChapterImages scrapes image URLs from a chapter page
func ScrapeChapterImages(chapterURL string) ([]string, error) {
	doc, err := fetchDoc(chapterURL)
	if err != nil {
		return nil, fmt.Errorf("fetching chapter page: %w", err)
	}

	var images []string
	seen := map[string]bool{}

	addImg := func(src string) {
		src = strings.TrimSpace(src)
		if src != "" && !strings.HasPrefix(src, "data:") && !seen[src] {
			seen[src] = true
			images = append(images, src)
		}
	}

	// hentai20.io puts images in #readerarea inside a <noscript> block.
	// goquery does NOT parse noscript content as HTML - it's treated as raw text.
	// We need to grab the raw text content and regex-parse it.
	doc.Find("#readerarea noscript").Each(func(_ int, s *goquery.Selection) {
		// .Text() gives us the raw noscript content as a string
		rawText := s.Text()
		// Try to find img src attributes within it
		matches := imgSrcRe.FindAllStringSubmatch(rawText, -1)
		for _, m := range matches {
			if len(m) >= 2 {
				addImg(m[1])
			}
		}
		// Also try Html() which may decode it differently
		if len(matches) == 0 {
			rawHTML, _ := s.Html()
			matches2 := imgSrcRe.FindAllStringSubmatch(rawHTML, -1)
			for _, m := range matches2 {
				if len(m) >= 2 {
					addImg(m[1])
				}
			}
		}
	})

	// Fallback: try direct img tags in readerarea (some themes render them directly)
	if len(images) == 0 {
		doc.Find("#readerarea img, .reading-content img, .page-break img, .wp-manga-chapter-img img").Each(func(_ int, s *goquery.Selection) {
			for _, attr := range []string{"src", "data-src", "data-lazy-src"} {
				if v, ok := s.Attr(attr); ok && v != "" && !strings.HasPrefix(v, "data:") {
					addImg(v)
					break
				}
			}
		})
	}

	// Fallback: blog style — manhwamyanmar.com puts images directly in .entry-content
	// using data-src lazy loading (a3-lazy-load plugin)
	if len(images) == 0 {
		doc.Find(".entry-content img, .post-content img, .post-body img").Each(func(_ int, s *goquery.Selection) {
			for _, attr := range []string{"data-src", "data-lazy-src", "src"} {
				v, ok := s.Attr(attr)
				if !ok || v == "" {
					continue
				}
				// Skip placeholders (gif, svg, tiny images)
				if strings.HasPrefix(v, "data:") || strings.Contains(v, "lazy_placeholder") || strings.HasSuffix(v, ".gif") {
					continue
				}
				// Make sure URL is absolute
				if strings.HasPrefix(v, "//") {
					v = "https:" + v
				}
				addImg(v)
				break
			}
		})
	}

	return images, nil
}
