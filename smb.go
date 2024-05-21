package sitemapbuilder

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	lp "github.com/ugurakn/go-html-link-parser"
)

type SMB struct {
	URL     *url.URL
	Sitemap []string
	Visited map[string]struct{}
}

func getClient() *http.Client {
	c := &http.Client{}

	return c
}

func New(rawURL string) (*SMB, error) {
	// remove trailing slash if it exists
	rawURL = removeTrailingSlash(rawURL)

	// validate root url. ensure it has a host and is absolute
	URL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing raw URL (%s): %v", rawURL, err)
	}

	if !URL.IsAbs() || URL.Hostname() == "" {
		return nil, fmt.Errorf("invalid root URL (%s): root URL must be absolute with a host, and not a relative path", rawURL)
	}

	var l []string
	v := make(map[string]struct{})
	return &SMB{URL, l, v}, nil
}

func (smb *SMB) Build() error {
	smb.getAll(smb.URL.String())

	return nil
}

func (smb *SMB) getAll(URL string) {
	if _, ok := smb.Visited[URL]; ok {
		return
	}
	smb.Sitemap = append(smb.Sitemap, URL)
	smb.Visited[URL] = struct{}{}

	doc, err := getDocFromURL(URL)
	if err != nil {
		panic(fmt.Errorf("can't get doc from URL (%v): %v", URL, err))
	}

	if doc == nil {
		return
	}

	links, err := smb.getLinksFromDoc(doc)
	if err != nil {
		panic(fmt.Errorf("can't get links from doc: %v", err))
	}

	for _, l := range links {
		smb.getAll(l)
	}
}

func getDocFromURL(URL string) (io.ReadCloser, error) {
	c := getClient()

	resp, err := c.Get(URL)
	if err != nil {
		return nil, err
	}

	cType := resp.Header.Get("content-type")

	if resp.StatusCode == 200 && strings.Contains(cType, "text/html") {
		return resp.Body, nil
	}

	return nil, nil
}

func (smb *SMB) getLinksFromDoc(doc io.ReadCloser) ([]string, error) {
	allLinks, err := lp.Parse(doc)
	doc.Close()
	if err != nil {
		return nil, fmt.Errorf("error getting links from doc: %v", err)
	}

	// only include links to the same host
	var links []string
	for _, link := range allLinks {
		linkURL, err := url.Parse(link.Href)
		if err != nil {
			return nil, fmt.Errorf("can't parse URL (%v): %v", link.Href, err)
		}

		if !linkURL.IsAbs() && linkURL.Host == "" {
			linkURL = smb.URL.ResolveReference(linkURL)
		}

		if linkURL.Scheme == smb.URL.Scheme && linkURL.Hostname() == smb.URL.Hostname() && !strings.Contains(link.Href, "#") {
			linkStr := removeTrailingSlash(linkURL.String())
			links = append(links, linkStr)
		}
	}

	return links, nil
}

func removeTrailingSlash(s string) string {
	return strings.TrimSuffix(s, "/")
}
