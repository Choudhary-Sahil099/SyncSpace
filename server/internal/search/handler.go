package search

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"syncspace/internal/auth"
	"syncspace/internal/storage"
)

func RegisterRoutes(router *gin.Engine, store storage.DocumentStore, authService *auth.Service) {
	router.GET("/api/search", func(c *gin.Context) {
		token := strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
		user, err := authService.AuthenticateAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "a valid access token is required"})
			return
		}
		query := c.Query("q")
		if len(query) < 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "q must contain at least 2 characters"})
			return
		}
		allowed := authService.AllowedRoomIDs(user.ID)
		all := store.SearchDocuments(query)
		results := make([]storage.SearchResult, 0, len(all))
		for _, result := range all {
			if allowed[result.RoomID] {
				results = append(results, result)
			}
		}
		c.JSON(http.StatusOK, gin.H{"query": query, "results": results})
	})
}
