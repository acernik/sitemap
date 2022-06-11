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
	"strings"
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
	var parallel int
	var outPutFile string
	var maxDepth int

	flag.StringVar(&websiteURL, "url", "https://www.coastal.edu", "Specify URL to be parsed. Default is https://www.coastal.edu/ .")
	flag.IntVar(&parallel, "parallel", 1, "Specify number of parallel workers to navigate through site. Default is 1.")
	flag.StringVar(&outPutFile, "output-file", "sitemap.xml", "Specify output file path. Default is sitemap.xml.")
	flag.IntVar(&maxDepth, "max-depth", 5, "Specify max depth of URL navigation recursion. Default is 5.")

	flag.Parse()

	u, err := url.Parse(websiteURL)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("u.Scheme: ", u.Scheme)
	fmt.Println("u.Host: ", u.Host)
	fmt.Println("u.Path: ", u.Path)
	fmt.Println("u.String(): ", u.String())

	response, err := http.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	sitemap, err := createSitemap(response.Body, u)
	if err != nil {
		log.Fatal(err)
	}

	err = writeSitemapToFile(outPutFile, sitemap)
	if err != nil {
		log.Fatal(err)
	}
}

func createSitemap(body io.Reader, u *url.URL) (Sitemap, error) {
	var baseElement string

	htmlTokens := html.NewTokenizer(body)

	var sitemap Sitemap

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

						if strings.HasPrefix(attr.Val, "http") || strings.HasPrefix(attr.Val, "www") {
							if !strings.HasPrefix(attr.Val, baseURL) {
								continue
							}

							fullURL, err := url.Parse(attr.Val)
							if err == nil {
								link = fullURL.String()
							}
						}

						sitemap.URLs = append(sitemap.URLs, URL{
							Location: link,
						})
					}
				}
			}

			if t.Data == "base" {
				for _, attr := range t.Attr {
					if attr.Key == "href" {
						baseElementURL, err := url.Parse(attr.Val)
						if err != nil {
							return sitemap, err
						}

						baseElement = baseElementURL.String()

						fmt.Println("Found base element: ", baseElement)
					}
				}
			}
		}
	}

	return sitemap, nil
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
