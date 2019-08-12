package main

import (
	"fmt"
	"github.com/gocolly/colly"
	//"github.com/gocolly/colly/debug"
	"regexp"
	"time"
)

const (
	//burst		= 20* time.Second
	duration	= 1 * time.Minute
	//duration	= 5 * time.Second
)


func crawler(url string, syn <-chan struct{}) {

	// Instantiate default collector
	c := colly.NewCollector(
		// Attach a debugger to the collector
		//colly.Debugger(&debug.LogDebugger{}),
		colly.Async(true),
		colly.URLFilters(
			regexp.MustCompile("http://.+"),
		),
	)

	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		_ = c.Visit(e.Request.AbsoluteURL(link))
	})

	c.OnRequest(func(r *colly.Request) {
		//fmt.Println("Visiting", r.URL)
	})

	for i := 0; i < 16; i++ {
		//_ = c.Visit(fmt.Sprintf("%s?n=%d", url, i))
		_ = c.Visit(fmt.Sprintf("%s?n=%d", url, i))
	}

	<-syn
}


func main() {

	url := "http://bbc.com/"
	syn := make(chan struct{})
	//nbburst := 3

	go crawler(url, syn)

	loop:
	for {
		select {

		/*
		case <-time.After(burst):
			if nbburst > 0 {
				nbburst -= 1
				syn <- struct{}{}
				//fmt.Println("Pause...")
				time.Sleep(burst)
				go crawler(url, syn)
			} else {
				break loop
			}

		 */

		case <-time.After(duration):
			break loop
		}
	}
	syn<- struct{}{}
}