// Package downloader provides functions for downloading files from URLs
// with support for progress tracking, batch downloads, and GitHub release assets.
package downloader

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	// Third-party dependencies
	"github.com/schollz/progressbar/v3"
	
	// Internal dependencies
	"github.com/hemzaz/lsweb/pkg/common"
)

// Default configuration values
var (
	defaultTimeout         = common.DefaultTimeout
	maxConcurrentDownloads = 5
	allowOverwriteFiles    = false
)

// SetTimeout sets the timeout for HTTP requests
func SetTimeout(timeout time.Duration) {
	defaultTimeout = timeout
}

// SetMaxConcurrent sets the maximum number of concurrent downloads
func SetMaxConcurrent(max int) {
	if max > 0 {
		maxConcurrentDownloads = max
	}
}

// SetOverwriteFiles sets whether to overwrite existing files
func SetOverwriteFiles(overwrite bool) {
	allowOverwriteFiles = overwrite
}

type GitHubRelease struct {
	URL  string `json:"html_url"`
	Name string `json:"name"`
}

// FetchGitHubReleases retrieves download URLs for assets from all releases in a GitHub repository.
// It parses the repository URL to extract owner and repo name, then queries the GitHub API.
// Returns a slice of all asset download URLs or an error if the fetch fails.
// The ignoreCert parameter can be used to skip TLS certificate validation.
func FetchGitHubReleases(repoURL string, ignoreCert bool) ([]string, error) {
	// Parse the URL properly
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	
	// Extract path components
	pathParts := strings.Split(strings.TrimPrefix(parsedURL.Path, "/"), "/")
	if len(pathParts) < 2 {
		return nil, fmt.Errorf("invalid GitHub repository URL: expected format github.com/{user}/{repo}")
	}
	
	user, repo := pathParts[0], pathParts[1]
	// Remove any trailing .git from repo name
	repo = strings.TrimSuffix(repo, ".git")

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", user, repo)

	// Set up the client with timeout
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreCert},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   defaultTimeout,
	}

	// Create request with context
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	
	// Add user-agent and accept headers required by GitHub API
	req.Header.Set("User-Agent", common.UserAgent)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching GitHub releases: %w", err)
	}
	defer resp.Body.Close()

	// Check for rate limiting
	if resp.StatusCode == 403 && resp.Header.Get("X-RateLimit-Remaining") == "0" {
		resetTime := resp.Header.Get("X-RateLimit-Reset")
		return nil, fmt.Errorf("GitHub API rate limit exceeded. Reset at %s", resetTime)
	}

	// Check for other error status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned non-success status: %d %s", resp.StatusCode, resp.Status)
	}

	// Limit body size for safety
	body, err := io.ReadAll(io.LimitReader(resp.Body, common.MaxContentSize))
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var releases []struct {
		Assets []struct {
			BrowserDownloadURL string `json:"browser_download_url"`
			Name              string `json:"name"`
			Size              int    `json:"size"`
		} `json:"assets"`
		TagName string `json:"tag_name"`
	}

	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("error parsing GitHub response: %w", err)
	}

	var downloadLinks []string
	for _, release := range releases {
		for _, asset := range release.Assets {
			downloadLinks = append(downloadLinks, asset.BrowserDownloadURL)
		}
	}

	if len(downloadLinks) == 0 {
		return nil, fmt.Errorf("no release assets found for %s/%s", user, repo)
	}

	return downloadLinks, nil
}

