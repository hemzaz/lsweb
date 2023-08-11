package main

import (
	"flag"
	"fmt"
	"lsweb/downloader"
	"lsweb/parser"
	"os"
)

func main() {
	// Define CLI flags
	listFlag := flag.Bool("L", false, "List downloadable links")
	downloadFlag := flag.Bool("D", false, "Download the files")
	simFlag := flag.Bool("S", false, "Download simultaneously")
	outputFlag := flag.String("O", "text", "Output format: json, txt, num, html")
	fileFlag := flag.String("F", "", "File to write the output")

	// Parse the flags
	flag.Parse()

	// Get the URL argument
	if len(flag.Args()) < 1 {
		fmt.Println("Error: URL not provided.")
		os.Exit(1)
	}
	url := flag.Args()[0]

	// Extract links from the URL
	links := parser.ExtractLinks(url)

	// Handle the flags
	if *listFlag || len(flag.Args()) == 1 {
		// Display the links
		for _, link := range links {
			fmt.Println(link)
		}
	} else if *downloadFlag {
		// Download the files
		downloader.DownloadFiles(links, *simFlag)
	}

	// Handle the output format and file output
	// This can be enhanced further based on the desired output format and file handling.
}
