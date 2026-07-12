package main

import (
	"fmt"

	"gosearch/crawler"
)

func main() {
	c := crawler.NewCrawler("https://en.wikipedia.org/wiki/Go_(programming_language)", 5)
	pages := c.Run()

	fmt.Println("\n=== Crawl complete ===")
	fmt.Println("Pages crawled:", len(pages))
}
