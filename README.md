crawler
=======

[![Build Status](https://travis-ci.org/rafaeljusto/crawler.png?branch=master)](https://travis-ci.org/rafaeljusto/crawler)
[![GoDoc](https://godoc.org/github.com/rafaeljusto/crawler?status.png)](https://godoc.org/github.com/rafaeljusto/crawler)
[![Download](https://api.bintray.com/packages/rafaeljusto/deb/crawler/images/download.png) ](https://bintray.com/rafaeljusto/deb/crawler/_latestVersion)

Web crawler tool limited to one domain. When crawling example.com it would crawl all pages
within the example.com domain, but not follow the links to Facebook or Instagram accounts
or subdomains like other.example.com. Given a URL, it should output a site map, showing
which static assets each page depends on, and the links between pages.

building
========

The Crawler project was developed using the Go language and it depends on the following Go packages:

* code.google.com/p/go.net/html

All the above packages can be installed using the command:

    go get -u <package_name>

Also, to easy run the project tests you will need the following:

* Python3 - http://www.python.org/

Finally, to download and build the command line tool just use the following commands:

    go get -u github.com/rafaeljusto/crawler
    go build -o crawler github.com/rafaeljusto/crawler/app

deploying
=========

To deploy the project you will need the program bellow.

* FPM - https://github.com/jordansissel/fpm (Debian packages)