package storage

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Document struct {
	Content           string
	Version           int64
	AppliedOperations map[string]int64
}

type SearchResult struct {
	RoomID  string  `json:"roomId"`
	Content string  `json:"content"`
	Rank    float32 `json:"rank,omitempty"`
}

type DocumentStore interface {
	SaveDocument(roomID, content, operationID string) int64
	GetOperationVersion(roomID, operationID string) (int64, bool)
	GetDocument(roomID string) *Document
	SearchDocuments(query string) []SearchResult
}

// MemoryDocumentStore is the zero-dependency development/test store.
type MemoryDocumentStore struct {
	mu        sync.RWMutex
	Documents map[string]*Document
}

func NewDocumentStore() *MemoryDocumentStore {
	return &MemoryDocumentStore{Documents: make(map[string]*Document)}
}

func (ds *MemoryDocumentStore) SaveDocument(roomID, content, operationID string) int64 {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	doc := ds.Documents[roomID]
	if doc == nil {
		doc = &Document{AppliedOperations: make(map[string]int64)}
		ds.Documents[roomID] = doc
	}
	doc.Content = content
	doc.Version++
	if operationID != "" {
		doc.AppliedOperations[operationID] = doc.Version
	}
	return doc.Version
}

func (ds *MemoryDocumentStore) GetOperationVersion(roomID, operationID string) (int64, bool) {
	if operationID == "" {
		return 0, false
	}
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	doc := ds.Documents[roomID]
	if doc == nil {
		return 0, false
	}
	version, ok := doc.AppliedOperations[operationID]
	return version, ok
}

func (ds *MemoryDocumentStore) GetDocument(roomID string) *Document {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	doc := ds.Documents[roomID]
	if doc == nil {
		return &Document{Content: "", Version: 0, AppliedOperations: make(map[string]int64)}
	}
	copyDoc := &Document{Content: doc.Content, Version: doc.Version, AppliedOperations: make(map[string]int64, len(doc.AppliedOperations))}
	for k, v := range doc.AppliedOperations {
		copyDoc.AppliedOperations[k] = v
	}
	return copyDoc
}

func (ds *MemoryDocumentStore) SearchDocuments(query string) []SearchResult {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	results := make([]SearchResult, 0)
	for roomID, doc := range ds.Documents {
		content := strings.ToLower(doc.Content)
		if strings.Contains(content, q) {
			results = append(results, SearchResult{RoomID: roomID, Content: doc.Content, Rank: 1})
		}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].RoomID < results[j].RoomID })
	return results
}

// PostgresDocumentStore persists collaborative state and provides PostgreSQL FTS.
type PostgresDocumentStore struct{ pool *pgxpool.Pool }

func NewPostgresDocumentStore(ctx context.Context, databaseURL string) (*PostgresDocumentStore, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}
	store := &PostgresDocumentStore{pool: pool}
	if err := store.Migrate(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return store, nil
}

func (s *PostgresDocumentStore) Close() { s.pool.Close() }

func (s *PostgresDocumentStore) Migrate(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS documents (
            room_id TEXT PRIMARY KEY,
            content TEXT NOT NULL DEFAULT '',
            version BIGINT NOT NULL DEFAULT 0,
            search_vector tsvector GENERATED ALWAYS AS (to_tsvector('english', content)) STORED,
            updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
        );
        CREATE INDEX IF NOT EXISTS idx_documents_search ON documents USING GIN(search_vector);
        CREATE TABLE IF NOT EXISTS applied_operations (
            room_id TEXT NOT NULL,
            operation_id TEXT NOT NULL,
            version BIGINT NOT NULL,
            PRIMARY KEY(room_id, operation_id)
        );
        CREATE INDEX IF NOT EXISTS idx_applied_operations_room ON applied_operations(room_id, version);
        CREATE EXTENSION IF NOT EXISTS vector;
        CREATE TABLE IF NOT EXISTS document_embeddings (
            room_id TEXT PRIMARY KEY REFERENCES documents(room_id) ON DELETE CASCADE,
            embedding vector(1536),
            model TEXT NOT NULL,
            updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
        );
        CREATE INDEX IF NOT EXISTS idx_document_embeddings_hnsw ON document_embeddings USING hnsw (embedding vector_cosine_ops);
    `)
	if err != nil {
		return fmt.Errorf("migrate documents: %w", err)
	}
	return nil
}

func (s *PostgresDocumentStore) SaveDocument(roomID, content, operationID string) int64 {
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0
	}
	defer tx.Rollback(ctx)
	var version int64
	err = tx.QueryRow(ctx, `
        INSERT INTO documents(room_id, content, version) VALUES($1,$2,1)
        ON CONFLICT(room_id) DO UPDATE SET content=EXCLUDED.content, version=documents.version+1, updated_at=NOW()
        RETURNING version`, roomID, content).Scan(&version)
	if err != nil {
		return 0
	}
	if operationID != "" {
		if _, err = tx.Exec(ctx, `INSERT INTO applied_operations(room_id, operation_id, version) VALUES($1,$2,$3) ON CONFLICT DO NOTHING`, roomID, operationID, version); err != nil {
			return 0
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return 0
	}
	return version
}

func (s *PostgresDocumentStore) GetOperationVersion(roomID, operationID string) (int64, bool) {
	if operationID == "" {
		return 0, false
	}
	var version int64
	err := s.pool.QueryRow(context.Background(), `SELECT version FROM applied_operations WHERE room_id=$1 AND operation_id=$2`, roomID, operationID).Scan(&version)
	return version, err == nil
}

func (s *PostgresDocumentStore) GetDocument(roomID string) *Document {
	doc := &Document{AppliedOperations: make(map[string]int64)}
	err := s.pool.QueryRow(context.Background(), `SELECT content, version FROM documents WHERE room_id=$1`, roomID).Scan(&doc.Content, &doc.Version)
	if err != nil {
		return doc
	}
	rows, err := s.pool.Query(context.Background(), `SELECT operation_id, version FROM applied_operations WHERE room_id=$1`, roomID)
	if err != nil {
		return doc
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var version int64
		if rows.Scan(&id, &version) == nil {
			doc.AppliedOperations[id] = version
		}
	}
	return doc
}

func (s *PostgresDocumentStore) SearchDocuments(query string) []SearchResult {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}
	rows, err := s.pool.Query(context.Background(), `
        SELECT room_id, content, ts_rank(search_vector, websearch_to_tsquery('english', $1))
        FROM documents WHERE search_vector @@ websearch_to_tsquery('english', $1)
        ORDER BY 3 DESC LIMIT 20`, query)
	if err != nil {
		return nil
	}
	defer rows.Close()
	results := make([]SearchResult, 0)
	for rows.Next() {
		var r SearchResult
		if rows.Scan(&r.RoomID, &r.Content, &r.Rank) == nil {
			results = append(results, r)
		}
	}
	return results
}
