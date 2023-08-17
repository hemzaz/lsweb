package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hemzaz/lsweb/pkg/downloader"
	"github.com/hemzaz/lsweb/pkg/parser"
)

func main() {
	urlFlag := flag.String("u", "", "URL to extract links from")
	fileFlag := flag.String("f", "", "File to extract links from")
	ghFlag := flag.String("gh", "", "GitHub repository to fetch releases from")
	downloadFlag := flag.Bool("d", false, "Download the files")
	simFlag := flag.Bool("s", false, "Download the files simultaneously")
	ignoreCertFlag := flag.Bool("ic", false, "Ignore certificate verification")
	limitFlag := flag.Int("l", 0, "Limit the number of links to extract")
	filterFlag := flag.String("filter", "", "Filter links using a regular expression")
	outputFlag := flag.String("o", "txt", "Specify the output format: json, txt, num, html")
	flag.Parse()

	var links []string
	var err error

	if *urlFlag != "" {
		links, err = parser.ExtractLinksFromURL(*urlFlag, *ignoreCertFlag)
		if err != nil {
			log.Fatal(err)
		}
	} else if *fileFlag != "" {
		links, err = parser.ExtractLinksFromFile(*fileFlag)
		if err != nil {
			log.Fatal(err)
		}
	} else if *ghFlag != "" {
		links, err = downloader.FetchGitHubReleases(*ghFlag, *ignoreCertFlag)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("Please provide a URL, file, or GitHub repository to extract links from.")
		os.Exit(1)
	}

	if *limitFlag > 0 && *limitFlag < len(links) {
		links = links[:*limitFlag]
	}

	if *filterFlag != "" {
		links, err = parser.FilterLinksByRegex(links, *filterFlag)
		if err != nil {
			log.Fatal(err)
		}
	}

	switch strings.ToLower(*outputFlag) {
	case "json":
		parser.PrintLinksAsJSON(links)
	case "txt":
		parser.PrintLinksAsText(links)
	case "num":
		parser.PrintLinksAsNumbered(links)
	case "html":
		parser.PrintLinksAsHTML(links)
	default:
		fmt.Println("Invalid output format. Available formats: json, txt, num, html.")
		os.Exit(1)
	}

	if *downloadFlag {
		if *simFlag {
			downloader.DownloadFilesSimultaneously(links, *ignoreCertFlag)
		} else {
			downloader.DownloadFiles(links, *ignoreCertFlag)
		}
	}
}
