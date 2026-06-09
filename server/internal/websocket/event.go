package websocket

type Event struct {
	Client  *Client
	Message Message
}