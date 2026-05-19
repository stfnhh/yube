package web

import (
	"context"
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
		r.FormValue("feed_url"),
	)

	channel, err := s.Refresher.ResolveChannelInput(
		r.Context(),
		rawURL,
	)
	if err != nil {
		http.Error(
			w,
			"Expected a channel URL or subscription feed URL",
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
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)

		return
	}

	for _, item := range feeds {
		if !strings.Contains(
			item.URL,
			feed.SubscriptionFeedPath(),
		) {
			continue
		}

		channelID, err := feed.ExtractChannelID(
			item.URL,
		)
		if err != nil {
			continue
		}

		title := item.Title

		if title == "" {
			title = channelID
		}

		_ = s.Store.UpsertFeed(
			r.Context(),
			title,
			channelID,
			item.URL,
			"",
		)
	}

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
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

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
