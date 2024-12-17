package crawler

import "testing"

func TestConfig_Validate(t *testing.T) {
	c := Config{
		Domain: "https://www.google.com/",
	}

	if e := c.Validate(); e != nil {
		t.Error(e)
	} else if c.Domain != "www.google.com" {
		t.Errorf("unexpected domain: %s", c.Domain)
	}

	c.Domain = "www.google.com"

	if e := c.Validate(); e != nil {
		t.Error(e)
	} else if c.Domain != "www.google.com" {
		t.Errorf("unexpected domain: %s", c.Domain)
	}

	c.Domain = "www.google.com/sitemap.xml"

	if e := c.Validate(); e != nil {
		t.Error(e)
	} else if c.Domain != "www.google.com" {
		t.Errorf("unexpected domain: %s", c.Domain)
	}

	c.Domain = "lorem ipsum blabla"

	if e := c.Validate(); e == nil {
		t.Error("invalid domain should have returned error")
	} else if len(c.Domain) > 0 {
		t.Errorf("domain should be empty, got: %s", c.Domain)
	}
}
