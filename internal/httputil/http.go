package httputil

import (
	"log/slog"
	"net"
	"net/http"
	"time"
)

var _ http.RoundTripper = (*LoggerTransport)(nil)

// LoggerTransport logs outgoing requests for debugging purposes.
type LoggerTransport struct {
	*http.Transport
}

func (t *LoggerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	slog.Debug("Performing request", slog.String("method", r.Method), slog.String("url", r.URL.String()))
	return t.Transport.RoundTrip(r)
}

// DefaultClient is a HTTP client with sane defaults.
var DefaultClient = &http.Client{
	Transport: &LoggerTransport{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
		},
	},
}
