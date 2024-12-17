package crawler

import (
	"strings"
	"testing"
)

func TestCrawler_Run(t *testing.T) {
	config := Config{
		//Domain: "www.lemonde.fr",
		Domain:      "www.champion-direct.com",
		MaxDuration: 10,
	}

	cr := NewCrawler(config)
	ch, err := cr.Run()

	if err != nil {
		t.Fatal(err)
	}

	foundPages := 0

	for ok := true; ok; {
		var page PageSummary
		select {
		case page, ok = <-ch:
			if ok {
				foundPages++
				if !strings.HasPrefix(page.Url, "https://") && !strings.HasPrefix(page.Url, "http://") {
					t.Errorf("invalid url for page crawled: %s", page.Url)
				}

				if len(page.Body) == 0 {
					t.Errorf("empty body for page crawled: %s", page.Url)
				}
				if len(page.Title) == 0 {
					t.Errorf("empty title for page crawled: %s", page.Title)
				}
			}
			break
		}
	}

	if foundPages == 0 {
		t.Errorf("no pages found")
	}
}
