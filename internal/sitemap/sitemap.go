package sitemap

import (
	"encoding/xml"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

// URL holds one links value found on the website.
type URL struct {
	XMLName  xml.Name `xml:"url"`
	Location string   `xml:"loc"`
}

// Sitemap holds a list of URL values.
type Sitemap struct {
	XMLName xml.Name `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
	URLs    []URL
}

// CreateSitemap recursively navigates through site pages and sends found links to the channel. Once received in the
// channel the links are stored in a map. These links are later written by WriteSitemapToFile function to an XML file specified.
func CreateSitemap(u *url.URL, sitemapChan chan<- URL, errChan chan<- error, parallelGroup *sync.WaitGroup, maxDepth int) {
	if u == nil {
		parallelGroup.Done()
		return
	}

	if maxDepth == 0 {
		parallelGroup.Done()
		return
	}

	response, err := http.Get(u.String())
	defer response.Body.Close()
	if err != nil {
		errChan <- err
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

						if maxDepth > 0 {
							maxDepth -= 1
							parallelGroup.Add(1)
							go CreateSitemap(nextURL, sitemapChan, errChan, parallelGroup, maxDepth)
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

// WriteSitemapToFile takes value of type Sitemap and writes it to the sitemap file.
func WriteSitemapToFile(fileName string, sitemap Sitemap) error {
	if _, err := os.Stat(fileName); err == nil {
		err := os.Remove(fileName)
		if err != nil {
			return err
		}
	}

	xmlFile, err := os.Create(fileName)
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
