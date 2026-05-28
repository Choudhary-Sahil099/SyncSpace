package websocket

import (
	"fmt"
)

type Hub struct {
	Rooms      map[string]*Room
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan Message
}

func NewHub() *Hub {

	return &Hub{
		Rooms:      make(map[string]*Room),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan Message),
	}
}

func (h *Hub) removeClient(
	room *Room,
	client *Client,
) {

	delete(room.Clients, client)

	client.Close()

	fmt.Println("CLIENT REMOVED:", client.Username)
}

func (h *Hub) Run() {

	for {

		select {

		case client := <-h.Register:

			fmt.Println("CLIENT JOINED:", client.Username)

			if h.Rooms[client.RoomID] == nil {

				h.Rooms[client.RoomID] = &Room{
					ID:       client.RoomID,
					Clients:  make(map[*Client]bool),
					Document: "",
				}
			}

			room := h.Rooms[client.RoomID]

			room.Clients[client] = true

			syncMessage := Message{
				Type:    "document_sync",
				RoomID:  room.ID,
				Content: room.Document,
			}

			client.Send <- syncMessage

			joinMessage := Message{
				Type:     "user_joined",
				RoomID:   client.RoomID,
				UserID:   client.ID,
				Username: client.Username,
				Content:  client.Username + " joined the room",
			}

			for c := range room.Clients {

				select {

				case c.Send <- joinMessage:

					fmt.Println("JOIN MESSAGE SENT TO:", c.Username)

				default:

					h.removeClient(room, c)
				}
			}

		case client := <-h.Unregister:

			if room, ok := h.Rooms[client.RoomID]; ok {

				if _, exists := room.Clients[client]; exists {

					h.removeClient(room, client)

					leaveMessage := Message{
						Type:     "user_left",
						RoomID:   client.RoomID,
						UserID:   client.ID,
						Username: client.Username,
						Content:  client.Username + " left the room",
					}

					for c := range room.Clients {

						select {

						case c.Send <- leaveMessage:

						default:

							h.removeClient(room, c)
						}
					}

					if len(room.Clients) == 0 {

						delete(h.Rooms, room.ID)

						fmt.Println("ROOM DELETED:", room.ID)
					}
				}
			}

		case message := <-h.Broadcast:

			fmt.Println("BROADCAST EVENT:", message.Content)

			if room, ok := h.Rooms[message.RoomID]; ok {

				if message.Type == "edit" {

					room.Document = message.Content

					fmt.Println(
						"DOCUMENT UPDATED:",
						room.Document,
					)
				}

				for client := range room.Clients {

					select {

					case client.Send <- message:

						fmt.Println(
							"MESSAGE SENT TO:",
							client.Username,
						)

					default:

						h.removeClient(room, client)
					}
				}
			}
		}
	}
}