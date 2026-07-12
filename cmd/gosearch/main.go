package main

import (
	"context"
	"fmt"

	"gosearch/storage"
)

func main() {
	pool, err := storage.Connect()
	if err != nil {
		fmt.Println("Connection failed:", err)
		return
	}
	defer pool.Close()

	ctx := context.Background()

	query := "go programming"
	results, err := storage.SearchDocuments(ctx, pool, query)
	if err != nil {
		fmt.Println("Search failed:", err)
		return
	}

	fmt.Printf("--- Search results for %q ---\n", query)
	for _, r := range results {
		fmt.Printf("%.4f  %s\n", r.Score, r.URL)
	}
}
