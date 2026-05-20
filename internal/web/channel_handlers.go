package web

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"yube/internal/feed"
	"yube/internal/opml"
)

func (s *Server) channels(
	w http.ResponseWriter,
	r *http.Request,
) {
	feeds, err := s.Store.ListChannelFeeds(
		r.Context(),
	)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	data := PageData{
		Title: "Channels · Yubè",

		ChannelFeeds: feeds,
	}

	s.renderPage(
		w,
		r,
		"channels.html",
		data,
	)
}

func (s *Server) addChannel(
	w http.ResponseWriter,
	r *http.Request,
) {
	if err := r.ParseForm(); err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)

		return
	}

	rawURL := strings.TrimSpace(
		r.FormValue("channel_source"),
	)

	channel, err := s.Refresher.ResolveChannelInput(
		r.Context(),
		rawURL,
	)
	if err != nil {
		log.Printf("add channel failed source=%q remote=%s error=%v", rawURL, clientIP(r), err)
		http.Error(
			w,
			"Expected a channel URL or Atom source URL",
			http.StatusBadRequest,
		)

		return
	}

	if err := s.Store.UpsertFeed(
		r.Context(),
		channel.Title,
		channel.ChannelID,
		channel.FeedURL,
		channel.IconURL,
	); err != nil {
		log.Printf("add channel store failed channel_id=%s title=%q remote=%s error=%v", channel.ChannelID, channel.Title, clientIP(r), err)
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	log.Printf("channel added channel_id=%s title=%q remote=%s", channel.ChannelID, channel.Title, clientIP(r))

	go s.Refresher.RefreshAll(
		context.Background(),
	)

	http.Redirect(
		w,
		r,
		"/channels",
		http.StatusSeeOther,
	)
}

func (s *Server) importOPML(
	w http.ResponseWriter,
	r *http.Request,
) {
	if err := r.ParseMultipartForm(
		8 << 20,
	); err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)

		return
	}

	file, _, err := r.FormFile("opml")
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)

		return
	}

	defer file.Close()

	feeds, err := opml.Parse(file)
	if err != nil {
		log.Printf("import subscriptions parse failed remote=%s error=%v", clientIP(r), err)
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)

		return
	}

	imported := 0
	skipped := 0

	for _, item := range feeds {
		if !strings.Contains(
			item.URL,
			feed.SubscriptionFeedPath(),
		) {
			skipped++
			continue
		}

		channelID, err := feed.ExtractChannelID(
			item.URL,
		)
		if err != nil {
			skipped++
			continue
		}

		title := item.Title

		if title == "" {
			title = channelID
		}

		if err := s.Store.UpsertFeed(
			r.Context(),
			title,
			channelID,
			item.URL,
			"",
		); err != nil {
			log.Printf("import subscription store failed channel_id=%s title=%q remote=%s error=%v", channelID, title, clientIP(r), err)
			skipped++
			continue
		}

		imported++
	}

	log.Printf("subscriptions imported imported=%d skipped=%d remote=%s", imported, skipped, clientIP(r))

	go s.Refresher.RefreshAll(
		context.Background(),
	)

	http.Redirect(
		w,
		r,
		"/channels",
		http.StatusSeeOther,
	)
}

func (s *Server) unsubscribeChannel(
	w http.ResponseWriter,
	r *http.Request,
) {
	channelID := strings.TrimSpace(
		chi.URLParam(r, "channelID"),
	)
	if channelID == "" {
		http.NotFound(w, r)
		return
	}

	if err := s.Store.DeleteFeedByChannelID(
		r.Context(),
		channelID,
	); err != nil {
		log.Printf("unsubscribe channel failed channel_id=%s remote=%s error=%v", channelID, clientIP(r), err)
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	log.Printf("channel unsubscribed channel_id=%s remote=%s", channelID, clientIP(r))

	http.Redirect(
		w,
		r,
		"/channels",
		http.StatusSeeOther,
	)
}

func (s *Server) refresh(
	w http.ResponseWriter,
	r *http.Request,
) {
	log.Printf("manual refresh requested remote=%s", clientIP(r))

	if err := s.Store.TouchRefreshTime(
		r.Context(),
	); err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	go s.Refresher.RefreshAll(
		context.Background(),
	)

	http.Redirect(
		w,
		r,
		"/",
		http.StatusSeeOther,
	)
}
