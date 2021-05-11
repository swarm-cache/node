package nodes

import (
	"fmt"
	"sync"
	"time"

	"github.com/tgbv/swarm-cache/bag"
	"github.com/tgbv/swarm-cache/glob"
	"github.com/tgbv/swarm-cache/lib"
	wconn "github.com/tgbv/swarm-cache/wrapped-conn"
)

// Broadcasts set-strict command to connected nodes.
// Is ran in case current bag has no more memory
func broadcastSetStrict(cw *wconn.WrappedConn, meta j, data *[]byte) error {
	succeeded := false
	errorMsg := "Swarm ran out of memory!"
	commID, _ := lib.GenerateRandomString(glob.COMM_ID_LENGTH)

	for _, n := range nodes {
		fMsgID, _ := lib.GenerateRandomString(glob.COMM_ID_LENGTH)
		fContinue := make(chan bool)

		n.PushCB(fMsgID, func(cbFName string, cbWC *wconn.WrappedConn, cbMeta j, cbData *[]byte) {
			if cbFName == cbMeta["msgID"].(string) {
				if cbMeta["code"].(float64) == glob.RES_SUCCESS {
					succeeded = true
					fContinue <- false
				} else {
					errorMsg = cbMeta["message"].(string)
					fContinue <- true
				}

				n.DelCB(cbFName)
			}
		})

		// send command
		n.Send(j{
			"msgID":  fMsgID,
			"type":   "comm",
			"commID": commID,
			"comm":   "set-strict",
			"key":    meta["key"].(string),
		}, data)

		// timeout handler
		go func() {
			time.Sleep(glob.NODE_TO_NODE_RES_TIMEOUT * time.Millisecond)

			n.DelCB(fMsgID)
			errorMsg = "Internode timeout occurred!"
			fContinue <- true
		}()

		if <-fContinue {
			continue
		} else {
			break
		}
	}

	if succeeded {
		return nil
	} else {
		return fmt.Errorf(errorMsg)
	}
}

// Attempts to set key in current bag and reply success message if ok
func attemptSet(cw *wconn.WrappedConn, meta j, data *[]byte) {
	err := bag.Set(meta["key"].(string), data)
	if err == nil {
		cw.Send(j{
			"type":  "res",
			"msgID": meta["msgID"].(string),
			"code":  glob.RES_SUCCESS,
		}, nil)
	} else {
		// expand only if allowed to.
		if !glob.F_OVERFLOW_EXPAND {
			cw.Send(j{
				"type":    "res",
				"msgID":   meta["msgID"].(string),
				"commID":  meta["commID"],
				"code":    glob.RES_ERROR,
				"message": err.Error(),
			}, nil)
		}

		if e := broadcastSetStrict(cw, meta, data); e == nil {
			cw.Send(j{
				"type":   "res",
				"msgID":  meta["msgID"].(string),
				"commID": meta["commID"],
				"code":   glob.RES_SUCCESS,
			}, nil)
		} else {
			cw.Send(j{
				"type":    "res",
				"msgID":   meta["msgID"].(string),
				"commID":  meta["commID"],
				"code":    glob.RES_ERROR,
				"message": e.Error(),
			}, nil)
		}
	}
}

// Sets a key only if present in current node.
//
// Otherwise attempts to find the key within other connected nodes. Returns 404 if not found
func commSet(cw *wconn.WrappedConn, meta j, data *[]byte) {
	// Check if key is set in current bag
	if _, exists := bag.Get(meta["key"].(string)); exists {
		attemptSet(cw, meta, data)
		return
	}

	wg := sync.WaitGroup{}
	localMux := sync.Mutex{}
	set := false
	remainingNodes := 0

	// broadcast it to all connected nodes from which command has not been received
	for _, n := range nodes {
		if n.GetComm(meta["commID"].(string)) != nil {
			continue
		}

		wg.Add(1)
		remainingNodes++

		// Create a random message ID
		fMsgID, _ := lib.GenerateRandomString(glob.MSG_ID_LENGTH)

		// First create a listener for all incoming messages from node connection
		n.PushCB(fMsgID, func(cbName string, cbCw *wconn.WrappedConn, cbMeta j, cbData *[]byte) {
			if cbMeta["msgID"] == fMsgID {

				if cbMeta["code"].(float64) == glob.RES_SUCCESS {
					cw.Send(j{
						"commID": meta["commID"],
						"msgID":  meta["msgID"],
						"type":   "res",
						"code":   glob.RES_SUCCESS,
					}, cbData)
					set = true
				}

				wg.Done()
				n.DelCB(cbName)

				localMux.Lock()
				defer localMux.Unlock()

				remainingNodes--
			}
		})

		n.Send(j{
			"msgID":  fMsgID,
			"commID": meta["commID"],
			"type":   "comm",
			"comm":   "set",
			"key":    meta["key"],
		}, data)
	}

	// wait for reply from all nodes
	go func() {
		wg.Wait()

		if !set {
			cw.Send(j{
				"commID": meta["commID"],
				"msgID":  meta["msgID"],
				"type":   "res",
				"code":   glob.RES_NOT_FOUND,
			}, nil)
		}
	}()

	// timeout func
	// it releases the waitgroup if timeout is encountered
	go func() {
		time.Sleep(glob.NODE_TO_NODE_RES_TIMEOUT * time.Millisecond)

		localMux.Lock()
		defer localMux.Unlock()

		if remainingNodes > 0 {
			for remainingNodes > 0 {
				wg.Done()
				remainingNodes--
			}
		}
	}()
}
