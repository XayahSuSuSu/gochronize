package util

import (
	"encoding/json"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

func getHttpClient(proxyHttp string) (*http.Client, error) {
	var client http.Client
	if proxyHttp != "" {
		proxy, err := url.Parse(proxyHttp)
		if err != nil {
			fmt.Printf("Failed to parse proxy: %s.\n", proxyHttp)
			return nil, err
		}
		client = http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxy),
			},
		}
	} else {
		client = http.Client{}
	}
	return &client, nil
}

func GetRelease(user, repo, proxyHttp string) *Release {
	api := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", user, repo)
	client, err := getHttpClient(proxyHttp)
	if err != nil {
		return nil
	}

	resp, err := client.Get(api)
	if err != nil {
		fmt.Printf("Failed to get: %s.\n", api)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to access, status code: %d.\n", resp.StatusCode)
		return nil
	}

	var release Release
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		fmt.Println("Failed to parse release body.")
		return nil
	}

	return &release
}

func Download(url string, dst, proxyHttp string) error {
	client, err := getHttpClient(proxyHttp)
	if err != nil {
		return err
	}

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
