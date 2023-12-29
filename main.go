package main

import (
	"flag"
	"fmt"
	"github.com/XayahSuSuSu/gochronize/util"
	"os"
)

func usage() {
	fmt.Println("Gochronize is a tool for synchronizing releases from GitHub with local.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("gochronize --config \"example.yml\"")
	fmt.Println()
	fmt.Println("Available arguments:")
	flag.PrintDefaults()
}

var args util.Args

func init() {
	flag.BoolVar(&args.Help, "help", false, "Print the usage.")
	flag.BoolVar(&args.Version, "version", false, "Print the version.")
	flag.StringVar(&args.Config, "config", "", "The configuration path of yaml file format.")
	flag.Usage = usage
}

func main() {
	flag.Parse()

	exitCode := util.ParseArgs(args)
	os.Exit(exitCode)
}
