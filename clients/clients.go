package clients

import (
	"sync"

	"github.com/gorilla/websocket"
	c "github.com/tgbv/swarm-cache/wrapped-conn"
)

// List of websocket connections.
//
// They're wrapped to contain a connection-custom mutex.
var clients = make([]*c.WrappedConn, 0)

// Clients default mutex.
//
// Used to prevent data race amid clients accessing.
var clientsMux = sync.Mutex{}

// Helper function to add a client.
//
// Mutex protected
func addClient(conn *websocket.Conn) *c.WrappedConn {
	cw := c.Init(conn)

	clientsMux.Lock()
	clients = append(clients, cw)
	clientsMux.Unlock()

	return cw
}

// Helper function to delete a client by index
//
// Mutex protected.
func delClientIndex(index uint) {
	clientsMux.Lock()
	defer clientsMux.Unlock()

	len := len(clients)
	clients[index] = clients[len-1]
	clients = clients[:len-1]
}

// Helper function to delete a client
//
// Mutex protected.
func delClient(cw *c.WrappedConn) {
	clientsMux.Lock()
	defer clientsMux.Unlock()

	l := len(clients)

	for i, tcw := range clients {
		if tcw == cw {
			clients[i] = clients[l-1]
			clients = clients[:l-1]
			return
		}
	}
}

// Helper function to close the connection of a client
//
// Mutex protected.
func closeConn(cw *c.WrappedConn) {
	delClient(cw)

	cw.Conn.Close()
}
