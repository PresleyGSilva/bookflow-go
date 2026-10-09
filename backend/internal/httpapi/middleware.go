package httpapi

import (
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"bookflow-go/internal/auth"
)

// JWT authenticates via Authorization: Bearer <access_token>.
// Falls back to X-User header (compat with frontend tenant headers).
func JWT(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var uid string
			if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
				if id, err := auth.Parse(secret, strings.TrimPrefix(h, "Bearer "), "access"); err == nil {
					uid = id
				} else {
					WriteErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
					return
				}
			} else if xu := r.Header.Get("X-User"); xu != "" {
				uid = xu
			} else {
				WriteErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing credentials")
				return
			}
			next.ServeHTTP(w, r.WithContext(WithUserID(r.Context(), uid)))
		})
	}
}

// RequireOrg resolves X-Org-ID and verifies the user is a member (403 otherwise).
func RequireOrg(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			org := r.Header.Get("X-Org-ID")
			if org == "" {
				WriteErr(w, http.StatusBadRequest, "ORG_REQUIRED", "missing X-Org-ID header")
				return
			}
			var ok bool
			err := pool.QueryRow(r.Context(),
				`SELECT EXISTS(SELECT 1 FROM org_members WHERE org_id=$1 AND user_id=$2)`, org, UserID(r)).Scan(&ok)
			if err != nil || !ok {
				WriteErr(w, http.StatusForbidden, "FORBIDDEN", "not a member of this organization")
				return
			}
			next.ServeHTTP(w, r.WithContext(WithOrgID(r.Context(), org)))
		})
	}
}
