package indexer

import (
	"math"
	"sort"
)

// Document represents one crawled page stored in the index.
type Document struct {
	ID   int
	URL  string
	Text string
}

// posting stores how many times a word appears in a specific document.
type posting struct {
	docID int
	freq  int
}

// Index is an inverted index: maps each word to postings (doc ID + frequency).
type Index struct {
	documents map[int]*Document
	postings  map[string][]posting
	nextID    int
}

// NewIndex creates an empty Index ready to use.
func NewIndex() *Index {
	return &Index{
		documents: make(map[int]*Document),
		postings:  make(map[string][]posting),
		nextID:    0,
	}
}

// Add tokenizes the given text and adds it to the index under a new document ID.
func (idx *Index) Add(url string, text string) int {
	id := idx.nextID
	idx.nextID++

	idx.documents[id] = &Document{ID: id, URL: url, Text: text}

	tokens := Tokenize(text)
	counts := make(map[string]int)
	for _, word := range tokens {
		counts[word]++
	}

	for word, freq := range counts {
		idx.postings[word] = append(idx.postings[word], posting{docID: id, freq: freq})
	}

	return id
}

// ScoredDocument pairs a Document with its relevance score for a query.
type ScoredDocument struct {
	Doc   *Document
	Score float64
}

// Search returns documents containing ALL query words, ranked by TF-IDF score (best match first).
func (idx *Index) Search(query string) []ScoredDocument {
	tokens := Tokenize(query)
	if len(tokens) == 0 {
		return nil
	}

	totalDocs := len(idx.documents)

	// Build docID -> matched-word-count, and docID -> score, only for docs matching ALL words
	matchCount := make(map[int]int)
	scores := make(map[int]float64)

	for _, word := range tokens {
		wordPostings := idx.postings[word]
		df := len(wordPostings) // document frequency: how many docs contain this word at all
		if df == 0 {
			continue
		}
		idf := math.Log(float64(totalDocs) / float64(df))

		for _, p := range wordPostings {
			tf := float64(p.freq)
			scores[p.docID] += tf * idf
			matchCount[p.docID]++
		}
	}

	var results []ScoredDocument
	for docID, count := range matchCount {
		if count == len(tokens) { // must match ALL query words (AND logic)
			results = append(results, ScoredDocument{
				Doc:   idx.documents[docID],
				Score: scores[docID],
			})
		}
	}

	// Sort by score descending — best matches first
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}
