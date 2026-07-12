package crawler

import "sync"

// Queue manages URLs to be crawled, avoiding duplicates.
type Queue struct {
	mu      sync.Mutex
	pending []string
	visited map[string]bool
}

// NewQueue creates an empty Queue ready to use.
func NewQueue() *Queue {
	return &Queue{
		pending: []string{},
		visited: make(map[string]bool),
	}
}

// Add adds a URL to the queue if it hasn't been visited or already queued.
func (q *Queue) Add(url string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.visited[url] {
		return
	}
	q.pending = append(q.pending, url)
	q.visited[url] = true
}

// Next removes and returns the next URL to crawl.
// The second return value is false if the queue is empty.
func (q *Queue) Next() (string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.pending) == 0 {
		return "", false
	}

	url := q.pending[0]
	q.pending = q.pending[1:]
	return url, true
}
