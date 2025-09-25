package middleware

import (
	"net/http"

	"github.com/rizkyharahap/swimo/config"
)

// CORSMiddleware creates middleware that handles CORS headers
func CORSMiddleware(cfg config.CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			if cfg.AllowOrigins != "" {
				w.Header().Set("Access-Control-Allow-Origin", cfg.AllowOrigins)
			}
			if cfg.AllowMethods != "" {
				w.Header().Set("Access-Control-Allow-Methods", cfg.AllowMethods)
			}
			if cfg.AllowHeaders != "" {
				w.Header().Set("Access-Control-Allow-Headers", cfg.AllowHeaders)
			}
			if cfg.ExposeHeaders != "" {
				w.Header().Set("Access-Control-Expose-Headers", cfg.ExposeHeaders)
			}
			if cfg.Credentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() config.CORSConfig {
	return config.CORSConfig{
		AllowOrigins:  "*",
		AllowMethods:  "GET, POST, PUT, DELETE, OPTIONS",
		AllowHeaders:  "Content-Type, Authorization",
		ExposeHeaders: "",
		Credentials:   false,
	}
}
