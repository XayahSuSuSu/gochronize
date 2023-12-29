package util

import (
	"encoding/json"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func GetHttpClient(proxyHttp string, timeout int) (*http.Client, error) {
	var client http.Client
	if proxyHttp != "" {
		proxy, err := url.Parse(proxyHttp)
		if err != nil {
			fmt.Printf("Failed to parse proxy: %s.\n", proxyHttp)
			return nil, err
		}
		client = http.Client{
			Timeout: time.Duration(timeout) * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxy),
			},
		}
	} else {
		client = http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		}
	}
	return &client, nil
}

func GetRelease(client *http.Client, user, repo string, page int) ([]Release, int) {
	api := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?page=%d", user, repo, page)

	resp, err := client.Get(api)
	if err != nil {
		fmt.Printf("* err: Failed to get: %s, %s.\n", api, err.Error())
		return nil, -1
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("* err: Failed to access, status code: %d.\n", resp.StatusCode)
		return nil, -1
	}

	nextPage := -1
	links := resp.Header.Get("Link")
	link := strings.Split(links, ", ")
	for _, item := range link {
		bin := strings.Split(item, "; ")
		if len(bin) == 2 {
			rel := strings.Replace(strings.Split(bin[1], "rel=")[1], "\"", "", -1)
			if rel == "next" {
				nextPage, _ = strconv.Atoi(strings.Replace(strings.Split(bin[0], "?page=")[1], ">", "", -1))
			}
		}
	}

	var releases []Release
	err = json.NewDecoder(resp.Body).Decode(&releases)
	if err != nil {
		fmt.Printf("* err: Failed to parse release body: %s.", err.Error())
		return nil, -1
	}

	return releases, nextPage
}

func GetLatestRelease(client *http.Client, user, repo string) *Release {
	api := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", user, repo)

	resp, err := client.Get(api)
	if err != nil {
		fmt.Printf("* err: Failed to get: %s, %s.\n", api, err.Error())
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("* err: Failed to access, status code: %d.\n", resp.StatusCode)
		return nil
	}

	var release Release
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		fmt.Printf("* err: Failed to parse release body: %s.", err.Error())
		return nil
	}

	return &release
}

func GetReleaseByTag(client *http.Client, user, repo, tag string) *Release {
	api := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", user, repo, tag)

	resp, err := client.Get(api)
	if err != nil {
		fmt.Printf("* err: Failed to get: %s, %s.\n", api, err.Error())
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		fmt.Printf("* err: Tag not found: %s.\n", tag)
		return nil
	} else if resp.StatusCode != http.StatusOK {
		fmt.Printf("* err: Failed to access, status code: %d.\n", resp.StatusCode)
		return nil
	}

	var release Release
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		fmt.Printf("* err: Failed to parse release body: %s.", err.Error())
		return nil
	}

	return &release
}

func Download(client *http.Client, url, dst string) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	bar := pb.Full.Start64(resp.ContentLength)
	bar.Set(pb.Bytes, true)
	bar.Set(pb.SIBytesPrefix, true)
	bar.SetRefreshRate(time.Second)

	reader := bar.NewProxyReader(resp.Body)
	_, err = io.Copy(dstFile, reader)
	if err != nil {
		return err
	}
	bar.Finish()

	return nil
}
