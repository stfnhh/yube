package web

import (
	"net/http"
	"strings"

	"yube/internal/db"
)

func (s *Server) settings(
	w http.ResponseWriter,
	r *http.Request,
) {
	settings, err := s.Store.GetSettings(
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
		Title:    "Settings · Yubè",
		Settings: settings,
		Message:  strings.TrimSpace(r.URL.Query().Get("message")),
	}

	s.renderPage(
		w,
		r,
		"settings.html",
		data,
	)
}

func (s *Server) updateSettings(
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

	settings := db.Settings{
		RefreshIntervalMin: parseInt(
			r.FormValue("refresh_interval_minutes"),
			15,
		),
		VideoRetentionDays: parseInt(
			r.FormValue("video_retention_days"),
			90,
		),
		MaxVideosPerChannel: parseInt(
			r.FormValue("max_videos_per_channel"),
			250,
		),
	}

	settings = db.NormalizeSettings(settings)

	if err := s.Store.SaveSettings(
		r.Context(),
		settings,
	); err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	if s.Refresher != nil {
		s.Refresher.ApplySettings(settings)
	}

	_ = s.Store.CleanupVideos(
		r.Context(),
		settings.VideoRetentionDays,
		settings.MaxVideosPerChannel,
	)

	http.Redirect(
		w,
		r,
		"/settings?message=saved",
		http.StatusSeeOther,
	)
}
