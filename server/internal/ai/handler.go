package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"syncspace/internal/auth"
	"syncspace/internal/storage"

	"github.com/gin-gonic/gin"
)

type request struct {
	Question string `json:"question"`
}
type openAIRequest struct {
	Model       string              `json:"model"`
	Messages    []map[string]string `json:"messages"`
	Temperature float32             `json:"temperature"`
}
type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func RegisterRoutes(router *gin.Engine, store storage.DocumentStore, authService *auth.Service) {
	router.POST("/api/ai/ask", func(c *gin.Context) {
		token := strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
		user, err := authService.AuthenticateAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "a valid access token is required"})
			return
		}
		var req request
		if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Question) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "question is required"})
			return
		}
		results := store.SearchDocuments(req.Question)
		if postgres, ok := store.(*storage.PostgresDocumentStore); ok && os.Getenv("OPENAI_API_KEY") != "" {
			if embedding, _, err := CreateEmbedding(context.Background(), req.Question); err == nil {
				if vectorResults, err := postgres.VectorSearch(context.Background(), embedding, 5); err == nil && len(vectorResults) > 0 {
					results = vectorResults
				}
			}
		}
		allowed := authService.AllowedRoomIDs(user.ID)
		filtered := make([]storage.SearchResult, 0, len(results))
		for _, result := range results {
			if allowed[result.RoomID] {
				filtered = append(filtered, result)
			}
		}
		results = filtered
		if len(results) > 5 {
			results = results[:5]
		}
		var context strings.Builder
		for _, result := range results {
			fmt.Fprintf(&context, "\n[Document %s]\n%s\n", result.RoomID, result.Content)
		}
		if context.Len() == 0 {
			context.WriteString("No matching workspace documents were retrieved.")
		}

		key := os.Getenv("OPENAI_API_KEY")
		if key == "" {
			c.JSON(http.StatusOK, gin.H{"answer": "AI is configured for the production profile. Set OPENAI_API_KEY on the server to generate an LLM answer. Retrieved context:\n" + context.String(), "sources": results})
			return
		}
		model := os.Getenv("OPENAI_MODEL")
		if model == "" {
			model = "gpt-4.1-mini"
		}
		payload, _ := json.Marshal(openAIRequest{Model: model, Temperature: 0.1, Messages: []map[string]string{{"role": "system", "content": "Answer only from the provided workspace context. If the context is insufficient, say so. Keep the answer concise and mention source document IDs."}, {"role": "user", "content": "Question: " + req.Question + "\nWorkspace context:" + context.String()}}})
		httpReq, _ := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(payload))
		httpReq.Header.Set("Authorization", "Bearer "+key)
		httpReq.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "AI provider unavailable"})
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode >= 300 {
			c.JSON(http.StatusBadGateway, gin.H{"error": "AI provider rejected the request"})
			return
		}
		var decoded openAIResponse
		if json.Unmarshal(body, &decoded) != nil || len(decoded.Choices) == 0 {
			c.JSON(http.StatusBadGateway, gin.H{"error": "invalid AI provider response"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"answer": decoded.Choices[0].Message.Content, "sources": results})
	})
}
