package crawler

type PageSummary struct {
	Description string `json:"description"`
	Body        string `json:"body"`
	Title       string `json:"title"`
	Keywords    string `json:"keywords"`
	Url         string `json:"url"`
}

type pageAttribute uint8

const (
	attrTitle pageAttribute = iota
	attrDesc
	attrKeywords
	attrBody
)
