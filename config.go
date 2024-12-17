package crawler

import (
	"cmp"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Config struct {
	Domain                string `json:"domain"`
	MaxPages              uint16 `json:"max_pages"`
	MaxDuration           uint8  `json:"max_duration"`
	MaxDepth              uint8  `json:"max_depth"`
	KeepHeaderFooterLinks bool   `json:"keep_header_links"`
}

func (c *Config) touch() {
	var u *url.URL
	c.MaxPages = cmp.Or(c.MaxPages, 100)
	c.MaxDuration = cmp.Or(c.MaxDuration, 30)
	c.MaxDepth = cmp.Or(c.MaxDepth, 4)

	if strings.Contains(c.Domain, "://") {
		u, _ = url.Parse(c.Domain)
	} else {
		u, _ = url.Parse("https://" + c.Domain)
	}

	if nil != u {
		c.Domain = u.Host
	} else {
		c.Domain = ""
	}
}

func (c *Config) Validate() error {
	c.touch()

	if len(c.Domain) == 0 {
		return errors.New("missing or invalid domain")
	} else if res, e := http.Get(fmt.Sprintf("https://%s/", c.Domain)); e != nil {
		return e
	} else if res.StatusCode >= http.StatusBadRequest {
		return errors.New(fmt.Sprintf("http status code: %d", res.StatusCode))
	}

	return nil
}
