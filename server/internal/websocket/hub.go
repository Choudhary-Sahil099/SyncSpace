package websocket

import (
	"fmt"
	"syncspace/internal/storage"
)

type Hub struct {
	Rooms map[string]*Room
	Store *storage.DocumentStore

	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan Message
}

func NewHub() *Hub {
	return &Hub{
		Rooms: make(map[string]*Room),
		Store: storage.NewDocumentStore(),

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

	fmt.Println(
		"CLIENT REMOVED:",
		client.Username,
	)
}

func (h *Hub) getUsers(room *Room) []string {

	users := []string{}

	for client := range room.Clients {
		users = append(users, client.Username)
	}

	return users
}

func (h *Hub) broadcastUsers(room *Room) {

	usersMessage := Message{
		Type:   "users_list",
		RoomID: room.ID,
		Users:  h.getUsers(room),
	}

	for client := range room.Clients {

		select {

		case client.Send <- usersMessage:

		default:

			fmt.Println(
				"CLIENT BUFFER FULL:",
				client.Username,
			)
		}
	}
}

func (h *Hub) Run() {

	for {

		select {

		case client := <-h.Register:

			fmt.Println(
				"CLIENT JOINED:",
				client.Username,
			)

			if h.Rooms[client.RoomID] == nil {

				h.Rooms[client.RoomID] = &Room{
					ID:      client.RoomID,
					Clients: make(map[*Client]bool),
				}
			}

			room := h.Rooms[client.RoomID]

			room.Clients[client] = true

			doc := h.Store.GetDocument(room.ID)

			syncMessage := Message{
				Type:    "document_sync",
				RoomID:  room.ID,
				Content: doc.Content,
				Version: doc.Version,
			}

			client.Send <- syncMessage

			joinMessage := Message{
				Type:     "user_joined",
				RoomID:   room.ID,
				UserID:   client.ID,
				Username: client.Username,
				Content:  client.Username + " joined the room",
			}

			for c := range room.Clients {

				select {

				case c.Send <- joinMessage:

					fmt.Println(
						"JOIN MESSAGE SENT TO:",
						c.Username,
					)

				default:

					fmt.Println(
						"CLIENT BUFFER FULL:",
						c.Username,
					)
				}
			}

			h.broadcastUsers(room)

		case client := <-h.Unregister:

			if room, ok := h.Rooms[client.RoomID]; ok {

				if _, exists := room.Clients[client]; exists {

					h.removeClient(
						room,
						client,
					)

					leaveMessage := Message{
						Type:     "user_left",
						RoomID:   room.ID,
						UserID:   client.ID,
						Username: client.Username,
						Content:  client.Username + " left the room",
					}

					for c := range room.Clients {

						select {

						case c.Send <- leaveMessage:

						default:

							fmt.Println(
								"CLIENT BUFFER FULL:",
								c.Username,
							)
						}
					}

					h.broadcastUsers(room)

					if len(room.Clients) == 0 {

						delete(
							h.Rooms,
							room.ID,
						)

						fmt.Println(
							"ROOM DELETED:",
							room.ID,
						)
					}
				}
			}

		case message := <-h.Broadcast:

			if room, ok := h.Rooms[message.RoomID]; ok {

				if message.Type == "edit" {

					version := h.Store.SaveDocument(
						message.RoomID,
						message.Content,
					)
					message.Version = version
					fmt.Println(
						"DOCUMENT VERSION:",
						version,
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

						fmt.Println(
							"CLIENT BUFFER FULL:",
							client.Username,
						)
					}
				}
			}
		}
	}
}
