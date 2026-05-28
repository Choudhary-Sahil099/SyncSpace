package websocket

type Room struct {
	ID       string
	Clients  map[*Client]bool
	Document string
}