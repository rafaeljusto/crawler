package crawler

import (
	"code.google.com/p/go.net/html"
	"strings"
)

var (
	// visitedPages store all pages already visited
	visitedPages []string

	// pagesToVisit store all the pages that need to be analyzed yet
	pagesToVisit chan *Page
)

func init() {
	// We will keep a waiting list in the channel with the size of the number of go routines
	// processing the pages
	pagesToVisit = make(chan *Page, 10)
}

// Crawl check all pages of the URL managing go routines
func Crawl(url string, fetcher Fetcher) (Page, error) {
	return crawlPage(url, fetcher)
}

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

					// TODO: Add pointer of the created page to pages to visit only if not yet visited and is
					// inside the same domain
					//pagesToVisit <- &link.Page
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
