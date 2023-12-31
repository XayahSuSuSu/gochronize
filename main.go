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
	fmt.Println("gochronize --config \"example.yml\" --history \"history.yml\"")
	fmt.Println()
	fmt.Println("Available arguments:")
	flag.PrintDefaults()
}

var args util.Args

func init() {
	flag.BoolVar(&args.Help, "help", false, "Print the usage.")
	flag.BoolVar(&args.Version, "version", false, "Print the version.")
	flag.StringVar(&args.Config, "config", "", "The configuration path of yaml file format.")
	flag.StringVar(&args.History, "history", "history.yml", "The history configuration path of yaml file format.")
	flag.Usage = usage
}

func main() {
	flag.Parse()

	util.ParseArgs(args)
}
