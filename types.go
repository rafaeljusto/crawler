package crawler

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

// Page describes the information stored after a webpage is crawled
type Page struct {
	URL          string   // Address of the page
	Links        []Link   // List of links for other URLs in this page
	StaticAssets []string // List of static dependencies of this page
}

// String transforms the Page into text mode to print the results
func (p Page) String() string {
	staticAssets := ""
	for _, staticAsset := range p.StaticAssets {
		if len(staticAssets) > 0 {
			staticAssets += "\n"
		}

		staticAssets += fmt.Sprintf(`  ▤  %s`, staticAsset)
	}

	links := ""
	for _, link := range p.Links {
		if len(links) > 0 {
			links += "\n"
		}

		// Add an identation level to the link content
		linkPage := strings.Replace(link.Page.String(), "\n", "\n    ", -1)

		links += fmt.Sprintf(`  ↳ "%s"
  %s`, link.Label, linkPage)
	}

	pageStr := fmt.Sprintf("\n❆ %s\n", p.URL)

	// Don't add unecessary spaces when there's no information
	if len(staticAssets) > 0 {
		pageStr += "\n" + staticAssets + "\n"
	}

	// Don't add unecessary spaces when there's no information
	if len(links) > 0 {
		pageStr += "\n" + links + "\n"
	}

	return pageStr
}

// Link stores information of other URL in this page
type Link struct {
	Label string // Context identification of the link
	Page  Page   // Page information about the other URL
}

// Fetcher creates an interface to allow a flexibility on how we retrieve the page data. For tests
// we will simulate the response while in production we will do a HTTP GET
type Fetcher interface {
	Fetch(url string) (io.Reader, error)
}

// HTTPFetcher will retrieve the page content via HTTP GET request
type HTTPFetcher struct {
}

func (f HTTPFetcher) Fetch(url string) (io.Reader, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(content), nil
}

// CrawlerContext stores all attributes used during a crawling execution
type CrawlerContext struct {
	Domain  string
	Fetcher Fetcher
	WG      sync.WaitGroup
	Fail    chan error

	// visitedPages store all pages already visited in a map, so that if we found a link for the same
	// page again, we just pick on the map the same object address. The function that prints the page
	// is responsable for detecting cycle loops
	visitedPages map[string]*Page

	// visitedPagesLock allows visitedPages to be manipulated safely by go routines
	visitedPagesLock sync.RWMutex
}

// NewCrawlerContext make it easy to initialize a new context
func NewCrawlerContext(domain string, fetcher Fetcher) *CrawlerContext {
	c := &CrawlerContext{
		Domain:  domain,
		Fetcher: fetcher,
	}

	c.Fail = make(chan error)
	c.visitedPages = make(map[string]*Page)
	return c
}

// VisitPage is a go routine safe way to add a new item in the visitedPages map
func (c *CrawlerContext) VisitPage(page *Page) {
	c.visitedPagesLock.Lock()
	defer c.visitedPagesLock.Unlock()
	c.visitedPages[page.URL] = page
}

// URLWasVisited is a go routine safe way to check if a page was alredy analyzed
func (c *CrawlerContext) URLWasVisited(url string) bool {
	c.visitedPagesLock.RLock()
	defer c.visitedPagesLock.RUnlock()
	_, visited := c.visitedPages[url]
	return visited
}
