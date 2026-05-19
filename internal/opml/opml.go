package opml

import (
	"encoding/xml"
	"io"
)

type document struct {
	Outlines []outline `xml:"body>outline"`
}

type outline struct {
	Text     string    `xml:"text,attr"`
	Title    string    `xml:"title,attr"`
	XMLURL   string    `xml:"xmlUrl,attr"`
	Outlines []outline `xml:"outline"`
}

type Feed struct {
	Title string
	URL   string
}

func Parse(r io.Reader) ([]Feed, error) {
	var doc document
	if err := xml.NewDecoder(r).Decode(&doc); err != nil { return nil, err }
	var feeds []Feed
	var walk func([]outline)
	walk = func(items []outline) {
		for _, item := range items {
			if item.XMLURL != "" {
				title := item.Title
				if title == "" { title = item.Text }
				feeds = append(feeds, Feed{Title: title, URL: item.XMLURL})
			}
			walk(item.Outlines)
		}
	}
	walk(doc.Outlines)
	return feeds, nil
}
