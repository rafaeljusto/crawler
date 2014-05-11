package main

import (
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

// FakeFetcher is a function that implements an interface using the same strategy of
// http.HandlerFunc. http://www.onebigfluke.com/2014/04/gos-power-is-in-emergent-behavior.html
type FakeFetcher func(url string) (io.Reader, error)

func (f FakeFetcher) Fetch(url string) (io.Reader, error) {
	return f(url)
}

func TestCrawlMustReturnPageWithInformation(t *testing.T) {
	testData := []struct {
		url          string
		data         string
		expectedPage Page
	}{
		{
			url: "example.com",
			data: `<html>
  <body>
    <a href="example.net">Example</a>
  </body>
</html>`,
			expectedPage: Page{
				URL: "example.com",
				Links: []string{
					"example.net",
				},
			},
		},
		{
			url: "example.com",
			data: `<html>
  <body>
    <A HREF="example.net">Example</A>
  </body>
</html>`,
			expectedPage: Page{
				URL: "example.com",
				Links: []string{
					"example.net",
				},
			},
		},
	}

	for _, testItem := range testData {
		page, err := Crawl(testItem.url, FakeFetcher(func(url string) (io.Reader, error) {
			return strings.NewReader(testItem.data), nil
		}))

		if err != nil {
			t.Fatalf("Unexpected error returned. Expected '%v' and got '%v'", nil, err)
		}

		if !reflect.DeepEqual(testItem.expectedPage, page) {
			t.Errorf("Unexpected page returned. Expected '%#v' and got '%#v'",
				testItem.expectedPage, page)
		}
	}
}

func TestCrawlMustReturnErrorOnFetchProblems(t *testing.T) {
	testData := []struct {
		url         string
		data        string
		expectedErr error
	}{
		{
			url:         "example.com",
			data:        "",
			expectedErr: http.ErrContentLength,
		},
	}

	for _, testItem := range testData {
		_, err := Crawl(testItem.url, FakeFetcher(func(url string) (io.Reader, error) {
			return strings.NewReader(testItem.data), http.ErrContentLength
		}))

		if testItem.expectedErr != err {
			t.Fatalf("Unexpected error returned. Expected '%v' and got '%v'",
				testItem.expectedErr, err)
		}
	}
}
