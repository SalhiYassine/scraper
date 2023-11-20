package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

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

		scrapeOutcome := ScrapeEntireSite(ScrapeOptions{
			StartingURLs:    sitemapUrls,
			ValidDomains:    options.ValidDomains,
			Concurrency:     options.Concurrency,
			MaxDepth:        options.MaxDepth,
			MinTimeBetween:  options.MinTimeBetween,
			MaxPagesToVisit: options.MaxPagesToVisit,
		})

		fmt.Println("Success:", len(scrapeOutcome.Success))
		fmt.Println("Error:", len(scrapeOutcome.Error))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(scrapeOutcome)
	})

	fmt.Println("Listening on", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}

}
