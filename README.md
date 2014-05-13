crawler
=======

Web crawler limited to one domain. When crawling example.com it would crawl all pages within the example.com domain, but not follow the links to Facebook or Instagram accounts or subdomains like other.example.com. Given a URL, it should output a site map, showing which static assets each page depends on, and the links between pages.

building
========

The Crawler project was developed using the Go language and it depends on the following Go packages:

* code.google.com/p/go.net/html

All the above packages can be installed using the command:

    go get -u <package_name>

Also, to easy run the project tests you will need the following:

* Python3 - http://www.python.org/

deploying
=========

To deploy the project you will need the program bellow.

* FPM - https://github.com/jordansissel/fpm (Debian packages)