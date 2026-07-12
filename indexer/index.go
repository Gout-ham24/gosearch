package indexer

// Document represents one crawled page stored in the index.
type Document struct {
	ID   int
	URL  string
	Text string
}

// Index is an inverted index: maps each word to the list of document IDs containing it.
type Index struct {
	documents map[int]*Document
	postings  map[string][]int
	nextID    int
}

// NewIndex creates an empty Index ready to use.
func NewIndex() *Index {
	return &Index{
		documents: make(map[int]*Document),
		postings:  make(map[string][]int),
		nextID:    0,
	}
}

// Add tokenizes the given text and adds it to the index under a new document ID.
// Returns the assigned document ID.
func (idx *Index) Add(url string, text string) int {
	id := idx.nextID
	idx.nextID++

	idx.documents[id] = &Document{ID: id, URL: url, Text: text}

	tokens := Tokenize(text)
	seen := make(map[string]bool) // avoid adding the same doc ID twice for repeated words

	for _, word := range tokens {
		if seen[word] {
			continue
		}
		seen[word] = true
		idx.postings[word] = append(idx.postings[word], id)
	}

	return id
}

// Search returns documents that contain ALL the given query words.
func (idx *Index) Search(query string) []*Document {
	tokens := Tokenize(query)
	if len(tokens) == 0 {
		return nil
	}

	// Start with the postings list of the first word
	matchIDs := make(map[int]bool)
	for _, id := range idx.postings[tokens[0]] {
		matchIDs[id] = true
	}

	// For each remaining word, keep only IDs that also appear in ITS postings list
	for _, word := range tokens[1:] {
		wordIDs := make(map[int]bool)
		for _, id := range idx.postings[word] {
			wordIDs[id] = true
		}

		for id := range matchIDs {
			if !wordIDs[id] {
				delete(matchIDs, id)
			}
		}
	}

	var results []*Document
	for id := range matchIDs {
		results = append(results, idx.documents[id])
	}

	return results
}
