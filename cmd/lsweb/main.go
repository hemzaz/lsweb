package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hemzaz/lsweb/pkg/downloader"
	"github.com/hemzaz/lsweb/pkg/parser"
)

func main() {
	// Setup flags
	urlFlag := flag.String("u", "", "URL to fetch links from")
	fileFlag := flag.String("f", "", "File to fetch links from")
	outputFlag := flag.String("o", "txt", "Output format (json, txt, num, html)")
	filterFlag := flag.String("filter", "", "Regex to filter links (can be specified multiple times)")
	limitFlag := flag.Int("limit", 0, "Limit the number of links to fetch")
	ignoreCertFlag := flag.Bool("ic", false, "Ignore certificate errors")
	ghFlag := flag.Bool("gh", false, "Fetch GitHub releases")
	downloadFlag := flag.Bool("download", false, "Download the files")
	simFlag := flag.Bool("sim", false, "Download files simultaneously")
	listFlag := flag.Bool("list", true, "List the links")
	maxConcurrentFlag := flag.Int("max-concurrent", 5, "Maximum number of concurrent downloads (with -sim)")
	overwriteFlag := flag.Bool("overwrite", false, "Overwrite existing files when downloading")
	timeoutFlag := flag.Int("timeout", 60, "Timeout in seconds for HTTP requests")
	versionFlag := flag.Bool("version", false, "Show version information")
	flag.Parse()

	// Show version and exit if requested
	if *versionFlag {
		fmt.Println("lsweb version 1.0.0")
		os.Exit(0)
	}

	// Configure logging
	log.SetPrefix("lsweb: ")
	log.SetFlags(0) // Don't show date/time in errors

	var links []string
	var err error

	// Require either URL or file input
	if *urlFlag == "" && *fileFlag == "" {
		fmt.Println("Error: Please provide a URL (-u) or file (-f) to fetch links from")
		flag.Usage()
		os.Exit(1)
	}

	// Set the timeout value for HTTP requests
	downloader.SetTimeout(time.Duration(*timeoutFlag) * time.Second)
	downloader.SetMaxConcurrent(*maxConcurrentFlag)
	downloader.SetOverwriteFiles(*overwriteFlag)

	// Fetch links from source
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
	}

	// Filter links if requested
	if *filterFlag != "" {
		links, err = parser.FilterLinksByRegex(links, *filterFlag)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Limit number of links if requested
	if *limitFlag > 0 && *limitFlag < len(links) {
		links = links[:*limitFlag]
	}

	// Show link count
	fmt.Printf("Found %d links\n", len(links))

	// Download files if requested
	if *downloadFlag {
		if len(links) == 0 {
			fmt.Println("No links to download")
		} else if *simFlag {
			downloader.DownloadFilesSimultaneously(links, *ignoreCertFlag)
		} else {
			downloader.DownloadFiles(links, *ignoreCertFlag, true)
		}
	}

	// List links if requested
	if *listFlag && len(links) > 0 {
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
			fmt.Printf("Invalid output format: %s\n", *outputFlag)
			fmt.Println("Valid formats: json, txt, num, html")
			os.Exit(1)
		}
	}
}
