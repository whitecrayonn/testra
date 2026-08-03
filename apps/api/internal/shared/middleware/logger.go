package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// RequestLogger returns a chi middleware that logs requests using the provided
// structured formatter.
func RequestLogger(f middleware.LogFormatter) func(http.Handler) http.Handler {
	return middleware.RequestLogger(f)
}

// NewStructuredLogFormatter returns a chi LogFormatter that writes each
// request as a single JSON line via slog. It is intended to be used with
// chi's RequestLogger middleware.
func NewStructuredLogFormatter(logger *slog.Logger) middleware.LogFormatter {
	return &structuredLogFormatter{logger: logger}
}

type structuredLogFormatter struct {
	logger *slog.Logger
}

func (f *structuredLogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	return &structuredLogEntry{
		logger:    f.logger,
		request:   r,
		requestID: middleware.GetReqID(r.Context()),
		start:     time.Now(),
	}
}

type structuredLogEntry struct {
	logger    *slog.Logger
	request   *http.Request
	requestID string
	start     time.Time
}

func (e *structuredLogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	attrs := []slog.Attr{
		slog.String("method", e.request.Method),
		slog.String("path", e.request.URL.EscapedPath()),
		slog.Int("status", status),
		slog.Int("bytes", bytes),
		slog.Int64("duration_ms", elapsed.Milliseconds()),
		slog.String("request_id", e.requestID),
	}

	if query := e.request.URL.RawQuery; query != "" {
		attrs = append(attrs, slog.String("query", RedactQuery(query)))
	}

	// Include the original client IP when chi's RealIP middleware is in use.
	// Mask the last octet to reduce PII exposure in logs.
	if realIP := e.request.Header.Get("X-Forwarded-For"); realIP != "" {
		attrs = append(attrs, slog.String("client_ip", MaskIP(realIP)))
	} else if realIP := e.request.Header.Get("X-Real-Ip"); realIP != "" {
		attrs = append(attrs, slog.String("client_ip", MaskIP(realIP)))
	}

	e.logger.LogAttrs(e.request.Context(), slog.LevelInfo, "request", attrs...)
}

func (e *structuredLogEntry) Panic(v interface{}, stack []byte) {
	e.logger.LogAttrs(e.request.Context(), slog.LevelError, "request panic",
		slog.Any("panic", v),
		slog.String("stack", string(stack)),
		slog.String("request_id", e.requestID),
	)
}
