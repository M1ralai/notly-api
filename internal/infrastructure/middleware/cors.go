package middleware

import "net/http"

// CORSMiddleware handles CORS preflight and response headers at the backend level.
// This acts as a safety net: even if Nginx CORS handling is bypassed or misconfigured,
// OPTIONS preflight requests return 204 immediately (before gorilla/mux can 405 them),
// and all responses carry the necessary Access-Control-* headers.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-Connection-ID, x-connection-id")
		w.Header().Set("Vary", "Origin")

		// Short-circuit OPTIONS — don't let gorilla/mux 405 it
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent) // 204
			return
		}

		next.ServeHTTP(w, r)
	})
}
