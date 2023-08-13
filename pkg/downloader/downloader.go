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
)

type GitHubRelease struct {
	Assets []struct {
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func DownloadFiles(links []string, ignoreCert bool) error {
	for _, link := range links {
		err := downloadFile(link, ignoreCert)
		if err != nil {
			return err
		}
	}
	return nil
}

func DownloadFilesSimultaneously(links []string, ignoreCert bool) error {
	var wg sync.WaitGroup
	errors := make(chan error, len(links))

	for _, link := range links {
		wg.Add(1)
		go func(link string) {
			defer wg.Done()
			err := downloadFile(link, ignoreCert)
			if err != nil {
				errors <- err
			}
		}(link)
	}

	wg.Wait()
	close(errors)

	if len(errors) > 0 {
		return fmt.Errorf("encountered multiple errors during download")
	}

	return nil
}

func downloadFile(link string, ignoreCert bool) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreCert},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	filename := filepath.Base(resp.Request.URL.Path)
	if filename == "/" || filename == "." {
		filename = "downloaded_file"
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func FetchGitHubReleases(repoURL string) ([]string, error) {
	apiURL := strings.Replace(repoURL, "github.com", "api.github.com/repos", 1) + "/releases"
	token := os.Getenv("GITHUB_TOKEN")

	client := &http.Client{}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	if token != "" {
		req.Header.Add("Authorization", "token "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var releases []GitHubRelease
	err = json.Unmarshal(body, &releases)
	if err != nil {
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
