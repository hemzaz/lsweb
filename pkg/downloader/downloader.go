package downloader

import (
	"fmt"
	// other necessary imports for downloading
)

// DownloadFiles downloads the provided links.
// If the 'simultaneous' flag is true, it will download the files simultaneously.
func DownloadFiles(links []string, simultaneous bool) {
	if simultaneous {
		DownloadFilesSimultaneously(links)
		return
	}
	// Logic for non-simultaneous download
	for _, link := range links {
		// Download logic for each link
		fmt.Println("Downloading:", link)
	}
}

// DownloadFilesSimultaneously downloads the provided links simultaneously.
func DownloadFilesSimultaneously(links []string) {
	// Logic for simultaneous download
	for _, link := range links {
		go func(link string) {
			// Download logic for each link
			fmt.Println("Simultaneously downloading:", link)
		}(link)
	}
}
