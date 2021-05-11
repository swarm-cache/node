package nodes

import (
	"time"

	"github.com/tgbv/swarm-cache/bag"
	"github.com/tgbv/swarm-cache/glob"
	"github.com/tgbv/swarm-cache/lib"
	wconn "github.com/tgbv/swarm-cache/wrapped-conn"
)

// Sets to current bag only if there is enough memory
//
// If there isn't, synchronously attempt to set data to all other connected nodes
// from which message hadn't been received
func commSetStrict(cw *wconn.WrappedConn, meta j, data *[]byte) {

	// set only if there is enough memory in current node
	if err := bag.Set(meta["key"].(string), data); err == nil {
		cw.Send(j{
			"type":   "res",
			"commID": meta["commID"].(string),
			"msgID":  meta["msgID"].(string),
			"code":   glob.RES_SUCCESS,
		}, nil)
		return
	}

	isSet := false
	errorMsg := "Swarm ran out of memory!"

	for _, n := range nodes {
		if n.GetComm(meta["commID"].(string)) != nil {
			continue
		}

		cContinue := make(chan bool)
		fMsgID, _ := lib.GenerateRandomString(glob.MSG_ID_LENGTH)

		n.PushCB(fMsgID, func(cMsgID string, cWc *wconn.WrappedConn, cMeta j, cData *[]byte) {
			if cMsgID == cMeta["msgID"].(string) {
				if cMeta["code"].(float64) == glob.RES_SUCCESS {
					isSet = true
					cContinue <- false
				} else {
					errorMsg = cMeta["message"].(string)
					cContinue <- true
				}

				n.DelCB(fMsgID)
			}
		})

		// send command
		n.Send(j{
			"msgID":  fMsgID,
			"type":   "comm",
			"commID": meta["commID"].(string),
			"comm":   "set-strict",
			"key":    meta["key"].(string),
		}, data)

		// handles timeout from node
		go func() {
			time.Sleep(glob.NODE_TO_NODE_RES_TIMEOUT * time.Millisecond)

			n.DelCB(fMsgID)
			errorMsg = "Internode timeout occurred!"
			cContinue <- true
		}()

		if <-cContinue {
			continue
		} else {
			break
		}
	}

	if isSet {
		cw.Send(j{
			"type":   "res",
			"commID": meta["commID"].(string),
			"msgID":  meta["msgID"].(string),
			"code":   glob.RES_SUCCESS,
		}, nil)
	} else {
		cw.Send(j{
			"type":    "res",
			"commID":  meta["commID"].(string),
			"msgID":   meta["msgID"].(string),
			"code":    glob.RES_ERROR,
			"message": errorMsg,
		}, nil)
	}
}
