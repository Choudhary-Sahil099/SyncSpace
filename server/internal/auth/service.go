package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

var (
	ErrEmailTaken        = errors.New("an account with this email already exists")
	ErrUsernameTaken     = errors.New("an account with this username already exists")
	ErrInvalidCredential = errors.New("invalid email or password")
	ErrInvalidToken      = errors.New("invalid or expired access token")
)

type User struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type Service struct {
	db        *sql.DB
	jwtSecret []byte
}

func NewService(databasePath string, jwtSecret string) (*Service, error) {
	if strings.TrimSpace(jwtSecret) == "" {
		return nil, errors.New("JWT secret must not be empty")
	}

	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("open account database: %w", err)
	}

	service := &Service{db: db, jwtSecret: []byte(jwtSecret)}
	if err := service.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return service, nil
}

func (s *Service) Close() error { return s.db.Close() }

func (s *Service) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE COLLATE NOCASE,
			username TEXT NOT NULL UNIQUE COLLATE NOCASE,
			password_hash TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		return fmt.Errorf("create users table: %w", err)
	}
	return nil
}

func (s *Service) Register(email string, username string, password string) (User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	username = strings.TrimSpace(username)
	if !strings.Contains(email, "@") || len(username) < 3 || len(username) > 40 || len(password) < 8 {
		return User{}, errors.New("use a valid email, a 3-40 character username, and an 8+ character password")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("hash password: %w", err)
	}

	result, err := s.db.Exec(
		"INSERT INTO users (email, username, password_hash) VALUES (?, ?, ?)",
		email, username, string(hash),
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			if strings.Contains(err.Error(), "users.email") {
				return User{}, ErrEmailTaken
			}
			if strings.Contains(err.Error(), "users.username") {
				return User{}, ErrUsernameTaken
			}
			return User{}, ErrEmailTaken
		}
		return User{}, fmt.Errorf("create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return User{}, fmt.Errorf("read user ID: %w", err)
	}
	return User{ID: id, Email: email, Username: username}, nil
}

func (s *Service) Login(email string, password string) (User, error) {
	var user User
	var passwordHash string

	err := s.db.QueryRow(
		"SELECT id, email, username, password_hash FROM users WHERE email = ?",
		strings.ToLower(strings.TrimSpace(email)),
	).Scan(&user.ID, &user.Email, &user.Username, &passwordHash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrInvalidCredential
		}
		return User{}, fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return User{}, ErrInvalidCredential
	}

	return user, nil
}

type tokenHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

type tokenClaims struct {
	Subject  int64  `json:"sub"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Expires  int64  `json:"exp"`
}

func (s *Service) CreateAccessToken(user User) (string, error) {
	headerBytes, _ := json.Marshal(tokenHeader{Algorithm: "HS256", Type: "JWT"})
	header := base64.RawURLEncoding.EncodeToString(headerBytes)

	claimsBytes, err := json.Marshal(tokenClaims{
		Subject:  user.ID,
		Email:    user.Email,
		Username: user.Username,
		Expires:  time.Now().Add(24 * time.Hour).Unix(),
	})
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(claimsBytes)
	signature := s.sign(header + "." + payload)

	return header + "." + payload + "." + signature, nil
}

func (s *Service) AuthenticateAccessToken(token string) (User, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || !hmac.Equal([]byte(parts[2]), []byte(s.sign(parts[0]+"."+parts[1]))) {
		return User{}, ErrInvalidToken
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return User{}, ErrInvalidToken
	}
	var header tokenHeader
	if json.Unmarshal(headerBytes, &header) != nil || header.Algorithm != "HS256" || header.Type != "JWT" {
		return User{}, ErrInvalidToken
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return User{}, ErrInvalidToken
	}
	var claims tokenClaims
	if json.Unmarshal(payloadBytes, &claims) != nil || claims.Subject == 0 || claims.Expires <= time.Now().Unix() {
		return User{}, ErrInvalidToken
	}

	return User{ID: claims.Subject, Email: claims.Email, Username: claims.Username}, nil
}

func (s *Service) sign(message string) string {
	mac := hmac.New(sha256.New, s.jwtSecret)
	mac.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func NewDevelopmentSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}