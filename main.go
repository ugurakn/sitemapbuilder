package main

import (
	"flag"
	"log"

	sitemapbuilder "github.com/ugurakn/sitemapbuilder/smb"
)

func main() {
	urlFlag := flag.String("u", "http://127.0.0.1:8080/", "the URL to build the sitemap for")
	depth := flag.Int("d", 5, "the maximum depth of links to traverse. setting it to 1 means extracting from only the root URL u. Panics if d < 1.")
	flag.Parse()

	// panic if maxDepth < 1
	if *depth < 1 {
		panic("maxDepth must be 1 or higher.")
	}

	smb, err := sitemapbuilder.New(*urlFlag, *depth)
	if err != nil {
		log.Fatal("error creating new smb: ", err)
	}

	err = smb.Build()
	if err != nil {
		log.Fatal("sitemap build failed: ", err)
	}
}
