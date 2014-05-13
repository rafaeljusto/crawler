// Copyright 2014 Rafael Dantas Justo. All rights reserved.
// Use of this source code is governed by a GPL
// license that can be found in the LICENSE file.

// crawler verify a HTML page and list the resources
package crawler

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

func TestCrawlPageMustReturnPageWithInformation(t *testing.T) {
	testData := []struct {
		url      string
		data     string
		expected Page
	}{
		// Lower case test
		{
			url: "example.com",
			data: `<html>
  <head>
    <link rel="stylesheet" type="text/css" href="example.css">
  </head>
  <body>
    <a href="example.net">Example</a>
    <img src="example.png" alt="example"/>
    <script type="text/javascript" src="example.js"/>
  </body>
</html>`,
			expected: Page{
				URL: "example.com",
				Links: []Link{
					{
						Label: "Example",
						Page:  Page{URL: "example.net"},
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.png",
					"example.js",
				},
			},
		},

		// Upper case test
		{
			url: "example.com",
			data: `<html>
  <head>
    <LINK rel="stylesheet" type="text/css" HREF="example.css">
  </head>
  <body>
    <A HREF="example.net">Example</A>
    <IMG SRC="example.png" alt="example"/>
    <SCRIPT type="text/javascript" SRC="example.js"/>
  </body>
</html>`,
			expected: Page{
				URL: "example.com",
				Links: []Link{
					{
						Label: "Example",
						Page:  Page{URL: "example.net"},
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.png",
					"example.js",
				},
			},
		},

		// No end-tag test
		{
			url: "example.com",
			data: `<html>
  <head>
    <link rel="stylesheet" type="text/css" href="example.css">
  </head>
  <body>
    <a href="example.net">Example
    <img src="example.png" alt="example">
    <script type="text/javascript" src="example.js">
  </body>
</html>`,
			expected: Page{
				URL: "example.com",
				Links: []Link{
					{
						Label: "Example",
						Page:  Page{URL: "example.net"},
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.png",
					"example.js",
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

		if !reflect.DeepEqual(testItem.expected, *page) {
			t.Errorf("Unexpected page returned. Expected '%#v' and got '%#v'",
				testItem.expected, page)
		}
	}
}

func TestCrawlPageMustReturnErrorOnFetchProblems(t *testing.T) {
	testData := []struct {
		url      string
		data     string
		expected error
	}{
		{
			url:      "example.com",
			data:     "",
			expected: http.ErrContentLength,
		},
	}

	for _, testItem := range testData {
		_, err := Crawl(testItem.url, FakeFetcher(func(url string) (io.Reader, error) {
			return strings.NewReader(testItem.data), http.ErrContentLength
		}))

		if testItem.expected != err {
			t.Fatalf("Unexpected error returned. Expected '%v' and got '%v'",
				testItem.expected, err)
		}
	}
}
