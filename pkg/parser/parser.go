package parser

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"os"
	"regexp"
)

var LinkRegex = regexp.MustCompile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\\(\\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)

// ExtractLinksFromURL fetches the content of a URL and extracts all links from it.
func ExtractLinksFromURL(targetURL string, ignoreCert bool) ([]string, error) {
	// If ignoreCert is true, create a custom client that ignores SSL verification
	var client *http.Client
	if ignoreCert {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{}
	}

	resp, err := client.Get(targetURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch the webpage")
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var links []string
	var malformedURLs []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					// Convert relative URLs to absolute URLs
					absoluteURL, err := url.Parse(a.Val)
					if err != nil {
						malformedURLs = append(malformedURLs, a.Val)
						continue
					}
					absoluteURL = resp.Request.URL.ResolveReference(absoluteURL)
					links = append(links, absoluteURL.String())
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if len(malformedURLs) > 0 {
		return links, fmt.Errorf("malformed URLs detected: %v", malformedURLs)
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
