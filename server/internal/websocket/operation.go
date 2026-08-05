package websocket

// better version update for the OT engine
type Operation struct {

	// identity
	ID     string `json:"id"`
	UserID string `json:"userId"`

	// metadata
	Type     string `json:"type"`
	Position int    `json:"position"`
	Text     string `json:"text,omitempty"`
	Length   int    `json:"length,omitempty"`

	// metaData
	BaseVersion int64 `json:"baseVersion"`
	Version     int64 `json:"version"`
	Timestamp   int64 `json:"timestamp"`
}
