package setup

import (
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/common/log"

	"github.com/tgbv/swarm-cache/glob"
)

// Iterates os.Args and assigns the required vars and parameters
func ParseFlags() {

	for i, arg := range os.Args {
		if i == 0 {
			continue
		}

		switch true {

		case hasFlag(arg, "--cl"):
			_, vars := splitComm(arg)
			glob.F_CLIENT_LISTEN = (*vars)[0]

		case hasFlag(arg, "--nl"):
			_, vars := splitComm(arg)
			glob.F_NODE_LISTEN = (*vars)[0]

		case hasFlag(arg, "--nc"):
			_, vars := splitComm(arg)
			glob.F_NODE_CONNECT = *vars

		case hasFlag(arg, "--log-io-msg"):
			_, vars := splitComm(arg)

			if state, err := strconv.Atoi((*vars)[0]); err != nil {
				log.Error("Invalid option for --log-io-msg. Accepted: 0,1")
			} else {
				if state > 0 {
					glob.F_LOG_IO_MSG = true
				}
			}

		}

		if glob.F_CLIENT_LISTEN == glob.F_NODE_LISTEN {
			log.Fatal("Client and Node server cannot listen on the same address:port!")
			os.Exit(0)
		}
	}
}

// Checks if arg is a flagName
func hasFlag(arg string, flagName string) bool {
	return strings.Contains(arg, flagName)
}

// Splits a flag from its parameters
func splitComm(flagName string) (string, *[]string) {
	splitted := strings.Split(flagName, "=")
	params := make([]string, 0)

	if len(splitted) > 1 {
		params = strings.Split(splitted[1], ",")
	}

	return splitted[0], &params
}
