package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"runtime"
	"sync"

	"github.com/acernik/sitemap/internal/sitemap"
)

func main() {
	var websiteURL string
	var maxParallel int
	var outPutFile string
	var maxDepth int

	flag.StringVar(&websiteURL, "url", "https://getaurox.com/", "Specify URL to be parsed. Default is https://www.coastal.edu/ .")
	flag.IntVar(&maxParallel, "parallel", 1, "Specify number of parallel workers to navigate through site. Default is 1.")
	flag.StringVar(&outPutFile, "output-file", "sitemap.xml", "Specify output file path. Default is sitemap.xml.")
	flag.IntVar(&maxDepth, "max-depth", 1, "Specify max depth of URL navigation recursion. Default is 1.")

	flag.Parse()

	runtime.GOMAXPROCS(maxParallel)

	u, err := url.Parse(websiteURL)
	if err != nil {
		log.Fatal(err)
	}

	sitemapURLs := make(map[string]sitemap.URL)

	sitemapChan := make(chan sitemap.URL)
	sitemapGroup := sync.WaitGroup{}
	sitemapGroup.Add(1)
	go func() {
		for sitemapURL := range sitemapChan {
			fmt.Printf("found sitemap URL: %s\n", sitemapURL.Location)
			sitemapURLs[sitemapURL.Location] = sitemapURL
		}
		sitemapGroup.Done()
	}()

	errChan := make(chan error)
	errGroup := sync.WaitGroup{}
	errGroup.Add(1)
	go func() {
		for err := range errChan {
			fmt.Printf("error creating sitemap: %s\n", err)
		}
		errGroup.Done()
	}()

	parallelGroup := sync.WaitGroup{}
	parallelGroup.Add(1)

	sitemap.CreateSitemap(u, sitemapChan, errChan, &parallelGroup, maxDepth)

	parallelGroup.Wait()

	close(sitemapChan)
	close(errChan)

	sitemapGroup.Wait()
	errGroup.Wait()

	var result sitemap.Sitemap

	for _, sitemapURL := range sitemapURLs {
		result.URLs = append(result.URLs, sitemapURL)
	}

	err = sitemap.WriteSitemapToFile(outPutFile, result)
	if err != nil {
		log.Fatal(err)
	}
}
