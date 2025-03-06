package parser

import (
	"bytes"
	"io"
	"net/url"
	"os"
	"strings"
	"testing"
	
	"golang.org/x/net/html"
)

func TestRemoveDuplicateLinks(t *testing.T) {
	tests := []struct {
		name     string
		links    []string
		expected []string
	}{
		{
			name:     "Empty links",
			links:    []string{},
			expected: []string{},
		},
		{
			name:     "No duplicates",
			links:    []string{"https://example.com/1", "https://example.com/2"},
			expected: []string{"https://example.com/1", "https://example.com/2"},
		},
		{
			name:     "With duplicates",
			links:    []string{"https://example.com/1", "https://example.com/1", "https://example.com/2"},
			expected: []string{"https://example.com/1", "https://example.com/2"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := removeDuplicateLinks(tc.links)
			
			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d links, got %d", len(tc.expected), len(result))
				return
			}
			
			// Create map from expected slice for easier comparison
			expectedMap := make(map[string]bool)
			for _, link := range tc.expected {
				expectedMap[link] = true
			}
			
			// Check each result is in expected map
			for _, link := range result {
				if !expectedMap[link] {
					t.Errorf("Unexpected link in result: %s", link)
				}
			}
		})
	}
}

func TestFilterLinksByRegex(t *testing.T) {
	tests := []struct {
		name      string
		links     []string
		regex     string
		expected  []string
		expectErr bool
	}{
		{
			name:      "Filter by file extension",
			links:     []string{"https://example.com/file.pdf", "https://example.com/file.txt", "https://example.com/image.jpg"},
			regex:     "\\.pdf$",
			expected:  []string{"https://example.com/file.pdf"},
			expectErr: false,
		},
		{
			name:      "Invalid regex",
			links:     []string{"https://example.com/file.pdf"},
			regex:     "[", // Missing closing bracket makes this a truly invalid regex
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := FilterLinksByRegex(tc.links, tc.regex)
			
			if tc.expectErr && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			
			if !tc.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if tc.expectErr {
				return
			}
			
			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d links, got %d", len(tc.expected), len(result))
				return
			}
			
			for i, link := range result {
				if link != tc.expected[i] {
					t.Errorf("Expected %s at position %d, got %s", tc.expected[i], i, link)
				}
			}
		})
	}
}

func TestPrintLinksAsFunctions(t *testing.T) {
	testLinks := []string{"https://example.com/1", "https://example.com/2"}
	
	// Capture standard output for testing
	// Save original stdout
	oldStdout := os.Stdout
	
	// Create a pipe to capture output
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Test PrintLinksAsJSON
	PrintLinksAsJSON(testLinks)
	
	// Close writer to get output
	w.Close()
	os.Stdout = oldStdout
	
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	
	// Verify JSON output
	expectedJSON := `["https://example.com/1","https://example.com/2"]`
	if strings.TrimSpace(buf.String()) != expectedJSON {
		t.Errorf("PrintLinksAsJSON output mismatch.\nExpected: %s\nGot: %s", 
			expectedJSON, strings.TrimSpace(buf.String()))
	}
	
	// Create a new pipe for the next test
	r, w, _ = os.Pipe()
	os.Stdout = w
	
	// Test PrintLinksAsText
	PrintLinksAsText(testLinks)
	
	// Close writer to get output
	w.Close()
	os.Stdout = oldStdout
	
	buf.Reset()
	_, _ = io.Copy(&buf, r)
	
	// Verify text output
	expectedText := "https://example.com/1\nhttps://example.com/2"
	if strings.TrimSpace(buf.String()) != expectedText {
		t.Errorf("PrintLinksAsText output mismatch.\nExpected: %s\nGot: %s", 
			expectedText, strings.TrimSpace(buf.String()))
	}
	
	// We don't test all output formats to keep the test simpler
}

func TestExtractLinksFromJSON(t *testing.T) {
	// Test with a map containing URLs
	jsonData := map[string]interface{}{
		"url": "https://example.com/1",
		"nested": map[string]interface{}{
			"url": "https://example.com/2",
		},
		"items": []interface{}{
			"https://example.com/3",
			map[string]interface{}{
				"url": "https://example.com/4",
			},
		},
	}
	
	links := extractLinksFromJSON(jsonData)
	
	// Should find all 4 URLs
	expectedCount := 4
	if len(links) != expectedCount {
		t.Errorf("Expected %d links, got %d", expectedCount, len(links))
	}
	
	// Check extracted URLs
	expectedLinks := map[string]bool{
		"https://example.com/1": true,
		"https://example.com/2": true,
		"https://example.com/3": true,
		"https://example.com/4": true,
	}
	
	for _, link := range links {
		if !expectedLinks[link] {
			t.Errorf("Unexpected link: %s", link)
		}
	}
}

func TestExtractLinksFromHTML(t *testing.T) {
	htmlContent := `
<!DOCTYPE html>
<html>
<head>
    <title>Test Page</title>
</head>
<body>
    <a href="https://example.com/page1">Page 1</a>
    <a href="https://example.com/page2">Page 2</a>
    <a href="#section">Section Link</a>
    <a href="javascript:void(0)">JavaScript Link</a>
    <a href="mailto:test@example.com">Email Link</a>
    <a href="page3">Relative Link</a>
</body>
</html>
`
	baseURL, _ := url.Parse("https://example.com")
	
	// Parse HTML document
	doc, err := html.Parse(bytes.NewReader([]byte(htmlContent)))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	
	links, malformed := extractLinksFromHTML(doc, baseURL)
	
	// Count all links found
	// Update test to expect 4 links (the section link is included as full URL)
	expectedLinks := map[string]bool{
		"https://example.com/page1": true,
		"https://example.com/page2": true,
		"https://example.com/page3": true,
		"https://example.com#section": true,
	}
	
	// Check if we got the right number of links
	expectedCount := len(expectedLinks)
	if len(links) != expectedCount {
		t.Errorf("Expected %d links, got %d", expectedCount, len(links))
	}
	
	for _, link := range links {
		if !expectedLinks[link] {
			t.Errorf("Unexpected link: %s", link)
		}
	}
	
	// No malformed URLs expected
	if len(malformed) > 0 {
		t.Errorf("Expected 0 malformed URLs, got %d", len(malformed))
	}
}

func TestExtractLinksFromFile(t *testing.T) {
	// Create a temporary HTML file
	tempFile, err := os.CreateTemp("", "test-*.html")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	
	// Write HTML content with links
	htmlContent := `<!DOCTYPE html>
<html>
<body>
  <a href="https://example.com/page1">Page 1</a>
  <a href="https://example.com/page2">Page 2</a>
</body>
</html>`
	
	if _, err := tempFile.Write([]byte(htmlContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	
	// Extract links from the file
	links, err := ExtractLinksFromFile(tempFile.Name())
	if err != nil {
		t.Fatalf("ExtractLinksFromFile failed: %v", err)
	}
	
	// Verify extracted links
	expectedCount := 2
	if len(links) != expectedCount {
		t.Errorf("Expected %d links, got %d", expectedCount, len(links))
	}
	
	// Check links content
	expectedLinks := map[string]bool{
		"https://example.com/page1": true,
		"https://example.com/page2": true,
	}
	
	for _, link := range links {
		if !expectedLinks[link] {
			t.Errorf("Unexpected link: %s", link)
		}
	}
	
	// Test with non-existent file
	_, err = ExtractLinksFromFile("/non/existent/file.html")
	if err == nil {
		t.Errorf("Expected error for non-existent file, got nil")
	}
}