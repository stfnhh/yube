package opml

import (
	"encoding/xml"
	"io"
	"time"
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

type exportDocument struct {
	XMLName xml.Name   `xml:"opml"`
	Version string     `xml:"version,attr"`
	Head    exportHead `xml:"head"`
	Body    exportBody `xml:"body"`
}

type exportHead struct {
	Title       string `xml:"title"`
	DateCreated string `xml:"dateCreated"`
}

type exportBody struct {
	Outlines []exportOutline `xml:"outline"`
}

type exportOutline struct {
	Text   string `xml:"text,attr"`
	Title  string `xml:"title,attr"`
	Type   string `xml:"type,attr"`
	XMLURL string `xml:"xmlUrl,attr"`
}

type Feed struct {
	Title string
	URL   string
}

func Parse(r io.Reader) ([]Feed, error) {
	var doc document
	if err := xml.NewDecoder(r).Decode(&doc); err != nil {
		return nil, err
	}

	var feeds []Feed

	var walk func([]outline)
	walk = func(items []outline) {
		for _, item := range items {
			if item.XMLURL != "" {
				title := item.Title
				if title == "" {
					title = item.Text
				}

				feeds = append(feeds, Feed{Title: title, URL: item.XMLURL})
			}

			walk(item.Outlines)
		}
	}

	walk(doc.Outlines)

	return feeds, nil
}

func Write(w io.Writer, title string, feeds []Feed) error {
	doc := exportDocument{
		Version: "2.0",
		Head: exportHead{
			Title:       title,
			DateCreated: time.Now().UTC().Format(time.RFC1123Z),
		},
		Body: exportBody{
			Outlines: make([]exportOutline, 0, len(feeds)),
		},
	}

	for _, feed := range feeds {
		doc.Body.Outlines = append(doc.Body.Outlines, exportOutline{
			Text:   feed.Title,
			Title:  feed.Title,
			Type:   "rss",
			XMLURL: feed.URL,
		})
	}

	if _, err := io.WriteString(w, xml.Header); err != nil {
		return err
	}

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")

	return encoder.Encode(doc)
}
