package main

import (
	"flag"
	"log"
	"os"

	sitemapbuilder "github.com/ugurakn/sitemapbuilder/smb"
)

func main() {
	urlFlag := flag.String("u", "http://127.0.0.1:8080/", "the URL to build the sitemap for")
	depth := flag.Int("d", 3, "the maximum number of links deep to traverse")
	flag.Parse()

	smb, err := sitemapbuilder.New(*urlFlag, *depth)
	if err != nil {
		log.Fatal("error creating new smb: ", err)
	}

	err = smb.Build()
	if err != nil {
		log.Fatal("sitemap build failed: ", err)
	}

	if smb.Sitemap == nil {
		log.Println("no links found in", *urlFlag)
	} else {
		var b []byte
		for _, l := range smb.Sitemap {
			b = append(b, []byte(l+"\n")...)
		}
		err = os.WriteFile("links.txt", b, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
}
