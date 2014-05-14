// crawler - Web crawler limited to one domain
//
// Copyright (C) 2014 Rafael Dantas Justo <adm@rafael.net.br>
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.

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

	fmt.Println("Analyzing domain...")
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
┃ ↺ Already visited    ┃
┃                      ┃
┗━━━━━━━━━━━━━━━━━━━━━━┛
`, url)

	fmt.Println("Building output...")
	fmt.Println(page)
}
