// Package parser provides functions for extracting and filtering links from various sources.
package parser

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"golang.org/x/net/html"

	"github.com/hemzaz/lsweb/pkg/common"
)

// ExtractLinksFromURL fetches a URL and extracts all links from its content.
// Supports HTML, JSON, XML content types.
// The ignoreCert parameter can be used to skip TLS certificate validation.
// Returns a slice of unique links found in the content or an error if the fetch or parsing fails.
func ExtractLinksFromURL(targetURL string, ignoreCert bool) ([]string, error) {
	// Set up a client with timeout
	client := &http.Client{
		Timeout: common.DefaultTimeout,
	}
	if ignoreCert {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	// Create context for the request
	ctx, cancel := context.WithTimeout(context.Background(), common.DefaultTimeout)
	defer cancel()

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add a user-agent to be polite
	req.Header.Set("User-Agent", common.UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching webpage: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Error closing response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned non-success status: %d %s", resp.StatusCode, resp.Status)
	}

	// Check content type - only process recognized types
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") &&
		!strings.Contains(contentType, "application/json") &&
		!strings.Contains(contentType, "application/xml") &&
		!strings.Contains(contentType, "text/xml") {
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}

	// Limit body size for safety
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, common.MaxContentSize))
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Different handling based on content type
	var links []string

	if strings.Contains(contentType, "application/json") {
		// For JSON content, try to extract URLs from JSON structure
		var jsonData interface{}
		if err := json.Unmarshal(bodyBytes, &jsonData); err != nil {
			return nil, fmt.Errorf("error parsing JSON: %w", err)
		}
		links = extractLinksFromJSON(jsonData)
	} else {
		// Create a new reader from the bytes
		bodyReader := bytes.NewReader(bodyBytes)

		// Parse HTML for links
		doc, err := html.Parse(bodyReader)
		if err != nil {
			return nil, fmt.Errorf("error parsing HTML: %w", err)
		}

		// Extract links from HTML
		var malformedURLs []string
		links, malformedURLs = extractLinksFromHTML(doc, resp.Request.URL)

		if len(malformedURLs) > 0 {
			// Continue with the links we found, but warn about malformed ones
			fmt.Printf("Warning: %d malformed URLs detected\n", len(malformedURLs))
		}
	}

	// Remove duplicates
	links = removeDuplicateLinks(links)

	return links, nil
}

// Helper function to extract links from HTML
func extractLinksFromHTML(doc *html.Node, baseURL *url.URL) ([]string, []string) {
	var links []string
	var malformedURLs []string

	// Use a map to track visited URLs for deduplication
	visited := make(map[string]bool)

	// Use a function to traverse the DOM
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					// Convert relative URLs to absolute URLs
					absoluteURL, err := url.Parse(a.Val)
					if err != nil {
						malformedURLs = append(malformedURLs, a.Val)
						continue
					}

					absoluteURL = baseURL.ResolveReference(absoluteURL)
					urlStr := absoluteURL.String()

					// Skip javascript: and mailto: links
					if strings.HasPrefix(urlStr, "javascript:") ||
						strings.HasPrefix(urlStr, "mailto:") ||
						strings.HasPrefix(urlStr, "#") {
						continue
					}

					// Add to links if not already visited
					if !visited[urlStr] {
						visited[urlStr] = true
						links = append(links, urlStr)
					}
				}
			}
		}

		// Traverse children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return links, malformedURLs
}

// Helper function to extract links from JSON
func extractLinksFromJSON(data interface{}) []string {
	var links []string
	var extract func(interface{})

	// Use a map to track visited URLs for deduplication
	visited := make(map[string]bool)

	// Define URL regex pattern
	urlPattern := regexp.MustCompile(`https?://[^\s"']+`)

	extract = func(v interface{}) {
		switch val := v.(type) {
		case map[string]interface{}:
			for _, value := range val {
				extract(value)
			}
		case []interface{}:
			for _, item := range val {
				extract(item)
			}
		case string:
			// Check if string is a URL
			if urlPattern.MatchString(val) {
				matches := urlPattern.FindAllString(val, -1)
				for _, match := range matches {
					if !visited[match] {
						visited[match] = true
						links = append(links, match)
					}
				}
			}
		}
	}

	extract(data)
	return links
}

