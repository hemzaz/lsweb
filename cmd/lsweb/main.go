package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"

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
	flag.StringVar(&urlFlag, "url", "", "The target URL to fetch and list/download files from.")
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

	links, err := parser.ExtractLinks(urlFlag)
	if err != nil {
		fmt.Println("Error extracting links:", err)
		return
	}

	if listFlag || (!listFlag && !downloadFlag) {
		displayLinks(links)
	}

	if downloadFlag {
		downloader.DownloadFiles(links, simFlag)
	}
}

func displayLinks(links []string) {
	switch outputFlag {
	case "json":
		data, err := json.Marshal(links)
		if err != nil {
			fmt.Println("Error marshaling links to JSON:", err)
			return
		}
		fmt.Println(string(data))
	case "txt":
		for _, link := range links {
			fmt.Println(link)
		}
	case "num":
		for i, link := range links {
			fmt.Printf("%d. %s\n", i+1, link)
		}
	case "html":
		fmt.Println("<ul>")
		for _, link := range links {
			fmt.Printf("<li><a href='%s'>%s</a></li>\n", link, link)
		}
		fmt.Println("</ul>")
	default:
		for _, link := range links {
			fmt.Println(link)
		}
	}

	// If fileFlag is provided, write the output to the specified file
	if fileFlag != "" {
		var content string
		switch outputFlag {
		case "json":
			data, err := json.Marshal(links)
			if err != nil {
				fmt.Println("Error marshaling links to JSON:", err)
				return
			}
			content = string(data)
		case "txt", "num", "html":
			for _, link := range links {
				content += link + "\n"
			}
		}
		err := ioutil.WriteFile(fileFlag, []byte(content), 0644)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}
	}
}
