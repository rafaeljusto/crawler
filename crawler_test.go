// Copyright 2014 Rafael Dantas Justo. All rights reserved.
// Use of this source code is governed by a GPL
// license that can be found in the LICENSE file.

// crawler verify a HTML page and list the resources
package crawler

import (
	"io"
	"net/http"
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
						Page:  &Page{URL: "example.net"},
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
						Page:  &Page{URL: "example.net"},
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
						Page:  &Page{URL: "example.net"},
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

		if !page.Equal(testItem.expected) {
			t.Errorf("Unexpected page returned. Expected '%#v' and got '%#v'",
				testItem.expected, page)
		}
	}
}

func TestCrawlMustReturnErrorOnFetchProblems(t *testing.T) {
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

func TestCrawlMustFollowLinks(t *testing.T) {
	testData := []struct {
		url      string
		data     map[string]string
		expected Page
	}{
		// One level link
		{
			url: "example.com",
			data: map[string]string{
				"example.com": `<html>
  <head>
    <link rel="stylesheet" type="text/css" href="example.css">
  </head>
  <body>
    <a href="example.com/link1.html">Link 1</a>
    <img src="example.png" alt="example"/>
    <script type="text/javascript" src="example.js"/>
  </body>
</html>`,
				"example.com/link1.html": `<html>
  <head>
    <link rel="stylesheet" type="text/css" href="link1.css">
  </head>
  <body>
    <img src="link1.png" alt="link1"/>
    <script type="text/javascript" src="link1.js"/>
  </body>
</html>`,
			},
			expected: Page{
				URL: "example.com",
				Links: []Link{
					{
						Label: "Link 1",
						Page: &Page{
							URL: "example.com/link1.html",
							StaticAssets: []string{
								"link1.css",
								"link1.png",
								"link1.js",
							},
						},
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.png",
					"example.js",
				},
			},
		},

		// Two levels link
		{
			url: "example.com",
			data: map[string]string{
				"example.com": `<html>
  <head>
    <link rel="stylesheet" type="text/css" href="example.css">
  </head>
  <body>
    <a href="example.com/link1.html">Link 1</a>
    <img src="example.png" alt="example"/>
    <script type="text/javascript" src="example.js"/>
  </body>
</html>`,
				"example.com/link1.html": `<html>
  <head>
    <link rel="stylesheet" type="text/css" href="link1.css">
  </head>
  <body>
    <a href="example.com/link2.html">Link 2</a>
    <img src="link1.png" alt="link1"/>
    <script type="text/javascript" src="link1.js"/>
  </body>
</html>`,
				"example.com/link2.html": `<html>
  <head>
    <link rel="stylesheet" type="text/css" href="link2.css">
  </head>
  <body>
    <img src="link2.png" alt="link2"/>
    <script type="text/javascript" src="link2.js"/>
  </body>
</html>`,
			},
			expected: Page{
				URL: "example.com",
				Links: []Link{
					{
						Label: "Link 1",
						Page: &Page{
							URL: "example.com/link1.html",
							Links: []Link{
								{
									Label: "Link 2",
									Page: &Page{
										URL: "example.com/link2.html",
										StaticAssets: []string{
											"link2.css",
											"link2.png",
											"link2.js",
										},
									},
								},
							},
							StaticAssets: []string{
								"link1.css",
								"link1.png",
								"link1.js",
							},
						},
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.png",
					"example.js",
				},
			},
		},

		// Cyclic link
		{
			url: "example.com",
			data: map[string]string{
				"example.com": `<html>
  <head>
    <link rel="stylesheet" type="text/css" href="example.css">
  </head>
  <body>
    <a href="example.com/link1.html">Link 1</a>
    <img src="example.png" alt="example"/>
    <script type="text/javascript" src="example.js"/>
  </body>
</html>`,
				"example.com/link1.html": `<html>
  <head>
    <link rel="stylesheet" type="text/css" href="link1.css">
  </head>
  <body>
    <a href="example.com/link2.html">Link 2</a>
    <img src="link1.png" alt="link1"/>
    <script type="text/javascript" src="link1.js"/>
  </body>
</html>`,
				"example.com/link2.html": `<html>
  <head>
    <link rel="stylesheet" type="text/css" href="link2.css">
  </head>
  <body>
    <a href="example.com">Example</a>
    <img src="link2.png" alt="link2"/>
    <script type="text/javascript" src="link2.js"/>
  </body>
</html>`,
			},
			expected: Page{
				URL: "example.com",
				Links: []Link{
					{
						Label: "Link 1",
						Page: &Page{
							URL: "example.com/link1.html",
							Links: []Link{
								{
									Label: "Link 2",
									Page: &Page{
										URL: "example.com/link2.html",
										Links: []Link{
											{
												Label:      "Example",
												CyclicPage: true,
											},
										},
										StaticAssets: []string{
											"link2.css",
											"link2.png",
											"link2.js",
										},
									},
								},
							},
							StaticAssets: []string{
								"link1.css",
								"link1.png",
								"link1.js",
							},
						},
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
			return strings.NewReader(testItem.data[url]), nil
		}))

		if err != nil {
			t.Fatalf("Unexpected error returned. Expected '%v' and got '%v'", nil, err)
		}

		if !page.Equal(testItem.expected) {
			t.Errorf("Unexpected page returned. Expected '%#v' and got '%#v'",
				testItem.expected, page)
		}
	}
}
