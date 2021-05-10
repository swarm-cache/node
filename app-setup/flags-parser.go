package setup

import (
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/prometheus/common/log"

	"github.com/tgbv/swarm-cache/glob"
)

// Iterates env file and os.Args and assigns the required vars and parameters
func ParseFlags() {
	hDir, _ := os.UserConfigDir()
	cfgPath := flag.String("cfg-path", hDir+"/swarm-cache/env.txt", "Path to configuration file.")

	tF_CLIENT_LISTEN := flag.String("cl", "", "Address on which to listen for incoming client connections. If blank it will not listen.")
	tF_NODE_LISTEN := flag.String("nl", "", "Address on which to listen for incoming node connections. If blank it will not listen.")
	tF_NODE_CONNECT := flag.String("nc", "", "Nodes to which this node will connect. Separated by commas (,). If blank it will not connect. ")
	tF_LOG_IO_MSG := flag.Bool("log-io-msg", false, "Log input/output websocket messages. Can be true/false")

	flag.Parse()

	// if there is a config file parse that first
	//

	log.Info("Loading env file...")
	err := godotenv.Load(*cfgPath)
	if err == nil {
		glob.F_CLIENT_LISTEN = os.Getenv("F_CLIENT_LISTEN")
		glob.F_NODE_LISTEN = os.Getenv("F_NODE_LISTEN")
		glob.F_NODE_CONNECT = strings.Split(os.Getenv("F_NODE_CONNECT"), ",")
		glob.F_LOG_IO_MSG, _ = strconv.ParseBool(os.Getenv("F_LOG_IO_MSG"))
		glob.F_MAX_USED_MEMORY, _ = strconv.ParseInt(os.Getenv("F_MAX_USED_MEMORY"), 10, 64)
		glob.F_MAX_CACHED_KEYS, _ = strconv.ParseInt(os.Getenv("F_MAX_CACHED_KEYS"), 10, 64)
		log.Info("OK")
	} else {
		log.Errorf("Could not load env file! %s", err)
	}

	// otherwise parse the flags and assign to glob anyway
	//

	if *tF_CLIENT_LISTEN != "" {
		glob.F_CLIENT_LISTEN = *tF_CLIENT_LISTEN
	}
	if *tF_NODE_LISTEN != "" {
		glob.F_NODE_LISTEN = *tF_NODE_LISTEN
	}
	if *tF_NODE_CONNECT != "" {
		glob.F_NODE_CONNECT = strings.Split(*tF_NODE_CONNECT, ",")
	}
	if *tF_CLIENT_LISTEN != "" {
		glob.F_LOG_IO_MSG = *tF_LOG_IO_MSG
	}

	// other checks
	//

	if glob.F_CLIENT_LISTEN == glob.F_NODE_LISTEN && glob.F_CLIENT_LISTEN != "" {
		log.Fatal("Client and Node servers cannot listen on the same address:port!")
		os.Exit(0)
	}
}
