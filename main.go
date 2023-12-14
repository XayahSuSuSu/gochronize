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
	help      bool
	version   bool
	user      string
	repo      string
	proxyHttp string
)

func init() {
	flag.BoolVar(&help, "help", false, "Print the usage.")
	flag.BoolVar(&version, "version", false, "Print the version.")
	flag.StringVar(&user, "user", "", "The target user to get synchronized.")
	flag.StringVar(&repo, "repo", "", "The target repo name to get synchronized.")
	flag.StringVar(&proxyHttp, "proxy-http", "", "The http proxy url to be used.")
	flag.Usage = usage
}

func main() {
	flag.Parse()

	if user != "" && repo != "" {
		release := util.GetRelease(user, repo, proxyHttp)
		if release != nil {
			for _, asset := range release.Assets {
				url := asset.BrowserDownloadURL
				name := asset.Name
				fmt.Printf("Download: %s\n", name)
				err := util.Download(url, name, proxyHttp)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println("")
			}
		}
	}

	if help {
		flag.Usage()
	} else if version {
		fmt.Printf("Gochronize version: %s\n", util.Version)
	}
}
