package websocket
type Cursor struct {
    Position       int `json:"position"`
    SelectionStart int `json:"selectionStart"`
    SelectionEnd   int `json:"selectionEnd"`
}

type Message struct {
	Type      string      `json:"type"`
	RoomID    string      `json:"roomId"`
	UserID    string      `json:"userId,omitempty"`
	Username  string      `json:"username,omitempty"`
	Content   string      `json:"content,omitempty"`
	Users     []string    `json:"users,omitempty"`
	Version   int64       `json:"version,omitempty"`
	Operation *Operation  `json:"operation,omitempty"`
	Cursor *Cursor `json:"cursor,omitempty"`
	Timestamp int64       `json:"timestamp,omitempty"`
}