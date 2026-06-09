package websocket

type Room struct {
    ID      string
    Clients map[*Client]bool
    Cursors map[string]int
}