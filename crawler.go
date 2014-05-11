package main

import (
	"code.google.com/p/go.net/html"
	"strings"
)

var (
	// visitedPages store all pages already visited in a map indexed by the page URL to allow a fast
	// detection of what page was already visited
	visitedPages map[string]Page
)

// crawlPage fetch the URL data and try to retrieve all the information from the page. On error a
// dummy Page struct is returned. We are not using pointer on page objects because we want them to
// be destroyed as soon as possible
func crawlPage(url string, fetcher Fetcher) (Page, error) {
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
	if node.Type == html.ElementNode {
		switch node.Data {
		case "a":
			var link Link
			for _, attr := range node.Attr {
				if attr.Key == "href" {
					link.Page = Page{
						URL: attr.Val,
					}

					// TODO: Add pointer of the created page to pages to visit (if not yet visited)
					break
				}
			}

			for child := node.FirstChild; child != nil; child = child.NextSibling {
				// For all texts direct inside a link, we are going to append as labels of this link
				if child.Type == html.TextNode {
					// Normalize the data to detect empty labels, this can occur when we don't close the a tag
					data := strings.TrimSpace(child.Data)
					if len(data) == 0 {
						continue
					}

					// Line break will be the label separator when more than one text is found inside the link
					// tag
					if len(link.Label) > 0 {
						link.Label += "\n"
					}
					link.Label += data
				}
			}

			page.Links = append(page.Links, link)

		case "link":
			for _, attr := range node.Attr {
				if attr.Key == "href" {
					page.StaticAssets = append(page.StaticAssets, attr.Val)
				}
			}

		case "img", "script":
			for _, attr := range node.Attr {
				if attr.Key == "src" {
					page.StaticAssets = append(page.StaticAssets, attr.Val)
				}
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
