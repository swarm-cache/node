package nodes

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/prometheus/common/log"
)

// Dialer which is used to connect to other nodes
var dialer = websocket.Dialer{}

// Connects to another node
func Connect(addr string) error {
	log.Infof("Connecting to ws://%s", addr)

	conn, res, err := dialer.Dial("ws://"+addr, http.Header{"X-Node-Id": []string{NodeID}})
	if err != nil {
		return err
	}

	if nid, exists := res.Header["X-Node-Id"]; !exists || nid == nil {
		return fmt.Errorf("'X-Node-Id' not present!")
	}

	wsConnectionHandler(conn, j{"id": res.Header["X-Node-Id"][0]})

	return nil
}
