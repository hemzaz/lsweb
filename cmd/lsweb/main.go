package main

import (
	"flag"
	"fmt"
	"github.com/hemzaz/lsweb/pkg/downloader"
	"github.com/hemzaz/lsweb/pkg/parser"
	"log"
	"os"
)

var (
	urlFlag        = flag.String("url", "", "URL to extract links from")
	ghFlag         = flag.Bool("gh", false, "Flag to indicate if the provided URL is a GitHub releases page")
	fileFlag       = flag.String("file", "", "File path to extract links from")
	outputFlag     = flag.String("output", "txt", "Output format: txt, num, json, html")
	downloadFlag   = flag.Bool("download", false, "Download the files from the extracted links")
	simFlag        = flag.Bool("sim", false, "Download files simultaneously")
	limitFlag      = flag.Int("limit", 0, "Limit the number of links to extract")
	filterFlag     = flag.String("filter", "", "Regex pattern to filter the links")
	ignoreCertFlag = flag.Bool("ic", false, "Ignore certificate errors")
)

func main() {
	flag.Parse()

	if *urlFlag == "" && *fileFlag == "" {
		fmt.Println("Please provide a URL or file path to extract links from.")
		os.Exit(1)
	}

	var links []string
	var err error

	if *ghFlag {
		releases, err := downloader.FetchGitHubReleases(*urlFlag, *ignoreCertFlag)
		if err != nil {
			log.Fatalf("Error fetching GitHub releases: %v", err)
		}

		// Extract download URLs from the GitHub releases
		for _, release := range releases {
			for _, asset := range release.Assets {
				links = append(links, asset.BrowserDownloadURL)
			}
		}
	} else if *urlFlag != "" {
		links, err = parser.ExtractLinksFromURL(*urlFlag, *ignoreCertFlag)
		if err != nil {
			log.Fatalf("Error extracting links from URL: %v", err)
		}
	} else if *fileFlag != "" {
		links, err = parser.ExtractLinksFromFile(*fileFlag)
		if err != nil {
			log.Fatalf("Error extracting links from file: %v", err)
		}
	}

	if *filterFlag != "" {
		links, err = parser.FilterLinksByRegex(links, *filterFlag)
		if err != nil {
			log.Fatalf("Error filtering links: %v", err)
		}
	}

	if *limitFlag > 0 && *limitFlag < len(links) {
		links = links[:*limitFlag]
	}

	switch *outputFlag {
	case "json":
		parser.PrintLinksAsJSON(links)
	case "num":
		parser.PrintLinksAsNumbered(links)
	case "html":
		parser.PrintLinksAsHTML(links)
	default:
		parser.PrintLinksAsText(links)
	}

	if *downloadFlag {
		if *simFlag {
			downloader.DownloadFilesSimultaneously(links, *ignoreCertFlag)
		} else {
			downloader.DownloadFiles(links, *ignoreCertFlag)
		}
	}
}
