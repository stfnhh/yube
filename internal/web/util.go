package web

import (
	"net/url"
	"strconv"
	"time"
)

func parseInt(
	s string,
	fallback int,
) int {
	v, err := strconv.Atoi(s)

	if err != nil || v < 1 {
		return fallback
	}

	return v
}

func pageURL(
	path string,
	page int,
	search string,
) string {
	values := url.Values{}

	if search != "" {
		values.Set(
			"q",
			search,
		)
	}

	values.Set(
		"page",
		strconv.Itoa(page),
	)

	return path + "?" + values.Encode()
}

func humanAgo(
	d time.Duration,
) string {
	if d < time.Minute {
		return "just now"
	}

	if d < time.Hour {
		return strconv.Itoa(
			int(d.Minutes()),
		) + "m ago"
	}

	if d < 24*time.Hour {
		return strconv.Itoa(
			int(d.Hours()),
		) + "h ago"
	}

	if d < 30*24*time.Hour {
		return strconv.Itoa(
			int(d.Hours()/24),
		) + "d ago"
	}

	return strconv.Itoa(
		int(d.Hours()/24/30),
	) + "mo ago"
}
