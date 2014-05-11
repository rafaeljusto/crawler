package main

import (
	"code.google.com/p/go.net/html"
	"io"
)

var (
	// visitedPages store all pages already visited in a map indexed by the page URL to allow a fast
	// detection of what page was already visited
	visitedPages map[string]Page
)

// Page describes the information stored after a webpage is crawled
type Page struct {
	URL          string   // Address of the page
	Links        []string // List of links for other URLs in this page
	StaticAssets []string // List of static dependencies of this page
}

// Fetcher creates an interface to allow a flexibility on how we retrieve the page data. For tests
// we will simulate the response while in production we will do a HTTP GET
type Fetcher interface {
	Fetch(url string) (io.Reader, error)
}

// Crawl fetch the URL data and try to retrieve all the information from the page. On error a dummy
// Page struct is returned. We are not using pointer on page objects because we want them to be
// destroyed as soon as possible
func Crawl(url string, fetcher Fetcher) (Page, error) {
	page := Page{
		URL: url,
	}

	r, err := fetcher.Fetch(url)
	if err != nil {
		return page, err
	}

	root, err := html.Parse(r)
	if err != nil {
		return page, err
	}

	parseHTML(root, &page)
	return page, nil
}

// parseHTML is an auxiliary function of Crawl function that will travel recursively around the HTML
// document identifying elements to populate the Page object
func parseHTML(node *html.Node, page *Page) {
	if node.Type == html.ElementNode && node.Data == "a" {
		for _, attr := range node.Attr {
			if attr.Key == "href" {
				page.Links = append(page.Links, attr.Val)
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		parseHTML(child, page)
	}
}

// main will control the flow of all go routines that retrieve each crawler
func main() {

}
