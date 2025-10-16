package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/rizkyharahap/swimo/pkg/response"
	"github.com/rizkyharahap/swimo/pkg/security"
)

type ctxKey string

const userClaimKey ctxKey = "userClaim"

func AuthMiddleware(secret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(response.Error{Message: "Missing Authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(response.Error{Message: "Invalid Authorization format"})
			return
		}

		token := parts[1]
		claims, err := security.VerifyJWT(token, secret)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(response.Error{Message: "Invalid or expired token"})
			return
		}

		ctx := context.WithValue(r.Context(), userClaimKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthFromContext extracts JWT claims from context
func AuthFromContext(ctx context.Context) *security.Claim {
	val := ctx.Value(userClaimKey)
	if claim, ok := val.(*security.Claim); ok {
		return claim
	}
	return nil
}
