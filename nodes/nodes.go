package nodes

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/tgbv/swarm-cache/glob"
	"github.com/tgbv/swarm-cache/lib"
	c "github.com/tgbv/swarm-cache/wrapped-conn"
)

// aliases
type j = glob.J

// Map of websocket nodes connections identified by their unique ID
// var nodes = make([]*c.WrappedConn, 0)
var nodes = map[string]*c.WrappedConn{}

// nodes default mutex.
//
// Used to prevent data race amid nodes accessing.
var nodesMux = sync.Mutex{}

// Node unique ID.
//
// Cryptographically generated when program starts up
var NodeID, _ = lib.GenerateRandomString(16)

// Helper function to add a Node.
//
// Mutex protected
func addNode(conn *websocket.Conn, meta j) *c.WrappedConn {
	cw := c.Init(conn)
	cw.Meta = meta

	nodesMux.Lock()
	nodes[meta["id"].(string)] = cw
	nodesMux.Unlock()

	go cw.RCCleanup()

	return cw
}

// Helper function to delete a Node
//
// Mutex protected.
func delNode(cw *c.WrappedConn) {
	nodesMux.Lock()
	defer nodesMux.Unlock()

	for id := range nodes {
		if nodes[id] == cw {
			delete(nodes, id)
			break
		}
	}

	return
}

// Helper function to close the connection of a Node
//
// Mutex protected.
func closeConn(cw *c.WrappedConn) {
	delNode(cw)

	cw.Conn.Close()
}

// Helper function to retrieve the nodes.
func GetNodes() map[string]*c.WrappedConn {
	nodesMux.Lock()
	defer nodesMux.Unlock()

	return nodes
}
