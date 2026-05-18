package router

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"mhentai-backend/internal/handlers"
)

func New(adminToken string) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:5174", "http://localhost:3000", "http://localhost:3001", "https://mhentai.pages.dev", "https://mybooks.sbs",},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Admin-Token"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Public API
	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		api.GET("/series", handlers.ListSeries)
		api.GET("/series/:id", handlers.GetSeries)
		api.GET("/series/:id/latest-chapters", handlers.GetSeriesLatestChapters)
		api.GET("/chapters/:slug", handlers.GetChapter)
		api.GET("/latest", handlers.GetLatestUpdates)
		api.GET("/proxy/img", handlers.ProxyImage)
	}

	// Admin API (simple token auth)
	admin := r.Group("/api/admin")
	admin.Use(func(c *gin.Context) {
		token := c.GetHeader("X-Admin-Token")
		if token == "" {
			// Also check query param for convenience
			token = c.Query("token")
		}
		if token != adminToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	})
	{
		admin.GET("/stats", handlers.AdminGetStats)
		admin.GET("/recent", handlers.AdminGetRecentSeries)
		admin.GET("/series", handlers.AdminListSeries)
		admin.POST("/series", handlers.AdminCreateSeries)
		admin.PUT("/series/:id", handlers.AdminUpdateSeries)
		admin.DELETE("/series/:id", handlers.AdminDeleteSeries)
		admin.GET("/chapters", handlers.AdminListChapters)
		admin.POST("/chapters", handlers.AdminCreateChapter)
		admin.DELETE("/chapters/:id", handlers.AdminDeleteChapter)
		admin.POST("/upload/presign", handlers.AdminPresignUpload)
		admin.POST("/import", handlers.AdminImport)
		admin.POST("/import/preview-url", handlers.AdminPreviewURL)
		admin.POST("/import/preview", handlers.AdminPreviewChapters)
		admin.POST("/import/selected", handlers.AdminImportSelectedChapters)
		admin.POST("/import/chapter-images", handlers.AdminImportChapterImages)
		admin.POST("/import/chapters", handlers.AdminScrapeChapterList)
		admin.POST("/upload/series-cover", handlers.AdminUploadSeriesCover)
		admin.POST("/upload/chapter-images", handlers.AdminUploadChapterImages)
		admin.POST("/import/mangaboost/preview", handlers.AdminPreviewMangaBoost)
		admin.POST("/import/mangaboost", handlers.AdminImportMangaBoost)
		admin.POST("/import/mangaboost/rescrape", handlers.AdminRescrapeSeriesImages)
	}

	return r
}
