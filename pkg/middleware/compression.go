package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

// CompressionMiddleware creates middleware that compresses HTTP responses
func CompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts compression
		acceptEncoding := r.Header.Get("Accept-Encoding")
		if !strings.Contains(acceptEncoding, "gzip") {
			// Client doesn't accept compression, proceed normally
			next.ServeHTTP(w, r)
			return
		}

		// Create gzip response writer
		gzWriter := gzip.NewWriter(w)
		defer gzWriter.Close()

		// Wrap response writer
		compressedWriter := &gzipResponseWriter{
			ResponseWriter: w,
			gzipWriter:     gzWriter,
		}

		// Set content encoding header
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length") // Content length will change after compression

		// Call next handler
		next.ServeHTTP(compressedWriter, r)
	})
}

// gzipResponseWriter wraps http.ResponseWriter to handle gzip compression
type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter *gzip.Writer
}

func (gz *gzipResponseWriter) Write(data []byte) (int, error) {
	return gz.gzipWriter.Write(data)
}

func (gz *gzipResponseWriter) WriteHeader(statusCode int) {
	gz.ResponseWriter.WriteHeader(statusCode)
}
