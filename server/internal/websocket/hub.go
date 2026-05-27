package websocket

import "fmt"

// active room management
type Hub struct {
	Rooms      map[string]map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan Message
}

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[string]map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan Message),
	}
}

func (h *Hub) Run() {

	for {

		select {

		case client := <-h.Register:

			fmt.Println("CLIENT JOINED:", client.Username)

			// Create room if not exists
			if h.Rooms[client.RoomID] == nil {
				h.Rooms[client.RoomID] = make(map[*Client]bool)
			}

			// Add client
			h.Rooms[client.RoomID][client] = true

			// Join message
			joinMessage := Message{
				Type:     "user_joined",
				RoomID:   client.RoomID,
				UserID:   client.ID,
				Username: client.Username,
				Content:  client.Username + " joined the room",
			}

			// Broadcast to room
			for c := range h.Rooms[client.RoomID] {

				fmt.Println("BROADCASTING TO:", c.Username)

				select {

				case c.Send <- joinMessage:
					fmt.Println("MESSAGE SENT")

				default:
					close(c.Send)
					delete(h.Rooms[client.RoomID], c)
				}
			}

		case client := <-h.Unregister:

			if _, ok := h.Rooms[client.RoomID][client]; ok {

				delete(h.Rooms[client.RoomID], client)

				close(client.Send)

				leaveMessage := Message{
					Type:     "user_left",
					RoomID:   client.RoomID,
					UserID:   client.ID,
					Username: client.Username,
					Content:  client.Username + " left the room",
				}

				for c := range h.Rooms[client.RoomID] {

					select {

					case c.Send <- leaveMessage:

					default:
						close(c.Send)
						delete(h.Rooms[client.RoomID], c)
					}
				}
			}

		case message := <-h.Broadcast:

			if clients, ok := h.Rooms[message.RoomID]; ok {

				for client := range clients {

					select {

					case client.Send <- message:

					default:
						close(client.Send)
						delete(clients, client)
					}
				}
			}
		}
	}
}