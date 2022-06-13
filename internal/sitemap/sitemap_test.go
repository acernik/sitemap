package sitemap

import (
	"fmt"
	"net/url"
	"os"
	"runtime"
	"sync"
	"testing"
)

func TestSitemap_WriteSitemapToFile(t *testing.T) {
	tests := []struct {
		fileName  string
		sitemap   Sitemap
		shouldErr bool
	}{
		{
			fileName: "sitemap_test.xml",
			sitemap: Sitemap{
				URLs: []URL{
					{
						Location: "https://www.test-1.com",
					},
					{
						Location: "https://www.test-2.com",
					},
					{
						Location: "https://www.test-3.com",
					},
				},
			},
			shouldErr: false,
		},
		{
			fileName: "",
			sitemap: Sitemap{
				URLs: []URL{
					{
						Location: "https://www.test-1.com",
					},
				},
			},
			shouldErr: true,
		},
		{
			fileName: "sitemap_test.xml",
			sitemap: Sitemap{
				URLs: nil,
			},
			shouldErr: true,
		},
	}

	for _, test := range tests {
		err := WriteSitemapToFile(test.fileName, test.sitemap)
		if err != nil && !test.shouldErr {
			t.Errorf("expected error to be nil, got %v", err)
		}
	}

	if _, err := os.Stat("sitemap_test.xml"); err == nil {
		err := os.Remove("sitemap_test.xml")
		if err != nil {
			t.Errorf("expected error to be nil, got %v", err)
		}
	}
}

func TestSitemap_CreateSitemap(t *testing.T) {
	websiteURL := "https://getaurox.com/"
	maxParallel := 4
	outPutFile := "test_sitemap_ok.xml"
	maxDepth := 1

	runtime.GOMAXPROCS(maxParallel)

	u, err := url.Parse(websiteURL)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}

	sitemapURLs := make(map[string]URL)

	sitemapChan := make(chan URL)
	sitemapGroup := sync.WaitGroup{}
	sitemapGroup.Add(1)
	go func() {
		for sitemapURL := range sitemapChan {
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

	CreateSitemap(u, sitemapChan, errChan, &parallelGroup, maxDepth)

	parallelGroup.Wait()

	close(sitemapChan)
	close(errChan)

	sitemapGroup.Wait()
	errGroup.Wait()

	var result Sitemap

	for _, sitemapURL := range sitemapURLs {
		result.URLs = append(result.URLs, sitemapURL)
	}

	err = WriteSitemapToFile(outPutFile, result)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}

	if _, err := os.Stat("test_sitemap_ok.xml"); err == nil {
		err := os.Remove("test_sitemap_ok.xml")
		if err != nil {
			t.Errorf("expected error to be nil, got %v", err)
		}
	}
}

func TestSitemap_CreateSitemap_InvalidURL(t *testing.T) {
	maxParallel := 4
	maxDepth := 1

	runtime.GOMAXPROCS(maxParallel)

	sitemapURLs := make(map[string]URL)

	sitemapChan := make(chan URL)
	sitemapGroup := sync.WaitGroup{}
	sitemapGroup.Add(1)
	go func() {
		for sitemapURL := range sitemapChan {
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

	CreateSitemap(nil, sitemapChan, errChan, &parallelGroup, maxDepth)

	parallelGroup.Wait()

	close(sitemapChan)
	close(errChan)

	sitemapGroup.Wait()
	errGroup.Wait()
}

func TestSitemap_CreateSitemap_MaxDepthZero(t *testing.T) {
	websiteURL := "https://getaurox.com/"
	maxParallel := 4
	maxDepth := 0

	runtime.GOMAXPROCS(maxParallel)

	u, err := url.Parse(websiteURL)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}

	sitemapURLs := make(map[string]URL)

	sitemapChan := make(chan URL)
	sitemapGroup := sync.WaitGroup{}
	sitemapGroup.Add(1)
	go func() {
		for sitemapURL := range sitemapChan {
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

	CreateSitemap(u, sitemapChan, errChan, &parallelGroup, maxDepth)

	parallelGroup.Wait()

	close(sitemapChan)
	close(errChan)

	sitemapGroup.Wait()
	errGroup.Wait()
}
