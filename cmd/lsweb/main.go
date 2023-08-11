package main

import (
	"flag"
	"fmt"

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
)

func init() {
	flag.StringVar(&urlFlag, "url", "", "The URL to target.")
	flag.BoolVar(&listFlag, "L", false, "List downloadable links. Default action when no flag is provided.")
	flag.BoolVar(&downloadFlag, "D", false, "Download the files. Default to non-simultaneously.")
	flag.BoolVar(&simFlag, "S", false, "Download simultaneously. Use with -D flag.")
	flag.StringVar(&outputFlag, "O", "text", "Output format. Defaults to text. Options: json, txt, num, html.")
	flag.StringVar(&fileFlag, "F", "", "File to write the output to.")
}

func main() {
	flag.Parse()

	if urlFlag == "" {
		fmt.Println("Please provide a URL using the -url flag.")
		return
	}

	if listFlag || (!listFlag && !downloadFlag) {
		links := parser.ExtractLinks(urlFlag)
		// Display links based on the output format specified
		for _, link := range links {
			fmt.Println(link)
		}
	}

	if downloadFlag {
		links := parser.ExtractLinks(urlFlag)
		downloader.DownloadFiles(links, simFlag)
	}
}
