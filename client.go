package sitemapbuilder

import (
	"bytes"
	"fmt"
	"io"
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
		return nil, fmt.Errorf("error parsing raw URL (%s): %v", rawURL, err)
	}

	if !url.IsAbs() || url.Hostname() == "" {
		return nil, fmt.Errorf("invalid root URL (%s): root URL must be absolute with a host, and not a relative path", rawURL)
	}

	var sitemap []lp.Link
	getAll(url, &sitemap)
	return sitemap, nil

	// doc, err := getDocFromURL(rawURL)
	// if err != nil {
	// 	return nil, err
	// }

	// if doc != nil {
	// 	links, err := getLinksFromDoc(url, doc)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	return links, nil
	// }
	// return nil, nil
}

func getAll(URL *url.URL, sm *[]lp.Link) {
	doc, err := getDocFromURL(URL.String())
	if err != nil {
		panic(fmt.Errorf("can't get doc from URL (%v): %v", URL, err))
	}

	if doc == nil {
		return
	}

	links, err := getLinksFromDoc(URL, doc)
	if err != nil {
		panic(fmt.Errorf("can't get links from doc: %v", err))
	}

	for _, l := range links {
		fmt.Printf("found: %s\n", l.Href)

		*sm = append(*sm, l)
		getAll(URL, sm)
	}

}

func getDocFromURL(URL string) ([]byte, error) {
	c := getClient()

	resp, err := c.Get(URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cType := resp.Header.Get("content-type")

	// fmt.Println(resp.StatusCode)
	// fmt.Println(cType)

	if resp.StatusCode == 200 && strings.Contains(cType, "text/html") {
		doc, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %v", err)
		}
		return doc, nil
	}

	return nil, nil
}

func getLinksFromDoc(url *url.URL, doc []byte) ([]lp.Link, error) {
	scheme := url.Scheme
	host := url.Hostname()

	allLinks, err := lp.Parse(bytes.NewReader(doc))
	if err != nil {
		return nil, fmt.Errorf("error getting links from doc: %v", err)
	}
	// only include links to the same host
	var links []lp.Link
	for _, l := range allLinks {
		linkURL, _ := url.Parse(l.Href)
		if linkURL.Scheme == scheme && linkURL.Hostname() == host && !strings.Contains(l.Href, "#") {
			l.Href = linkURL.String()
			links = append(links, l)
		}
	}

	return links, nil
}
