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
