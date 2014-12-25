package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
)

type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	//User name
	userName string

	// Buffered channel of outbound messages.
	send chan outgoing
}

const (
	// Commands
	JOIN  = "J"
	LEAVE = "L"
	SEND  = "S"

	//Message types
	DEFAULT = "D"
	MESSAGE = "M"
	LEAVING = "L"
	JOINING = "J"
)

//Outgoing messages structure to the client
type outgoing struct {
	T string //Message type
	F string //From
	M string //Message
	R string //Room name
}

//Incoming messages structure from the client
type incoming struct {
	C string //Command
	T string //Message type
	M string //Message
	R string //Room name
}

//The main upgrader instance
var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//This function is used to read incomming messages from the clients to the server
func (c *connection) reader(ws *websocket.Conn) {
	var in incoming
	for {
		err := c.ws.ReadJSON(&in)
		if err != nil {
			log.Println("Parsing Error: " + err.Error() + " while reading from: " + c.userName)
			break
		}

		switch in.C {
		case SEND:
			bc := broadcastStruct{
				src:  c,
				msg:  in.M,
				room: in.R,
				typ:  in.T,
			}
			h.broadcastChannel <- bc
		case JOIN:
			cs := channelSubScribtion{
				src:  c,
				room: in.R,
			}
			h.joinChannel <- cs
		case LEAVE:
			cs := channelSubScribtion{
				src:  c,
				room: in.R,
			}
			h.leaveChannel <- cs
		}

	}
	c.ws.Close()
}

//This function is used to write ougoing messages from the server to the client
func (c *connection) writer() {
	for message := range c.send {
		if msg, err := json.Marshal(message); err == nil {
			err := c.ws.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("writer: err " + err.Error() + " while sending to: " + c.userName)
				break
			}
			log.Print("writer: sent " + string(msg) + " to: " + c.userName)
		}
	}
	c.ws.Close()
}
