package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	utils "github.com/aorith/whoip/internal"
	"github.com/aorith/whoip/pkg/whoip"
)

var (
	showVersion    bool
	showCategories bool
)

func main() {
	flag.BoolVar(&showVersion, "version", false, "show version information and exit")
	flag.BoolVar(&showCategories, "categories", false, "show available categories and exit")
	flag.Parse()

	if showVersion {
		utils.ShowVersion("whoip-cli")
		os.Exit(0)
	}

	if showCategories {
		fmt.Printf("%s\n", whoip.Categories())
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: whoip-cli [IP Address]")
		os.Exit(1)
	}

	ipStr := args[0]
	ip := net.ParseIP(ipStr)
	if ip == nil {
		fmt.Printf("Invalid IP address: %s\n", ipStr)
		os.Exit(1)
	}

	fmt.Printf("%s\n", whoip.FindIP(ip))
}
