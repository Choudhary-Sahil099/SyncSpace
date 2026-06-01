package storage

import "sync"
type Document struct {
	Content string
	Version int64
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
) int64 {

	ds.mu.Lock()
	defer ds.mu.Unlock()

	doc, exists := ds.Documents[roomID]

	if !exists {

		doc = &Document{
			Content: "",
			Version: 0,
		}

		ds.Documents[roomID] = doc
	}

	doc.Content = content
	doc.Version++

	return doc.Version
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
