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
	urlFlag := flag.String("u", "", "URL to fetch links from")
	fileFlag := flag.String("f", "", "File to fetch links from")
	outputFlag := flag.String("o", "txt", "Output format (json, txt, num, html)")
	filterFlag := flag.String("filter", "", "Regex to filter links")
	limitFlag := flag.Int("limit", 0, "Limit the number of links to fetch")
	ignoreCertFlag := flag.Bool("ic", false, "Ignore certificate errors")
	ghFlag := flag.Bool("gh", false, "Fetch GitHub releases")
	downloadFlag := flag.Bool("download", false, "Download the files")
	simFlag := flag.Bool("sim", false, "Download files simultaneously")
	listFlag := flag.Bool("list", true, "List the links")
	flag.Parse()

	var links []string
	var err error

	if *urlFlag != "" {
		if *ghFlag {
			links, err = downloader.FetchGitHubReleases(*urlFlag, *ignoreCertFlag)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			links, err = parser.ExtractLinksFromURL(*urlFlag, *ignoreCertFlag)
			if err != nil {
				log.Fatal(err)
			}
		}
	} else if *fileFlag != "" {
		links, err = parser.ExtractLinksFromFile(*fileFlag)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("Please provide a URL or file to fetch links from")
		os.Exit(1)
	}

	if *filterFlag != "" {
		links, err = parser.FilterLinksByRegex(links, *filterFlag)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *limitFlag > 0 && *limitFlag < len(links) {
		links = links[:*limitFlag]
	}

	if *downloadFlag {
		if *simFlag {
			downloader.DownloadFilesSimultaneously(links, *ignoreCertFlag)
		} else {
			downloader.DownloadFiles(links, *ignoreCertFlag, true)
		}
	}

	if *listFlag {
		switch strings.ToLower(*outputFlag) {
		case "json":
			parser.PrintLinksAsJSON(links)
		case "num":
			parser.PrintLinksAsNumbered(links)
		case "html":
			parser.PrintLinksAsHTML(links)
		case "txt":
			parser.PrintLinksAsText(links)
		default:
			fmt.Println("Invalid output format")
			os.Exit(1)
		}
	}
}
