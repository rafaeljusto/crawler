package crawler

import (
	"code.google.com/p/go.net/html"
	"strings"
	"sync"
)

var (
	// visitedPages store all pages already visited in a map, so that if we found a link for the same
	// page again, we just pick on the map the same object address. The function that prints the page
	// is responsable for detecting cycle loops
	visitedPages map[string]*Page

	// visitedPagesLock allows visitedPages to be manipulated safely by go routines
	visitedPagesLock sync.Mutex
)

func init() {
	visitedPages = make(map[string]*Page)
}

// Crawl check all pages of the URL managing go routines
func Crawl(url string, fetcher Fetcher) (*Page, error) {
	page := &Page{
		URL: url,
	}

	var wg sync.WaitGroup
	fail := make(chan error)

	wg.Add(1)
	go crawlPage(url, page, fetcher, &wg, fail)

	done := make(chan bool)
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Everything worked fine

	case err := <-fail:
		return nil, err
	}

	return page, nil
}

// Crawl fetch the URL data and try to retrieve all the information from the page,
// filling the page pointer on successful return
func crawlPage(url string, page *Page, fetcher Fetcher, wg *sync.WaitGroup, fail chan error) {
	defer wg.Done()

	visitedPagesLock.Lock()
	visitedPages[page.URL] = page
	visitedPagesLock.Unlock()

	r, err := fetcher.Fetch(page.URL)
	if err != nil {
		fail <- err
		return
	}

	root, err := html.Parse(r)
	if err != nil {
		fail <- err
		return
	}

	parseHTML(url, root, page, fetcher, wg, fail)
}

// parseHTML is an auxiliary function of Crawl function that will travel recursively around the HTML
// document identifying elements to populate the Page object
func parseHTML(url string, node *html.Node, page *Page, fetcher Fetcher, wg *sync.WaitGroup, fail chan error) {
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
					linkURL = url + "/" + linkURL
				}

				// Check if we already processed this page, if so add the pointer of the page, otherwise
				// set the page to be processed if is in the same domain
				if strings.HasPrefix(linkURL, url) {
					if p, found := visitedPages[linkURL]; found {
						link.Page = p

					} else {
						link.Page = &Page{
							URL: linkURL,
						}

						wg.Add(1)
						go crawlPage(url, link.Page, fetcher, wg, fail)
					}

				} else {
					// Outside the domain
					link.Page = &Page{
						URL: linkURL,
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
		parseHTML(url, child, page, fetcher, wg, fail)
	}
}
