package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// tsReaderRe matches ts_reader.run({...}) and captures the JSON object
var tsReaderRe = regexp.MustCompile(`ts_reader\.run\((\{[\s\S]*?\})\)`)

// ScrapeMangaBoostSeries scrapes series info + chapter list from a mangaboost.com manga URL.
func ScrapeMangaBoostSeries(seriesURL string) (*SeriesInfo, error) {
	doc, err := fetchMangaBoostDoc(seriesURL)
	if err != nil {
		return nil, fmt.Errorf("fetching series page: %w", err)
	}

	info := &SeriesInfo{SourceURL: seriesURL}

	// Title
	info.Title = strings.TrimSpace(doc.Find("h1.entry-title").First().Text())
	if info.Title == "" {
		info.Title = strings.TrimSpace(doc.Find("h1").First().Text())
	}

	// Slug from URL
	info.Slug = extractSlugFromURL(seriesURL)

	// Cover — direct src attribute (not data-src)
	doc.Find("img.wp-post-image").Each(func(_ int, s *goquery.Selection) {
		if info.CoverURL != "" {
			return
		}
		if v, ok := s.Attr("src"); ok && v != "" && !strings.HasPrefix(v, "data:") {
			info.CoverURL = strings.TrimSpace(v)
		}
	})

	// Description
	doc.Find(".entry-content-single p").Each(func(_ int, s *goquery.Selection) {
		if info.Description == "" {
			t := strings.TrimSpace(s.Text())
			if len(t) > 10 {
				info.Description = t
			}
		}
	})

	// Status — first .tsinfo .imptdt i
	doc.Find(".tsinfo .imptdt").Each(func(i int, s *goquery.Selection) {
		if i == 0 && info.Status == "" {
			lv := strings.ToLower(strings.TrimSpace(s.Find("i").Text()))
			if strings.Contains(lv, "ongoing") || strings.Contains(lv, "on going") {
				info.Status = "ongoing"
			} else if strings.Contains(lv, "complet") {
				info.Status = "completed"
			}
		}
	})
	if info.Status == "" {
		info.Status = "ongoing"
	}

	// Author — look for second imptdt or a specific label
	doc.Find(".tsinfo .imptdt").Each(func(_ int, s *goquery.Selection) {
		label := strings.ToLower(strings.TrimSpace(s.Find("h3, b, .imptdt-tx").Text()))
		if strings.Contains(label, "author") || strings.Contains(label, "artist") {
			if info.Author == "" {
				info.Author = strings.TrimSpace(s.Find("a, i").Last().Text())
			}
		}
	})

	// Genres
	doc.Find(".mgen a").Each(func(_ int, s *goquery.Selection) {
		g := strings.TrimSpace(s.Text())
		if g != "" {
			info.Genres = append(info.Genres, g)
		}
	})

	// Chapter list — .eph-num a contains .chapternum for title
	doc.Find(".eph-num a").Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if href == "" {
			return
		}
		title := strings.TrimSpace(s.Find(".chapternum").Text())
		if title == "" {
			title = strings.TrimSpace(s.Text())
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

	// Mangaboost lists chapters newest-first in the HTML, so reverse to get
	// chronological order (Ch.1 → Ch.2 → ... → Epilogue). Sorting by number
	// would break epilogue/special chapters which parse to wrong numbers.
	for i, j := 0, len(info.Chapters)-1; i < j; i, j = i+1, j-1 {
		info.Chapters[i], info.Chapters[j] = info.Chapters[j], info.Chapters[i]
	}

	return info, nil
}

// ScrapeMangaBoostChapterImages extracts image URLs from a mangaboost.com chapter page.
// Images are embedded in a ts_reader.run({...}) JS call — not in HTML img tags.
func ScrapeMangaBoostChapterImages(chapterURL string) ([]string, error) {
	doc, err := fetchMangaBoostDoc(chapterURL)
	if err != nil {
		return nil, fmt.Errorf("fetching chapter page: %w", err)
	}

	var jsonStr string
	doc.Find("script").Each(func(_ int, s *goquery.Selection) {
		if jsonStr != "" {
			return
		}
		if m := tsReaderRe.FindStringSubmatch(s.Text()); len(m) >= 2 {
			jsonStr = m[1]
		}
	})

	if jsonStr == "" {
		return nil, fmt.Errorf("ts_reader.run() not found in chapter page")
	}

	// ts_reader embeds JS booleans (!1 = false, !0 = true) which aren't valid JSON
	jsonStr = strings.ReplaceAll(jsonStr, ":!1", ":false")
	jsonStr = strings.ReplaceAll(jsonStr, ":!0", ":true")

	var tsData struct {
		Sources []struct {
			Images []string `json:"images"`
		} `json:"sources"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &tsData); err != nil {
		return nil, fmt.Errorf("parsing ts_reader data: %w", err)
	}

	if len(tsData.Sources) == 0 || len(tsData.Sources[0].Images) == 0 {
		return nil, fmt.Errorf("no images found in ts_reader data")
	}

	return tsData.Sources[0].Images, nil
}

func fetchMangaBoostDoc(rawURL string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://mangaboost.com/")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")

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
