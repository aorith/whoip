package utils

import (
	"fmt"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func ShowVersion(progName string) {
	fmt.Printf("%s %s\ncommit %s\nbuilt at %s\n", progName, version, commit, buildDate)
}
