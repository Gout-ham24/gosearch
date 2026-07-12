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

// Search returns the documents that contain the given word.
func (idx *Index) Search(word string) []*Document {
	tokens := Tokenize(word) // normalize the query the same way we normalize indexed text
	if len(tokens) == 0 {
		return nil
	}

	ids := idx.postings[tokens[0]]
	var results []*Document
	for _, id := range ids {
		results = append(results, idx.documents[id])
	}

	return results
}
