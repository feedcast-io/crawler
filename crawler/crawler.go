package crawler

import (
	"cmp"
	"errors"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/microcosm-cc/bluemonday"
	"html"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Domain                string `json:"domain"`
	MaxPages              uint16 `json:"max_pages"`
	MaxDuration           uint8  `json:"max_duration"`
	MaxDepth              uint8  `json:"max_depth"`
	WithPageContent       bool   `json:"with_page_content"`
	KeepHeaderFooterLinks bool   `json:"keep_header_links"`
}

func (c *Config) Touch() {
	c.MaxPages = cmp.Or(c.MaxPages, 100)
	c.MaxDuration = cmp.Or(c.MaxDuration, 30)
	c.MaxDepth = cmp.Or(c.MaxDepth, 4)

	// Sanitize domain if complete url furnished
	c.Domain = strings.ReplaceAll(c.Domain, "https://", "")
	c.Domain = strings.ReplaceAll(c.Domain, "http://", "")
	c.Domain = strings.TrimRight(c.Domain, "/")
}

func (c *Config) Validate() error {
	if ok, err := regexp.Match(`^[a-z0-9\.\-]+(\.[a-z0-9\.\-]+){1,}$`, []byte(c.Domain)); !ok || err != nil {
		return cmp.Or(err, errors.New("Invalid domain"))
	}

	return nil
}

type crawler struct {
	config    Config
	collector *colly.Collector
	m         *sync.Mutex
	p         *bluemonday.Policy
}

func NewCrawler(config Config) *crawler {
	p := bluemonday.UGCPolicy()

	p = bluemonday.NewPolicy()

	p.AllowImages()
	p.AllowElements("br")
	p.SkipElementsContent("header", "footer", "a", "script", "object")

	return &crawler{
		config: config,
		collector: colly.NewCollector(
			colly.Async(true),
			colly.MaxDepth(int(config.MaxDepth)),
		),
		m: &sync.Mutex{},
		p: p,
	}
}

type SiteContent struct {
	Description string `json:"description"`
	Body        string `json:"body"`
	Title       string `json:"title"`
	Keywords    string `json:"keywords"`
}

func (c *crawler) Run() map[string]SiteContent {
	result := make(map[string]SiteContent)

	maxBeforeSkip := time.Now().Add(time.Duration(c.config.MaxDuration) * time.Second)

	processed := 0

	c.collector.Limit(&colly.LimitRule{
		DomainGlob:  fmt.Sprintf("*%s*", c.config.Domain),
		Parallelism: 3,
	})

	c.collector.OnHTML("body", func(e *colly.HTMLElement) {
		raw, _ := e.DOM.Html()
		c.m.Lock()
		key := e.Request.URL.String()
		o, ok := result[key]
		if !ok {
			o = SiteContent{}
		}
		o.Body = c.p.Sanitize(raw)
		o.Body = strings.ReplaceAll(o.Body, "<br>", "\n")
		o.Body = strings.ReplaceAll(o.Body, "<br/>", "\n")
		o.Body = strings.ReplaceAll(o.Body, "<br />", "\n")
		o.Body = html.UnescapeString(o.Body)
		re, _ := regexp.Compile(`(\s\s)\s*`)
		o.Body = re.ReplaceAllString(o.Body, "$1")
		result[key] = o
		c.m.Unlock()
	})

	c.collector.OnHTML("title", func(e *colly.HTMLElement) {
		c.m.Lock()
		key := e.Request.URL.String()
		o, ok := result[key]
		if !ok {
			o = SiteContent{}
		}

		o.Title, _ = e.DOM.Html()
		result[key] = o

		c.m.Unlock()
	})

	c.collector.OnHTML("meta[name=keywords]", func(e *colly.HTMLElement) {
		c.m.Lock()
		key := e.Request.URL.String()
		o, ok := result[key]
		if !ok {
			o = SiteContent{}
		}
		o.Keywords, _ = e.DOM.Attr("content")
		result[key] = o

		c.m.Unlock()
	})

	c.collector.OnHTML("meta[name=description]", func(e *colly.HTMLElement) {
		c.m.Lock()
		key := e.Request.URL.String()
		o, ok := result[key]
		if !ok {
			o = SiteContent{}
		}
		o.Description, _ = e.DOM.Attr("content")
		result[key] = o
		c.m.Unlock()
	})

	c.collector.OnScraped(func(response *colly.Response) {
		log.Println("End page crawl", response.Request.URL.String())
	})

	c.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		parts := strings.Split(e.Attr("href"), "#")
		link := parts[0]

		//isRelative := 0 == strings.Index(link, "/")

		ignoreLink := "nofollow" == e.Attr("rel") ||
			0 == len(link) ||
			strings.Contains(link, "javascript:") ||
			strings.Contains(link, "mailto:")

		// Ignore link if parent is header or footer depending on config
		if !c.config.KeepHeaderFooterLinks {
			for _, n := range e.DOM.Parents().Nodes {
				ignoreLink = ignoreLink ||
					"header" == n.Data ||
					"footer" == n.Data
			}
		}

		c.m.Lock()

		linkUrl := strings.Split(link, "?")[0]
		linkUrl = c.getUrlForCrawl(linkUrl)

		if _, alreadyVisited := result[linkUrl]; !alreadyVisited &&
			!ignoreLink &&
			len(linkUrl) > 0 &&
			processed < int(c.config.MaxPages) &&
			maxBeforeSkip.After(time.Now()) {
			processed++
			result[linkUrl] = SiteContent{}
			e.Request.Visit(linkUrl)
		}
		c.m.Unlock()
	})

	c.collector.Visit(fmt.Sprintf("https://%s", c.config.Domain))
	c.collector.Wait()

	for k, v := range result {
		if 0 == len(v.Body) {
			delete(result, k)
		}
	}

	return result
}

func (c *crawler) getUrlForCrawl(linkUrl string) string {
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
