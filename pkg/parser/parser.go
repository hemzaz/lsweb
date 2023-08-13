package parser

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
)

var LinkRegex = regexp.MustCompile(`https?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*(),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)

func ExtractLinksFromWebPage(body []byte) []string {
	links := LinkRegex.FindAllString(string(body), -1)
	return links
}

func ExtractLinksFromURL(url string, ignoreCert bool) ([]string, error) {
	client := &http.Client{}
	if ignoreCert {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	links := ExtractLinksFromWebPage(body)
	return links, nil
}

func SaveLinksToFile(links []string, format string) error {
	var data []byte
	var err error

	switch format {
	case "json":
		data, err = json.Marshal(links)
		if err != nil {
			return err
		}
	case "txt":
		for _, link := range links {
			data = append(data, link...)
			data = append(data, '\n')
		}
	case "num":
		for i, link := range links {
			line := fmt.Sprintf("%d. %s\n", i+1, link)
			data = append(data, line...)
		}
	case "html":
		data = []byte("<html><body><ul>")
		for _, link := range links {
			line := fmt.Sprintf("<li><a href='%s'>%s</a></li>", link, link)
			data = append(data, line...)
		}
		data = append(data, []byte("</ul></body></html>")...)
	default:
		return errors.New("unsupported format")
	}

	err = os.WriteFile("links."+format, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
