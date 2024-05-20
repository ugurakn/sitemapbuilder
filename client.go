package sitemapbuilder

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	lp "github.com/ugurakn/go-html-link-parser"
)

func getClient() *http.Client {
	c := &http.Client{}

	return c
}

// Build builds a sitemap using the rawURL as the root URL.
// It only includes links to the same scheme and host.
// NB for now, it returns a []Link.
func Build(rawURL string) ([]lp.Link, error) {
	// validate main url. ensure it has a host and is abs,
	// and is not a relative path
	url, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing raw URL %s: %v", rawURL, err)
	}

	if !url.IsAbs() || url.Hostname() == "" {
		return nil, fmt.Errorf("invalid root URL (%s): root URL must be absolute with a host, and not a relative path", rawURL)
	}

	links, err := getLinksFromURL(url)
	if err != nil {
		return nil, err
	}
	return links, nil
}

func getLinksFromURL(url *url.URL) ([]lp.Link, error) {
	scheme := url.Scheme
	host := url.Hostname()
	// fmt.Printf("scheme: %s, host: %s\n", scheme, host)

	c := getClient()

	resp, err := c.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cType := resp.Header.Get("content-type")

	// fmt.Println(resp.StatusCode)
	// fmt.Println(cType)

	if resp.StatusCode == 200 && strings.Contains(cType, "text/html") {
		allLinks, err := lp.Parse(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error parsing response body: %v", err)
		}
		// only include links to the same host
		var links []lp.Link
		for _, l := range allLinks {
			linkURL, _ := url.Parse(l.Href)
			if linkURL.Scheme == scheme && linkURL.Hostname() == host && !strings.Contains(l.Href, "#") {
				links = append(links, l)
			}
		}

		return links, nil
	}

	return nil, nil
}
