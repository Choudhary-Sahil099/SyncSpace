package auth

import "testing"

func TestRegisterLoginAndAccessToken(t *testing.T) {
	service, err := NewService(":memory:", "test-secret")
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()

	registered, err := service.Register("ada@example.com", "ada", "password123")
	if err != nil {
		t.Fatal(err)
	}

	loggedIn, err := service.Login("ada@example.com", "password123")
	if err != nil || loggedIn.ID != registered.ID {
		t.Fatalf("login failed: %v", err)
	}

	token, err := service.CreateAccessToken(loggedIn)
	if err != nil {
		t.Fatal(err)
	}
	user, err := service.AuthenticateAccessToken(token)
	if err != nil || user.Username != "ada" {
		t.Fatalf("token authentication failed: %v", err)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	service, err := NewService(":memory:", "test-secret")
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()
	if _, err := service.Register("ada@example.com", "ada", "password123"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Login("ada@example.com", "incorrect"); err != ErrInvalidCredential {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestWorkspaceTasksAndDocuments(t *testing.T) {
	service, err := NewService(":memory:", "test-secret")
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()

	user, err := service.Register("test@example.com", "testuser", "password123")
	if err != nil {
		t.Fatal(err)
	}

	if err := service.EnsureDefaultWorkspace(user); err != nil {
		t.Fatal(err)
	}

	workspaces, err := service.ListWorkspaces(user.ID)
	if err != nil || len(workspaces) != 1 {
		t.Fatalf("expected 1 workspace, got %v", len(workspaces))
	}

	ws := workspaces[0]
	if len(ws.Documents) != 1 {
		t.Fatalf("expected 1 default document, got %v", len(ws.Documents))
	}
	if len(ws.Tasks) != 3 {
		t.Fatalf("expected 3 default tasks, got %v", len(ws.Tasks))
	}

	// Test Document Update & Delete
	doc := ws.Documents[0]
	updatedDoc, err := service.UpdateDocument(user.ID, ws.ID, doc.ID, "Updated Title", "Updated Description")
	if err != nil || updatedDoc.Title != "Updated Title" {
		t.Fatalf("failed to update document: %v", err)
	}

	newDoc, err := service.CreateDocument(user.ID, ws.ID, "Second Doc", "Desc")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.DeleteDocument(user.ID, ws.ID, newDoc.ID); err != nil {
		t.Fatalf("failed to delete document: %v", err)
	}

	// Test Task CRUD
	createdTask, err := service.CreateTask(user.ID, ws.ID, "Write unit tests")
	if err != nil || createdTask.Title != "Write unit tests" {
		t.Fatalf("failed to create task: %v", err)
	}

	updatedTask, err := service.UpdateTask(user.ID, ws.ID, createdTask.ID, true)
	if err != nil || !updatedTask.Completed {
		t.Fatalf("failed to update task: %v", err)
	}

	tasks, err := service.ListTasks(user.ID, ws.ID)
	if err != nil || len(tasks) != 4 { // 3 default + 1 created
		t.Fatalf("expected 4 tasks, got %v", len(tasks))
	}

	if err := service.DeleteTask(user.ID, ws.ID, createdTask.ID); err != nil {
		t.Fatalf("failed to delete task: %v", err)
	}

	tasksAfterDelete, err := service.ListTasks(user.ID, ws.ID)
	if err != nil || len(tasksAfterDelete) != 3 {
		t.Fatalf("expected 3 tasks after deletion, got %v", len(tasksAfterDelete))
	}
}
