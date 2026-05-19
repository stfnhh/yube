package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()

	r.Handle(
		"/static/*",
		http.StripPrefix(
			"/static/",
			http.FileServer(
				http.Dir("static"),
			),
		),
	)

	r.Get("/", s.index)

	r.Get("/search", s.search)

	r.Get("/settings", s.settings)
	r.Post("/settings", s.updateSettings)

	r.Route("/channels", func(r chi.Router) {
		r.Get("/", s.channels)

		r.Post("/", s.addChannel)

		r.Post(
			"/import",
			s.importOPML,
		)

		r.Delete("/{channelID}", s.unsubscribeChannel)

		r.Get("/{channelID}", s.index)

		r.Get("/{channelID}/icon", s.feedIcon)
	})

	r.Route("/videos", func(r chi.Router) {
		r.Get("/{videoID}/thumbnail", s.videoThumbnail)
	})

	r.Get("/watch/{videoID}", s.watchVideo)

	r.Post("/refresh", s.refresh)

	return r
}
