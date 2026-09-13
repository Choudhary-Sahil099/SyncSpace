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

	"github.com/google/uuid"

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

type Workspace struct {
	ID          int64               `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Role        string              `json:"role"`
	InviteCode  string              `json:"inviteCode,omitempty"`
	Members     []WorkspaceMember   `json:"members,omitempty"`
	Documents   []WorkspaceDocument `json:"documents,omitempty"`
	Tasks       []WorkspaceTask     `json:"tasks,omitempty"`
}

type WorkspaceTask struct {
	ID          int64  `json:"id"`
	WorkspaceID int64  `json:"workspaceId"`
	Title       string `json:"title"`
	Completed   bool   `json:"completed"`
	CreatedBy   int64  `json:"createdBy"`
	CreatedAt   string `json:"createdAt"`
}

type WorkspaceMember struct {
	UserID   int64  `json:"userId"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type WorkspaceDocument struct {
	ID          int64  `json:"id"`
	RoomID      string `json:"roomId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	UpdatedAt   string `json:"updatedAt"`
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
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS workspaces (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			invite_code TEXT NOT NULL UNIQUE,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS workspace_members (
			workspace_id INTEGER NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role TEXT NOT NULL DEFAULT 'member',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY(workspace_id, user_id)
		);
		CREATE TABLE IF NOT EXISTS workspace_documents (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			room_id TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			created_by INTEGER NOT NULL REFERENCES users(id),
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS workspace_invites (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			code TEXT NOT NULL UNIQUE,
			created_by INTEGER NOT NULL REFERENCES users(id),
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS workspace_tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			title TEXT NOT NULL,
			completed BOOLEAN NOT NULL DEFAULT 0,
			created_by INTEGER NOT NULL REFERENCES users(id),
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("create workspace tables: %w", err)
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

func (s *Service) EnsureDefaultWorkspace(user User) error {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM workspace_members WHERE user_id=?`, user.ID).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	code := uuid.NewString()
	result, err := tx.Exec(`INSERT INTO workspaces(name, description, invite_code) VALUES(?,?,?)`, user.Username+"'s Workspace", "Your personal SyncSpace workspace.", code)
	if err != nil {
		return err
	}
	wid, err := result.LastInsertId()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO workspace_members(workspace_id,user_id,role) VALUES(?,?,?)`, wid, user.ID, "owner"); err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO workspace_documents(workspace_id,room_id,title,description,created_by) VALUES(?,?,?,?,?)`, wid, "room1", "System Architecture", "Real-time architecture and infrastructure notes for SyncSpace.", user.ID); err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO workspace_tasks(workspace_id,title,completed,created_by) VALUES(?,?,?,?)`, wid, "Explore real-time collaborative editing", 0, user.ID); err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO workspace_tasks(workspace_id,title,completed,created_by) VALUES(?,?,?,?)`, wid, "Invite teammates with the workspace code", 0, user.ID); err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO workspace_tasks(workspace_id,title,completed,created_by) VALUES(?,?,?,?)`, wid, "Ask SyncSpace AI about architecture notes", 0, user.ID); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *Service) ListWorkspaces(userID int64) ([]Workspace, error) {
	rows, err := s.db.Query(`SELECT w.id,w.name,w.description,wm.role,w.invite_code FROM workspaces w JOIN workspace_members wm ON wm.workspace_id=w.id WHERE wm.user_id=? ORDER BY w.id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Workspace
	for rows.Next() {
		var w Workspace
		if err := rows.Scan(&w.ID, &w.Name, &w.Description, &w.Role, &w.InviteCode); err != nil {
			return nil, err
		}
		result = append(result, w)
	}
	for i := range result {
		if err := s.populateWorkspace(&result[i]); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (s *Service) GetWorkspace(userID, workspaceID int64) (Workspace, error) {
	var w Workspace
	err := s.db.QueryRow(`SELECT w.id,w.name,w.description,wm.role,w.invite_code FROM workspaces w JOIN workspace_members wm ON wm.workspace_id=w.id WHERE w.id=? AND wm.user_id=?`, workspaceID, userID).Scan(&w.ID, &w.Name, &w.Description, &w.Role, &w.InviteCode)
	if err != nil {
		return Workspace{}, errors.New("workspace not found")
	}
	if err := s.populateWorkspace(&w); err != nil {
		return Workspace{}, err
	}
	return w, nil
}

func (s *Service) populateWorkspace(w *Workspace) error {
	memberRows, err := s.db.Query(`SELECT u.id,u.username,u.email,wm.role FROM workspace_members wm JOIN users u ON u.id=wm.user_id WHERE wm.workspace_id=? ORDER BY CASE wm.role WHEN 'owner' THEN 0 ELSE 1 END,u.username`, w.ID)
	if err != nil {
		return err
	}
	for memberRows.Next() {
		var m WorkspaceMember
		if err := memberRows.Scan(&m.UserID, &m.Username, &m.Email, &m.Role); err != nil {
			memberRows.Close()
			return err
		}
		w.Members = append(w.Members, m)
	}
	memberRows.Close()
	docRows, err := s.db.Query(`SELECT id,room_id,title,description,updated_at FROM workspace_documents WHERE workspace_id=? ORDER BY updated_at DESC,id DESC`, w.ID)
	if err != nil {
		return err
	}
	for docRows.Next() {
		var d WorkspaceDocument
		if err := docRows.Scan(&d.ID, &d.RoomID, &d.Title, &d.Description, &d.UpdatedAt); err != nil {
			docRows.Close()
			return err
		}
		w.Documents = append(w.Documents, d)
	}
	docRows.Close()

	taskRows, err := s.db.Query(`SELECT id,workspace_id,title,completed,created_by,created_at FROM workspace_tasks WHERE workspace_id=? ORDER BY completed ASC, id DESC`, w.ID)
	if err == nil {
		for taskRows.Next() {
			var t WorkspaceTask
			if err := taskRows.Scan(&t.ID, &t.WorkspaceID, &t.Title, &t.Completed, &t.CreatedBy, &t.CreatedAt); err == nil {
				w.Tasks = append(w.Tasks, t)
			}
		}
		taskRows.Close()
	}

	return nil
}

func (s *Service) CreateWorkspace(userID int64, name, description string) (Workspace, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	if len(name) < 2 || len(name) > 80 {
		return Workspace{}, errors.New("workspace name must be 2-80 characters")
	}
	user, err := s.userByID(userID)
	if err != nil {
		return Workspace{}, err
	}
	code := uuid.NewString()
	tx, err := s.db.Begin()
	if err != nil {
		return Workspace{}, err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`INSERT INTO workspaces(name,description,invite_code) VALUES(?,?,?)`, name, description, code)
	if err != nil {
		return Workspace{}, err
	}
	wid, err := res.LastInsertId()
	if err != nil {
		return Workspace{}, err
	}
	if _, err = tx.Exec(`INSERT INTO workspace_members(workspace_id,user_id,role) VALUES(?,?,?)`, wid, userID, "owner"); err != nil {
		return Workspace{}, err
	}
	if _, err = tx.Exec(`INSERT INTO workspace_documents(workspace_id,room_id,title,description,created_by) VALUES(?,?,?,?,?)`, wid, uuid.NewString(), "Untitled document", "Start writing with your team.", userID); err != nil {
		return Workspace{}, err
	}
	if err = tx.Commit(); err != nil {
		return Workspace{}, err
	}
	return s.GetWorkspace(user.ID, wid)
}

func (s *Service) CreateDocument(userID, workspaceID int64, title, description string) (WorkspaceDocument, error) {
	if _, err := s.GetWorkspace(userID, workspaceID); err != nil {
		return WorkspaceDocument{}, err
	}
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if len(title) < 1 || len(title) > 120 {
		return WorkspaceDocument{}, errors.New("document title must be 1-120 characters")
	}
	room := uuid.NewString()
	var d WorkspaceDocument
	err := s.db.QueryRow(`INSERT INTO workspace_documents(workspace_id,room_id,title,description,created_by) VALUES(?,?,?,?,?) RETURNING id,room_id,title,description,updated_at`, workspaceID, room, title, description, userID).Scan(&d.ID, &d.RoomID, &d.Title, &d.Description, &d.UpdatedAt)
	return d, err
}

func (s *Service) UpdateDocument(userID, workspaceID, docID int64, title, description string) (WorkspaceDocument, error) {
	if _, err := s.GetWorkspace(userID, workspaceID); err != nil {
		return WorkspaceDocument{}, err
	}
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if len(title) < 1 || len(title) > 120 {
		return WorkspaceDocument{}, errors.New("document title must be 1-120 characters")
	}
	var d WorkspaceDocument
	err := s.db.QueryRow(`UPDATE workspace_documents SET title=?, description=?, updated_at=CURRENT_TIMESTAMP WHERE id=? AND workspace_id=? RETURNING id,room_id,title,description,updated_at`, title, description, docID, workspaceID).Scan(&d.ID, &d.RoomID, &d.Title, &d.Description, &d.UpdatedAt)
	if err != nil {
		return WorkspaceDocument{}, errors.New("document not found")
	}
	return d, nil
}

func (s *Service) DeleteDocument(userID, workspaceID, docID int64) error {
	if _, err := s.GetWorkspace(userID, workspaceID); err != nil {
		return err
	}
	res, err := s.db.Exec(`DELETE FROM workspace_documents WHERE id=? AND workspace_id=?`, docID, workspaceID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("document not found")
	}
	return nil
}

func (s *Service) ListTasks(userID, workspaceID int64) ([]WorkspaceTask, error) {
	if _, err := s.GetWorkspace(userID, workspaceID); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`SELECT id,workspace_id,title,completed,created_by,created_at FROM workspace_tasks WHERE workspace_id=? ORDER BY completed ASC, id DESC`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []WorkspaceTask
	for rows.Next() {
		var t WorkspaceTask
		if err := rows.Scan(&t.ID, &t.WorkspaceID, &t.Title, &t.Completed, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (s *Service) CreateTask(userID, workspaceID int64, title string) (WorkspaceTask, error) {
	if _, err := s.GetWorkspace(userID, workspaceID); err != nil {
		return WorkspaceTask{}, err
	}
	title = strings.TrimSpace(title)
	if len(title) < 1 || len(title) > 200 {
		return WorkspaceTask{}, errors.New("task title must be 1-200 characters")
	}
	var t WorkspaceTask
	err := s.db.QueryRow(`INSERT INTO workspace_tasks(workspace_id,title,completed,created_by) VALUES(?,?,?,?) RETURNING id,workspace_id,title,completed,created_by,created_at`, workspaceID, title, false, userID).Scan(&t.ID, &t.WorkspaceID, &t.Title, &t.Completed, &t.CreatedBy, &t.CreatedAt)
	return t, err
}

func (s *Service) UpdateTask(userID, workspaceID, taskID int64, completed bool) (WorkspaceTask, error) {
	if _, err := s.GetWorkspace(userID, workspaceID); err != nil {
		return WorkspaceTask{}, err
	}
	var t WorkspaceTask
	err := s.db.QueryRow(`UPDATE workspace_tasks SET completed=? WHERE id=? AND workspace_id=? RETURNING id,workspace_id,title,completed,created_by,created_at`, completed, taskID, workspaceID).Scan(&t.ID, &t.WorkspaceID, &t.Title, &t.Completed, &t.CreatedBy, &t.CreatedAt)
	if err != nil {
		return WorkspaceTask{}, errors.New("task not found")
	}
	return t, nil
}

func (s *Service) DeleteTask(userID, workspaceID, taskID int64) error {
	if _, err := s.GetWorkspace(userID, workspaceID); err != nil {
		return err
	}
	res, err := s.db.Exec(`DELETE FROM workspace_tasks WHERE id=? AND workspace_id=?`, taskID, workspaceID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("task not found")
	}
	return nil
}

func (s *Service) CreateInvite(userID, workspaceID int64) (string, error) {
	w, err := s.GetWorkspace(userID, workspaceID)
	if err != nil {
		return "", err
	}
	code := uuid.NewString()
	_, err = s.db.Exec(`INSERT INTO workspace_invites(workspace_id,code,created_by) VALUES(?,?,?)`, w.ID, code, userID)
	return code, err
}

func (s *Service) JoinWorkspace(userID int64, code string) (Workspace, error) {
	var wid int64
	if err := s.db.QueryRow(`SELECT workspace_id FROM workspace_invites WHERE code=?`, strings.TrimSpace(code)).Scan(&wid); err != nil {
		return Workspace{}, errors.New("invalid invite code")
	}
	if _, err := s.db.Exec(`INSERT OR IGNORE INTO workspace_members(workspace_id,user_id,role) VALUES(?,?,?)`, wid, userID, "member"); err != nil {
		return Workspace{}, err
	}
	return s.GetWorkspace(userID, wid)
}

func (s *Service) userByID(id int64) (User, error) {
	var u User
	err := s.db.QueryRow(`SELECT id,email,username FROM users WHERE id=?`, id).Scan(&u.ID, &u.Email, &u.Username)
	return u, err
}

func (s *Service) HasDocumentAccess(userID int64, roomID string) bool {
	var exists int
	err := s.db.QueryRow(`SELECT 1 FROM workspace_documents d JOIN workspace_members m ON m.workspace_id=d.workspace_id WHERE d.room_id=? AND m.user_id=? LIMIT 1`, roomID, userID).Scan(&exists)
	return err == nil && exists == 1
}

func (s *Service) AllowedRoomIDs(userID int64) map[string]bool {
	rows, err := s.db.Query(`SELECT d.room_id FROM workspace_documents d JOIN workspace_members m ON m.workspace_id=d.workspace_id WHERE m.user_id=?`, userID)
	if err != nil {
		return map[string]bool{}
	}
	defer rows.Close()
	rooms := make(map[string]bool)
	for rows.Next() {
		var room string
		if rows.Scan(&room) == nil {
			rooms[room] = true
		}
	}
	return rooms
}
