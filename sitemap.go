package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/benjaminestes/robots"

	"github.com/temoto/robotstxt"
)

func fetchRobotsTxt(domain string) (*robotstxt.RobotsData, error) {
	robotsLink, err := robots.Locate(domain)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(robotsLink)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := robotstxt.FromResponse(resp)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func extractSitemapURLs(data *robotstxt.RobotsData) []string {
	return data.Sitemaps
}

func fetchAndParseXML(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func parseSitemap(url string) ([]string, error) {
	data, err := fetchAndParseXML(url)
	if err != nil {
		return nil, err
	}

	var urlSet URLSet
	err = xml.Unmarshal(data, &urlSet)
	if err == nil {
		var urls []string
		for _, url := range urlSet.URLs {
			urls = append(urls, url.Loc)
		}
		return urls, nil
	}

	var sitemapIndex SitemapIndex
	err = xml.Unmarshal(data, &sitemapIndex)
	if err == nil {
		var urls []string
		for _, sitemap := range sitemapIndex.Sitemaps {
			childURLs, err := parseSitemap(sitemap.Loc)
			if err != nil {
				return nil, err
			}
			urls = append(urls, childURLs...)
		}
		return urls, nil
	}

	return nil, fmt.Errorf("unable to parse sitemap or sitemap index: %v", err)
}

func parseSitemaps(sitemapURLs []string) ([]string, error) {
	var urls []string
	for _, sitemapURL := range sitemapURLs {
		childURLs, err := parseSitemap(sitemapURL)
		if err != nil {
			return nil, err
		}
		urls = append(urls, childURLs...)
	}
	return urls, nil
}

func GetSitemapUrls(domain string) ([]string, error) {
	// time the function
	defer TimeTrack(time.Now(), "GetSitemapUrls")

	robotsData, err := fetchRobotsTxt(domain)
	if err != nil {
		fmt.Println("Error fetching robots.txt:", err)
		return nil, err
	}

	sitemapURLs := extractSitemapURLs(robotsData)
	urls, err := parseSitemaps(sitemapURLs)
	if err != nil {
		fmt.Println("Error parsing sitemaps:", err)
		return nil, err
	}

	var sanitisedUrls []string
	urlInList := make(map[string]bool)

	for _, url := range urls {
		sanitisedUrl, error := SanitiseURL(url)
		if error != nil {
			fmt.Println("Error sanitising url:", error)
			return nil, error
		}

		if _, ok := urlInList[sanitisedUrl]; ok {
			continue
		}

		sanitisedUrls = append(sanitisedUrls, sanitisedUrl)
		urlInList[sanitisedUrl] = true

	}

	return sanitisedUrls, nil
}

func TimeTrack(start time.Time, functionName string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s\n", functionName, elapsed)
}
