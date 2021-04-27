package main

import (
	"bufio"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

type Feed struct {
	Title    string
	URL      string
	Date     string
	Articles []Article
}

type Article struct {
	Title       string
	URL         string
	Description string
}

type HTMLDocument struct {
	Body struct {
		H1 string
	}
}

var articlesFile string

//go:embed rss.xml.tmpl
var rssTemplate string

func main() {
	feed := Feed{
		Date: time.Now().Format(time.RFC822Z),
	}

	flag.StringVar(&feed.Title, "title", os.Getenv("RSS_TITLE"), "title of the RSS feed")
	flag.StringVar(&feed.URL, "url", os.Getenv("RSS_URL"), "URL where the feed can be downloaded (required)")
	flag.StringVar(&articlesFile, "articles", os.Getenv("RSS_ARTICLES"), "file containing a list of article URLs (required)")
	flag.Parse()

	if feed.Title == "" {
		feed.Title = "RSS Feed"
	}
	if feed.URL == "" {
		fmt.Println("Error: a URL is required")
		flag.Usage()
		os.Exit(2)
	}
	if articlesFile == "" {
		fmt.Println("Error: an articles file is required")
		flag.Usage()
		os.Exit(2)
	}

	if input, err := os.Open(articlesFile); err == nil {
		scanner := bufio.NewScanner(input)
		for scanner.Scan() {
			url := scanner.Text()
			resp, err := http.Get(url)
			if err != nil {
				log.Error(err)
				continue
			}

			doc, err := html.Parse(resp.Body)
			if err != nil {
				log.Error(err)
				continue
			}

			feed.Articles = append(feed.Articles,
				Article{
					URL:   url,
					Title: getTitle(doc),
				},
			)
		}
		check(scanner.Err())
	} else {
		if errors.Is(err, fs.ErrNotExist) {
			log.Warn(err)
		} else {
			log.Fatal(err)
		}
	}

	tmpl, err := template.New("rss").Parse(rssTemplate)
	check(err)

	err = tmpl.Execute(os.Stdout, feed)
	check(err)
}

func getTitle(root *html.Node) string {
	for n := root.FirstChild; n != nil; {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "html", "head":
				n = n.FirstChild
				continue
			case "title":
				return n.FirstChild.Data
			}
		}
		n = n.NextSibling
	}
	return ""
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
