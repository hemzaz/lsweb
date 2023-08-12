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
		links, err := parser.ExtractLinks(urlFlag)
		if err != nil {
			fmt.Println("Error extracting links:", err)
			return
		}

		// Display links based on the output format specified
		displayLinks(links)
	}

	if downloadFlag {
		links, err := parser.ExtractLinks(urlFlag)
		if err != nil {
			fmt.Println("Error extracting links:", err)
			return
		}
		downloader.DownloadFiles(links, simFlag)
	}
}

func displayLinks(links []string) {
	switch outputFlag {
	case "json":
		data, _ := json.Marshal(links)
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
			data, _ := json.Marshal(links)
			content = string(data)
		case "txt", "num", "html":
			for _, link := range links {
				content += link + "\n"
			}
		}
		ioutil.WriteFile(fileFlag, []byte(content), 0644)
	}
}
