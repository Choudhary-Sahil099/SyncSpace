package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type embeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}
type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

func CreateEmbedding(ctx context.Context, input string) ([]float32, string, error) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		return nil, "", fmt.Errorf("OPENAI_API_KEY is not configured")
	}
	model := os.Getenv("OPENAI_EMBEDDING_MODEL")
	if model == "" {
		model = "text-embedding-3-small"
	}
	payload, _ := json.Marshal(embeddingRequest{Model: model, Input: input})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/embeddings", bytes.NewReader(payload))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("embedding provider returned %s", resp.Status)
	}
	var decoded embeddingResponse
	if err := json.Unmarshal(body, &decoded); err != nil || len(decoded.Data) == 0 {
		return nil, "", fmt.Errorf("invalid embedding response")
	}
	return decoded.Data[0].Embedding, model, nil
}
