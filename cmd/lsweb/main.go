package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/hemzaz/lsweb/pkg/downloader"
	"github.com/hemzaz/lsweb/pkg/parser"
)

var (
	urlFlag      string
	listFlag     bool
	downloadFlag bool
	simFlag      bool
	outputFlag   string
	fileFlag     string
	ghFlag       bool
)

func init() {
	flag.StringVar(&urlFlag, "url", "", "the URL we are targeting")
	flag.BoolVar(&listFlag, "L", false, "list downloadable links")
	flag.BoolVar(&downloadFlag, "D", false, "download the files")
	flag.BoolVar(&simFlag, "S", false, "download simultaneously")
	flag.StringVar(&outputFlag, "O", "txt", "output format")
	flag.StringVar(&fileFlag, "F", "", "file to write the output")
	flag.BoolVar(&ghFlag, "gh", false, "fetch all releases from a GitHub URL")
	flag.Parse()
}

func fetchGitHubReleases(repoURL string) ([]string, error) {
	parts := strings.Split(repoURL, "/")
	if len(parts) < 5 {
		return nil, fmt.Errorf("invalid GitHub URL")
	}
	user, repo := parts[3], parts[4]

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", user, repo)
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
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

func main() {
	if ghFlag {
		links, err := fetchGitHubReleases(urlFlag)
		if err != nil {
			fmt.Println("Error fetching GitHub releases:", err)
			return
		}

		if listFlag || (!listFlag && !downloadFlag) {
			// Print the links
			for _, link := range links {
				fmt.Println(link)
			}
		}

		if downloadFlag {
			if simFlag {
				downloader.DownloadFilesSimultaneously(links)
			} else {
				downloader.DownloadFiles(links, true)
			}
		}
		return
	}

	links, err := parser.ExtractLinks(urlFlag)
	if err != nil {
		fmt.Println("Error extracting links:", err)
		return
	}

	if listFlag || (!listFlag && !downloadFlag) {
		// Print the links
		for _, link := range links {
			fmt.Println(link)
		}
	}

	if downloadFlag {
		if simFlag {
			downloader.DownloadFilesSimultaneously(links)
		} else {
			downloader.DownloadFiles(links, true)
		}
	}
}
