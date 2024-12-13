package crawler

import (
	"cmp"
	"errors"
	"regexp"
	"strings"
)

type Config struct {
	Domain      string `json:"domain"`
	MaxPages    uint16 `json:"max_pages"`
	MaxDuration uint8  `json:"max_duration"`
	MaxDepth    uint8  `json:"max_depth"`
	//WithPageContent       bool   `json:"with_page_content"`
	KeepHeaderFooterLinks bool `json:"keep_header_links"`
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
