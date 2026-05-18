package repository

import (
	"sort"

	"mhentai-backend/internal/models"
)

func sortSeries(items []*models.Series, by string) {
	switch by {
	case "title":
		sort.Slice(items, func(i, j int) bool {
			return items[i].Title < items[j].Title
		})
	case "views":
		sort.Slice(items, func(i, j int) bool {
			return items[i].ViewCount > items[j].ViewCount
		})
	case "created_at":
		sort.Slice(items, func(i, j int) bool {
			return items[i].CreatedAt > items[j].CreatedAt
		})
	default: // "updated_at"
		sort.Slice(items, func(i, j int) bool {
			return items[i].UpdatedAt > items[j].UpdatedAt
		})
	}
}