// DownloadFile downloads a single file from the specified URL to the current directory.
// The file is named based on the last part of the URL path.
// If showProgress is true, it displays a progress bar during download.
// The ignoreCert parameter can be used to skip TLS certificate validation.
// Returns an error if download fails, file already exists, or file is too large.
func DownloadFile(url string, ignoreCert bool, showProgress bool) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	
	// Create a client with timeout
	client := &http.Client{
		Timeout: defaultTimeout,
	}
	if ignoreCert {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	// Create a request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	
	// Add a user-agent to be polite
	req.Header.Set("User-Agent", common.UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error downloading %s: %w", url, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Error closing response body: %v\n", closeErr)
		}
	}()
	
	// Check for successful status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("server returned non-success status: %d %s", resp.StatusCode, resp.Status)
	}

	// Check content size if available
	if resp.ContentLength > 1024*1024*1000 { // 1GB
		return fmt.Errorf("file too large (%.2f GB). Use a dedicated download tool instead", float64(resp.ContentLength)/(1024*1024*1024))
	}

	filename := filepath.Base(url)
	
	// Check if file already exists
	if !allowOverwriteFiles {
		if _, err := os.Stat(filename); err == nil {
			return fmt.Errorf("file %s already exists, skipping download (use -overwrite to override)", filename)
		}
	}
	
	// Create the file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file %s: %w", filename, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Error closing file: %v\n", closeErr)
		}
	}()

	if showProgress {
		bar := progressbar.DefaultBytes(
			resp.ContentLength,
			"downloading "+filename,
		)
		_, err = io.Copy(io.MultiWriter(file, bar), resp.Body)
	} else {
		_, err = io.Copy(file, resp.Body)
	}

	if err != nil {
		// On error, clean up the partial file
		os.Remove(filename)
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

// DownloadFiles downloads multiple files sequentially from the provided URLs.
// If showProgress is true, it displays a progress bar for each download.
// The ignoreCert parameter can be used to skip TLS certificate validation.
// The function continues to the next URL if a download fails and returns an error
// at the end if any downloads failed.
func DownloadFiles(urls []string, ignoreCert bool, showProgress bool) error {
	if len(urls) == 0 {
		return fmt.Errorf("no URLs to download")
	}
	
	// Create a context with timeout for the entire operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	
	var failedCount int
	
	for i, url := range urls {
		// Check for context cancellation between downloads
		select {
		case <-ctx.Done():
			return fmt.Errorf("download operation timed out after %d/%d files", i, len(urls))
		default:
			// Continue with download
		}
		
		fmt.Printf("[%d/%d] Downloading: %s\n", i+1, len(urls), url)
		err := DownloadFile(url, ignoreCert, showProgress)
		if err != nil {
			fmt.Printf("Error downloading %s: %v\n", url, err)
			failedCount++
			// Continue with next URL rather than stopping
		} else if showProgress {
			// Add a newline after progress bar completes
			fmt.Println()
		}
		
		// Add a small delay between downloads to be kind to servers
		if i < len(urls)-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}
	
	fmt.Printf("Download complete: %d/%d files\n", len(urls)-failedCount, len(urls))
	
	if failedCount > 0 {
		return fmt.Errorf("%d/%d downloads failed", failedCount, len(urls))
	}
	
	return nil
}

// DownloadFilesSimultaneously downloads multiple files concurrently from the provided URLs.
// It uses a semaphore to limit the number of concurrent downloads to maxConcurrentDownloads.
// The ignoreCert parameter can be used to skip TLS certificate validation.
// The showProgress parameter determines whether to display progress bars (defaults to true).
// Returns an error if any download fails, including the count of failed downloads.
func DownloadFilesSimultaneously(urls []string, ignoreCert bool, showProgress bool) error {
	if len(urls) == 0 {
		return fmt.Errorf("no URLs to download")
	}

	// Create a semaphore to limit concurrency
	maxConcurrent := maxConcurrentDownloads
	sem := make(chan struct{}, maxConcurrent)
	
	// Use a mutex to protect file creation
	var mu sync.Mutex
	
	// Track errors
	errorChan := make(chan error, len(urls))
	
	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			// Acquire semaphore
			sem <- struct{}{}
			defer func() {
				// Release semaphore when done
				<-sem
				wg.Done()
			}()
			
			// Use a more atomic file creation approach
			mu.Lock()
			filename := filepath.Base(url)
			
			// Check if file already exists
			if !allowOverwriteFiles {
				if _, err := os.Stat(filename); err == nil {
					// File exists, create a unique name
					for i := 1; ; i++ {
						newName := fmt.Sprintf("%s.%d", filename, i)
						if _, err := os.Stat(newName); os.IsNotExist(err) {
							filename = newName
							break
						}
					}
				}
			}
			mu.Unlock()
			
			// Custom download to use our unique filename
			client := &http.Client{
				Timeout: defaultTimeout,
			}
			if ignoreCert {
				client.Transport = &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
			}
			
			// Create a request with context
			ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
			defer cancel()
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				errorChan <- fmt.Errorf("error creating request for %s: %w", url, err)
				return
			}
			
			// Add a user-agent to be polite
			req.Header.Set("User-Agent", common.UserAgent)
			
			resp, err := client.Do(req)
			if err != nil {
				errorChan <- fmt.Errorf("error downloading %s: %w", url, err)
				return
			}
			defer func() {
				if closeErr := resp.Body.Close(); closeErr != nil {
					fmt.Printf("Error closing response body: %v\n", closeErr)
				}
			}()
			
			// Check for successful status code
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				errorChan <- fmt.Errorf("server returned non-success status for %s: %d %s", url, resp.StatusCode, resp.Status)
				return
			}
			
			file, err := os.Create(filename)
			if err != nil {
				errorChan <- fmt.Errorf("error creating file %s: %w", filename, err)
				return
			}
			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					fmt.Printf("Error closing file: %v\n", closeErr)
				}
			}()
			
			if showProgress {
				bar := progressbar.DefaultBytes(
					resp.ContentLength,
					"downloading "+filename,
				)
				_, err = io.Copy(io.MultiWriter(file, bar), resp.Body)
			} else {
				_, err = io.Copy(file, resp.Body)
			}
			if err != nil {
				// Clean up partial file
				os.Remove(filename)
				errorChan <- fmt.Errorf("error writing to file %s: %w", filename, err)
			}
		}(url)
	}
	
	// Wait for all downloads to complete
	wg.Wait()
	close(errorChan)
	
	// Collect errors
	var downloadErrors []string
	for err := range errorChan {
		downloadErrors = append(downloadErrors, err.Error())
	}
	
	if len(downloadErrors) > 0 {
		// Return a concatenated error with all details
		return fmt.Errorf("%d download(s) failed. Errors: %s", 
			len(downloadErrors), 
			strings.Join(downloadErrors, "; "))
	}
	
	return nil
}
