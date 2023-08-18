package downloader

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/schollz/progressbar/v3"
)

type GitHubRelease struct {
	URL  string `json:"html_url"`
	Name string `json:"name"`
}

func FetchGitHubReleases(repoURL string, ignoreCert bool) ([]string, error) {
	parts := strings.Split(repoURL, "/")
	if len(parts) < 5 {
		return nil, fmt.Errorf("invalid GitHub URL")
	}
	user, repo := parts[3], parts[4]

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", user, repo)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreCert},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var releases []struct {
		Assets []struct {
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, err
	}

	var downloadLinks []string
	for _, release := range releases {
		for _, asset := range release.Assets {
			downloadLinks = append(downloadLinks, asset.BrowserDownloadURL)
		}
	}

	return downloadLinks, nil
}

func DownloadFile(url string, ignoreCert bool, showProgress bool) error {
	client := &http.Client{}
	if ignoreCert {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
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

	filename := filepath.Base(url)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)

	if showProgress {
		bar := progressbar.DefaultBytes(
			resp.ContentLength,
			"downloading",
		)
		_, err = io.Copy(io.MultiWriter(file, bar), resp.Body)
	} else {
		_, err = io.Copy(file, resp.Body)
	}

	return err
}

func DownloadFiles(urls []string, ignoreCert bool, showProgress bool) {
	for _, url := range urls {
		err := DownloadFile(url, ignoreCert, showProgress)
		if err != nil {
			fmt.Println("Error downloading file:", err)
		}
	}
}

func DownloadFilesSimultaneously(urls []string, ignoreCert bool) {
	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			err := DownloadFile(url, ignoreCert, true)
			if err != nil {
				fmt.Println("Error downloading file:", err)
			}
		}(url)
	}
	wg.Wait()
}
