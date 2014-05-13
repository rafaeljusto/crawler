package crawler

import (
	"code.google.com/p/go.net/html"
	"strings"
)

// Crawl check all pages of the URL managing go routines
func Crawl(url string, fetcher Fetcher) (*Page, error) {
	page := &Page{
		URL: url,
	}

	context := NewCrawlerContext(url, fetcher)

	context.WG.Add(1)
	go crawlPage(context, page)

	done := make(chan bool)
	go func() {
		context.WG.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Everything worked fine

	case err := <-context.Fail:
		return nil, err
	}

	return page, nil
}

// Crawl fetch the URL data and try to retrieve all the information from the page,
// filling the page pointer on successful return
func crawlPage(context *CrawlerContext, page *Page) {
	defer context.WG.Done()

	context.VisitPage(page)

	r, err := context.Fetcher.Fetch(page.URL)
	if err != nil {
		context.Fail <- err
		return
	}

	root, err := html.Parse(r)
	if err != nil {
		context.Fail <- err
		return
	}

	parseHTML(context, root, page)
}

// parseHTML is an auxiliary function of Crawl function that will travel recursively around the HTML
// document identifying elements to populate the Page object
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
					linkURL = context.Domain + "/" + linkURL
				}

				// Check if we already processed this page, if so add the pointer of the page, otherwise
				// set the page to be processed if is in the same domain
				if strings.HasPrefix(linkURL, context.Domain) {
					if p, found := context.URLWasVisited(linkURL); found {
						link.Page = p

					} else {
						link.Page = &Page{
							URL: linkURL,
						}

						context.WG.Add(1)
						go crawlPage(context, link.Page)
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
		parseHTML(context, child, page)
	}
}
