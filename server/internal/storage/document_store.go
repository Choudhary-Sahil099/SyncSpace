package storage

import "sync"

type DocumentStore struct {
	mu        sync.RWMutex
	Documents map[string]string
}

func NewDocumentStore() *DocumentStore {
	return &DocumentStore{
		Documents: make(map[string]string),
	}
}
func (ds *DocumentStore) SaveDocument(
	roomID string,
	content string,
) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.Documents[roomID] = content
}
func (ds *DocumentStore) GetDocument(
	roomID string,
) string {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return ds.Documents[roomID]
}
