package web

import (
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"yube/internal/db"
)

const rssItemLimit = 100

type rssDocument struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Atom    string     `xml:"xmlns:atom,attr,omitempty"`
	Media   string     `xml:"xmlns:media,attr,omitempty"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	AtomLink    atomLink  `xml:"atom:link"`
	Items       []rssItem `xml:"item"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type rssItem struct {
	Title       string         `xml:"title"`
	Link        string         `xml:"link"`
	GUID        guid           `xml:"guid"`
	PubDate     string         `xml:"pubDate"`
	Author      string         `xml:"author,omitempty"`
	Description cdata          `xml:"description"`
	Thumbnail   mediaThumbnail `xml:"media:thumbnail"`
}

type guid struct {
	IsPermaLink bool   `xml:"isPermaLink,attr"`
	Value       string `xml:",chardata"`
}

type mediaThumbnail struct {
	URL string `xml:"url,attr"`
}

type cdata struct {
	Value string `xml:",innerxml"`
}

func (s *Server) rssRoute(
	r chi.Router,
	pattern string,
) {
	r.Get(pattern, s.rss)
	r.Head(pattern, s.rss)
}

func (s *Server) rss(
	w http.ResponseWriter,
	r *http.Request,
) {
	videos, _, err := s.Store.ListVideos(
		r.Context(),
		1,
		rssItemLimit,
		"",
		"",
	)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	baseURL := requestBaseURL(r)
	feedURL := baseURL + "/rss.xml"

	doc := rssDocument{
		Version: "2.0",
		Atom:    "http://www.w3.org/2005/Atom",
		Media:   "http://search.yahoo.com/mrss/",
		Channel: rssChannel{
			Title:       "Yubè",
			Link:        baseURL + "/",
			Description: "Latest videos from your Yubè channels.",
			AtomLink: atomLink{
				Href: feedURL,
				Rel:  "self",
				Type: "application/rss+xml",
			},
			Items: rssItems(
				baseURL,
				videos,
			),
		},
	}

	w.Header().Set(
		"Content-Type",
		"application/rss+xml; charset=utf-8",
	)
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte(xml.Header))

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(doc); err != nil {
		// Headers are already sent, so the best we can do is stop writing.
		return
	}
}

func rssItems(
	baseURL string,
	videos []db.Video,
) []rssItem {
	items := make([]rssItem, 0, len(videos))

	for _, video := range videos {
		thumbnailURL := rssThumbnailURL(baseURL, video)

		items = append(items, rssItem{
			Title: video.Title,
			Link:  video.VideoURL,
			GUID: guid{
				IsPermaLink: false,
				Value:       video.VideoID,
			},
			PubDate: video.PublishedAt.Format(time.RFC1123Z),
			Author:  video.ChannelName,
			Description: rssDescription(
				thumbnailURL,
				video,
			),
			Thumbnail: mediaThumbnail{
				URL: thumbnailURL,
			},
		})
	}

	return items
}

func rssDescription(
	thumbnailURL string,
	video db.Video,
) cdata {
	descriptionHTML := fmt.Sprintf(
		`<p><img src="%s" alt=""></p><p><strong>%s</strong></p><p>%s</p>`,
		html.EscapeString(thumbnailURL),
		html.EscapeString(video.ChannelName),
		html.EscapeString(video.Title),
	)

	return cdata{
		Value: "<![CDATA[" + strings.ReplaceAll(
			descriptionHTML,
			"]]>",
			"]]]]><![CDATA[>",
		) + "]]>",
	}
}

func rssThumbnailURL(
	baseURL string,
	video db.Video,
) string {
	return fmt.Sprintf(
		"%s/videos/%s/thumbnail",
		baseURL,
		url.PathEscape(video.VideoID),
	)
}

func requestBaseURL(r *http.Request) string {
	scheme := strings.TrimSpace(
		r.Header.Get("X-Forwarded-Proto"),
	)
	if scheme == "" {
		scheme = "http"
		if r.TLS != nil {
			scheme = "https"
		}
	}

	host := strings.TrimSpace(
		r.Header.Get("X-Forwarded-Host"),
	)
	if host == "" {
		host = r.Host
	}

	return scheme + "://" + host
}
