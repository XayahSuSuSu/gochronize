package main

import (
	"flag"
	"fmt"
	"github.com/XayahSuSuSu/gochronize/util"
)

func usage() {
	fmt.Println("Gochronize is a tool for synchronizing releases from GitHub with local.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println()
	fmt.Println("Available arguments:")
	flag.PrintDefaults()
}

var (
	help    bool
	version bool
	user    string
	repo    string
)

func init() {
	flag.BoolVar(&help, "help", false, "Print the usage.")
	flag.BoolVar(&version, "version", false, "Print the version.")
	flag.StringVar(&user, "user", "", "The target user to get synchronized.")
	flag.StringVar(&repo, "repo", "", "The target repo name to get synchronized.")
	flag.Usage = usage
}

func main() {
	flag.Parse()

	if help {
		flag.Usage()
	} else if version {
		fmt.Printf("Gochronize version: %s\n", util.Version)
	}
}
