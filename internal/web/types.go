package web

import (
	"html/template"
	"time"

	"tubehive/internal/db"
	"tubehive/internal/feed"
)

type Server struct {
	Store     *db.Store
	Refresher *feed.Refresher
	Templates *template.Template
}

type PageData struct {
	Title string

	Channels       []db.RecentChannel
	ChannelResults []db.RecentChannel
	Videos         []db.Video
	ChannelFeeds   []db.ChannelFeed
	Settings       db.Settings

	Page    int
	PerPage int
	Total   int

	HasPrev bool
	HasNext bool

	PrevPage    int
	NextPage    int
	PrevPageURL string
	NextPageURL string

	Now     time.Time
	Message string

	LastUpdated    time.Time
	HasLastUpdated bool

	SelectedChannelID    string
	SelectedChannelTitle string
	VideoPath            string
	Search               string
	ActiveNav            string
}
