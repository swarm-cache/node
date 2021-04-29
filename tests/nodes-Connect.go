package tests

import (
	"github.com/prometheus/common/log"
	"github.com/tgbv/swarm-cache/nodes"
)

func TestNodesConnect() {
	err := nodes.Connect("ws://127.0.0.1:40000")
	if err != nil {
		log.Errorf("Could not connect to server: %s", err)
	}
}
