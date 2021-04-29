package nodes

import (
	"sync"
	"time"

	"github.com/tgbv/swarm-cache/bag"
	"github.com/tgbv/swarm-cache/glob"
	"github.com/tgbv/swarm-cache/lib"
	wconn "github.com/tgbv/swarm-cache/wrapped-conn"
)

// Sets a key only if present in current node.
//
// Otherwise attempts to find the key within other connected nodes. Returns 404 if not found
func commSet(cw *wconn.WrappedConn, meta j, data *[]byte) {
	// Check if key is set in current bag
	if _, exists := bag.Get(meta["key"].(string)); exists {
		bag.Set(meta["key"].(string), data)
		cw.Send(j{
			"msgID":  meta["msgID"],
			"commID": meta["commID"],
			"type":   "res",
			"code":   glob.RES_SUCCESS,
		}, nil)
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
