package wconn

import (
	"time"

	"github.com/tgbv/swarm-cache/glob"
)

// A typicall structure containing the time at which command has arrived into node
type comm struct {
	Arrived time.Duration
}

// Pushes a new command ID
func (wc *WrappedConn) PushComm(id string) {
	wc.MuxRC.Lock()
	defer wc.MuxRC.Unlock()

	wc.ReceivedComms[id] = &comm{time.Duration(time.Now().UTC().UnixNano())}
}

// Delete command
func (wc *WrappedConn) DelComm(id string) {
	wc.MuxRC.Lock()
	defer wc.MuxRC.Unlock()

	delete(wc.ReceivedComms, id)
}

// Traverses commands with callback
func (wc *WrappedConn) Traverse(f func(string, *comm)) {
	wc.MuxRC.Lock()
	defer wc.MuxRC.Unlock()

	for cid, meta := range wc.ReceivedComms {
		go f(cid, meta)
	}
}

// Retrieve command received by node by id
func (wc *WrappedConn) GetComm(id string) *comm {
	wc.MuxRC.Lock()
	defer wc.MuxRC.Unlock()

	return wc.ReceivedComms[id]
}

// Retrieve all comms mutex protected
func (wc *WrappedConn) GetComms() map[string]*comm {
	wc.MuxRC.Lock()
	defer wc.MuxRC.Unlock()

	return wc.ReceivedComms
}

// Starts the cleanup
//
// Removes commands after certain milliseconds
func (wc *WrappedConn) RCCleanup() {

	for i, c := range wc.GetComms() {
		if c.Arrived+time.Duration(glob.F_CONN_RC_CLEANUP_INTERVAL) <= time.Duration(time.Now().UnixNano()) {
			wc.DelComm(i)
		}
	}

	time.Sleep(glob.F_CONN_RC_CLEANUP_INTERVAL * time.Millisecond)
	go wc.RCCleanup()
}
