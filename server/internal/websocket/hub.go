package websocket

import (
	"fmt"
	"syncspace/internal/storage"
)

type Hub struct {
	Rooms map[string]*Room
	Store *storage.DocumentStore

	OTEngine *OTEngine

	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan Event
}

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[string]*Room),
		Store:      storage.NewDocumentStore(),
		OTEngine:   NewOTEngine(),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan Event),
	}
}

func (h *Hub) removeClient(
	room *Room,
	client *Client,
) {

	delete(room.Clients, client)

	delete(
		room.Cursors,
		client.ID,
	)

	fmt.Println(
		"CURSOR REMOVED:",
		client.Username,
	)

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
					Cursors: make(map[string]CursorState),
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
					cursorRemove := Message{
						Type:   "cursor_remove",
						RoomID: room.ID,
						UserID: client.ID,
					}

					for c := range room.Clients {

						select {

						case c.Send <- cursorRemove:

						default:

							fmt.Println(
								"CLIENT BUFFER FULL:",
								c.Username,
							)
						}
					}
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

		case event := <-h.Broadcast:

			message := event.Message
			sender := event.Client

			if room, ok := h.Rooms[message.RoomID]; ok {
				if message.Type == "cursor_move" {

					if message.Cursor != nil {

						room.Cursors[sender.ID] = CursorState{
							Username:       sender.Username,
							Position:       message.Cursor.Position,
							SelectionStart: message.Cursor.SelectionStart,
							SelectionEnd:   message.Cursor.SelectionEnd,
						}
						message.UserID = sender.ID
						message.Username = sender.Username

						fmt.Printf(
							"CURSOR STATE: %+v\n",
							room.Cursors[sender.ID],
						)
					}
				}

				if message.Type == "edit" {

					doc := h.Store.GetDocument(message.RoomID)

					if message.Operation == nil {
						continue
					}

					// Build the authoritative operation from the sender.
					operation := *message.Operation

					operation.UserID = sender.ID
					operation.BaseVersion = message.Version

					fmt.Println("CONTENT:", message.Content)
					fmt.Printf("INCOMING OPERATION: %+v\n", operation)
					fmt.Println("SERVER VERSION:", doc.Version)
					fmt.Println("CLIENT BASE VERSION:", operation.BaseVersion)

					// if client behind transform the operation
					if operation.BaseVersion < doc.Version {

						fmt.Println(
							"TRANSFORMING AGAINST HISTORY FROM VERSION:",
							operation.BaseVersion,
						)

						operation = h.OTEngine.TransformAgainstHistory(
							operation,
						)
					}

					message.Operation = &operation

					fmt.Printf(
						"TRANSFORMED OPERATION: %+v\n",
						operation,
					)

					//transformed operation to current server 
					fmt.Println(
						"CURRENT DOCUMENT:",
						doc.Content,
					)

					updatedDoc, err := ApplyOperation(
						doc.Content,
						message.Operation,
					)

					if err != nil {

						fmt.Println(
							"OPERATION ERROR:",
							err,
						)

						continue
					}

					fmt.Println(
						"UPDATED DOCUMENT:",
						updatedDoc,
					)

					message.Content = updatedDoc

					// Save the new document
					version := h.Store.SaveDocument(
						message.RoomID,
						message.Content,
					)

					message.Version = version

					// The operation belongs To the server version or not
					message.Operation.Version = version

					//transformed operation to OT history.
					h.OTEngine.History.Add(
						*message.Operation,
					)

					fmt.Println(
						"DOCUMENT VERSION:",
						version,
					)

					// Ack the sender
					ack := Message{
						Type:      "edit_ack",
						RoomID:    message.RoomID,
						Version:   version,
						Operation: message.Operation,
					}

					select {

					case sender.Send <- ack:

						fmt.Println(
							"EDIT ACK SENT TO:",
							sender.Username,
							"VERSION:",
							version,
						)

					default:

						fmt.Println(
							"CLIENT BUFFER FULL:",
							sender.Username,
						)
					}
				}

				for client := range room.Clients {

					if client == sender &&
						(message.Type == "cursor_move" ||
							message.Type == "edit") {
						continue
					}

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
