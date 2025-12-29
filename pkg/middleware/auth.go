package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/rizkyharahap/swimo/pkg/response"
	"github.com/rizkyharahap/swimo/pkg/security"
)

type ctxKey string

const UserClaimKey ctxKey = "userClaim"

func Auth(secret string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorization := r.Header.Get("Authorization")

			if authorization == "" {
				response.JSON(w, http.StatusUnauthorized, response.Message{Message: "Missing Authorization header"})
				return
			}

			parts := strings.SplitN(authorization, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				response.JSON(w, http.StatusUnauthorized, response.Message{Message: "Invalid Authorization format"})
				return
			}

			token := parts[1]
			claims, err := security.VerifyJWT(token, secret)
			if err != nil {
				response.JSON(w, http.StatusUnauthorized, response.Message{Message: "Invalid or expired token"})
				return
			}

			ctx := context.WithValue(r.Context(), UserClaimKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthFromContext extracts JWT claims from context
func AuthFromContext(ctx context.Context) *security.Claim {
	val := ctx.Value(UserClaimKey)
	if claim, ok := val.(*security.Claim); ok {
		return claim
	}
	return nil
}
