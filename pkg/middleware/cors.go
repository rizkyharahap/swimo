package middleware

import (
	"net/http"

	"github.com/rizkyharahap/swimo/config"
)

// CORS creates middleware that handles CORS headers
func CORS(cfg config.CORSConfig) Middleware {
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
