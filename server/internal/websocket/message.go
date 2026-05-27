package websocket

type Message struct {
	Type      string   `json:"type"`
	RoomID    string   `json:"roomId"`
	UserID    string   `json:"userId"`
	Username  string   `json:"username"`
	Content   string   `json:"content"`
	Users     []string `json:"users,omitempty"`
	Timestamp int64    `json:"timestamp"`
}