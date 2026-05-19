package web

import "net/http"

func (s *Server) notFound(
	w http.ResponseWriter,
	r *http.Request,
) {
	s.renderPageWithStatus(
		w,
		r,
		"not-found.html",
		PageData{
			Title:     "Page not found · Yubè",
			ActiveNav: "none",
		},
		http.StatusNotFound,
	)
}