// Helper function to remove duplicate links
func removeDuplicateLinks(links []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, link := range links {
		if !seen[link] {
			seen[link] = true
			result = append(result, link)
		}
	}

	return result
}

// ExtractLinksFromFile reads a file and extracts all links from its content.
// Supports HTML, JSON, and plain text files.
// The file size is limited to 10MB for safety.
// Returns a slice of unique links found in the file or an error if reading or parsing fails.
func ExtractLinksFromFile(filePath string) ([]string, error) {
	// Check file size before opening to prevent loading large files
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("error checking file: %w", err)
	}

	// Limit file size to 10MB
	if fileInfo.Size() > 10*1024*1024 {
		return nil, fmt.Errorf("file too large (%.2f MB). Maximum size is 10MB", float64(fileInfo.Size())/(1024*1024))
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Error closing file: %v\n", closeErr)
		}
	}()

	// Read the first few bytes to detect file type
	header := make([]byte, 512)
	_, err = file.Read(header)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error reading file header: %w", err)
	}

	// Reset file position
	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("error resetting file position: %w", err)
	}

	// Detect content type
	contentType := http.DetectContentType(header)

	// Read the entire file
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file content: %w", err)
	}

	var links []string

	// Process based on content type
	if strings.Contains(contentType, "text/html") {
		// Parse HTML
		doc, err := html.Parse(bytes.NewReader(content))
		if err != nil {
			return nil, fmt.Errorf("error parsing HTML: %w", err)
		}

		// Create a base URL for resolving relative links
		baseURL, _ := url.Parse("file://" + filePath)

		// Extract links
		links, _ = extractLinksFromHTML(doc, baseURL)

	} else if strings.Contains(contentType, "application/json") {
		// Parse JSON
		var jsonData interface{}
		if err := json.Unmarshal(content, &jsonData); err != nil {
			return nil, fmt.Errorf("error parsing JSON: %w", err)
		}

		// Extract links from JSON
		links = extractLinksFromJSON(jsonData)

	} else if strings.Contains(contentType, "text/plain") {
		// For plain text, look for URLs using regex
		urlPattern := regexp.MustCompile(`https?://[^\s"']+`)
		matches := urlPattern.FindAllString(string(content), -1)

		// Remove duplicates
		seen := make(map[string]bool)
		for _, match := range matches {
			if !seen[match] {
				seen[match] = true
				links = append(links, match)
			}
		}
	} else {
		return nil, fmt.Errorf("unsupported file type: %s", contentType)
	}

	return links, nil
}

// FilterLinksByRegex filters a slice of links using a regular expression pattern.
// Only links that match the pattern are returned.
// Returns an error if the regex pattern is invalid.
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

// PrintLinksAsJSON prints the links as a JSON array to stdout.
// If JSON marshaling fails, an error message is printed.
func PrintLinksAsJSON(links []string) {
	data, err := json.Marshal(links)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(string(data))
}

// PrintLinksAsNumbered prints the links as a numbered list to stdout.
// Each link is prefixed with its position number in the list.
func PrintLinksAsNumbered(links []string) {
	for i, link := range links {
		fmt.Printf("%d. %s\n", i+1, link)
	}
}

// PrintLinksAsHTML prints the links as an HTML unordered list to stdout.
// Each link is wrapped in an anchor tag that links to itself.
func PrintLinksAsHTML(links []string) {
	fmt.Println("<ul>")
	for _, link := range links {
		fmt.Printf("<li><a href=\"%s\">%s</a></li>\n", link, link)
	}
	fmt.Println("</ul>")
}

// PrintLinksAsText prints the links as plain text to stdout.
// Each link is printed on a new line.
func PrintLinksAsText(links []string) {
	for _, link := range links {
		fmt.Println(link)
	}
}
