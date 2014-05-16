// Copyright 2014 Rafael Dantas Justo. All rights reserved.
// Use of this source code is governed by a GPL
// license that can be found in the LICENSE file.

// crawler verify a HTML page and list the resources
package crawler

import (
	"code.google.com/p/go.net/html"
	"strings"
)

const (
	// Maximum number of running go routines on the a crawl. Got code from
	// http://golang.org/doc/effective_go.html#channels. This is necessary because we can go
	// out of descriptors if we start creating go routines with no limit. There's also a
	// great post about this on http://burke.libbey.me/conserving-file-descriptors-in-go/
	maxOutstanding = 200
)

var (
	// Semaphore to control go routines execution
	sem = make(chan int, maxOutstanding)
)

func init() {
	for i := 0; i < maxOutstanding; i++ {
		sem <- 1
	}
}

// Crawl check all pages of the URL managing go routines
func Crawl(url string, fetcher Fetcher) (*Page, error) {
	page := &Page{
		URL: url,
	}

	context := NewCrawlerContext(url, fetcher)

	context.WG.Add(1)
	go crawlPage(context, page)
	context.WG.Wait()

	return page, nil
}

// Crawl fetch the URL data and try to retrieve all the information from the page,
// filling the page pointer on successful return
func crawlPage(context *CrawlerContext, page *Page) {
	<-sem

	defer func() {
		sem <- 1
		context.WG.Done()
	}()

	context.VisitPage(page)

	r, err := context.Fetcher.Fetch(page.URL)
	if err != nil {
		page.Fail = true
		return
	}

	root, err := html.Parse(r)
	if err != nil {
		page.Fail = true
		return
	}

	parseHTML(context, root, page)
}

// parseHTML is an auxiliary function of Crawl function that will travel recursively
// around the HTML document identifying elements to populate the Page object
func parseHTML(context *CrawlerContext, node *html.Node, page *Page) {
	if node.Type == html.ElementNode {
		switch node.Data {
		case "a":
			var link Link
			for _, attr := range node.Attr {
				if attr.Key != "href" {
					continue
				}

				linkURL := strings.TrimSpace(attr.Val)
				if strings.HasPrefix(linkURL, "/") {
					linkURL = context.Domain + linkURL
				}

				// Check if we already processed this page, to avoid a cyclic recursion when
				// showing the results we aren't going to add a reference for the already analyzed
				// page
				if page, visited := context.URLWasVisited(linkURL); visited {
					link.Page = page
					link.CyclicPage = true

				} else {
					link.Page = &Page{
						URL: linkURL,
					}

					if strings.HasPrefix(linkURL, context.Domain) {
						context.WG.Add(1)
						go crawlPage(context, link.Page)
					}
				}

				// TODO: Not checking when the link has a relative path
				break
			}

			for child := node.FirstChild; child != nil; child = child.NextSibling {
				// For all texts direct inside a link, we are going to append as labels of this link
				if child.Type != html.TextNode {
					continue
				}

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

			if len(link.Label) == 0 {
				link.Label = "<no label>"
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
		parseHTML(context, child, page)
	}
}
