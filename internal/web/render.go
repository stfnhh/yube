package web

import (
	"bytes"
	"net/http"
	"strings"
	"time"
)

func (s *Server) render(
	w http.ResponseWriter,
	name string,
	data PageData,
) {
	s.renderWithStatus(
		w,
		name,
		data,
		http.StatusOK,
	)
}

func (s *Server) renderWithStatus(
	w http.ResponseWriter,
	name string,
	data PageData,
	status int,
) {
	var buf bytes.Buffer

	err := s.Templates.ExecuteTemplate(
		&buf,
		name,
		data,
	)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set(
		"Content-Type",
		"text/html; charset=utf-8",
	)

	w.WriteHeader(status)

	_, _ = buf.WriteTo(w)
}

func (s *Server) renderPage(
	w http.ResponseWriter,
	r *http.Request,
	name string,
	data PageData,
) {
	data, err := s.sharedPageData(
		r,
		data,
	)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	s.render(
		w,
		name,
		data,
	)
}

func (s *Server) renderPageWithStatus(
	w http.ResponseWriter,
	r *http.Request,
	name string,
	data PageData,
	status int,
) {
	data, err := s.sharedPageData(
		r,
		data,
	)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	s.renderWithStatus(
		w,
		name,
		data,
		status,
	)
}

func (s *Server) sharedPageData(
	r *http.Request,
	data PageData,
) (PageData, error) {
	if data.Channels == nil {
		channels, err := s.Store.RecentChannels(
			r.Context(),
			7,
		)
		if err != nil {
			return PageData{}, err
		}

		data.Channels = channels
	}

	if data.Now.IsZero() {
		data.Now = time.Now()
	}

	if !data.HasLastUpdated {
		lastUpdated, ok, err := s.Store.LastRefreshTime(
			r.Context(),
		)
		if err != nil {
			return PageData{}, err
		}

		data.LastUpdated = lastUpdated
		data.HasLastUpdated = ok
	}

	if data.ActiveNav == "" {
		data.ActiveNav = activeNav(
			r.URL.Path,
		)
	}

	return data, nil
}

func activeNav(path string) string {
	switch {
	case path == "/channels" ||
		strings.HasPrefix(path, "/channels/"):
		return "channels"

	case path == "/settings" ||
		strings.HasPrefix(path, "/settings/"):
		return "settings"

	default:
		return "watch"
	}
}
