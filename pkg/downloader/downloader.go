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

	"github.com/schollz/progressbar/v3"
)

// Default configuration values
var (
	defaultTimeout     = 30 * time.Second
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
	req.Header.Set("User-Agent", "lsweb/1.0")
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
	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024)) // 10MB limit
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

func DownloadFile(url string, ignoreCert bool, showProgress bool) error {
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
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	
	// Add a user-agent to be polite
	req.Header.Set("User-Agent", "lsweb/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error downloading %s: %w", url, err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)
	
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
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)

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

func DownloadFiles(urls []string, ignoreCert bool, showProgress bool) {
	if len(urls) == 0 {
		fmt.Println("No URLs to download")
		return
	}
	
	// Create a context with timeout for the entire operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	
	for i, url := range urls {
		// Check for context cancellation between downloads
		select {
		case <-ctx.Done():
			fmt.Println("Download operation timed out")
			return
		default:
			// Continue with download
		}
		
		fmt.Printf("[%d/%d] Downloading: %s\n", i+1, len(urls), url)
		err := DownloadFile(url, ignoreCert, showProgress)
		if err != nil {
			fmt.Printf("Error downloading %s: %v\n", url, err)
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
	
	fmt.Printf("Download complete: %d/%d files\n", len(urls), len(urls))
}

func DownloadFilesSimultaneously(urls []string, ignoreCert bool) {
	// Create a semaphore to limit concurrency
	maxConcurrent := maxConcurrentDownloads
	sem := make(chan struct{}, maxConcurrent)
	
	// Use a mutex to protect file creation
	var mu sync.Mutex
	
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
			mu.Unlock()
			
			// Custom download to use our unique filename
			client := &http.Client{}
			if ignoreCert {
				client.Transport = &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
			}
			
			resp, err := client.Get(url)
			if err != nil {
				fmt.Println("Error downloading file:", err)
				return
			}
			defer resp.Body.Close()
			
			file, err := os.Create(filename)
			if err != nil {
				fmt.Println("Error creating file:", err)
				return
			}
			defer file.Close()
			
			bar := progressbar.DefaultBytes(
				resp.ContentLength,
				"downloading "+filename,
			)
			_, err = io.Copy(io.MultiWriter(file, bar), resp.Body)
			if err != nil {
				fmt.Println("Error writing file:", err)
			}
		}(url)
	}
	wg.Wait()
}
