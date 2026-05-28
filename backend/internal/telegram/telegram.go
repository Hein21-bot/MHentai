package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	botToken  string
	channelID string
	client    = &http.Client{Timeout: 10 * time.Second}
)

func Init() {
	botToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	channelID = os.Getenv("TELEGRAM_CHANNEL_ID")
}

func Enabled() bool {
	return botToken != "" && channelID != ""
}

type SeriesInfo struct {
	Title        string
	Description  string
	CoverURL     string
	ChapterCount int
	Status       string
	Genres       string
	Slug         string
}

// NotifyNewSeries sends a photo+caption message if cover exists, otherwise text.
// Falls back to text-only if sendPhoto fails (e.g. hotlink-protected image URL).
func NotifyNewSeries(s SeriesInfo) {
	if !Enabled() {
		return
	}
	caption := buildCaption(s)
	if s.CoverURL != "" {
		if err := sendPhoto(s.CoverURL, caption); err != nil {
			fmt.Printf("[telegram] sendPhoto failed (%v), falling back to text\n", err)
			if err2 := sendMessage(caption); err2 != nil {
				fmt.Printf("[telegram] sendMessage fallback failed: %v\n", err2)
			}
		}
	} else {
		if err := sendMessage(caption); err != nil {
			fmt.Printf("[telegram] sendMessage failed: %v\n", err)
		}
	}
}

func buildCaption(s SeriesInfo) string {
	status := "Ongoing"
	if s.Status == "completed" {
		status = "Completed"
	}

	desc := strings.TrimSpace(s.Description)
	if len([]rune(desc)) > 200 {
		runes := []rune(desc)
		desc = string(runes[:200]) + "..."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📖 <b>%s</b>\n", escapeHTML(s.Title)))
	sb.WriteString(fmt.Sprintf("🔖 %s", status))
	if s.Genres != "" {
		sb.WriteString(fmt.Sprintf(" · %s", escapeHTML(s.Genres)))
	}
	sb.WriteString("\n")
	if s.ChapterCount > 0 {
		sb.WriteString(fmt.Sprintf("📚 %d chapters\n", s.ChapterCount))
	}
	if desc != "" {
		sb.WriteString(fmt.Sprintf("\n%s\n", escapeHTML(desc)))
	}
	if s.Slug != "" {
		sb.WriteString(fmt.Sprintf("\n🔗 <a href=\"https://mybooks.sbs/my/series/%s\">Read on mybooks.sbs</a>\n", s.Slug))
	}
	return sb.String()
}

func sendPhoto(photoURL, caption string) error {
	payload := map[string]interface{}{
		"chat_id":    channelID,
		"photo":      photoURL,
		"caption":    caption,
		"parse_mode": "HTML",
	}
	return post("sendPhoto", payload)
}

func sendMessage(text string) error {
	payload := map[string]interface{}{
		"chat_id":    channelID,
		"text":       text,
		"parse_mode": "HTML",
	}
	return post("sendMessage", payload)
}

func post(method string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", botToken, method)
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned %d", resp.StatusCode)
	}
	return nil
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
