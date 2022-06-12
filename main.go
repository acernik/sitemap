package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
)

type URL struct {
	XMLName  xml.Name `xml:"url"`
	Location string   `xml:"loc"`
}

type Sitemap struct {
	XMLName xml.Name `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
	URLs    []URL
}

func main() {
	var websiteURL string
	var maxGoroutines int
	var outPutFile string
	var maxDepth int

	flag.StringVar(&websiteURL, "url", "https://www.coastal.edu", "Specify URL to be parsed. Default is https://www.coastal.edu/ .")
	flag.IntVar(&maxGoroutines, "parallel", 1, "Specify number of parallel workers to navigate through site. Default is 1.")
	flag.StringVar(&outPutFile, "output-file", "sitemap.xml", "Specify output file path. Default is sitemap.xml.")
	flag.IntVar(&maxDepth, "max-depth", 5, "Specify max depth of URL navigation recursion. Default is 5.")

	flag.Parse()

	runtime.GOMAXPROCS(maxGoroutines)

	u, err := url.Parse(websiteURL)
	if err != nil {
		log.Fatal(err)
	}

	response, err := http.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}

	sitemapURLs := make(map[string]URL)

	sitemapChan := make(chan URL)
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

	createSitemap(response, u, sitemapChan, errChan, &parallelGroup, maxDepth)

	parallelGroup.Wait()

	close(sitemapChan)
	close(errChan)

	sitemapGroup.Wait()
	errGroup.Wait()

	var sitemap Sitemap

	for _, sitemapURL := range sitemapURLs {
		sitemap.URLs = append(sitemap.URLs, sitemapURL)
	}

	err = writeSitemapToFile(outPutFile, sitemap)
	if err != nil {
		log.Fatal(err)
	}
}

func createSitemap(response *http.Response, u *url.URL, sitemapChan chan<- URL, errChan chan<- error, parallelGroup *sync.WaitGroup, maxDepth int) {
	defer response.Body.Close()

	if maxDepth == 0 {
		parallelGroup.Done()
		return
	}

	var baseElement string

	htmlTokens := html.NewTokenizer(response.Body)

loop:
	for {
		tt := htmlTokens.Next()

		switch tt {
		case html.ErrorToken:
			fmt.Println("End")
			break loop
		case html.StartTagToken:
			t := htmlTokens.Token()

			if t.Data == "a" {
				for _, attr := range t.Attr {
					if attr.Key == "href" && len(attr.Val) > 0 && attr.Val != "/" && !strings.HasPrefix(attr.Val, "#") {
						if u.Path == attr.Val {
							continue
						}

						baseURL := u.Scheme + "://" + u.Host

						if len(baseElement) > 0 {
							baseURL = baseElement
						}

						link := baseURL + attr.Val

						if strings.HasPrefix(attr.Val, "http") {
							if !strings.HasPrefix(attr.Val, baseURL) {
								continue
							}

							fullURL, err := url.Parse(attr.Val)
							if err == nil {
								link = fullURL.String()
							}
						}

						sitemapChan <- URL{Location: link}

						nextURL, err := url.Parse(link)
						if err != nil {
							errChan <- err
							continue
						}

						resp, err := http.Get(nextURL.String())
						if err != nil {
							errChan <- err
							break
						}

						if maxDepth > 0 {
							maxDepth -= 1
							parallelGroup.Add(1)
							go createSitemap(resp, nextURL, sitemapChan, errChan, parallelGroup, maxDepth)
						}
					}
				}
			}

			if t.Data == "base" {
				for _, attr := range t.Attr {
					if attr.Key == "href" {
						baseElementURL, err := url.Parse(attr.Val)
						if err != nil {
							errChan <- err
							continue
						}

						baseElement = baseElementURL.String()

						fmt.Println("Found base element: ", baseElement)
					}
				}
			}
		}
	}

	parallelGroup.Done()
}

func writeSitemapToFile(filename string, sitemap Sitemap) error {
	if _, err := os.Stat(filename); err == nil {
		err := os.Remove(filename)
		if err != nil {
			return err
		}
	}

	xmlFile, err := os.Create(filename)
	if err != nil {
		return err
	}

	_, err = xmlFile.WriteString(xml.Header)
	if err != nil {
		return err
	}

	xmlWriter := io.Writer(xmlFile)

	enc := xml.NewEncoder(xmlWriter)
	enc.Indent("  ", "    ")

	err = enc.Encode(sitemap)
	if err != nil {
		return err
	}

	return nil
}
