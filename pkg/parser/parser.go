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

var LinkRegex = regexp.MustCompile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\\(\\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)

// ExtractLinksFromURL fetches the content of a URL and extracts all links from it.
func ExtractLinksFromURL(url string, ignoreCert bool) ([]string, error) {
	var links []string

	// Create an HTTP client that ignores certificate errors if the flag is set.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreCert},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

// ExtractLinksFromFile extracts all links from a given file.
func ExtractLinksFromFile(filePath string) ([]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	matches := LinkRegex.FindAllString(string(content), -1)
	return matches, nil
}

// FilterLinksByRegex filters the provided links based on a regex pattern.
func FilterLinksByRegex(links []string, pattern string) ([]string, error) {
	var filteredLinks []string
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	for _, link := range links {
		if re.MatchString(link) {
			filteredLinks = append(filteredLinks, link)
		}
	}

	return filteredLinks, nil
}

// PrintLinksAsJSON prints the links in JSON format.
func PrintLinksAsJSON(links []string) {
	data, _ := json.Marshal(links)
	fmt.Println(string(data))
}

// PrintLinksAsNumbered prints the links in a numbered list.
func PrintLinksAsNumbered(links []string) {
	for i, link := range links {
		fmt.Printf("%d. %s\n", i+1, link)
	}
}

// PrintLinksAsHTML prints the links in an HTML list format.
func PrintLinksAsHTML(links []string) {
	fmt.Println("<ul>")
	for _, link := range links {
		fmt.Printf("  <li><a href=\"%s\">%s</a></li>\n", link, link)
	}
	fmt.Println("</ul>")
}

// PrintLinksAsText prints the links as plain text.
func PrintLinksAsText(links []string) {
	for _, link := range links {
		fmt.Println(link)
	}
}
