package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("GoSearch API running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
