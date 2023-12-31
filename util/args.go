package util

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

func ParseArgs(args Args) {
	if args.Config != "" {
		// Config mode

		// Parse config
		config := ReadFromConfig(args.Config)

		// Get http client
		httpClient := GetHttpClient(config.ProxyHttp, config.Timeout)

		// Download for each config
		exitCode := Success
		for _, target := range config.Targets {
			Printfln("********************************************")
			Printfln("* user: %s", target.User)
			Printfln("* repo: %s", target.Repo)
			Printfln("* sync: %s", target.Sync)
			Printfln("********************************************")

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
		os.Exit(exitCode)
	} else if args.Version {
		// Print the version

		Printfln("Gochronize version: %s.", Version)
	} else if args.Help {
		// Print the usage

		flag.Usage()
	} else {
		// Unknown cmd

		flag.Usage()
		os.Exit(ErrorUnknownCmd)
	}

	os.Exit(Success)
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
		Fprintfln("* err: There's nothing to download.")
		return fmt.Errorf("")
	}
	return nil
}

func syncAll(client *http.Client, target *Target, config *Config) error {
	var mErr error = nil
	currentPage := 1
	for currentPage != -1 {
		Printfln("* page: %d", currentPage)
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
			Fprintfln("* err: There's nothing to download.")
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
	Printfln("********************************************")
	Printfln("* release: %s", release.Name)
	Printfln("* tag: %s", release.TagName)
	Printfln("* exclusion: [%s]", strings.Join(target.Exclusion, ", "))
	Printfln("*")

	if release != nil {
		parentDir := target.ParentDir
		if parentDir == "" {
			parentDir = fmt.Sprintf("./repos/%s/%s", RepoName, TagName)
		}
		parentDir = strings.ReplaceAll(parentDir, RepoName, target.Repo)
		parentDir = strings.ReplaceAll(parentDir, TagName, release.TagName)
		parentDir = strings.TrimSuffix(parentDir, "/")

		Printfln("* info: Trying to create: %s.", parentDir)
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
				Fprintfln("* err: Failed to parse date: %s, %v", asset.CreatedAt, err)
			}
			updatedAt, err := time.Parse(time.RFC3339, asset.UpdatedAt)
			if err != nil {
				Fprintfln("* err: Failed to parse date: %s, %v", asset.CreatedAt, err)
			}
			fileName = strings.ReplaceAll(fileName, CreatedAt, createdAt.Format(config.TimeFormat))
			fileName = strings.ReplaceAll(fileName, UpdatedAt, updatedAt.Format(config.TimeFormat))

			skip := false
			for _, s := range target.Exclusion {
				matched, err := MatchString(name, s)
				if err != nil {
					Fprintfln("* err: Failed to match: %v", err)
				}
				if matched {
					Printfln("* info: \"%s\" Matched: \"%s\", skip.", s, name)
					skip = true
				}
			}
			if skip {
				continue
			}

			count := config.Retries
			for count > 0 {
				Printfln("* info: Download: %s.", name)
				err := Download(client, url, fmt.Sprintf("%s/%s", parentDir, fileName))
				if err != nil {
					Fprintfln("%v", err)
					Printfln("* info: Retry: %d\n", config.Retries-count+1)
					count--
				} else {
					break
				}
			}

			Printfln("*")
		}
	} else {
		return fmt.Errorf("failed to get the latest release")
	}

	Printfln("********************************************")
	return nil
}
