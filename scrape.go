package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gocolly/colly"
)

func sanitiseURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	parsedURL.RawQuery = ""

	return parsedURL.String(), nil
}

func ScrapeEntireSite(options ScrapeOptions) ScrapeResult {
	defer TimeTrack(time.Now(), "ScrapeEntireSite")

	var successPages []PageSuccess
	var errorPages []PageError
	visited := NewSafeMap()
	pagesVisited := 0

	c := colly.NewCollector(
		colly.AllowedDomains(options.ValidDomains...),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: options.Concurrency,
		Delay:       time.Duration(options.MinTimeBetween),
	})

	c.SetRequestTimeout(5 * time.Second)

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		IdleConnTimeout:       5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	c.WithTransport(client.Transport)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if pagesVisited >= options.MaxPagesToVisit {
			return
		}

		link := e.Request.AbsoluteURL(e.Attr("href"))
		link, _ = sanitiseURL(link)

		if _, found := visited.Get(link); found {
			return
		}

		c.Visit(link)
	})

	c.OnRequest(func(r *colly.Request) {
		if pagesVisited >= options.MaxPagesToVisit {
			fmt.Println("Max pages to visit reached, aborting:", r.URL)
			r.Abort()
			return
		}
		r.URL.RawQuery = ""
		r.Headers.Set("User-Agent", googleUserAgent)

		if _, found := visited.Get(r.URL.String()); found {
			fmt.Println("Already visited, aborting:", r.URL)
			r.Abort()
			return
		}
	})

	c.OnResponse(func(r *colly.Response) {
		if pagesVisited >= options.MaxPagesToVisit {
			return
		}

		pagesVisited++
		initalURL := r.Request.URL.String()
		finalURL := initalURL
		headerLocation := r.Headers.Get("Location")

		if len(headerLocation) != 0 && (initalURL != headerLocation) {
			fmt.Println("Redirected URL:", initalURL, "to:", headerLocation)
			finalURL = r.Headers.Get("Location")
		}

		if _, found := visited.Get(finalURL); found {
			fmt.Println("Already visited, discarding:", finalURL)
			return
		}

		fmt.Println("Success URL:", finalURL, "Content length:", len(r.Body))

		visited.Set(finalURL, true)
		successPages = append(successPages, PageSuccess{Url: finalURL, Content: string(r.Body)})
	})

	c.OnError(func(r *colly.Response, err error) {
		pagesVisited++
		fmt.Println("Failed URL:", r.Request.URL, "Error:", err)

		errorPages = append(errorPages, PageError{Url: r.Request.URL.String(), Err: err})
	})

	for _, url := range options.StartingURLs {
		url, _ = sanitiseURL(url)

		if _, found := visited.Get(url); !found {
			c.Visit(url)
		}
	}

	c.Wait()

	return ScrapeResult{Success: successPages, Error: errorPages}
}
