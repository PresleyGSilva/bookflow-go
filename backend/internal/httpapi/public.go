package httpapi

import (
	"net/http"

	"github.com/jackc/pgx/v5"
)

// Public catalog + availability for the no-login booking page.
// Tenant scope: org_id query param for listings; professional's org for availability.

func (s *Server) handlePublicOrgs(w http.ResponseWriter, r *http.Request) {
	listRows(w, r, s.Pool, `SELECT id,name FROM orgs ORDER BY name`, nil,
		func(rows pgx.Rows) (map[string]any, error) {
			var id, name string
			if err := rows.Scan(&id, &name); err != nil {
				return nil, err
			}
			return map[string]any{"id": id, "name": name}, nil
		})
}

func (s *Server) handlePublicServices(w http.ResponseWriter, r *http.Request) {
	org := r.URL.Query().Get("org_id")
	if org == "" {
		WriteErr(w, http.StatusBadRequest, "BAD_REQUEST", "org_id is required")
		return
	}
	listRows(w, r, s.Pool,
		`SELECT id,name,duration_min,price_cents,active FROM services WHERE org_id=$1 AND active=TRUE ORDER BY name`,
		[]any{org},
		func(rows pgx.Rows) (map[string]any, error) {
			var id, name string
			var dur, price int
			var active bool
			if err := rows.Scan(&id, &name, &dur, &price, &active); err != nil {
				return nil, err
			}
			return map[string]any{"id": id, "name": name, "duration_min": dur, "price_cents": price, "active": active}, nil
		})
}

func (s *Server) handlePublicProfessionals(w http.ResponseWriter, r *http.Request) {
	org := r.URL.Query().Get("org_id")
	if org == "" {
		WriteErr(w, http.StatusBadRequest, "BAD_REQUEST", "org_id is required")
		return
	}
	listRows(w, r, s.Pool,
		`SELECT id,name,email,phone,timezone,active FROM professionals WHERE org_id=$1 AND active=TRUE ORDER BY name`,
		[]any{org},
		func(rows pgx.Rows) (map[string]any, error) {
			var id, name string
			var email, phone, tz *string
			var active bool
			if err := rows.Scan(&id, &name, &email, &phone, &tz, &active); err != nil {
				return nil, err
			}
			tzName := "America/Sao_Paulo"
			if tz != nil && *tz != "" {
				tzName = *tz
			}
			return map[string]any{"id": id, "name": name, "email": email, "phone": phone, "timezone": tzName, "active": active}, nil
		})
}

func (s *Server) handlePublicAvailability(w http.ResponseWriter, r *http.Request) {
	pid := r.URL.Query().Get("professional_id")
	sid := r.URL.Query().Get("service_id")
	if pid == "" || sid == "" {
		WriteErr(w, http.StatusBadRequest, "BAD_REQUEST", "professional_id and service_id are required")
		return
	}
	org, _, ok := s.orgOfProfessional(r.Context(), pid)
	if !ok {
		WriteErr(w, http.StatusNotFound, "NOT_FOUND", "professional not found")
		return
	}
	var sameOrg bool
	if err := s.Pool.QueryRow(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM services WHERE id=$1 AND org_id=$2)`, sid, org).Scan(&sameOrg); err != nil || !sameOrg {
		WriteErr(w, http.StatusNotFound, "NOT_FOUND", "service not found")
		return
	}
	s.serveAvailability(w, r, org)
}
