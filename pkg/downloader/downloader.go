package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

func DownloadFile(url string, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func DownloadFiles(urls []string, simultaneously bool) {
	if simultaneously {
		var wg sync.WaitGroup
		for _, url := range urls {
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				err := DownloadFile(url, "./") // Filepath can be enhanced
				if err != nil {
					fmt.Println("Error downloading:", url, err)
				}
			}(url)
		}
		wg.Wait()
	} else {
		for _, url := range urls {
			err := DownloadFile(url, "./") // Filepath can be enhanced
			if err != nil {
				fmt.Println("Error downloading:", url, err)
			}
		}
	}
}
