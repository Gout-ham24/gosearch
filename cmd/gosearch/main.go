package main

import (
	"fmt"
	"log"
	"net/http"

	"gosearch/api"
	"gosearch/storage"
)

func main() {
	pool, err := storage.Connect()
	if err != nil {
		fmt.Println("Connection failed:", err)
		return
	}
	defer pool.Close()

	server := api.NewServer(pool)

	http.HandleFunc("/search", api.EnableCORS(server.SearchHandler))
	http.HandleFunc("/crawl", api.EnableCORS(server.CrawlHandler))

	fmt.Println("GoSearch API running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
