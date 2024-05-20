package sitemapbuilder

import (
	"fmt"
	"net/http"
	"strings"

	lp "github.com/ugurakn/go-html-link-parser"
)

func getClient() *http.Client {
	c := &http.Client{}

	return c
}

func GetLinksFromURL(url string) ([]lp.Link, error) {
	c := getClient()

	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// get content-type from header
	cType := resp.Header.Get("content-type")

	// fmt.Println(resp.StatusCode)
	// fmt.Println(cType)

	if resp.StatusCode == 200 && strings.Contains(cType, "text/html") {
		links, err := lp.Parse(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error parsing response body: %v", err)
		}
		return links, nil
	}

	return nil, nil
}

// func ExtractHostFromURL(rawURL string) {
// 	URL, err := url.Parse(rawURL)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Println(URL.Hostname())
// 	fmt.Println(URL.Host)
// 	fmt.Println(URL.Path)
// }
