package main

import (
	"flag"
	"os"

	utils "github.com/aorith/whoip/internal"
)

var showVersion bool

func main() {
	flag.BoolVar(&showVersion, "version", false, "show version information and exit")
	flag.Parse()

	if showVersion {
		utils.ShowVersion("whoip-server")
		os.Exit(0)
	}
}
