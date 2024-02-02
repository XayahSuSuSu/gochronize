package util

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	history   *History
	repoIndex = -1
)

func ParseArgs(args Args) {
	if args.Config != "" {
		// Config mode

		// Parse config
		config := ReadFromConfig(args.Config)
		history = ReadFromHistory(args.History)

		// Get http client
		httpClient := GetHttpClient(config.ProxyHttp, config.Token, config.Timeout)

		// Download for each config
		exitCode := Success
		for _, target := range config.Targets {
			if target.Url != "" {
				urlSplit := strings.Split(target.Url, "/")
				if len(urlSplit) > 2 {
					target.Repo = strings.TrimSuffix(urlSplit[len(urlSplit)-1], ".git")
					target.User = urlSplit[len(urlSplit)-2]
				} else {
					Fprintfln("Failed to parse url: %s", target.Url)
					continue
				}
			}

			Printfln("********************************************")
			Printfln("* user: %s", target.User)
			Printfln("* repo: %s", target.Repo)
			Printfln("* sync: %s", target.Sync)
			Printfln("********************************************")

			repoIndex = -1
			for i, r := range history.Repos {
				if r.User == target.User && r.Repo == target.Repo {
					repoIndex = i
				}
			}
			if repoIndex == -1 {
				repoIndex = len(history.Repos)
				history.Repos = append(history.Repos, HistoryRepo{
					User: target.User,
					Repo: target.Repo,
				})
			}

			switch target.Sync {
			case SyncLatestRelease:
				err := syncLatestRelease(httpClient, &target, config, &args)
				if err != nil {
					exitCode = ErrorDownload
					continue
				}
			case SyncLatestPrerelease:
				err := syncLatestPrerelease(httpClient, &target, config, &args)
				if err != nil {
					exitCode = ErrorDownload
					continue
				}
			case SyncLatest:
				err := syncLatest(httpClient, &target, config, &args)
				if err != nil {
					exitCode = ErrorDownload
					continue
				}
			case SyncFromLatestLocal:
				err := syncFromLatestLocal(httpClient, &target, config, &args)
				if err != nil {
					exitCode = ErrorDownload
					continue
				}
			case SyncAll:
				err := syncAll(httpClient, &target, config, &args)
				if err != nil {
					exitCode = ErrorDownload
					continue
				}
			default:
				err := syncByTag(httpClient, &target, config, &args)
				if err != nil {
					exitCode = ErrorDownload
					continue
				}
			}
		}

		err := SaveHistoryToYaml(args.History, history)
		if err != nil {
			Fprintfln("Failed to save history config, %v", err)
			os.Exit(ErrorIO)
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

func syncLatestRelease(client *http.Client, target *Target, config *Config, args *Args) error {
	latestRelease := GetLatestRelease(client, target.User, target.Repo)
	err := downloadRelease(client, latestRelease, target, config, args)
	return err
}

func syncLatestPrerelease(client *http.Client, target *Target, config *Config, args *Args) error {
	var prerelease *Release = nil
	currentPage := 1
	for currentPage != -1 {
		isSuccess := false
		Printfln("* page: %d", currentPage)
		var releases []Release
		releases, currentPage = GetRelease(client, target.User, target.Repo, currentPage)
		if len(releases) >= 1 {
			for _, release := range releases {
				if release.Prerelease {
					prerelease = &release
					isSuccess = true
					break
				}
			}
		} else {
			Fprintfln("* err: There's nothing to download.")
			return fmt.Errorf("")
		}
		if isSuccess {
			break
		}
	}

	if prerelease != nil {
		err := downloadRelease(client, prerelease, target, config, args)
		if err != nil {
			return err
		}
	} else {
		Fprintfln("* err: There's no any prerelease to download.")
		return fmt.Errorf("")
	}
	return nil
}

func syncLatest(client *http.Client, target *Target, config *Config, args *Args) error {
	releases, _ := GetRelease(client, target.User, target.Repo, 1)
	if len(releases) >= 1 {
		err := downloadRelease(client, &releases[0], target, config, args)
		if err != nil {
			return err
		}
	} else {
		Fprintfln("* err: There's nothing to download.")
		return fmt.Errorf("")
	}
	return nil
}

func syncFromLatestLocal(client *http.Client, target *Target, config *Config, args *Args) error {
	var mErr error = nil
	currentPage := 1
	for currentPage != -1 {
		Printfln("* page: %d", currentPage)
		var releases []Release
		releases, currentPage = GetRelease(client, target.User, target.Repo, currentPage)
		if len(releases) >= 1 {
			var localRepo *HistoryRepo = nil
			for _, r := range history.Repos {
				if r.User == target.User && r.Repo == target.Repo {
					localRepo = &r
				}
			}
			for _, release := range releases {
				var latestReleaseId int64 = -1
				if localRepo != nil {
					for _, r := range localRepo.Releases {
						if latestReleaseId < r.Id {
							latestReleaseId = r.Id
						}
					}
				}

				if latestReleaseId != release.Id {
					Printfln("* info: This release is newer than latest local release.")
					Printfln("* info: Current release id: %d.", release.Id)
					Printfln("* info: The latest local release id: %d.", latestReleaseId)
					Printfln("* info: -1 means that there's no latest local release.")
					err := downloadRelease(client, &release, target, config, args)
					if err != nil {
						mErr = err
					}
				} else {
					return mErr
				}
			}
		} else {
			Fprintfln("* err: There's nothing to download.")
			mErr = fmt.Errorf("")
		}
	}
	return mErr
}

func syncAll(client *http.Client, target *Target, config *Config, args *Args) error {
	var mErr error = nil
	currentPage := 1
	for currentPage != -1 {
		Printfln("* page: %d", currentPage)
		var releases []Release
		releases, currentPage = GetRelease(client, target.User, target.Repo, currentPage)
		if len(releases) >= 1 {
			for _, release := range releases {
				err := downloadRelease(client, &release, target, config, args)
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

func syncByTag(client *http.Client, target *Target, config *Config, args *Args) error {
	latestRelease := GetReleaseByTag(client, target.User, target.Repo, target.Sync)
	var err error
	if latestRelease != nil {
		err = downloadRelease(client, latestRelease, target, config, args)
	} else {
		err = fmt.Errorf("failed to get the release by tag: %s", target.Sync)
	}
	return err
}

func handleVars(old, fileName, repoName, tagName, releaseName, createdAtStr, updatedAtStr, timeFormat string) string {
	str := strings.ReplaceAll(old, FileName, fileName)
	str = strings.ReplaceAll(str, RepoName, repoName)
	str = strings.ReplaceAll(str, TagName, tagName)
	str = strings.ReplaceAll(str, ReleaseName, releaseName)
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		Fprintfln("* err: Failed to parse date: %s, %v", createdAtStr, err)
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		Fprintfln("* err: Failed to parse date: %s, %v", updatedAtStr, err)
	}
	str = strings.ReplaceAll(str, CreatedAt, createdAt.Format(timeFormat))
	str = strings.ReplaceAll(str, UpdatedAt, updatedAt.Format(timeFormat))
	matchedVar, matchedStr, err := MatchCustomRegex(FileName, str, fileName)
	if err == nil {
		str = strings.ReplaceAll(str, matchedVar, matchedStr)
	}
	matchedVar, matchedStr, err = MatchCustomRegex(RepoName, str, fileName)
	if err == nil {
		str = strings.ReplaceAll(str, matchedVar, matchedStr)
	}
	matchedVar, matchedStr, err = MatchCustomRegex(TagName, str, fileName)
	if err == nil {
		str = strings.ReplaceAll(str, matchedVar, matchedStr)
	}
	matchedVar, matchedStr, err = MatchCustomRegex(ReleaseName, str, fileName)
	if err == nil {
		str = strings.ReplaceAll(str, matchedVar, matchedStr)
	}
	matchedVar, matchedStr, err = MatchCustomRegex(CreatedAt, str, createdAt.Format(timeFormat))
	if err == nil {
		str = strings.ReplaceAll(str, matchedVar, matchedStr)
	}
	matchedVar, matchedStr, err = MatchCustomRegex(UpdatedAt, str, updatedAt.Format(timeFormat))
	if err == nil {
		str = strings.ReplaceAll(str, matchedVar, matchedStr)
	}
	return str
}

func downloadRelease(client *http.Client, release *Release, target *Target, config *Config, args *Args) error {
	Printfln("********************************************")
	Printfln("* release: %s", release.Name)
	Printfln("* tag: %s", release.TagName)
	Printfln("* exclusion: [%s]", strings.Join(target.Exclusion, ", "))

	historyRelease := HistoryRelease{Name: release.Name, TagName: release.TagName}
	historyReleaseIndex := -1
	for i, r := range history.Repos[repoIndex].Releases {
		if r.Name == release.Name && r.TagName == release.TagName {
			historyRelease = r
			historyReleaseIndex = i
		}
	}
	historyRelease.Id = release.Id
	historyRelease.Prerelease = release.Prerelease
	historyRelease.CreatedAt = release.CreatedAt
	historyRelease.PublishedAt = release.PublishedAt

	if release != nil {
		for _, asset := range release.Assets {
			url := asset.BrowserDownloadURL
			name := asset.Name
			fileName := target.FileName
			parentDir := target.ParentDir

			if parentDir == "" {
				parentDir = fmt.Sprintf("./repos/%s/%s", RepoName, TagName)
			}
			parentDir = handleVars(parentDir, name, target.Repo, release.TagName, release.Name, asset.CreatedAt, asset.UpdatedAt, config.TimeFormat)
			parentDir = strings.TrimSuffix(parentDir, "/")

			Printfln("* info: Trying to create: %s.", parentDir)
			err := os.MkdirAll(parentDir, os.ModePerm)
			if err != nil {
				return err
			}

			if fileName == "" {
				fileName = FileName
			}
			fileName = handleVars(fileName, name, target.Repo, release.TagName, release.Name, asset.CreatedAt, asset.UpdatedAt, config.TimeFormat)

			skip := false
			for _, s := range target.Exclusion {
				matched, err := MatchString(name, s)
				if err != nil {
					Printfln("* info: \"%s\" Not matched: \"%s\", continue.", s, name)
				}
				if matched {
					Printfln("* info: \"%s\" Matched: \"%s\", skip.", s, name)
					skip = true
				}
			}
			if skip {
				continue
			}

			historyAsset := HistoryAsset{
				Name:               asset.Name,
				BrowserDownloadURL: asset.BrowserDownloadURL,
				CreatedAt:          asset.CreatedAt,
				UpdatedAt:          asset.UpdatedAt,
				ParentDir:          parentDir,
				FileName:           fileName,
			}
			historyAssetIndex := -1
			for i, a := range historyRelease.Assets {
				if a.Name == asset.Name && a.BrowserDownloadURL == asset.BrowserDownloadURL && a.ParentDir == parentDir && a.FileName == fileName {
					historyAsset = a
					historyAssetIndex = i
				}
			}

			if historyAssetIndex != -1 && (historyAsset.CreatedAt != asset.CreatedAt || historyAsset.UpdatedAt != asset.UpdatedAt) {
				Fprintfln("%s has new update!", historyAsset.Name)
				Fprintfln("Old: name: %s, url: %s, created at: %s, updated at: %s, parent dir: %s, file name: %s.", historyAsset.Name, historyAsset.BrowserDownloadURL, historyAsset.CreatedAt, historyAsset.UpdatedAt, historyAsset.ParentDir, historyAsset.FileName)
				Fprintfln("New: name: %s, url: %s, created at: %s, updated at: %s, parent dir: %s, file name: %s.", asset.Name, asset.BrowserDownloadURL, asset.CreatedAt, asset.UpdatedAt, parentDir, fileName)
				historyAsset.CreatedAt = asset.CreatedAt
				historyAsset.UpdatedAt = asset.UpdatedAt
			}

			if historyAssetIndex == -1 {
				historyRelease.Assets = append(historyRelease.Assets, historyAsset)
			} else {
				historyRelease.Assets[historyAssetIndex] = historyAsset
				Printfln("%s has already been in history config.", historyAsset.Name)
				if !target.Overwrite {
					continue
				}
			}

			if !args.DryRun {
				count := config.Retries
				for count > 0 {
					dst := fmt.Sprintf("%s/%s", parentDir, fileName)
					Printfln("* info: Download: %s to %s.", name, dst)
					err := Download(client, url, dst)
					if err != nil {
						Fprintfln("%v", err)
						Printfln("* info: Retry: %d", config.Retries-count+1)
						count--
						if count == 0 {
							Fprintfln("* err: Failed to download %s within %d times, trying to delete tmp file: %s.", name, config.Retries, dst)
							err := os.Remove(dst)
							if err != nil {
								Fprintfln("* err: Failed delete: %s.", name)
							}
						}
					} else {
						break
					}
				}
			} else {
				Printfln("* info: Dry-run is enabled and skip download.")
			}
		}

		if historyReleaseIndex == -1 {
			history.Repos[repoIndex].Releases = append(history.Repos[repoIndex].Releases, historyRelease)
		} else {
			history.Repos[repoIndex].Releases[historyReleaseIndex] = historyRelease
		}
	} else {
		return fmt.Errorf("failed to get the latest release")
	}

	Printfln("********************************************")
	return nil
}
