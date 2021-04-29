package clients

import (
	"sync"
	"time"

	"github.com/tgbv/swarm-cache/bag"
	"github.com/tgbv/swarm-cache/glob"
	"github.com/tgbv/swarm-cache/lib"
	"github.com/tgbv/swarm-cache/nodes"
	wconn "github.com/tgbv/swarm-cache/wrapped-conn"
)

func commDel(cw *wconn.WrappedConn, meta j) {
	key := meta["key"].(string)

	// Check if data exists in current node bag
	if _, exists := bag.Get(key); exists {
		bag.Del(key)
		cw.Send(j{
			"msgID": meta["msgID"],
			"code":  glob.RES_SUCCESS,
		}, nil)
		return
	}

	// Wait group to sync all nodes
	wg := sync.WaitGroup{}
	localMux := sync.Mutex{}
	found := false
	remainingNodes := 0
	commID, _ := lib.GenerateRandomString(glob.COMM_ID_LENGTH)

	// Push deletion signal among all other nodes
	for _, n := range nodes.GetNodes() {
		wg.Add(1)
		remainingNodes++

		fMsgID, _ := lib.GenerateRandomString(glob.MSG_ID_LENGTH)

		// Push callback into queue
		n.PushCB(fMsgID, func(fName string, fWc *wconn.WrappedConn, fMeta j, _ *[]byte) {
			if fName == fMeta["msgID"].(string) {
				localMux.Lock()
				if fMeta["code"].(float64) == glob.RES_SUCCESS && !found {
					cw.ReplyMsg(meta["msgID"].(string), glob.RES_SUCCESS, nil, nil)
					found = true
				}
				localMux.Unlock()

				wg.Done()
				n.DelCB(fName)

				localMux.Lock()
				defer localMux.Unlock()
				remainingNodes--
			}

		})

		// Send command message to node
		n.PushComm(commID)
		n.Send(j{
			"msgID":  fMsgID,
			"type":   "comm",
			"commID": commID,
			"comm":   "del",
			"key":    key,
		}, nil)
	}

	// wait for all responses from all nodes
	go func() {
		wg.Wait()

		if !found {
			cw.ReplyMsg(meta["msgID"].(string), glob.RES_ERROR, nil, nil)
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
