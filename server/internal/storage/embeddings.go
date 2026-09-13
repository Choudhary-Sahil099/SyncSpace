package storage

import (
	"context"
	"fmt"
	"strings"
)

func formatVector(values []float32) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = fmt.Sprintf("%g", v)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func (s *PostgresDocumentStore) SaveEmbedding(ctx context.Context, roomID, content, model string, embedding []float32) error {
	if len(embedding) == 0 {
		return fmt.Errorf("empty embedding")
	}
	_, err := s.pool.Exec(ctx, `
        INSERT INTO document_embeddings(room_id, embedding, model, updated_at)
        VALUES($1, $2::vector, $3, NOW())
        ON CONFLICT(room_id) DO UPDATE SET embedding=EXCLUDED.embedding, model=EXCLUDED.model, updated_at=NOW()`,
		roomID, formatVector(embedding), model)
	return err
}

func (s *PostgresDocumentStore) VectorSearch(ctx context.Context, embedding []float32, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 5
	}
	rows, err := s.pool.Query(ctx, `
        SELECT d.room_id, d.content, 1 - (e.embedding <=> $1::vector) AS rank
        FROM document_embeddings e
        JOIN documents d ON d.room_id=e.room_id
        ORDER BY e.embedding <=> $1::vector
        LIMIT $2`, formatVector(embedding), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]SearchResult, 0, limit)
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.RoomID, &r.Content, &r.Rank); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}
