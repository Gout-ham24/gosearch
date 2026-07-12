package main

import (
	"fmt"

	"gosearch/crawler"
	"gosearch/indexer"
)

func main() {
	// Step 1: Crawl a few pages
	c := crawler.NewCrawler("https://en.wikipedia.org/wiki/Go_(programming_language)", 5)
	pages := c.Run()

	// Step 2: Feed crawled pages into the index
	idx := indexer.NewIndex()
	for _, page := range pages {
		idx.Add(page.URL, page.Text)
	}
	fmt.Println("\nIndexed", len(pages), "pages")

	// Step 3: Search for something
	query := "go programming"
	results := idx.Search(query)

	fmt.Printf("\n--- Search results for %q ---\n", query)
	for _, doc := range results {
		fmt.Println(doc.URL)
	}
}
