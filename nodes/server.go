package nodes

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/prometheus/common/log"
	"github.com/tgbv/swarm-cache/glob"
	"github.com/tgbv/swarm-cache/lib"
	s "github.com/tgbv/swarm-cache/server"
	c "github.com/tgbv/swarm-cache/wrapped-conn"
)

// Holds Nodes server instance
var server = s.Init()

// Starts listening for incoming Nodes websocket connections.
func StartListening() error {
	server.Mux.HandleFunc("/", httpRequestHandler)
	server.Instance.Addr = glob.F_NODE_LISTEN

	log.Infof("Starting Node INBOUND server @ %s. Node ID is: %s", glob.F_NODE_LISTEN, NodeID)
	return server.Instance.ListenAndServe()
}

// Handles the incoming HTTP requests and fires the upgrader
func httpRequestHandler(w http.ResponseWriter, r *http.Request) {

	// Each node connection must have a special ID
	nid, exists := r.Header["X-Node-Id"]
	if !exists || nid == nil {
		log.Errorf("Could not establish WS connection: %s", "'X-Node-Id' not present!")
		w.WriteHeader(400)
		w.Write([]byte("'X-Node-Id' not present!"))
		return
	}

	// Upgrade the connection to WS
	conn, err := server.Upgrader.Upgrade(
		w,
		r,
		http.Header{"X-Node-Id": []string{NodeID}},
	)
	if err != nil {
		log.Errorf("Could not upgrade WS connection: %s", err)
		return
	}

	wsConnectionHandler(conn, j{"id": nid[0]})
}

// Handles the websocket connection
func wsConnectionHandler(conn *websocket.Conn, nodeMeta j) {
	cw := addNode(conn, nodeMeta)
	defer closeConn(cw)

	log.Infof("NODES: connection  Meta: %s ", nodeMeta)

	cw.PushCB("__default", defaultWSMessageHandler)

	for {
		_, in, err := cw.Conn.ReadMessage()
		if err != nil {
			log.Errorf("Could not read ws message: %s", err)
			return
		}

		// Decode the message, so handlers receive only meta and data
		err, meta, data := lib.DecodeMessage(&in)
		if err != nil {
			cw.Send(j{
				"msgID":   meta["msgID"],
				"message": err.Error(),
				"code":    glob.RES_BAD_REQ,
				"type":    "res",
			}, nil)
			continue
		}

		// Check message for required meta keys
		if _, exists := meta["msgID"]; !exists {
			cw.Send(j{
				"message": "'msgID' is required!",
				"code":    glob.RES_BAD_REQ,
				"type":    "res",
			}, nil)
			continue
		}
		if _, exists := meta["type"]; !exists {
			cw.Send(j{
				"msgID":   meta["msgID"],
				"message": "'type' is required!",
				"code":    glob.RES_BAD_REQ,
				"type":    "res",
			}, nil)
			fmt.Printf("Decoded message: %q", string(in))
			continue
		}

		// traverse the callbacks
		go cw.TraverseCB(meta, data)

		if err != nil {
			log.Errorf("Could not read ws message: %s", err)
			return
		}
	}
}

// Handles the websocket message
func defaultWSMessageHandler(_ string, cw *c.WrappedConn, meta j, input *[]byte) {

	if glob.F_LOG_IO_MSG {
		log.Infof("NODES: Incoming message \nMeta: %s \nBin:  %s", meta, input)
	}

	// Must be a command
	if meta["type"].(string) != "comm" {
		return
	}

	// Command field must be present
	if _, exists := meta["comm"]; !exists {
		cw.Send(j{
			"msgID":   meta["msgID"],
			"code":    glob.RES_BAD_REQ,
			"message": "'comm' is required",
			"type":    "res",
		}, nil)
		return
	}

	// Command ID must be present in message meta
	if _, exists := meta["commID"]; !exists {
		cw.Send(j{
			"msgID":   meta["msgID"],
			"code":    glob.RES_BAD_REQ,
			"message": "'commID' is required",
			"type":    "res",
		}, nil)
		return
	}

	// Command ID must not be received already
	if cw.GetComm(meta["commID"].(string)) != nil {
		cw.Send(j{
			"msgID":   meta["msgID"],
			"code":    glob.RES_CONFLICT,
			"message": "Command ID already passed this node.",
			"type":    "res",
		}, nil)
		return
	}

	// Must have key
	_, exists := meta["key"].(string)
	if !exists {
		cw.Send(j{
			"msgID":   meta["msgID"],
			"message": "Please provide a key!",
			"code":    glob.RES_BAD_REQ,
			"type":    "res",
		}, nil)
		return
	}

	// Register command ID
	cw.PushComm(meta["commID"].(string))

	switch meta["comm"].(string) {

	case "get":
		commGet(cw, meta)

	case "set":
		commSet(cw, meta, input)
	case "del":
		commDel(cw, meta)
	case "delcb":
		//cw.DelCB(key)

	default:
		cw.Send(j{
			"msgID":   meta["msgID"],
			"message": "Invalid command",
			"code":    glob.RES_BAD_REQ,
			"type":    "res",
		}, nil)
	}
}
