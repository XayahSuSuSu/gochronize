package util

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

func ParseArgs(args Args) int {
	if args.Config != "" {
		// Config mode

		// Parse config
		config, err := ReadFromConfig(args.Config)
		if err != nil {
			return ErrorIO
		}

		// Get http client
		httpClient, err := GetHttpClient(config.ProxyHttp, config.Timeout)
		if err != nil {
			return Error
		}

		// Download for each config
		exitCode := Success
		for _, target := range config.Targets {
			fmt.Printf("********************************************\n")
			fmt.Printf("* user: %s\n", target.User)
			fmt.Printf("* repo: %s\n", target.Repo)
			fmt.Printf("* sync: %s\n", target.Sync)
			fmt.Printf("********************************************\n")

			// The target local path to get sync with
			repoDir := target.Repo
			rootDir := target.RootDir
			if target.RepoDir != "" {
				repoDir = target.RepoDir
			}
			if rootDir == "" {
				rootDir = "."
			}
			dstDir := fmt.Sprintf("%s/%s", rootDir, repoDir)

			switch target.Sync {
			case SyncLatestRelease:
				err := syncLatestRelease(httpClient, target.User, target.Repo, dstDir, config.Retries)
				if err != nil {
					exitCode = ErrorDownload
					continue
				}
			case SyncLatest:
				err := syncLatest(httpClient, target.User, target.Repo, dstDir, config.Retries)
				if err != nil {
					exitCode = ErrorDownload
					continue
				}
			case SyncAll:
				err := syncAll(httpClient, target.User, target.Repo, dstDir, config.Retries)
				if err != nil {
					exitCode = ErrorDownload
					continue
				}
			default:
				err := syncByTag(httpClient, target.User, target.Repo, target.Sync, dstDir, config.Retries)
				if err != nil {
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

func syncLatestRelease(client *http.Client, user, repo, dstDir string, retries int) error {
	latestRelease := GetLatestRelease(client, user, repo)
	err := downloadRelease(client, latestRelease, dstDir, retries)
	return err
}

func syncLatest(client *http.Client, user, repo, dstDir string, retries int) error {
	releases, _ := GetRelease(client, user, repo, 1)
	if len(releases) >= 1 {
		err := downloadRelease(client, &releases[0], dstDir, retries)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("* err: There's nothing to download.\n")
		return fmt.Errorf("")
	}
	return nil
}

func syncAll(client *http.Client, user, repo, dstDir string, retries int) error {
	var mErr error = nil
	currentPage := 1
	for currentPage != -1 {
		fmt.Printf("* page: %d\n", currentPage)
		var releases []Release
		releases, currentPage = GetRelease(client, user, repo, currentPage)
		if len(releases) >= 1 {
			for _, release := range releases {
				err := downloadRelease(client, &release, dstDir, retries)
				if err != nil {
					mErr = err
				}
			}
		} else {
			fmt.Printf("* err: There's nothing to download.\n")
			mErr = fmt.Errorf("")
		}
	}
	return mErr
}

func syncByTag(client *http.Client, user, repo, tag, dstDir string, retries int) error {
	latestRelease := GetReleaseByTag(client, user, repo, tag)
	var err error
	if latestRelease != nil {
		err = downloadRelease(client, latestRelease, dstDir, retries)
	} else {
		err = fmt.Errorf("failed to get the release by tag: %s", tag)
	}
	return err
}

func downloadRelease(client *http.Client, release *Release, dstDir string, retries int) error {
	fmt.Printf("********************************************\n")
	fmt.Printf("* release: %s\n", release.Name)
	fmt.Printf("* tag: %s\n", release.TagName)
	println("*")

	err := os.MkdirAll(dstDir, os.ModePerm)
	if err != nil {
		return err
	}

	if release != nil {
		dstTagDir := fmt.Sprintf("%s/%s", dstDir, release.TagName)
		fmt.Printf("* info: Tring to create: %s.\n", dstTagDir)
		err := os.MkdirAll(dstTagDir, os.ModePerm)
		if err != nil {
			return err
		}

		for _, asset := range release.Assets {
			url := asset.BrowserDownloadURL
			name := asset.Name

			count := retries
			for count > 0 {
				fmt.Printf("* info: Download: %s.\n", name)
				err := Download(client, url, fmt.Sprintf("%s/%s", dstTagDir, name))
				if err != nil {
					fmt.Println(err)
					fmt.Printf("* info: Retry: %d\n", retries-count+1)
					count--
				} else {
					break
				}
			}

			fmt.Println("*")
		}
	} else {
		return fmt.Errorf("failed to get the latest release")
	}

	fmt.Printf("********************************************\n")
	return nil
}
