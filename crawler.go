package crawler

import (
	"fmt"
	"github.com/gocolly/colly"
	"github.com/microcosm-cc/bluemonday"
	"html"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Crawler struct {
	config    Config
	collector *colly.Collector
	m         *sync.Mutex
	p         *bluemonday.Policy
	res       map[string]PageSummary
}

func NewCrawler(config Config) *Crawler {
	config.touch()

	p := bluemonday.NewPolicy()
	p.AllowImages()
	p.AllowElements("br")
	p.SkipElementsContent("header", "footer", "a", "script", "object")

	return &Crawler{
		config: config,
		collector: colly.NewCollector(
			colly.Async(true),
			colly.MaxDepth(int(config.MaxDepth)),
		),
		m: &sync.Mutex{},
		p: p,
	}
}

func (c *Crawler) savePageAttribute(url, value string, attr pageAttribute) {
	c.m.Lock()
	o, ok := c.res[url]
	if !ok {
		o = PageSummary{}
	}

	switch attr {
	case attrBody:
		o.Body = value
		break
	case attrTitle:
		o.Title = value
		break
	case attrKeywords:
		o.Keywords = value
		break
	case attrDesc:
		o.Description = value
		break
	}

	c.res[url] = o
	c.m.Unlock()
}

func (c *Crawler) Run() (chan PageSummary, error) {
	c.res = make(map[string]PageSummary)

	chPages := make(chan PageSummary)

	if e := c.config.Validate(); nil != e {
		close(chPages)
		return chPages, e
	}

	endCrawlLimit := time.Now().Add(time.Duration(c.config.MaxDuration) * time.Second)

	processed := 0

	c.collector.Limit(&colly.LimitRule{
		DomainGlob:  fmt.Sprintf("*%s*", c.config.Domain),
		Delay:       time.Millisecond * 100,
		Parallelism: 4,
	})

	c.collector.OnHTML("body", func(e *colly.HTMLElement) {
		re, _ := regexp.Compile(`(\s\s)\s*`)
		raw, _ := e.DOM.Html()
		raw = c.p.Sanitize(raw)
		raw = strings.ReplaceAll(raw, "<br>", "\n")
		raw = strings.ReplaceAll(raw, "<br/>", "\n")
		raw = strings.ReplaceAll(raw, "<br />", "\n")
		raw = html.UnescapeString(raw)
		raw = re.ReplaceAllString(raw, "$1")

		c.savePageAttribute(e.Request.URL.String(), raw, attrBody)
	})

	c.collector.OnHTML("title", func(e *colly.HTMLElement) {
		title, _ := e.DOM.Html()
		c.savePageAttribute(e.Request.URL.String(), title, attrTitle)
	})

	c.collector.OnHTML("meta[name=keywords]", func(e *colly.HTMLElement) {
		keywords, _ := e.DOM.Attr("content")
		c.savePageAttribute(e.Request.URL.String(), keywords, attrKeywords)
	})

	c.collector.OnHTML("meta[name=description]", func(e *colly.HTMLElement) {
		desc, _ := e.DOM.Attr("content")
		c.savePageAttribute(e.Request.URL.String(), desc, attrDesc)
	})

	c.collector.OnScraped(func(response *colly.Response) {
		if o, ok := c.res[response.Request.URL.String()]; ok && len(o.Body) > 0 {
			o.Url = response.Request.URL.String()
			chPages <- o
		}
	})

	c.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		parts := strings.Split(e.Attr("href"), "#")
		link := parts[0]

		ignoreLink := "nofollow" == e.Attr("rel") ||
			0 == len(link) ||
			strings.Contains(link, "javascript:") ||
			strings.Contains(link, "mailto:")

		// Ignore link if parent is header or footer depending on config
		if !c.config.KeepHeaderFooterLinks {
			for _, n := range e.DOM.Parents().Nodes {
				ignoreLink = ignoreLink || "header" == n.Data || "footer" == n.Data
			}
		}

		c.m.Lock()

		linkUrl := strings.Split(link, "?")[0]
		linkUrl = c.getUrlForCrawl(linkUrl)

		if _, alreadyVisited := c.res[linkUrl]; !alreadyVisited &&
			!ignoreLink &&
			len(linkUrl) > 0 &&
			processed < int(c.config.MaxPages) &&
			endCrawlLimit.After(time.Now()) {
			processed++
			c.res[linkUrl] = PageSummary{}
			e.Request.Visit(linkUrl)
		}
		c.m.Unlock()
	})

	go func() {
		c.collector.Visit(fmt.Sprintf("https://%s", c.config.Domain))
		c.collector.Wait()
		close(chPages)
	}()

	return chPages, nil
}

func (c *Crawler) getUrlForCrawl(linkUrl string) string {
	// If relative link (without-domain), add protocol & domain
	if strings.HasPrefix(linkUrl, "/") {
		return fmt.Sprintf("https://%s%s", c.config.Domain, linkUrl)
	}

	allowedDomains := []string{
		c.config.Domain,
	}

	// Allow www.domain.com if domain.com provided
	if !strings.HasPrefix(c.config.Domain, "www.") {
		allowedDomains = append(allowedDomains, fmt.Sprintf("www.%s", c.config.Domain))
	}

	for _, d := range allowedDomains {
		if strings.HasPrefix(linkUrl, fmt.Sprintf("https://%s", d)) {
			return linkUrl
		}
	}

	return ""
}
