package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

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

	input, err := os.Open(articlesFile)
	check(err)

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		url := scanner.Text()
		resp, err := http.Get(url)
		check(err)

		doc, err := html.Parse(resp.Body)
		check(err)

		feed.Articles = append(feed.Articles,
			Article{
				URL:   url,
				Title: getTitle(doc),
			},
		)
	}
	check(scanner.Err())

	rssTemplateFile, err := os.Open("rss.xml.tmpl")
	check(err)

	rssTemplate, err := io.ReadAll(rssTemplateFile)
	check(err)

	tmpl, err := template.New("rss").Parse(string(rssTemplate))
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
