// Copyright 2014 Rafael Dantas Justo. All rights reserved.
// Use of this source code is governed by a GPL
// license that can be found in the LICENSE file.

package crawler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
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

		// No href test
		{
			url: "example.com",
			data: `<html>
  <head>
    <link rel="stylesheet" type="text/css" href="example.css">
  </head>
  <body>
    <a id="example">Example</a>
    <img src="example.png" alt="example">
    <script type="text/javascript" src="example.js">
  </body>
</html>`,
			expected: Page{
				URL: "example.com",
				Links: []Link{
					{
						Label: "Example",
						Page:  nil,
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.png",
					"example.js",
				},
			},
		},

		// Compose label test
		{
			url: "example.com",
			data: `<html>
  <head>
    <link rel="stylesheet" type="text/css" href="example.css">
  </head>
  <body>
    <a href="example.net">Example<span>Test</span>Link</a>
    <img src="example.png" alt="example">
    <script type="text/javascript" src="example.js">
  </body>
</html>`,
			expected: Page{
				URL: "example.com",
				Links: []Link{
					{
						Label: "Example\nLink",
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

		// No label test
		{
			url: "example.com",
			data: `<html>
  <head>
    <link rel="stylesheet" type="text/css" href="example.css">
  </head>
  <body>
    <a href="example.net"></a>
    <img src="example.png" alt="example">
    <script type="text/javascript" src="example.js">
  </body>
</html>`,
			expected: Page{
				URL: "example.com",
				Links: []Link{
					{
						Label: "<no label>",
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
			t.Errorf("Unexpected page returned. Expected '%s' and got '%s'",
				testItem.expected, page)
		}
	}
}

func TestCrawlMustReturnFailPageOnFetchProblems(t *testing.T) {
	testData := []struct {
		url      string
		data     string
		expected Page
	}{
		{
			url:  "example.com",
			data: "",
			expected: Page{
				URL:  "example.com",
				Fail: true,
			},
		},
	}

	for _, testItem := range testData {
		page, err := Crawl(testItem.url, FakeFetcher(func(url string) (io.Reader, error) {
			return strings.NewReader(testItem.data), http.ErrContentLength
		}))

		if err != nil {
			t.Fatal(err)
		}

		if !testItem.expected.Equal(*page) {
			t.Fatalf("Unexpected error returned. Expected '%s' and got '%s'",
				testItem.expected, page)
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

		// One level link with partial link
		{
			url: "example.com",
			data: map[string]string{
				"example.com": `<html>
  <head>
    <link rel="stylesheet" type="text/css" href="example.css">
  </head>
  <body>
    <a href="/link1.html">Link 1</a>
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
												Page: &Page{
													URL: "example.com",
												},
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

		// Link on subdomain
		{
			url: "example.com",
			data: map[string]string{
				"example.com": `<html>
  <head>
    <link rel="stylesheet" type="text/css" href="example.css">
  </head>
  <body>
    <a href="test.example.com/link1.html">Link 1</a>
    <img src="example.png" alt="example"/>
    <script type="text/javascript" src="example.js"/>
  </body>
</html>`,
				"test.example.com/link1.html": `<html>
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
							URL: "test.example.com/link1.html",
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
			t.Errorf("Unexpected page returned. Expected '%s' and got '%s'",
				testItem.expected, page)
		}
	}
}

func TestCrawlStress(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	index := ""

	httpTestServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/" {
			fmt.Fprintf(w, index)

		} else {
			fmt.Fprintf(w, "<html><body></body></html>")
		}
	}))
	defer httpTestServer.Close()

	domain := fmt.Sprintf("http://%s", httpTestServer.Listener.Addr().String())
	links := ""

	for i := 0; i < 20000; i++ {
		url := fmt.Sprintf("%s/test%d.html", domain, i)
		links += fmt.Sprintf("<a href=\"%s\">Test %d</a>\n", url, i)
	}

	index += fmt.Sprintf("<html><body>%s</body></html>", links)

	if _, err := Crawl(domain, HTTPFetcher{}); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkCrawl(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Crawl("example.com", FakeFetcher(func(url string) (io.Reader, error) {
			return strings.NewReader("<html><body></body></html>"), nil
		}))
	}
}
