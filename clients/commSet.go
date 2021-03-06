package clients

import (
	"fmt"
	"sync"
	"time"

	"github.com/tgbv/swarm-cache/bag"
	"github.com/tgbv/swarm-cache/glob"
	"github.com/tgbv/swarm-cache/lib"
	"github.com/tgbv/swarm-cache/nodes"
	wconn "github.com/tgbv/swarm-cache/wrapped-conn"
)

// Broadcasts set-strict command to connected nodes.
// Is ran in case current bag has no more memory
func broadcastSetStrict(cw *wconn.WrappedConn, meta j, data *[]byte) error {
	commID, _ := lib.GenerateRandomString(glob.COMM_ID_LENGTH)
	succeeded := false
	errorMsg := ""

	for _, n := range nodes.GetNodes() {
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
func attemptSet(cw *wconn.WrappedConn, meta j, data *[]byte, broadcast bool) {
	err := bag.Set(meta["key"].(string), data)
	fmt.Println(err)
	if err == nil {
		cw.ReplyMsg(meta["msgID"].(string), glob.RES_SUCCESS, nil, nil)
	} else {
		// broadcast only if allowed to
		if !broadcast || !glob.F_OVERFLOW_EXPAND {
			cw.Send(j{
				"code":    glob.RES_ERROR,
				"msgID":   meta["msgID"].(string),
				"message": err.Error(),
			}, nil)
			return
		}

		err = broadcastSetStrict(cw, meta, data)
		if err == nil {
			bag.Del(meta["key"].(string))

			cw.Send(j{
				"code":  glob.RES_SUCCESS,
				"msgID": meta["msgID"].(string),
			}, nil)
		} else {
			cw.Send(j{
				"code":    glob.RES_ERROR,
				"msgID":   meta["msgID"].(string),
				"message": err.Error(),
			}, nil)
		}

	}
}

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
		attemptSet(cw, meta, data, true)
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
							if nFMeta["code"].(float64) == glob.RES_SUCCESS {
								cw.ReplyMsg(meta["msgID"].(string), glob.RES_SUCCESS, nil, nil)
							}

							// Due to asynchronism it is possible that one client from another node
							// deletes the key from that node after we send the get command AND BEFORE we send the set command.
							//
							// In which case it means we must set the key in our node bag.
							if nFMeta["code"].(float64) == glob.RES_NOT_FOUND && !replied {
								attemptSet(cw, meta, data, true)
							}

							if nFMeta["code"].(float64) == glob.RES_ERROR && !replied {
								attemptSet(cw, meta, data, false)
							}

							fWc.DelCB(nFName)
							replied = true
						}
					})

					// Send command to set it!
					fWc.Send(j{
						"msgID":  cbMsgID,
						"type":   "comm",
						"commID": cbCommID,
						"comm":   "set",
						"key":    key,
					}, data)

					// timeout handler
					// set the key in our node if no reply in mean time
					go func() {
						time.Sleep(glob.NODE_TO_NODE_RES_TIMEOUT * time.Millisecond)

						if !replied {
							attemptSet(cw, meta, data, true)
						}
					}()
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
			attemptSet(cw, meta, data, true)
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
