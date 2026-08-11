package storage

import "sync"

type Document struct {
	Content           string
	Version           int64
	AppliedOperations map[string]int64
}
type DocumentStore struct {
	mu        sync.RWMutex
	Documents map[string]*Document
}

func NewDocumentStore() *DocumentStore {
	return &DocumentStore{
		Documents: make(map[string]*Document),
	}
}
func (ds *DocumentStore) SaveDocument(
	roomID string,
	content string,
	operationID string,
) int64 {

	ds.mu.Lock()
	defer ds.mu.Unlock()

	doc, exists := ds.Documents[roomID]

	if !exists {

		doc = &Document{
			Content:           "",
			Version:           0,
			AppliedOperations: make(map[string]int64),
		}

		ds.Documents[roomID] = doc
	}

	doc.Content = content
	doc.Version++

	if operationID != "" {
		doc.AppliedOperations[operationID] = doc.Version
	}

	return doc.Version
}
func (ds *DocumentStore) GetOperationVersion(roomID string, operationID string) (int64, bool) {
	if operationID == "" {
		return 0, false
	}

	ds.mu.RLock()
	defer ds.mu.RUnlock()

	doc, exists := ds.Documents[roomID]
	if !exists {
		return 0, false
	}

	version, exists := doc.AppliedOperations[operationID]
	return version, exists
}
func (ds *DocumentStore) GetDocument(
	roomID string,
) *Document {

	ds.mu.RLock()
	defer ds.mu.RUnlock()

	doc, exists := ds.Documents[roomID]

	if !exists {

		return &Document{
			Content: "",
			Version: 0,
		}
	}

	return doc
}