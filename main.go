package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "5007" // Default port if not specified
	}

	addr := "0.0.0.0:" + port

	http.Handle("/metrics", promhttp.Handler())

	// post request to /scrape with body
	http.HandleFunc("/scrape", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "POST" {
			http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		type ScrapeRequestBody struct {
			RootUrl         string   `json:"rootUrl"`
			ValidDomains    []string `json:"validDomains"`
			Concurrency     int      `json:"concurrency"`
			MaxDepth        int      `json:"maxDepth"`
			MinTimeBetween  int      `json:"minTimeBetween"`
			MaxPagesToVisit int      `json:"maxPagesToVisit"`
		}

		var options ScrapeRequestBody
		err := json.NewDecoder(r.Body).Decode(&options)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println(options)

		sitemapUrls, err := GetSitemapUrls(options.RootUrl)
		if err != nil {
			sitemapUrls = []string{options.RootUrl}
		}

		fmt.Println("Sitemap URLs:", len(sitemapUrls))

		scrapedChan := make(chan PageSuccess)
		failedChan := make(chan PageError)

		go ScrapeEntireSite(ScrapeOptions{
			StartingURLs:    sitemapUrls,
			ValidDomains:    options.ValidDomains,
			Concurrency:     options.Concurrency,
			MaxDepth:        options.MaxDepth,
			MinTimeBetween:  options.MinTimeBetween,
			MaxPagesToVisit: options.MaxPagesToVisit,
		}, scrapedChan, failedChan)

		w.Header().Set("Content-Type", "application/json")

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			for pageResult := range scrapedChan {
				json.NewEncoder(w).Encode(pageResult)
				flusher.Flush()
			}
			wg.Done()
		}()

		go func() {
			for pageError := range failedChan {
				json.NewEncoder(w).Encode(pageError)
				flusher.Flush()
			}
			wg.Done()
		}()

		wg.Wait() // Wait for both channels to close
		fmt.Println("Finished scraping")
	})

	fmt.Println("Listening on", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}

}
