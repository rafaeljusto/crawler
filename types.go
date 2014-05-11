package crawler

import (
	"fmt"
	"io"
	"net/http"
	"strings"
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

		// Add identation for the current level
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

	// TODO: Close the body?
	return response.Body, nil
}
