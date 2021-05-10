package setup

import (
	"strings"
	"sync"

	"github.com/prometheus/common/log"
	"github.com/tgbv/swarm-cache/clients"
	"github.com/tgbv/swarm-cache/glob"
	"github.com/tgbv/swarm-cache/nodes"
)

// Executes commands in liniar manner based on flags setup
func Exec() {
	wg := sync.WaitGroup{}

	// start client server
	if glob.F_CLIENT_LISTEN != "" {
		wg.Add(1)
		go func() {
			clients.StartListening()
			wg.Done()
		}()
	}

	// start node server
	if glob.F_NODE_LISTEN != "" {
		wg.Add(1)
		go func() {
			nodes.StartListening()
			wg.Done()
		}()
	}

	// connect to nodes if specified
	for _, addr := range glob.F_NODE_CONNECT {
		if addr == "" {
			continue
		}
		addr = strings.ReplaceAll(addr, " ", "")

		wg.Add(1)
		go func(_addr string) {
			nodes.Connect(_addr)
			wg.Done()
		}(addr)
	}

	log.Info("Waiting for all routines to end...")
	wg.Wait()
	log.Info("Program exited.")
}
