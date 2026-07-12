package storage

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var stopWords = map[string]bool{
	"the": true, "is": true, "at": true, "a": true, "an": true,
	"and": true, "or": true, "of": true, "to": true, "in": true,
	"it": true, "for": true, "on": true, "with": true, "as": true,
	"this": true, "that": true, "be": true, "are": true, "was": true,
}

// SaveDocument inserts a crawled page and its word frequencies into the database.
// If the URL already exists, it's skipped (no duplicate crawling).
func SaveDocument(ctx context.Context, pool *pgxpool.Pool, url string, text string) error {
	var docID int
	err := pool.QueryRow(ctx,
		`INSERT INTO documents (url, text) VALUES ($1, $2)
		 ON CONFLICT (url) DO NOTHING
		 RETURNING id`,
		url, text,
	).Scan(&docID)

	if err != nil {
		return nil
	}

	counts := make(map[string]int)
	var current strings.Builder
	for _, r := range text + " " {
		if isWordChar(r) {
			current.WriteRune(toLower(r))
		} else if current.Len() > 0 {
			word := current.String()
			if !stopWords[word] {
				counts[word]++
			}
			current.Reset()
		}
	}

	batch := &pgx.Batch{}
	for word, freq := range counts {
		batch.Queue(
			`INSERT INTO postings (word, document_id, frequency) VALUES ($1, $2, $3)
			 ON CONFLICT (word, document_id) DO UPDATE SET frequency = $3`,
			word, docID, freq,
		)
	}

	br := pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

func toLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}

// SearchResult pairs a document with its TF-IDF relevance score.
type SearchResult struct {
	URL   string
	Score float64
}

// SearchDocuments finds documents matching ALL words in the query, ranked by TF-IDF score.
func SearchDocuments(ctx context.Context, pool *pgxpool.Pool, query string) ([]SearchResult, error) {
	words := tokenizeQuery(query)
	if len(words) == 0 {
		return nil, nil
	}

	// This query:
	// 1. Filters postings to only the query words
	// 2. Computes TF-IDF per word per document
	// 3. Sums scores per document
	// 4. Keeps only documents matching ALL query words (HAVING COUNT = total words)
	// 5. Orders by total score, highest first
	sql := `
		WITH doc_totals AS (
			SELECT COUNT(*) AS total FROM documents
		),
		word_stats AS (
			SELECT
				p.document_id,
				p.word,
				p.frequency,
				(SELECT COUNT(DISTINCT document_id) FROM postings WHERE word = p.word) AS doc_freq
			FROM postings p
			WHERE p.word = ANY($1)
		)
		SELECT
			d.url,
			SUM(ws.frequency * LN((SELECT total FROM doc_totals)::float / ws.doc_freq)) AS score
		FROM word_stats ws
		JOIN documents d ON d.id = ws.document_id
		GROUP BY d.url
		HAVING COUNT(DISTINCT ws.word) = $2
		ORDER BY score DESC
	`

	rows, err := pool.Query(ctx, sql, words, len(words))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.URL, &r.Score); err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	return results, nil
}

// tokenizeQuery normalizes a search query the same way indexed text was normalized.
func tokenizeQuery(query string) []string {
	var tokens []string
	var current strings.Builder

	for _, r := range query + " " {
		if isWordChar(r) {
			current.WriteRune(toLower(r))
		} else if current.Len() > 0 {
			word := current.String()
			if !stopWords[word] {
				tokens = append(tokens, word)
			}
			current.Reset()
		}
	}

	return tokens
}
