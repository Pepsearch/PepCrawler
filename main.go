package main

import (
	"context"
	"fmt"
	"os"
	"github.com/go-redis/redis/v8"
	"github.com/gocolly/colly"
)

func main() {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),  // replace with your Redis server address and port without "redis://default:"
		Password: os.Getenv("REDIS_PASSWORD"), // replace with your password if any
		DB:       0,                                  // use default DB
	})
  
	c := colly.NewCollector(
		colly.Async(true),
	)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		fmt.Printf("Link found: %q -> %s\n", e.Text, link)
		c.Visit(e.Request.AbsoluteURL(link))
	})

	var title, description, ogTitle, ogDescription string

	c.OnHTML("title", func(e *colly.HTMLElement) {
		title = e.Text
	})

	c.OnHTML("meta[name=description]", func(e *colly.HTMLElement) {
		description = e.Attr("content")
	})

	c.OnHTML("meta[property=og:title]", func(e *colly.HTMLElement) {
		ogTitle = e.Attr("content")
	})

	c.OnHTML("meta[property=og:description]", func(e *colly.HTMLElement) {
		ogDescription = e.Attr("content")
	})

	c.OnScraped(func(r *colly.Response) {
		if title == "" {
			title = ogTitle
		}
		if description == "" {
			description = ogDescription
		}
		rdb.HSet(ctx, "page:"+r.Request.URL.String(), "title", title, "description", description, "og:title", ogTitle, "og:description", ogDescription)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit("https://wikipedia.org") // add a URL to crawl here.

	c.Wait()
}
