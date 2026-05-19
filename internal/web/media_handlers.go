package web

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (s *Server) channelIcon(
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

	iconURL, err := s.Store.ChannelIconURL(
		r.Context(),
		channelID,
	)
	if err != nil || iconURL == "" {
		http.NotFound(w, r)
		return
	}

	parsed, err := url.Parse(iconURL)
	if err != nil ||
		parsed.Scheme != "https" ||
		!strings.HasSuffix(parsed.Hostname(), "googleusercontent.com") {
		http.NotFound(w, r)
		return
	}

	req, err := http.NewRequestWithContext(
		r.Context(),
		http.MethodGet,
		iconURL,
		nil,
	)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	client := http.DefaultClient
	if s.Refresher != nil &&
		s.Refresher.HTTPClient != nil {
		client = s.Refresher.HTTPClient
	}

	resp, err := client.Do(req)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 ||
		resp.StatusCode >= 300 {
		http.NotFound(w, r)
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)

	_, _ = io.Copy(w, resp.Body)
}

func (s *Server) videoThumbnail(
	w http.ResponseWriter,
	r *http.Request,
) {
	videoID := strings.TrimSpace(
		chi.URLParam(r, "videoID"),
	)
	if videoID == "" {
		http.NotFound(w, r)
		return
	}

	thumbnailURL, err := s.Store.VideoThumbnailURL(
		r.Context(),
		videoID,
	)
	if err != nil || thumbnailURL == "" {
		http.NotFound(w, r)
		return
	}

	parsed, err := url.Parse(thumbnailURL)
	if err != nil ||
		parsed.Scheme != "https" ||
		!strings.HasSuffix(parsed.Hostname(), thumbnailHost()) {
		http.NotFound(w, r)
		return
	}

	client := http.DefaultClient
	if s.Refresher != nil &&
		s.Refresher.HTTPClient != nil {
		client = s.Refresher.HTTPClient
	}

	var resp *http.Response

	for _, candidate := range thumbnailCandidates(
		videoID,
		thumbnailURL,
	) {
		req, err := http.NewRequestWithContext(
			r.Context(),
			http.MethodGet,
			candidate,
			nil,
		)
		if err != nil {
			continue
		}

		resp, err = client.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode >= 200 &&
			resp.StatusCode < 300 {
			break
		}

		resp.Body.Close()
		resp = nil
	}

	if resp == nil {
		http.NotFound(w, r)
		return
	}

	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)

	_, _ = io.Copy(w, resp.Body)
}

func thumbnailCandidates(
	videoID,
	storedURL string,
) []string {
	base := "https://" + thumbnailHost() + "/vi/" +
		url.PathEscape(videoID) +
		"/"

	candidates := []string{
		base + "hq720.jpg",
		base + "maxresdefault.jpg",
		base + "sddefault.jpg",
	}

	if storedURL != "" {
		candidates = append(
			candidates,
			storedURL,
		)
	}

	return candidates
}

func thumbnailHost() string {
	return "yt" + "img.com"
}

func (s *Server) watchVideo(
	w http.ResponseWriter,
	r *http.Request,
) {
	videoID := strings.TrimSpace(
		chi.URLParam(r, "videoID"),
	)
	if videoID == "" {
		http.NotFound(w, r)
		return
	}

	videoURL, err := s.Store.MarkVideoWatchedByVideoID(
		r.Context(),
		videoID,
	)
	if err != nil || videoURL == "" {
		http.NotFound(w, r)
		return
	}

	http.Redirect(
		w,
		r,
		videoURL,
		http.StatusSeeOther,
	)
}
