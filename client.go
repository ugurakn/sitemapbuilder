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

type SMB struct {
	URL     *url.URL
	Sitemap []lp.Link
}

func getClient() *http.Client {
	c := &http.Client{}

	return c
}

func New(rawURL string) (*SMB, error) {
	// validate main url. ensure it has a host and is abs,
	// and is not a relative path
	URL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing raw URL (%s): %v", rawURL, err)
	}

	if !URL.IsAbs() || URL.Hostname() == "" {
		return nil, fmt.Errorf("invalid root URL (%s): root URL must be absolute with a host, and not a relative path", rawURL)
	}

	var l []lp.Link
	return &SMB{URL, l}, nil
}

// Build builds a sitemap using the rawURL as the root URL.
// It only includes links to the same scheme and host.
// NB for now, it returns a []Link.
func (smb *SMB) Build() error {
	smb.getAll(smb.URL.String())
	return nil
}

func (smb *SMB) getAll(URL string) {
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
		fmt.Printf("found: %s\n", l.Href)

		smb.Sitemap = append(smb.Sitemap, l)
		smb.getAll(l.Href)
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

func (smb *SMB) getLinksFromDoc(doc []byte) ([]lp.Link, error) {
	allLinks, err := lp.Parse(bytes.NewReader(doc))
	if err != nil {
		return nil, fmt.Errorf("error getting links from doc: %v", err)
	}

	// only include links to the same host
	var links []lp.Link
	for _, l := range allLinks {
		linkURL, _ := url.Parse(l.Href)

		if !linkURL.IsAbs() && linkURL.Host == "" {
			linkURL = smb.URL.ResolveReference(linkURL)
		}

		if linkURL.Scheme == smb.URL.Scheme && linkURL.Hostname() == smb.URL.Hostname() && !strings.Contains(l.Href, "#") {
			l.Href = linkURL.String()
			links = append(links, l)
		}
	}

	return links, nil
}
