package server

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func WithHTTPTracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := otelhttp.NewHandler(
			next,
			"HTTP_"+r.Method,
			otelhttp.WithTracerProvider(OtelTracerProvider()),
		)
		handler.ServeHTTP(w, r)
	})
}
