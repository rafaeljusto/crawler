package main

import (
	"flag"
	"fmt"
	"github.com/rafaeljusto/crawler"
	"os"
	"runtime"
)

// List of possible return codes of the program
const (
	NoError = iota
	ErrInputParameters
	ErrCrawlerExecution
)

// main will control the flow of all go routines that retrieve each crawler
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var url string
	flag.StringVar(&url, "url", "", "URL to build the site map")
	flag.StringVar(&url, "u", "", "URL to build the site map")
	flag.Parse()

	if len(url) == 0 {
		fmt.Println("URL parameter is mandatory")
		flag.PrintDefaults()
		os.Exit(ErrInputParameters)
	}

	page, err := crawler.Crawl(url, crawler.HTTPFetcher{})
	if err != nil {
		fmt.Println(err)
		os.Exit(ErrCrawlerExecution)
	}

	fmt.Printf(`
ＷＥＢ ＣＲＡＷＬＥＲ - %s

┏━━━━━━━━━━━━━━━━━━━━━━┓
┃ Legend               ┃
┃──────────────────────┃
┃                      ┃
┃ ❆ Page               ┃
┃ ↳ Link               ┃
┃ ▤ Static Asset       ┃
┃                      ┃
┗━━━━━━━━━━━━━━━━━━━━━━┛
`, url)

	fmt.Println(page)
}
