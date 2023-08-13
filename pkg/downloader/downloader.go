package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/schollz/progressbar/v3"
)

// DownloadFile downloads a file from the given URL and saves it to the current directory.
func DownloadFile(url string, showProgress bool) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	filename := filepath.Base(url)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if showProgress {
		bar := progressbar.DefaultBytes(
			resp.ContentLength,
			"downloading",
		)
		_, err = io.Copy(io.MultiWriter(file, bar), resp.Body)
	} else {
		_, err = io.Copy(file, resp.Body)
	}

	return err
}

// DownloadFiles downloads multiple files.
func DownloadFiles(urls []string, showProgress bool) {
	for _, url := range urls {
		err := DownloadFile(url, showProgress)
		if err != nil {
			fmt.Println("Error downloading file:", err)
		}
	}
}

// DownloadFilesSimultaneously downloads multiple files in parallel.
func DownloadFilesSimultaneously(urls []string) {
	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			err := DownloadFile(url, true)
			if err != nil {
				fmt.Println("Error downloading file:", err)
			}
		}(url)
	}
	wg.Wait()
}
