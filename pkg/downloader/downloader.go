package downloader

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/schollz/progressbar/v3"
)

const MaxConcurrentDownloads = 255

var errorsMutex sync.Mutex

func DownloadFiles(urls []string, simultaneously bool) error {
	if simultaneously {
		return downloadFilesConcurrently(urls)
	}
	for _, url := range urls {
		err := downloadFile(url)
		if err != nil {
			log.Printf("Failed to download %s: %v", url, err)
			return fmt.Errorf("failed to download %s: %w", url, err)
		}
	}
	return nil
}

func downloadFilesConcurrently(urls []string) error {
	sem := make(chan struct{}, MaxConcurrentDownloads)
	var wg sync.WaitGroup
	var errors []error

	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			sem <- struct{}{}
			err := downloadFile(u)
			if err != nil {
				errorsMutex.Lock()
				errors = append(errors, fmt.Errorf("failed to download %s concurrently: %w", u, err))
				errorsMutex.Unlock()
			}
			<-sem
		}(url)
	}
	wg.Wait()

	if len(errors) > 0 {
		for _, err := range errors {
			log.Println(err)
		}
		return fmt.Errorf("multiple errors occurred during concurrent downloads: %v", errors[0]) // return the first error for simplicity
	}
	return nil
}

func downloadFile(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s: HTTP %d", url, resp.StatusCode)
	}

	filename := filepath.Base(url)
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		filename = "new_" + filename
	}
	out, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer out.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"downloading",
	)
	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filename, err)
	}
	return nil
}
