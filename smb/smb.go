package sitemapbuilder

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	lp "github.com/ugurakn/go-html-link-parser"
)

const xmlns = "http://www.sitemaps.org/schemas/sitemap/0.9"

type SMB struct {
	URL      *url.URL
	maxDepth int
	Sitemap  []string
	Visited  map[string]struct{}
}

func New(rawURL string, maxDepth int) (*SMB, error) {
	// remove trailing slash if it exists
	rawURL = removeTrailingSlash(rawURL)

	// validate root url. ensure it has a host and is absolute
	URL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing URL string (%s): %v", rawURL, err)
	}

	if !URL.IsAbs() || URL.Hostname() == "" {
		return nil, fmt.Errorf("invalid root URL (%s): root URL must be a valid absolute path", rawURL)
	}

	var l []string
	v := make(map[string]struct{})
	return &SMB{URL, maxDepth, l, v}, nil
}

type loc struct {
	Value string `xml:"loc"`
}

type urlset struct {
	Urls  []loc  `xml:"url"`
	Xmlns string `xml:"xmlns,attr"`
}

func (smb *SMB) Build() error {
	smb.getAll(smb.URL.String(), 0)

	// build the xml
	toXml := urlset{
		Urls:  make([]loc, len(smb.Sitemap)),
		Xmlns: xmlns,
	}

	for i, link := range smb.Sitemap {
		toXml.Urls[i] = loc{link}
	}

	fmt.Print(xml.Header)
	enc := xml.NewEncoder(os.Stdout)
	enc.Indent("", "  ")
	err := enc.Encode(toXml)
	if err != nil {
		return fmt.Errorf("xml marshal failed: %v", err)
	}

	return nil
}

func (smb *SMB) getAll(URL string, depth int) {
	if _, ok := smb.Visited[URL]; ok || depth > smb.maxDepth {
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

	depth++
	for _, l := range links {
		smb.getAll(l, depth)
	}
}

func getDocFromURL(URL string) (io.ReadCloser, error) {
	resp, err := http.Get(URL)
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
	links := make([]string, 0, len(allLinks))
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
