package main

import (
	"testing"
)

func TestPageString(t *testing.T) {
	p := Page{
		URL: "index.html",
		Links: []Link{
			{
				Label: "Example 1",
				Page:  Page{URL: "example1.html"},
			},
			{
				Label: "Example 2",
				Page:  Page{URL: "example2.html"},
			},
		},
		StaticAssets: []string{
			"example.css",
			"example.js",
			"example.png",
		},
	}

	expected := `
❆ URL index.html

  ▤  example.css
  ▤  example.js
  ▤  example.png

  ↳ "Example 1"
  
    ❆ URL example1.html
    
  ↳ "Example 2"
  
    ❆ URL example2.html
    
`

	if p.String() != expected {
		t.Errorf("Page text format was different from the expected. Expected %s and got %s",
			expected, p)
	}
}
