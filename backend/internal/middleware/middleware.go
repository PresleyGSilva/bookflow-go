package middleware

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"bookflow-go/internal/httpapi"
)

// Thin wrappers around the httpapi middlewares (single implementation there
// to avoid an import cycle: httpapi must not import this package).
func JWT(secret string) func(http.Handler) http.Handler {
	return httpapi.JWT(secret)
}

func RequireOrg(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return httpapi.RequireOrg(pool)
}

func UserID(r *http.Request) string { return httpapi.UserID(r) }

func OrgID(r *http.Request) string { return httpapi.OrgID(r) }
