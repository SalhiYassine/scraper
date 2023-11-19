package main

import (
	"fmt"
	"math"
	"time"
)

func main() {
	sitemapUrls, err := GetSitemapUrls("https://algomo.com")
	if err != nil {
		panic(err)
	}

	fmt.Println("Sitemap URLs:", len(sitemapUrls))

	scrapeOutcome := ScrapeEntireSite(ScrapeOptions{
		StartingURLs:    sitemapUrls,
		ValidDomains:    []string{"www.algomo.com", "algomo.com", "help.algomo.com"},
		Concurrency:     15,
		MaxDepth:        int(math.Inf(1)),
		MinTimeBetween:  int(time.Second),
		MaxPagesToVisit: 5000,
	})

	fmt.Println("Success:", len(scrapeOutcome.Success))
	fmt.Println("Error:", len(scrapeOutcome.Error))

}
