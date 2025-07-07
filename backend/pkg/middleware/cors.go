package middleware

import (
	"net/http"
)

// CORSMiddleware adds CORS headers to responses
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Allow requests from frontend and desktop-messenger
		allowedOrigins := []string{
			"http://localhost:3000", // Frontend URL
			"file://",               // Electron app
			"http://localhost",      // Local development
			"https://localhost",     // HTTPS local development
		}

		// Check if the origin is allowed or if it's a file:// protocol request
		isAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin || (allowedOrigin == "file://" && (origin == "" || origin == "null")) {
				isAllowed = true
				break
			}
		}

		// For Electron apps, origin might be null or empty, so allow those too
		if origin == "" || origin == "null" || isAllowed {
			w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all for desktop app
		} else {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
