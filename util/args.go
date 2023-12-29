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
			fmt.Printf("Current user: %s, repo: %s, sync: %s.\n", target.User, target.Repo, target.Sync)
			switch target.Sync {
			case SyncLatestRelease:
				err := downloadLatest(target.User, target.Repo, config.ProxyHttp, fmt.Sprintf("%s/%s", "repo", target.Repo))
				if err != nil {
					fmt.Printf("Failed to download latest release: %s.\n", err.Error())
					exitCode = ErrorDownload
					continue
				}
			case SyncLatest:
				releases, _ := GetRelease(target.User, target.Repo, config.ProxyHttp, 1)
				if len(releases) >= 1 {
					err := downloadRelease(&releases[0], config.ProxyHttp, fmt.Sprintf("%s/%s", "repo", target.Repo))
					if err != nil {
						fmt.Printf("Failed to download latest: %s.\n", err.Error())
						exitCode = ErrorDownload
						continue
					}
				} else {
					fmt.Printf("There's nothing to download.")
					exitCode = ErrorDownload
					continue
				}
			case SyncAll:
				currentPage := 1
				for currentPage != -1 {
					fmt.Printf("Current page: %d.\n", currentPage)
					var releases []Release
					releases, currentPage = GetRelease(target.User, target.Repo, config.ProxyHttp, currentPage)
					if len(releases) >= 1 {
						for _, release := range releases {
							err := downloadRelease(&release, config.ProxyHttp, fmt.Sprintf("%s/%s", "repo", target.Repo))
							if err != nil {
								fmt.Printf("Failed to download latest: %s.\n", err.Error())
								exitCode = ErrorDownload
								continue
							}
						}
					} else {
						fmt.Printf("There's nothing to download.")
						exitCode = ErrorDownload
						continue
					}
				}
			default:
				err := downloadByTag(target.User, target.Repo, target.Sync, config.ProxyHttp, fmt.Sprintf("%s/%s", "repo", target.Repo))
				if err != nil {
					fmt.Printf("Failed to download release: %s.\n", err.Error())
					exitCode = ErrorDownload
					continue
				}
			}
		}
		return exitCode
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
	latestRelease := GetLatestRelease(user, repo, proxyHttp)
	err := downloadRelease(latestRelease, proxyHttp, dstDir)
	return err
}

func downloadByTag(user, repo, tag, proxyHttp, dstDir string) error {
	latestRelease := GetReleaseByTag(user, repo, tag, proxyHttp)
	var err error
	if latestRelease != nil {
		err = downloadRelease(latestRelease, proxyHttp, dstDir)
	} else {
		err = fmt.Errorf("failed to get the release by tag: %s", tag)
	}
	return err
}

func downloadRelease(release *Release, proxyHttp, dstDir string) error {
	fmt.Printf("Current release: %s, tag: %s.\n", release.Name, release.TagName)

	err := os.MkdirAll(dstDir, os.ModePerm)
	if err != nil {
		return err
	}

	if release != nil {
		dstTagDir := fmt.Sprintf("%s/%s", dstDir, release.TagName)
		fmt.Printf("Tring to create: %s.\n", dstTagDir)
		err := os.MkdirAll(dstTagDir, os.ModePerm)
		if err != nil {
			return err
		}

		for _, asset := range release.Assets {
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
