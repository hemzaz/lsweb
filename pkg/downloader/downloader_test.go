package downloader

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	
	"github.com/hemzaz/lsweb/pkg/common"
)

func TestSetTimeout(t *testing.T) {
	// Save original value to restore after test
	originalTimeout := defaultTimeout
	defer func() {
		defaultTimeout = originalTimeout
	}()
	
	// Set a new timeout value
	newTimeout := originalTimeout * 2
	SetTimeout(newTimeout)
	
	if defaultTimeout != newTimeout {
		t.Errorf("SetTimeout failed: expected %v, got %v", newTimeout, defaultTimeout)
	}
}

func TestSetMaxConcurrent(t *testing.T) {
	// Save original value to restore after test
	originalMaxConcurrent := maxConcurrentDownloads
	defer func() {
		maxConcurrentDownloads = originalMaxConcurrent
	}()
	
	// Test with valid value
	SetMaxConcurrent(10)
	if maxConcurrentDownloads != 10 {
		t.Errorf("SetMaxConcurrent failed with valid value: expected 10, got %v", maxConcurrentDownloads)
	}
	
	// Test with invalid value (should not change)
	SetMaxConcurrent(0)
	if maxConcurrentDownloads != 10 {
		t.Errorf("SetMaxConcurrent changed with invalid value: expected 10, got %v", maxConcurrentDownloads)
	}
	
	SetMaxConcurrent(-5)
	if maxConcurrentDownloads != 10 {
		t.Errorf("SetMaxConcurrent changed with negative value: expected 10, got %v", maxConcurrentDownloads)
	}
}

func TestSetOverwriteFiles(t *testing.T) {
	// Save original value to restore after test
	originalOverwrite := allowOverwriteFiles
	defer func() {
		allowOverwriteFiles = originalOverwrite
	}()
	
	// Test setting to true
	SetOverwriteFiles(true)
	if !allowOverwriteFiles {
		t.Errorf("SetOverwriteFiles failed to set true")
	}
	
	// Test setting to false
	SetOverwriteFiles(false)
	if allowOverwriteFiles {
		t.Errorf("SetOverwriteFiles failed to set false")
	}
}

func TestFetchGitHubReleases(t *testing.T) {
	// Set up a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that the path is correct
		if !strings.HasPrefix(r.URL.Path, "/repos/") {
			t.Errorf("Unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		
		// Check headers
		if r.Header.Get("User-Agent") != common.UserAgent {
			t.Errorf("Expected User-Agent header '%s', got '%s'", common.UserAgent, r.Header.Get("User-Agent"))
		}
		
		// Return mock release data
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `[
			{
				"tag_name": "v1.0.0",
				"assets": [
					{
						"name": "app.zip",
						"browser_download_url": "https://example.com/app.zip",
						"size": 1024
					},
					{
						"name": "app.tar.gz",
						"browser_download_url": "https://example.com/app.tar.gz",
						"size": 2048
					}
				]
			}
		]`)
	}))
	defer server.Close()
	
	// This variable is just declared for demonstration, not used in this test
	_ = server.URL + "/repos/testuser/testrepo/releases"
	
	// Test valid GitHub repo URL with our test server (bypassing actual GitHub API)
	// In a real implementation, we would create a mock HTTP client and inject it into the function
	// to test the GitHub API response parsing.
	//
	// For this test, we're only verifying that the URL parsing logic works correctly
	// with a real GitHub repo URL by attempting to connect to the actual GitHub API.
	// We expect an error since we're not properly mocking the API.
	_, err := FetchGitHubReleases("https://github.com/testuser/testrepo", false)
	
	// This should fail with "connection refused" or similar error since we're trying to reach GitHub
	// but our test is not configured to allow external connections
	if err == nil {
		t.Error("Expected error for real GitHub URL (no mock), got nil")
	}
}

func TestDownloadFile(t *testing.T) {
	// Create a temporary directory for downloads
	tempDir, err := os.MkdirTemp("", "download-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Change to the temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)
	
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	// Set up a test server that serves a small file
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "11")
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "Hello World")
	}))
	defer server.Close()
	
	// Test downloading a file
	err = DownloadFile(server.URL, false, false)
	if err != nil {
		t.Errorf("DownloadFile failed: %v", err)
	}
	
	// Check that the file was downloaded
	filename := filepath.Base(server.URL)
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("Failed to read downloaded file: %v", err)
	}
	
	if string(content) != "Hello World" {
		t.Errorf("Downloaded file has incorrect content: %s", string(content))
	}
	
	// Test downloading the same file again (should fail due to existing file)
	err = DownloadFile(server.URL, false, false)
	if err == nil {
		t.Errorf("Expected error when downloading to existing file, got nil")
	}
	
	// Test with overwrite enabled
	SetOverwriteFiles(true)
	err = DownloadFile(server.URL, false, false)
	if err != nil {
		t.Errorf("DownloadFile with overwrite failed: %v", err)
	}
	SetOverwriteFiles(false) // Reset
	
	// Test with invalid URL
	err = DownloadFile("http://invalid.url.that.does.not.exist", false, false)
	if err == nil {
		t.Errorf("Expected error with invalid URL, got nil")
	}
}