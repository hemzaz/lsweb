package parser

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
)

var LinkRegex = regexp.MustCompile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*(),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)

func ExtractLinksFromURL(url string, ignoreCert bool) ([]string, error) {
	var links []string

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreCert},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	matches := LinkRegex.FindAllString(string(body), -1)
	for _, match := range matches {
		links = append(links, match)
	}

	return links, nil
}

func FilterLinksByRegex(links []string, filter string) ([]string, error) {
	var filteredLinks []string
	filterRegex, err := regexp.Compile(filter)
	if err != nil {
		return nil, err
	}

	for _, link := range links {
		if filterRegex.MatchString(link) {
			filteredLinks = append(filteredLinks, link)
		}
	}

	return filteredLinks, nil
}

func PrintLinksAsJSON(links []string) {
	data, err := json.MarshalIndent(links, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(string(data))
}

func PrintLinksAsText(links []string) {
	for _, link := range links {
		fmt.Println(link)
	}
}

func PrintLinksAsNumbered(links []string) {
	for i, link := range links {
		fmt.Printf("%d. %s\n", i+1, link)
	}
}

func PrintLinksAsHTML(links []string) {
	fmt.Println("<ul>")
	for _, link := range links {
		fmt.Printf("  <li><a href=\"%s\">%s</a></li>\n", link, link)
	}
	fmt.Println("</ul>")
}

func ExtractLinksFromFile(filePath string) ([]string, error) {
	var links []string

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	matches := LinkRegex.FindAllString(string(data), -1)
	for _, match := range matches {
		links = append(links, match)
	}

	return links, nil
}
