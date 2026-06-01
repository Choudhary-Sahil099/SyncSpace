package websocket

type Message struct {
	Type      string   `json:"type"`
	RoomID    string   `json:"roomId"`
	UserID    string   `json:"userId,omitempty"`
	Username  string   `json:"username,omitempty"`
	Content   string   `json:"content,omitempty"`
	Users     []string `json:"users,omitempty"`
	Version int64 `json:"version,omitempty"`
	Timestamp int64    `json:"timestamp,omitempty"`
}