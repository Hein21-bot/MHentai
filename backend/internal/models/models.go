package models

// Series represents a manga/novel series stored in DynamoDB.
//
// DynamoDB table: mhentai_series
//   PK:   id  (S)
//   GSI1: slug-index           — PK: slug
//   GSI2: type-updated_at-index — PK: #type, SK: updated_at  (type = "SERIES")
//   GSI3: status-updated_at-index — PK: status, SK: updated_at
type Series struct {
	ID           string `dynamodbav:"id"            json:"id"`
	Type         string `dynamodbav:"#type"         json:"-"`          // always "SERIES"
	Slug         string `dynamodbav:"slug"          json:"slug"`
	Title        string `dynamodbav:"title"         json:"title"`
	CoverURL     string `dynamodbav:"cover_url"     json:"cover_url"`
	Description  string `dynamodbav:"description"   json:"description"`
	Status       string `dynamodbav:"status"        json:"status"` // ongoing | completed
	Author       string `dynamodbav:"author"        json:"author"`
	Genres       string `dynamodbav:"genres"        json:"genres"` // comma-separated
	Language     string `dynamodbav:"language"      json:"language"` // "en" | "my"
	ViewCount    int64  `dynamodbav:"view_count"    json:"view_count"`
	ChapterCount int    `dynamodbav:"chapter_count" json:"chapter_count"`
	SourceURL    string `dynamodbav:"source_url"    json:"source_url"`
	CreatedAt    string `dynamodbav:"created_at"    json:"created_at"` // RFC3339
	UpdatedAt    string `dynamodbav:"updated_at"    json:"updated_at"` // RFC3339

	// Populated at query time, not stored
	Chapters []Chapter `dynamodbav:"-" json:"chapters,omitempty"`
}

// Chapter represents a manga chapter stored in DynamoDB.
// Images are stored as a string list inside the chapter item itself.
//
// DynamoDB table: mhentai_chapters
//   PK:   id  (S)
//   GSI1: slug-index              — PK: slug
//   GSI2: series_id-number-index  — PK: series_id, SK: number (N)
//   GSI3: type-created_at-index   — PK: #type, SK: created_at  (type = "CHAPTER")
type Chapter struct {
	ID        string   `dynamodbav:"id"         json:"id"`
	Type      string   `dynamodbav:"#type"      json:"-"` // always "CHAPTER"
	SeriesID  string   `dynamodbav:"series_id"  json:"series_id"`
	Slug      string   `dynamodbav:"slug"       json:"slug"`
	Title     string   `dynamodbav:"title"      json:"title"`
	Number    float64  `dynamodbav:"number"     json:"number"`
	Images     []string `dynamodbav:"images"      json:"images,omitempty"`
	ImagesInR2 bool     `dynamodbav:"images_in_r2" json:"images_in_r2,omitempty"`
	Language   string   `dynamodbav:"language"    json:"language"` // "en" | "my"
	ViewCount int64    `dynamodbav:"view_count" json:"view_count"`
	SourceURL string   `dynamodbav:"source_url" json:"source_url,omitempty"`
	CreatedAt string   `dynamodbav:"created_at" json:"created_at"` // RFC3339
	UpdatedAt string   `dynamodbav:"updated_at" json:"updated_at"` // RFC3339

	// Populated at query time, not stored
	Series *Series `dynamodbav:"-" json:"series,omitempty"`
}
