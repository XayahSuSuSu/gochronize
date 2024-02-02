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

var token = ""

func Get(client *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
		println(token)
	}
	return client.Do(req)
}

func GetHttpClient(proxyHttp string, t string, timeout int) *http.Client {
	var client http.Client
	if proxyHttp != "" {
		proxy, err := url.Parse(proxyHttp)
		if err != nil {
			Fprintfln("Failed to parse proxy: %s, %v", proxyHttp, err)
			os.Exit(Error)
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
	token = t
	return &client
}

func GetRelease(client *http.Client, user, repo string, page int) ([]Release, int) {
	api := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?page=%d", user, repo, page)

	resp, err := Get(client, api)
	if err != nil {
		Fprintfln("* err: Failed to get: %s, %v", api, err)
		return nil, -1
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		Fprintfln("* err: Failed to access, status code: %d.", resp.StatusCode)
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
		Fprintfln("* err: Failed to parse release body: %v", err)
		return nil, -1
	}

	return releases, nextPage
}

func GetLatestRelease(client *http.Client, user, repo string) *Release {
	api := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", user, repo)

	resp, err := Get(client, api)
	if err != nil {
		Fprintfln("* err: Failed to get: %s, %v", api, err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		Fprintfln("* err: Failed to access, status code: %d.", resp.StatusCode)
		return nil
	}

	var release Release
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		Fprintfln("* err: Failed to parse release body: %v", err)
		return nil
	}

	return &release
}

func GetReleaseByTag(client *http.Client, user, repo, tag string) *Release {
	api := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", user, repo, tag)

	resp, err := Get(client, api)
	if err != nil {
		Fprintfln("* err: Failed to get: %s, %v", api, err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		Fprintfln("* err: Tag not found: %s.", tag)
		return nil
	} else if resp.StatusCode != http.StatusOK {
		Fprintfln("* err: Failed to access, status code: %d.", resp.StatusCode)
		return nil
	}

	var release Release
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		Fprintfln("* err: Failed to parse release body: %v", err)
		return nil
	}

	return &release
}

func Download(client *http.Client, url, dst string) error {
	resp, err := Get(client, url)
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
