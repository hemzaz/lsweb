package downloader

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

type GitHubRelease struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// FetchGitHubReleases fetches the releases of a GitHub repository.
// FetchGitHubReleases fetches the releases of a GitHub repository.
func FetchGitHubReleases(repoURL string, ignoreCert bool) ([]GitHubRelease, error) {
	parts := strings.Split(repoURL, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid GitHub repository URL")
	}

	owner := parts[len(parts)-2]
	repo := parts[len(parts)-1]

	// Create an HTTP client that ignores certificate errors if the flag is set.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreCert},
	}
	httpClient := &http.Client{Transport: tr}

	var client *github.Client
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(context.Background(), ts)
		client = github.NewClient(tc)
	} else {
		client = github.NewClient(httpClient)
	}

	releases, _, err := client.Repositories.ListReleases(context.Background(), owner, repo, nil)
	if err != nil {
		return nil, err
	}

	var ghReleases []GitHubRelease
	for _, release := range releases {
		var assets []Asset
		for _, asset := range release.Assets {
			assets = append(assets, Asset{
				Name:               asset.GetName(),
				BrowserDownloadURL: asset.GetBrowserDownloadURL(),
			})
		}
		ghReleases = append(ghReleases, GitHubRelease{
			TagName: release.GetTagName(),
			Assets:  assets,
		})
	}

	return ghReleases, nil
}

// DownloadFilesSimultaneously downloads files concurrently.
func DownloadFilesSimultaneously(urls []string, ignoreCert bool) {
	for _, url := range urls {
		go DownloadFile(url, ignoreCert)
	}
}

// DownloadFiles downloads files one by one.
func DownloadFiles(urls []string, ignoreCert bool) {
	for _, url := range urls {
		DownloadFile(url, ignoreCert)
	}
}

// DownloadFile downloads a file from a given URL.
func DownloadFile(url string, ignoreCert bool) error {
	// Create an HTTP client that ignores certificate errors if the flag is set.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreCert},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	parts := strings.Split(url, "/")
	filename := parts[len(parts)-1]

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
