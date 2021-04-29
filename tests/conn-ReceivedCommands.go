package tests

import (
	"fmt"

	"github.com/gorilla/websocket"
	wconn "github.com/tgbv/swarm-cache/wrapped-conn"
)

func TestConnReceivedCommands() {
	cw := wconn.Init(&websocket.Conn{})

	cw.PushComm("some ID")

	fmt.Println(cw.ReceivedComms)
	fmt.Println(cw.ReceivedComms["some ID"].Arrived)
}
