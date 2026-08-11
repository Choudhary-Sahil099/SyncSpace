package storage

import "testing"

func TestDocumentStoreTracksAppliedOperationIDs(t *testing.T) {
	store := NewDocumentStore()

	version := store.SaveDocument("room-1", "hello", "operation-1")
	if version != 1 {
		t.Fatalf("expected version 1, got %d", version)
	}

	gotVersion, exists := store.GetOperationVersion("room-1", "operation-1")
	if !exists || gotVersion != version {
		t.Fatalf("expected operation to be recorded at version %d, got %d (exists: %t)", version, gotVersion, exists)
	}

	if _, exists := store.GetOperationVersion("room-1", "missing-operation"); exists {
		t.Fatal("unexpected operation match")
	}
}
