package crawler

import "fmt"

type Crawler struct {
	Fetcher  *Fetcher
	Queue    *Queue
	MaxPages int
}

func NewCrawler(seedURL string, maxPages int) *Crawler {
	q := NewQueue()
	q.Add(seedURL)

	return &Crawler{
		Fetcher:  NewFetcher(),
		Queue:    q,
		MaxPages: maxPages,
	}
}

func (c *Crawler) Run() []*ParsedPage {
	var results []*ParsedPage

	for len(results) < c.MaxPages {
		url, ok := c.Queue.Next()
		if !ok {
			break
		}

		html, err := c.Fetcher.Fetch(url)
		if err != nil {
			fmt.Println("Skipping", url, "-", err)
			continue
		}

		page, err := Parse(html, url)
		if err != nil {
			fmt.Println("Parse failed for", url, "-", err)
			continue
		}

		fmt.Println("Crawled:", url)
		results = append(results, page)

		for _, link := range page.Links {
			c.Queue.Add(link)
		}
	}

	return results
}