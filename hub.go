package main

import (
	"log"
	"strconv"
)

type hub struct {
	// Registered connections and rooms connected to
	connectionsMap map[*connection]map[string]bool

	// Inbound messages from the connections.
	broadcastChannel chan broadcastStruct

	// Join requests from the connections.
	joinChannel chan channelSubScribtion

	// Leave requests from the connections.
	leaveChannel chan channelSubScribtion

	// Unregister requests from connections.
	disconnectChannel chan *connection

	//Room connections
	roomConnections map[string][]*connection
}

//The main Hub instance
var h = hub{
	broadcastChannel:  make(chan broadcastStruct),
	joinChannel:       make(chan channelSubScribtion),
	leaveChannel:      make(chan channelSubScribtion),
	disconnectChannel: make(chan *connection),
	connectionsMap:    make(map[*connection]map[string]bool),
	roomConnections:   make(map[string][]*connection),
}

var isRunning = false

func (h *hub) run() {
	isRunning = true
	for {
		select {
		case sub := <-h.joinChannel:
			go h.join(sub)
		case sub := <-h.leaveChannel:
			go h.leave(sub)
		case c := <-h.disconnectChannel:
			go h.disconnect(c)
		case b := <-h.broadcastChannel:
			go h.broadcast(b.src.userName, b.room, b.msg, b.typ)
		}
	}
	isRunning = false
}

//Sends messages to all the connections in a specified room
func (h *hub) broadcast(username string, room string, msg string, typ string) {
	for _, c := range h.roomConnections[room] {
		select {
		case c.send <- outgoing{F: username, R: room, M: msg, T: typ}:
		default:
			h.disconnect(c)
		}
	}
}

//Checks if the user is already connected or not
func (h *hub) isConnected(userName string) bool {
	for x, _ := range h.connectionsMap {
		if x.userName == userName {
			return true
		}
	}
	return false
}

//Function to disconnect a user connection
func (h *hub) disconnect(c *connection) {
	if conn := h.connectionsMap[c]; conn != nil {
		delete(h.connectionsMap, c)
		for room, _ := range conn {
			h.roomConnections[room], _ = removeConnection(h.roomConnections[room], c)
			log.Println("default-leave: " + c.userName + " from: " + room)
			h.broadcast(c.userName, room, "", LEAVING)
		}
		log.Println("disconnect: " + c.userName)
		close(c.send)
	}
}

//Used to join a user to a room
func (h *hub) join(sub channelSubScribtion) bool {

	var res bool
	subs := h.connectionsMap[sub.src]
	if subs == nil {
		subs = make(map[string]bool)
	}
	subs[sub.room] = true
	h.connectionsMap[sub.src] = subs
	conns := h.roomConnections[sub.room]
	h.roomConnections[sub.room], res = addConnection(conns, sub.src)

	log.Println("join: " + sub.src.userName + " toRoom: " + sub.room + " " + strconv.FormatBool(res))

	//Tell everyone that a new connection has joined the room
	if res {
		h.broadcast(sub.src.userName, sub.room, "", JOINING)
	}

	return res
}

//Called when the user dicides to leave the room
func (h *hub) leave(sub channelSubScribtion) bool {
	var res bool
	if _, ok := h.connectionsMap[sub.src][sub.room]; ok {
		delete(h.connectionsMap, sub.src)
		h.roomConnections[sub.room], res = removeConnection(h.roomConnections[sub.room], sub.src)
	}
	log.Println("leave: " + sub.src.userName + " fromRoom: " + sub.room + " " + strconv.FormatBool(res))

	//Tell everyone that a connection has left the room
	if res {
		h.broadcast(sub.src.userName, sub.room, "", LEAVING)
	}

	return res
}

//Helper func to remove a connection
func removeConnection(connections []*connection, connection *connection) ([]*connection, bool) {
	for p, c := range connections {
		if c == connection {
			log.Println("removeConnection: " + c.userName)
			connections = append(connections[:p], connections[p+1:]...)
			return connections, true
		}
	}
	log.Println("removeConnection: attempt not found " + connection.userName)
	cons := len(connections)
	log.Println("removeConnection: current connections " + strconv.Itoa(cons))
	return connections, false
}

//Helper function to add a connection
func addConnection(connections []*connection, connection *connection) ([]*connection, bool) {
	for _, c := range connections {
		if c.userName == connection.userName {
			log.Println("addConnection: attempt already added " + c.userName)
			return connections, false
		}
	}
	connections = append(connections, connection)
	log.Println("addConnection: " + connection.userName)
	cons := len(connections)
	log.Println("addConnection: current connections " + strconv.Itoa(cons))
	return connections, true
}

//Broadcast channel type
type broadcastStruct struct {
	src  *connection
	msg  string
	room string
	typ  string
}

//Subscribtion channel type
type channelSubScribtion struct {
	src  *connection
	room string
}
