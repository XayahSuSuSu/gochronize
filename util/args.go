package util

import (
	"flag"
	"fmt"
	"os"
)

func ParseArgs(args Args) int {
	if args.Config != "" {
		// Config mode

		// Parse config
		config, err := ReadFromConfig(args.Config)
		if err != nil {
			fmt.Printf("Failed to read from config: %s.\n", err.Error())
			return ErrorIo
		}

		// Download for each config
		exitCode := Success
		for _, target := range config.Targets {
			err := downloadLatest(target.User, target.Repo, config.ProxyHttp, fmt.Sprintf("%s/%s", "repo", target.Repo))
			if err != nil {
				fmt.Printf("Failed to downloadLatest: %s.\n", err.Error())
				exitCode = ErrorDownload
				continue
			}
		}
		return exitCode
	} else if args.User != "" && args.Repo != "" {
		// Cmd mode

		// Download directly
		err := downloadLatest(args.User, args.Repo, args.ProxyHttp, fmt.Sprintf("%s/%s", "repo", args.Repo))
		if err != nil {
			fmt.Printf("Failed to downloadLatest: %s.\n", err.Error())
			return ErrorDownload
		}
	} else if args.Version {
		// Print the version

		fmt.Printf("Gochronize version: %s.\n", Version)
	} else if args.Help {
		// Print the usage

		flag.Usage()
	} else {
		// Unknown cmd

		flag.Usage()
		return ErrorUnknownCmd
	}

	return Success
}

func downloadLatest(user, repo, proxyHttp, dstDir string) error {
	err := os.MkdirAll(dstDir, os.ModePerm)
	if err != nil {
		return err
	}

	latestRelease := GetLatestRelease(user, repo, proxyHttp)
	if latestRelease != nil {
		dstTagDir := fmt.Sprintf("%s/%s", dstDir, latestRelease.TagName)
		err := os.MkdirAll(dstTagDir, os.ModePerm)
		if err != nil {
			return err
		}

		for _, asset := range latestRelease.Assets {
			url := asset.BrowserDownloadURL
			name := asset.Name
			fmt.Printf("Download: %s.\n", name)
			err := Download(url, fmt.Sprintf("%s/%s", dstTagDir, name), proxyHttp)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println()
		}
	} else {
		return fmt.Errorf("failed to get the latest release")
	}

	return nil
}
