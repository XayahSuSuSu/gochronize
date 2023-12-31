package util

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
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

			switch target.Sync {
			case SyncLatestRelease:
				err := syncLatestRelease(httpClient, &target, config)
				if err != nil {
					exitCode = ErrorDownload
					continue
				}
			case SyncLatest:
				err := syncLatest(httpClient, &target, config)
				if err != nil {
					exitCode = ErrorDownload
					continue
				}
			case SyncAll:
				err := syncAll(httpClient, &target, config)
				if err != nil {
					exitCode = ErrorDownload
					continue
				}
			default:
				err := syncByTag(httpClient, &target, config)
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

func syncLatestRelease(client *http.Client, target *Target, config *Config) error {
	latestRelease := GetLatestRelease(client, target.User, target.Repo)
	err := downloadRelease(client, latestRelease, target, config)
	return err
}

func syncLatest(client *http.Client, target *Target, config *Config) error {
	releases, _ := GetRelease(client, target.User, target.Repo, 1)
	if len(releases) >= 1 {
		err := downloadRelease(client, &releases[0], target, config)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("* err: There's nothing to download.\n")
		return fmt.Errorf("")
	}
	return nil
}

func syncAll(client *http.Client, target *Target, config *Config) error {
	var mErr error = nil
	currentPage := 1
	for currentPage != -1 {
		fmt.Printf("* page: %d\n", currentPage)
		var releases []Release
		releases, currentPage = GetRelease(client, target.User, target.Repo, currentPage)
		if len(releases) >= 1 {
			for _, release := range releases {
				err := downloadRelease(client, &release, target, config)
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

func syncByTag(client *http.Client, target *Target, config *Config) error {
	latestRelease := GetReleaseByTag(client, target.User, target.Repo, target.Sync)
	var err error
	if latestRelease != nil {
		err = downloadRelease(client, latestRelease, target, config)
	} else {
		err = fmt.Errorf("failed to get the release by tag: %s", target.Sync)
	}
	return err
}

func downloadRelease(client *http.Client, release *Release, target *Target, config *Config) error {
	fmt.Printf("********************************************\n")
	fmt.Printf("* release: %s\n", release.Name)
	fmt.Printf("* tag: %s\n", release.TagName)
	fmt.Printf("* exclusion: %s\n", strings.Join(target.Exclusion, ", "))
	println("*")

	if release != nil {
		parentDir := target.ParentDir
		if parentDir == "" {
			parentDir = fmt.Sprintf("./repos/%s/%s", RepoName, TagName)
		}
		parentDir = strings.ReplaceAll(parentDir, RepoName, target.Repo)
		parentDir = strings.ReplaceAll(parentDir, TagName, release.TagName)
		parentDir = strings.TrimSuffix(parentDir, "/")

		fmt.Printf("* info: Tring to create: %s.\n", parentDir)
		err := os.MkdirAll(parentDir, os.ModePerm)
		if err != nil {
			return err
		}

		for _, asset := range release.Assets {
			url := asset.BrowserDownloadURL
			name := asset.Name
			fileName := target.FileName
			matchedVar, matchedStr, err := MatchCustomRegex(FileName, fileName, name)
			if err == nil {
				fileName = strings.ReplaceAll(fileName, matchedVar, matchedStr)
			}
			if fileName == "" {
				fileName = FileName
			}
			fileName = strings.ReplaceAll(fileName, FileName, name)
			createdAt, err := time.Parse(time.RFC3339, asset.CreatedAt)
			if err != nil {
				fmt.Printf("* err: Failed to parse date: %s, %s.\n", asset.CreatedAt, err.Error())
			}
			updatedAt, err := time.Parse(time.RFC3339, asset.UpdatedAt)
			if err != nil {
				fmt.Printf("* err: Failed to parse date: %s, %s.\n", asset.CreatedAt, err.Error())
			}
			fileName = strings.ReplaceAll(fileName, CreatedAt, createdAt.Format(config.TimeFormat))
			fileName = strings.ReplaceAll(fileName, UpdatedAt, updatedAt.Format(config.TimeFormat))

			skip := false
			for _, s := range target.Exclusion {
				matched, err := MatchString(name, s)
				if err != nil {
					fmt.Printf("* err: Failed to match: %s.\n", err.Error())
				}
				if matched {
					fmt.Printf("* info: \"%s\" Matched: \"%s\", skip.\n", s, name)
					skip = true
				}
			}
			if skip {
				continue
			}

			count := config.Retries
			for count > 0 {
				fmt.Printf("* info: Download: %s.\n", name)
				err := Download(client, url, fmt.Sprintf("%s/%s", parentDir, fileName))
				if err != nil {
					fmt.Println(err)
					fmt.Printf("* info: Retry: %d\n", config.Retries-count+1)
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
