package downloader

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type GitHubRelease struct {
	URL     string `json:"url"`
	HTMLURL string `json:"html_url"`
	Assets  []struct {
		URL                string `json:"url"`
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func DownloadFile(filepath string, url string, ignoreCert bool) error {
	client := &http.Client{}
	if ignoreCert {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = tr
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(out)

	_, err = io.Copy(out, resp.Body)
	return err
}

func DownloadFiles(urls []string, ignoreCert bool) {
	for _, url := range urls {
		parts := strings.Split(url, "/")
		filename := parts[len(parts)-1]
		err := DownloadFile(filename, url, ignoreCert)
		if err != nil {
			fmt.Println("Error downloading file:", err)
		}
	}
}

func DownloadFilesSimultaneously(urls []string, ignoreCert bool) {
	ch := make(chan string)
	for _, url := range urls {
		go func(url string) {
			parts := strings.Split(url, "/")
			filename := parts[len(parts)-1]
			err := DownloadFile(filename, url, ignoreCert)
			if err != nil {
				ch <- fmt.Sprintf("Error downloading file %s: %s", filename, err)
			} else {
				ch <- fmt.Sprintf("Downloaded file %s", filename)
			}
		}(url)
	}

	for range urls {
		fmt.Println(<-ch)
	}
}

func FetchGitHubReleases(repoURL string, ignoreCert bool) ([]string, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases", repoURL)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	if ignoreCert {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = tr
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	var releases []GitHubRelease
	err = json.NewDecoder(resp.Body).Decode(&releases)
	if err != nil {
		return nil, err
	}

	var urls []string
	for _, release := range releases {
		for _, asset := range release.Assets {
			urls = append(urls, asset.BrowserDownloadURL)
		}
	}

	return urls, nil
}
