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

// Sets data by key
//
// 1) check if key exists in current bag
//
// 2) if yes, set it here and return
//
// 3) if not check if key exists anywere else in network
//
// 4) if yes, set it there
//
// 5) if not set it in current bag
func commSet(cw *wconn.WrappedConn, meta j, data *[]byte) {
	key := meta["key"].(string)

	// Check if data exists in current node bag.
	// Set it if so.
	if _, i := bag.Get(key); i {
		bag.Set(key, data)
		cw.ReplyMsg(meta["msgID"].(string), glob.RES_SUCCESS, nil, nil)
		return
	}

	// Wait group to sync all nodes
	wg := sync.WaitGroup{}
	localMux := sync.Mutex{}
	found := false
	remainingNodes := 0
	commID, _ := lib.GenerateRandomString(glob.COMM_ID_LENGTH)

	// check if it exists among other nodes bags
	for _, n := range nodes.GetNodes() {
		wg.Add(1)
		remainingNodes++

		fMsgID, _ := lib.GenerateRandomString(glob.MSG_ID_LENGTH)

		// Push callback into queue
		n.PushCB(fMsgID, func(fName string, fWc *wconn.WrappedConn, fMeta j, fData *[]byte) {

			if fName == fMeta["msgID"].(string) {

				localMux.Lock()
				if fMeta["code"].(float64) == glob.RES_SUCCESS && !found {
					found = true

					// Generate new IDs
					cbMsgID, _ := lib.GenerateRandomString(glob.MSG_ID_LENGTH)
					cbCommID, _ := lib.GenerateRandomString(glob.COMM_ID_LENGTH)

					// in case it's ok
					replied := false

					// Push waiter callback
					fWc.PushCB(cbMsgID, func(nFName string, nFWc *wconn.WrappedConn, nFMeta j, nFData *[]byte) {
						if nFName == nFMeta["msgID"].(string) {
							if fMeta["code"].(float64) == glob.RES_SUCCESS {
								cw.ReplyMsg(meta["msgID"].(string), glob.RES_SUCCESS, nil, nil)
							}

							// Due to asynchronism it is possible that one client from another node
							// deletes the key from that node after we send the get command AND BEFORE we send the set command.
							//
							// In which case it means we must set the key in our node bag.
							if fMeta["code"].(float64) == glob.RES_NOT_FOUND && !replied {
								bag.Set(key, data)

								cw.ReplyMsg(meta["msgID"].(string), glob.RES_SUCCESS, nil, nil)
							}

							fWc.DelCB(nFName)
							replied = true
						}
					})

					// timeout handler
					// set the key in our node if no reply in mean time
					go func() {
						time.Sleep(glob.NODE_TO_NODE_RES_TIMEOUT * time.Millisecond)

						if !replied {
							bag.Set(key, data)
							cw.ReplyMsg(meta["msgID"].(string), glob.RES_SUCCESS, nil, nil)
						}
					}()

					// Send command to set it!
					fWc.Send(j{
						"msgID":  cbMsgID,
						"type":   "comm",
						"commID": cbCommID,
						"comm":   "set",
						"key":    key,
					}, data)
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
			"comm":   "get",
			"key":    key,
		}, nil)
	}

	// wait for all responses from all nodes
	go func() {
		wg.Wait()

		if !found {
			bag.Set(key, data)
			cw.ReplyMsg(meta["msgID"].(string), glob.RES_SUCCESS, nil, nil)
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
