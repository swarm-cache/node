package main

import (
	setup "github.com/tgbv/swarm-cache/app-setup"
)

func main() {

	setup.ParseCfg()
	setup.ParseFlags()
	setup.Exec()
}
