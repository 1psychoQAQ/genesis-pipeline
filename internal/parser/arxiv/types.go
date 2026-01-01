package arxiv

import "time"

// Atom feed XML structures for ArXiv API responses.

type atomFeed struct {
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	ID         string         `xml:"id"`
	Title      string         `xml:"title"`
	Summary    string         `xml:"summary"`
	Published  time.Time      `xml:"published"`
	Updated    time.Time      `xml:"updated"`
	Authors    []atomAuthor   `xml:"author"`
	Categories []atomCategory `xml:"category"`
	Links      []atomLink     `xml:"link"`
}

type atomAuthor struct {
	Name string `xml:"name"`
}

type atomCategory struct {
	Term string `xml:"term,attr"`
}

type atomLink struct {
	Href  string `xml:"href,attr"`
	Rel   string `xml:"rel,attr"`
	Type  string `xml:"type,attr"`
	Title string `xml:"title,attr"`
}
