package main

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/gocolly/colly"
	"golang.org/x/net/html"
)

func SanitiseURL(rawURL string) (string, error) {

	trimmedURL := strings.TrimSpace(rawURL)

	parsedURL, err := url.Parse(trimmedURL)
	if err != nil {
		return "", err
	}

	parsedURL.RawQuery = ""

	return parsedURL.String(), nil
}

func cleanNode(n *html.Node) {
	var next *html.Node
	for c := n.FirstChild; c != nil; c = next {

		next = c.NextSibling

		if c.Type == html.ElementNode && (c.Data == "script" || c.Data == "style") {
			n.RemoveChild(c)
			continue
		}

		if c.Type == html.CommentNode {
			n.RemoveChild(c)
			continue
		}

		if c.Type == html.ElementNode {
			c.Attr = nil
		}

		cleanNode(c)
	}

	if n.Type == html.ElementNode && isEmptyNode(n) {
		n.Parent.RemoveChild(n)
	}
}

func isEmptyNode(n *html.Node) bool {
	if n.Type == html.TextNode {

		return strings.TrimSpace(n.Data) == ""
	}
	return n.FirstChild == nil
}

func cleanHTML(htmlContent string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", err
	}

	cleanNode(doc)

	var buf bytes.Buffer
	err = html.Render(&buf, doc)
	if err != nil {
		return "", err
	}

	cleanedHTML := removeExtraWhitespace(buf.String())

	return cleanedHTML, nil
}

func removeExtraWhitespace(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	wasSpace := false

	for _, r := range s {
		if unicode.IsSpace(r) {
			if !wasSpace {
				b.WriteRune(' ')
				wasSpace = true
			}
		} else {
			b.WriteRune(r)
			wasSpace = false
		}
	}

	return b.String()
}

func ScrapeEntireSite(options ScrapeOptions, successChan chan<- PageSuccess, failChan chan<- PageError) {
	defer TimeTrack(time.Now(), "ScrapeEntireSite")

	fmt.Println("Starting scrape with options:", options)

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
		link, _ = SanitiseURL(link)

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

		cleanedHTML, err := cleanHTML(string(r.Body))
		if err != nil {
			fmt.Println("Error cleaning HTML:", err)
			return
		}

		fmt.Println("Success URL:", finalURL, "Content length:", len(cleanedHTML))

		visited.Set(finalURL, true)
		successChan <- PageSuccess{Url: finalURL, Content: string(r.Body)}
	})

	c.OnError(func(r *colly.Response, err error) {
		pagesVisited++
		fmt.Println("Failed URL:", r.Request.URL, "Error:", err)

		failChan <- PageError{Url: r.Request.URL.String(), Err: err}
	})

	for _, url := range options.StartingURLs {
		url, err := SanitiseURL(url)
		if err != nil {
			fmt.Println("Error sanitising URL:", url, "Error:", err)
			continue
		}

		if _, found := visited.Get(url); found {
			fmt.Println("Already visited, don't queue:", url)
			continue
		}

		fmt.Println("Starting URL:", url)

		c.Visit(url)

	}

	c.Wait()
	close(successChan)
	close(failChan)
}
