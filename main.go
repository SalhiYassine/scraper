package main

import (
	"fmt"
	"math"
	"time"
)

func main() {
	sitemapUrls, err := GetSitemapUrls("https://www.elitegarages.co.uk/")
	if err != nil {
		panic(err)
	}

	fmt.Println("Sitemap URLs:", len(sitemapUrls))

	scrapeOutcome := ScrapeEntireSite(ScrapeOptions{
		StartingURLs:    sitemapUrls,
		ValidDomains:    []string{"www.elitegarages.co.uk", "elitegarages.co.uk"},
		Concurrency:     25,
		MaxDepth:        int(math.Inf(1)),
		MinTimeBetween:  int(time.Second),
		MaxPagesToVisit: 5000,
	})

	fmt.Println("Success:", len(scrapeOutcome.Success))
	fmt.Println("Error:", len(scrapeOutcome.Error))

}
