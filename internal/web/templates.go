package web

import (
	"html/template"
	"strings"
	"time"

	"tubehive/internal/db"
	"tubehive/internal/feed"
)

func New(
	store *db.Store,
	refresher *feed.Refresher,
) *Server {
	funcs := template.FuncMap{
		"ago": func(t time.Time) string {
			return humanAgo(
				time.Since(t),
			)
		},
		"initial": func(s string) string {
			s = strings.TrimSpace(s)
			if s == "" {
				return "?"
			}

			return strings.ToUpper(
				string([]rune(s)[0]),
			)
		},
	}

	t := template.Must(
		template.
			New("").
			Funcs(funcs).
			ParseGlob("templates/*.html"),
	)

	return &Server{
		Store:     store,
		Refresher: refresher,
		Templates: t,
	}
}
