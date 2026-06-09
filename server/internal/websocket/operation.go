package websocket

type Operation struct {
	Type     string `json:"type"`
	Position int    `json:"position"`
	Text     string `json:"text,omitempty"`
	Length   int    `json:"length,omitempty"`
}