package web

import (
	"log"
	"net/http"
	"strings"
	"time"
)

type responseLogger struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *responseLogger) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseLogger) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}

	written, err := w.ResponseWriter.Write(data)
	w.bytes += written
	return written, err
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if skipRequestLog(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		started := time.Now()
		logger := &responseLogger{ResponseWriter: w}

		next.ServeHTTP(logger, r)

		status := logger.status
		if status == 0 {
			status = http.StatusOK
		}

		log.Printf(
			"http request method=%s path=%s status=%d bytes=%d duration=%s remote=%s",
			r.Method,
			r.URL.RequestURI(),
			status,
			logger.bytes,
			time.Since(started).Round(time.Millisecond),
			clientIP(r),
		)
	})
}

func skipRequestLog(path string) bool {
	if strings.HasPrefix(path, "/static/") ||
		strings.HasPrefix(path, "/videos/") {
		return true
	}

	return strings.HasPrefix(path, "/channels/") &&
		strings.HasSuffix(path, "/icon")
}

func clientIP(r *http.Request) string {
	if forwardedFor := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwardedFor != "" {
		if comma := strings.IndexByte(forwardedFor, ','); comma >= 0 {
			return strings.TrimSpace(forwardedFor[:comma])
		}

		return forwardedFor
	}

	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}

	return r.RemoteAddr
}
