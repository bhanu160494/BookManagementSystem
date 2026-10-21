package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type contextKey string

const claimsContextKey contextKey = "auth.claims"

// Middleware authenticates requests and enforces role based authorization.
type Middleware struct {
	verifier *TokenVerifier
}

// NewMiddleware builds middleware backed by the given verifier.
func NewMiddleware(verifier *TokenVerifier) *Middleware {
	return &Middleware{verifier: verifier}
}

// Authenticate rejects requests without a valid Keycloak access token.
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawToken, ok := bearerToken(r)
		if !ok {
			writeAuthError(w, http.StatusUnauthorized, "missing or malformed Authorization header")
			return
		}

		claims, err := m.verifier.Verify(r.Context(), rawToken)
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, "invalid or expired access token")
			return
		}

		ctx := context.WithValue(r.Context(), claimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole allows the request only when the token carries one of the roles.
// It must be chained after Authenticate.
func (m *Middleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				writeAuthError(w, http.StatusUnauthorized, "missing authentication context")
				return
			}
			for _, required := range roles {
				for _, granted := range claims.Roles {
					if granted == required {
						next.ServeHTTP(w, r)
						return
					}
				}
			}
			writeAuthError(w, http.StatusForbidden, "account lacks the required role for this operation")
		})
	}
}

// ClaimsFromContext returns the authenticated user's claims, if any.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*Claims)
	return claims, ok
}

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", false
	}
	const prefix = "bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	return token, token != ""
}

func writeAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	if status == http.StatusUnauthorized {
		w.Header().Set("WWW-Authenticate", `Bearer realm="bookstore"`)
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
