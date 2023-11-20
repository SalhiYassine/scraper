package main

import (
	"encoding/xml"
	"sync"
)

type PageSuccess struct {
	Url     string `json:"url"`
	Content string `json:"content"`
}

type PageError struct {
	Url string `json:"url"`
	Err error  `json:"error"`
}

type ScrapeResult struct {
	Success []PageSuccess `json:"successful"`
	Error   []PageError   `json:"failed"`
}

var googleUserAgent = "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/W.X.Y.Z Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"

type ScrapeOptions struct {
	StartingURLs    []string
	ValidDomains    []string
	Concurrency     int
	MaxDepth        int
	MinTimeBetween  int
	MaxPagesToVisit int
}

// SafeMap wraps a map with a mutex to enable safe concurrent access.
type SafeMap struct {
	mu    sync.Mutex
	items map[string]interface{}
}

// NewSafeMap initializes and returns a new SafeMap.
func NewSafeMap() *SafeMap {
	return &SafeMap{
		items: make(map[string]interface{}),
	}
}

// Set safely sets a key-value pair in the map.
func (sm *SafeMap) Set(key string, value interface{}) {
	sm.mu.Lock() // Lock the mutex before accessing the map

	sm.items[key] = value

	sm.mu.Unlock() // Unlock the mutex after accessing the map
}

// Get safely retrieves a value from the map by key.
func (sm *SafeMap) Get(key string) (interface{}, bool) {
	sm.mu.Lock() // Lock the mutex before accessing the map

	value, exists := sm.items[key]

	sm.mu.Unlock() // Unlock the mutex after accessing the map
	return value, exists
}

type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []URL    `xml:"url"`
}

type URL struct {
	Loc string `xml:"loc"`
}

type SitemapIndex struct {
	XMLName  xml.Name  `xml:"sitemapindex"`
	Sitemaps []Sitemap `xml:"sitemap"`
}

type Sitemap struct {
	Loc string `xml:"loc"`
}
