package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hemzaz/lsweb/pkg/downloader"
	"github.com/hemzaz/lsweb/pkg/parser"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "lsweb",
		Usage: "List all links from a web page",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "url",
				Aliases:  []string{"u"},
				Usage:    "URL to extract links from",
				Required: true,
			},
			&cli.BoolFlag{
				Name:    "download",
				Aliases: []string{"d"},
				Usage:   "Download files from the extracted links",
			},
			&cli.BoolFlag{
				Name:    "simultaneous",
				Aliases: []string{"s"},
				Usage:   "Download files simultaneously",
			},
			&cli.BoolFlag{
				Name:    "ignore-cert",
				Aliases: []string{"ic"},
				Usage:   "Ignore SSL certificate verification",
			},
			&cli.BoolFlag{
				Name:  "gh",
				Usage: "Fetch releases from a GitHub repository",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"O"},
				Usage:   "Specify the output format. Available formats: json, txt, num, html. Default is txt",
				Value:   "txt",
			},
		},
		Action: func(c *cli.Context) error {
			url := c.String("url")
			downloadFlag := c.Bool("download")
			simFlag := c.Bool("simultaneous")
			ignoreCertFlag := c.Bool("ignore-cert")
			ghFlag := c.Bool("gh")
			outputFormat := c.String("output")

			var links []string
			var err error

			if ghFlag {
				links, err = downloader.FetchGitHubReleases(url)
				if err != nil {
					return err
				}
			} else {
				links, err = parser.ExtractLinksFromURL(url, ignoreCertFlag)
				if err != nil {
					return err
				}
			}

			for _, link := range links {
				fmt.Println(link)
			}

			if outputFormat != "" {
				err = parser.SaveLinksToFile(links, outputFormat)
				if err != nil {
					return fmt.Errorf("failed to save links to file: %v", err)
				}
			}

			if downloadFlag {
				if simFlag {
					err = downloader.DownloadFilesSimultaneously(links, ignoreCertFlag)
				} else {
					err = downloader.DownloadFiles(links, ignoreCertFlag)
				}
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
