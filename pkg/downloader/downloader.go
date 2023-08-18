package downloader

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/schollz/progressbar/v3"
)

type GitHubRelease struct {
	URL  string `json:"html_url"`
	Name string `json:"name"`
}

func FetchGitHubReleases(repoURL string, ignoreCert bool) ([]string, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases", repoURL)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	client := &http.Client{}
	if ignoreCert {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)

	var releases []GitHubRelease
	err = json.NewDecoder(resp.Body).Decode(&releases)
	if err != nil {
		return nil, err
	}

	var links []string
	for _, release := range releases {
		links = append(links, release.URL)
	}

	return links, nil
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
