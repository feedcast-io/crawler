package main

import (
	"feedcast.crawler/crawler"
	"log/slog"
	"os"
)

func main() {
	testAkoutic()
}

func testAkoutic() {
	c := crawler.Config{
		Domain:   "https://www.akoustik-online.com",
		MaxPages: 100,
		//MaxDuration: 5,
		MaxDepth: 5,
	}

	h := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(h)
	slog.SetDefault(logger)

	//c.Domain = "https://airtable.com"
	c.Touch()

	cr := crawler.NewCrawler(c)

	data := cr.Run()

	slog.Info("End crawl", "domain", c.Domain, "found", len(data))
}
