package main

import (
	"github.com/feedcast-io/crawler"
	"log"
	"log/slog"
)

func main() {
	testAkoutic()
}

func testAkoutic() {
	c := crawler.Config{
		Domain:   "https://www.lemonde.fr/",
		MaxPages: 100,
		//MaxDuration: 5,
		MaxDepth: 5,
	}

	cr := crawler.NewCrawler(c)

	if ch, e := cr.Run(); nil != e {
		log.Fatal(e)
	} else {
		log.Println("Start process")

		found := 0
		for page := range ch {
			log.Println("Fetch success", page.Url)
			found++
		}

		slog.Info("End crawl")
	}
}
