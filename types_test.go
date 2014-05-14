// Copyright 2014 Rafael Dantas Justo. All rights reserved.
// Use of this source code is governed by a GPL
// license that can be found in the LICENSE file.

// crawler verify a HTML page and list the resources
package crawler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPageString(t *testing.T) {
	testData := []struct {
		page     Page
		expected string
	}{
		// Page with links test
		{
			page: Page{
				URL: "index.html",
				Links: []Link{
					{
						Label: "Example 1",
						Page:  &Page{URL: "example1.html"},
					},
					{
						Label: "Example 2",
						Page:  &Page{URL: "example2.html"},
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.js",
					"example.png",
				},
			},
			expected: `
❆ index.html

  ▤  example.css
  ▤  example.js
  ▤  example.png

  ↳ "Example 1"
  
    ❆ example1.html
    
  ↳ "Example 2"
  
    ❆ example2.html
    
`,
		},

		// Page with cyclic links test
		{
			page: Page{
				URL: "index.html",
				Links: []Link{
					{
						Label: "Example 1",
						Page: &Page{
							URL: "example1.html",
							Links: []Link{
								{
									Label:      "index",
									CyclicPage: true,
									Page: &Page{
										URL: "index.html",
									},
								},
							},
						},
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.js",
					"example.png",
				},
			},
			expected: `
❆ index.html

  ▤  example.css
  ▤  example.js
  ▤  example.png

  ↳ "Example 1"
  
    ❆ example1.html
    
      ↳ "index"
      
        ❆ index.html ↺
    
`,
		},
	}

	for _, testItem := range testData {
		if testItem.page.String() != testItem.expected {
			t.Errorf("Page text format was different from the expected. Expected %s and got %s",
				testItem.expected, testItem.page)
		}
	}
}

func TestPageComparison(t *testing.T) {
	testData := []struct {
		page1    Page
		page2    Page
		expected bool
	}{
		// Equal pages test
		{
			page1: Page{
				URL: "index.html",
				Links: []Link{
					{
						Label: "Example 1",
						Page:  &Page{URL: "example1.html"},
					},
					{
						Label: "Example 2",
						Page:  &Page{URL: "example2.html"},
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.js",
					"example.png",
				},
			},
			page2: Page{
				URL: "index.html",
				Links: []Link{
					{
						Label: "Example 1",
						Page:  &Page{URL: "example1.html"},
					},
					{
						Label: "Example 2",
						Page:  &Page{URL: "example2.html"},
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.js",
					"example.png",
				},
			},
			expected: true,
		},

		// Different static assets test
		{
			page1: Page{
				URL: "index.html",
				Links: []Link{
					{
						Label: "Example 1",
						Page:  &Page{URL: "example1.html"},
					},
					{
						Label: "Example 2",
						Page:  &Page{URL: "example2.html"},
					},
				},
				StaticAssets: []string{
					"example1.css",
					"example1.js",
					"example1.png",
				},
			},
			page2: Page{
				URL: "index.html",
				Links: []Link{
					{
						Label: "Example 1",
						Page:  &Page{URL: "example1.html"},
					},
					{
						Label: "Example 2",
						Page:  &Page{URL: "example2.html"},
					},
				},
				StaticAssets: []string{
					"example2.css",
					"example2.js",
					"example2.png",
				},
			},
			expected: false,
		},

		// Different number of links test
		{
			page1: Page{
				URL: "index.html",
				Links: []Link{
					{
						Label: "Example 1",
						Page:  &Page{URL: "example1.html"},
					},
					{
						Label: "Example 2",
						Page:  &Page{URL: "example2.html"},
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.js",
					"example.png",
				},
			},
			page2: Page{
				URL: "index.html",
				Links: []Link{
					{
						Label: "Example 1",
						Page:  &Page{URL: "example1.html"},
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.js",
					"example.png",
				},
			},
			expected: false,
		},

		// Different labels test
		{
			page1: Page{
				URL: "index.html",
				Links: []Link{
					{
						Label: "Example 1",
						Page:  &Page{URL: "example1.html"},
					},
					{
						Label: "Example 2",
						Page:  &Page{URL: "example2.html"},
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.js",
					"example.png",
				},
			},
			page2: Page{
				URL: "index.html",
				Links: []Link{
					{
						Label: "Example 3",
						Page:  &Page{URL: "example1.html"},
					},
					{
						Label: "Example 4",
						Page:  &Page{URL: "example2.html"},
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.js",
					"example.png",
				},
			},
			expected: false,
		},

		// Different cyclic page flag test
		{
			page1: Page{
				URL: "index.html",
				Links: []Link{
					{
						Label:      "Example 1",
						CyclicPage: true,
						Page:       &Page{URL: "example1.html"},
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.js",
					"example.png",
				},
			},
			page2: Page{
				URL: "index.html",
				Links: []Link{
					{
						Label:      "Example 1",
						CyclicPage: false,
						Page:       &Page{URL: "example1.html"},
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.js",
					"example.png",
				},
			},
			expected: false,
		},

		// Nil page pointer test
		{
			page1: Page{
				URL: "index.html",
				Links: []Link{
					{
						Label: "Example 1",
						Page:  &Page{URL: "example1.html"},
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.js",
					"example.png",
				},
			},
			page2: Page{
				URL: "index.html",
				Links: []Link{
					{
						Label: "Example 1",
						Page:  nil,
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.js",
					"example.png",
				},
			},
			expected: false,
		},

		// Nil page pointer test 2
		{
			page1: Page{
				URL: "index.html",
				Links: []Link{
					{
						Label: "Example 1",
						Page:  nil,
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.js",
					"example.png",
				},
			},
			page2: Page{
				URL: "index.html",
				Links: []Link{
					{
						Label: "Example 1",
						Page:  &Page{URL: "example1.html"},
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.js",
					"example.png",
				},
			},
			expected: false,
		},

		// Nil page pointer test 3
		{
			page1: Page{
				URL: "index.html",
				Links: []Link{
					{
						Label: "Example 1",
						Page:  nil,
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.js",
					"example.png",
				},
			},
			page2: Page{
				URL: "index.html",
				Links: []Link{
					{
						Label: "Example 1",
						Page:  nil,
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.js",
					"example.png",
				},
			},
			expected: true,
		},

		// Link page comparision test
		{
			page1: Page{
				URL: "index.html",
				Links: []Link{
					{
						Label: "Example 1",
						Page: &Page{
							URL: "example1.html",
							StaticAssets: []string{
								"example1.css",
								"example1.js",
								"example1.png",
							},
						},
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.js",
					"example.png",
				},
			},
			page2: Page{
				URL: "index.html",
				Links: []Link{
					{
						Label: "Example 1",
						Page: &Page{
							URL: "example2.html",
							StaticAssets: []string{
								"example2.css",
								"example2.js",
								"example2.png",
							},
						},
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.js",
					"example.png",
				},
			},
			expected: false,
		},

		// Link with cyclic page comparision test
		{
			page1: Page{
				URL: "index.html",
				Links: []Link{
					{
						Label:      "Example 1",
						CyclicPage: true,
						Page: &Page{
							URL: "example1.html",
							StaticAssets: []string{
								"example1.css",
								"example1.js",
								"example1.png",
							},
						},
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.js",
					"example.png",
				},
			},
			page2: Page{
				URL: "index.html",
				Links: []Link{
					{
						Label:      "Example 1",
						CyclicPage: true,
						Page: &Page{
							URL: "example2.html",
							StaticAssets: []string{
								"example2.css",
								"example2.js",
								"example2.png",
							},
						},
					},
				},
				StaticAssets: []string{
					"example.css",
					"example.js",
					"example.png",
				},
			},
			expected: true,
		},
	}

	for _, testItem := range testData {
		if testItem.page1.Equal(testItem.page2) != testItem.expected {
			t.Errorf("Page comparision was different from the expected. For page1: %s and "+
				"page2 %s expected %v",
				testItem.page1, testItem.page2, testItem.expected)
		}
	}
}

func TestHTTPFetcher(t *testing.T) {
	httpTestServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<html>
  <head>
    <link rel="stylesheet" type="text/css" href="example.css">
  </head>
  <body>
    <a href="example.net">Example</a>
    <img src="example.png" alt="example"/>
    <script type="text/javascript" src="example.js"/>
  </body>
</html>`)
	}))
	defer httpTestServer.Close()

	url := fmt.Sprintf("http://%s", httpTestServer.Listener.Addr().String())
	page, err := Crawl(url, HTTPFetcher{})
	if err != nil {
		t.Fatal(err)
	}

	page.Equal(Page{
		URL: url,
		Links: []Link{
			{
				Label: "Example",
				Page: &Page{
					URL: "example.net",
				},
			},
		},
		StaticAssets: []string{
			"example.css",
			"example.png",
			"example.js",
		},
	})

	page, err = Crawl("http://unknownurl.unknown", HTTPFetcher{})
	if err == nil {
		t.Error("Not detecting HTTP fetcher errors")
	}
}
