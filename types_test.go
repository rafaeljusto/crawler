package crawler

import (
	"testing"
)

func TestPageString(t *testing.T) {
	p := Page{
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
	}

	expected := `
❆ index.html

  ▤  example.css
  ▤  example.js
  ▤  example.png

  ↳ "Example 1"
  
    ❆ example1.html
    
  ↳ "Example 2"
  
    ❆ example2.html
    
`

	if p.String() != expected {
		t.Errorf("Page text format was different from the expected. Expected %s and got %s",
			expected, p)
	}
}
