# Feedcast Domain Crawler

## Intro

This project is a simple web crawler designed to scrape and index data from websites. The crawler follows links across a site, extracting relevant information like html body, meta keywords, title, and description. It is built with efficiency in mind and supports multiple configurations for customization.

## Requirements
- Go 1.22
- Libraries:
  - github.com/gocolly/colly 
  - github.com/microcosm-cc/bluemonday

## Usage

```go
import "github.com/feedcast-io/crawler"

c := crawler.Config{
    Domain:   "https://www.example.com",
    MaxPages: 100,
    MaxDepth: 5,
}

cr := crawler.NewCrawler(c)
ch, _ := cr.Run()

for page := range ch {
    // Process page content
    // log.Println(page.Url, page.Body)
}


```
