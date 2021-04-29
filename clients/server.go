package clients

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/prometheus/common/log"
	"github.com/tgbv/swarm-cache/glob"
	"github.com/tgbv/swarm-cache/lib"
	s "github.com/tgbv/swarm-cache/server"
	c "github.com/tgbv/swarm-cache/wrapped-conn"
)

// Aliases
type j = glob.J

// Holds clients server instance
var server = s.Init()

// Starts listening for incoming clients websocket connections.
func StartListening() {
	server.Mux.HandleFunc("/", httpRequestHandler)
	server.Instance.Addr = glob.F_CLIENT_LISTEN

	log.Infof("Starting Client INBOUND server @ %s\n", glob.F_CLIENT_LISTEN)
	server.Instance.ListenAndServe()
}

// Handles the incoming HTTP requests and fires the upgrader
func httpRequestHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := server.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Could not establish WS connection: %s", err)
		return
	}

	wsRequestHandler(conn)
}

// Handles the websocket connection
func wsRequestHandler(conn *websocket.Conn) {
	cw := addClient(conn)
	defer closeConn(cw)

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
			}, nil)
			continue
		}

		// Check message for required meta
		if _, exists := meta["msgID"]; !exists {
			cw.Send(j{
				"message": "'msgID' is required!",
				"code":    glob.RES_BAD_REQ,
			}, nil)
			continue
		}

		// The implicit websocket message handler
		go wsMessageHandler(cw, meta, data)

		if err != nil {
			log.Errorf("Could not read ws message: %s", err)
			return
		}
	}
}

// Handles the websocket message
func wsMessageHandler(cw *c.WrappedConn, meta j, input *[]byte) {

	if glob.F_LOG_IO_MSG {
		log.Infof("CLIENTS: Incoming message \nMeta: %s \nBin:  %s", meta, input)
	}

	// Check message for required metas
	if _, exists := meta["comm"]; !exists {
		return
	}
	// Must have key
	_, exists := meta["key"].(string)
	if !exists {
		cw.Send(j{
			"msgID":   meta["msgID"],
			"message": "Please provide a key!",
			"code":    glob.RES_BAD_REQ,
		}, nil)
		return
	}

	switch meta["comm"].(string) {

	case "get":
		commGet(cw, meta)

	case "set":
		commSet(cw, meta, input)

	case "del":
		commDel(cw, meta)

	default:
		cw.Send(j{
			"msgID":   meta["msgID"],
			"message": "Invalid command!",
			"code":    glob.RES_BAD_REQ,
		}, nil)
	}
}
