package parser

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"golang.org/x/net/html"
)

func ExtractLinksFromURL(targetURL string, ignoreCert bool) ([]string, error) {
	client := &http.Client{}
	if ignoreCert {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	resp, err := client.Get(targetURL)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)

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

func ExtractLinksFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)

	doc, err := html.Parse(file)
	if err != nil {
		return nil, err
	}

	var links []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					links = append(links, a.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return links, nil
}

func FilterLinksByRegex(links []string, regex string) ([]string, error) {
	re, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}

	var filteredLinks []string
	for _, link := range links {
		if re.MatchString(link) {
			filteredLinks = append(filteredLinks, link)
		}
	}

	return filteredLinks, nil
}

func PrintLinksAsJSON(links []string) {
	data, err := json.Marshal(links)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(string(data))
}

func PrintLinksAsNumbered(links []string) {
	for i, link := range links {
		fmt.Printf("%d. %s\n", i+1, link)
	}
}

func PrintLinksAsHTML(links []string) {
	fmt.Println("<ul>")
	for _, link := range links {
		fmt.Printf("<li><a href=\"%s\">%s</a></li>\n", link, link)
	}
	fmt.Println("</ul>")
}

func PrintLinksAsText(links []string) {
	for _, link := range links {
		fmt.Println(link)
	}
}
