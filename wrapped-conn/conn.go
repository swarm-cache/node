package wconn

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/prometheus/common/log"
	"github.com/tgbv/swarm-cache/glob"
	"github.com/tgbv/swarm-cache/lib"
)

// Aliases
type j = glob.J

// Implicit callbacks type
type callback func(string, *WrappedConn, j, *[]byte)
type callbacks map[string]callback

// A template holding the websocket connection with its mutex and metadata
type WrappedConn struct {
	Conn *websocket.Conn

	// Holds the callbacks ran after each received message
	Callbacks callbacks

	// Holds metadata information of node
	Meta j

	// Holds the recent commands received by this node
	ReceivedComms map[string]*comm

	// Mutexes to prevent racing
	MuxCB   sync.Mutex
	MuxRC   sync.Mutex // ReceivedComms
	MuxConn sync.Mutex
}

// Initializes a wrapped connection
func Init(conn *websocket.Conn) *WrappedConn {
	return &WrappedConn{
		MuxCB:   sync.Mutex{},
		MuxRC:   sync.Mutex{},
		MuxConn: sync.Mutex{},

		Conn:          conn,
		Meta:          j{},
		Callbacks:     make(callbacks),
		ReceivedComms: map[string]*comm{},
	}
}

// Create callback launched asynchronously once a message is received via connection
func (wc *WrappedConn) PushCB(name string, f callback) {
	wc.MuxCB.Lock()
	defer wc.MuxCB.Unlock()

	wc.Callbacks[name] = f
}

// Delete callback
func (wc *WrappedConn) DelCB(name string) {
	wc.MuxCB.Lock()
	defer wc.MuxCB.Unlock()

	delete(wc.Callbacks, name)
}

// Traverses callbacks with given input data
func (wc *WrappedConn) TraverseCB(meta j, data *[]byte) {
	wc.MuxCB.Lock()
	defer wc.MuxCB.Unlock()

	for name, f := range wc.Callbacks {
		go f(name, wc, meta, data)
	}
}

// Sends a message to current websocket connection.
func (wc *WrappedConn) Send(meta glob.J, data *[]byte) error {
	wc.MuxConn.Lock()
	defer wc.MuxConn.Unlock()

	err, out := lib.EncodeMessage(meta, data)
	if err != nil {
		log.Errorf("wrapped-con@Send - Error occurred: %s", err)
		return err
	}

	wc.Conn.WriteMessage(websocket.BinaryMessage, *out)

	if glob.F_LOG_IO_MSG {
		log.Infof("Message sent! \nMeta: %s \nData: %s", meta, data)
	}

	return nil
}

// Sends a typical "reply" message to current websocket connection.
func (wc *WrappedConn) ReplyMsg(msgID string, code int, data *[]byte, meta j) error {

	metaMap := j{
		"msgID": msgID,
		"code":  code,
	}

	if meta != nil {
		for key, val := range meta {
			metaMap[key] = val
		}
	}

	return wc.Send(metaMap, data)
}
