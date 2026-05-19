package web

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"

	"tubehive/internal/db"
)

func (s *Server) index(
	w http.ResponseWriter,
	r *http.Request,
) {
	page := parseInt(
		r.URL.Query().Get("page"),
		1,
	)

	selectedChannelID := strings.TrimSpace(
		chi.URLParam(r, "channelID"),
	)

	videoPath := "/"
	if selectedChannelID != "" {
		videoPath = "/channels/" +
			url.PathEscape(selectedChannelID)
	}

	selectedChannelTitle := ""
	if selectedChannelID != "" {
		selectedChannel, err := s.Store.FeedByChannelID(
			r.Context(),
			selectedChannelID,
		)
		if err != nil {
			http.Error(
				w,
				err.Error(),
				http.StatusInternalServerError,
			)

			return
		}

		selectedChannelTitle = selectedChannel.Title
	}

	perPage := 30

	videos, total, err := s.Store.ListVideos(
		r.Context(),
		page,
		perPage,
		selectedChannelID,
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

	s.renderIndex(
		w,
		r,
		PageData{
			Title: "TubeDeck",

			Videos: videos,

			Page:    page,
			PerPage: perPage,
			Total:   total,

			HasPrev: page > 1,
			HasNext: page*perPage < total,

			PrevPage: page - 1,
			NextPage: page + 1,
			PrevPageURL: pageURL(
				videoPath,
				page-1,
				"",
			),
			NextPageURL: pageURL(
				videoPath,
				page+1,
				"",
			),

			SelectedChannelID:    selectedChannelID,
			SelectedChannelTitle: selectedChannelTitle,
			VideoPath:            videoPath,
		},
	)
}

func (s *Server) search(
	w http.ResponseWriter,
	r *http.Request,
) {
	page := parseInt(
		r.URL.Query().Get("page"),
		1,
	)

	search := strings.TrimSpace(
		r.URL.Query().Get("q"),
	)

	perPage := 30

	videos, total, err := s.Store.ListVideos(
		r.Context(),
		page,
		perPage,
		"",
		search,
	)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	var channelResults []db.RecentChannel

	if search != "" && page == 1 {
		channelResults, err = s.Store.SearchChannels(
			r.Context(),
			search,
			8,
		)
		if err != nil {
			http.Error(
				w,
				err.Error(),
				http.StatusInternalServerError,
			)

			return
		}
	}

	s.renderPage(
		w,
		r,
		"search.html",
		PageData{
			Title: "Search · TubeHive",

			Videos:         videos,
			ChannelResults: channelResults,

			Page:    page,
			PerPage: perPage,
			Total:   total,

			HasPrev: page > 1,
			HasNext: page*perPage < total,

			PrevPage: page - 1,
			NextPage: page + 1,
			PrevPageURL: pageURL(
				"/search",
				page-1,
				search,
			),
			NextPageURL: pageURL(
				"/search",
				page+1,
				search,
			),

			VideoPath: "/search",
			Search:    search,
		},
	)
}

func (s *Server) renderIndex(
	w http.ResponseWriter,
	r *http.Request,
	data PageData,
) {
	s.renderPage(
		w,
		r,
		"index.html",
		data,
	)
}
