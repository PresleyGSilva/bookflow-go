package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)
// generic CRUD helpers

func listRows[T any](w http.ResponseWriter, r *http.Request, pool *pgxpool.Pool, query string, args []any, scan func(pgx.Rows) (T, error)) {
	rows, err := pool.Query(r.Context(), query, args...)
	if err != nil {
		WriteErr(w, http.StatusInternalServerError, "INTERNAL", "db error")
		return
	}
	defer rows.Close()
	out := []T{}
	for rows.Next() {
		v, err := scan(rows)
		if err != nil {
			WriteErr(w, http.StatusInternalServerError, "INTERNAL", "db error")
			return
		}
		out = append(out, v)
	}
	WriteData(w, http.StatusOK, out, nil)
}
func (s *Server) ownsProfessional(r *http.Request, org, pid string) bool {
	var ok bool
	err := s.Pool.QueryRow(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM professionals WHERE id=$1 AND org_id=$2)`, pid, org).Scan(&ok)
	return err == nil && ok
}


func (s *Server) handleProfessionals(w http.ResponseWriter, r *http.Request) {
	org := OrgID(r)
	switch r.Method {
	case http.MethodGet:
		listRows(w, r, s.Pool, `SELECT id,name,email,phone,timezone,active FROM professionals WHERE org_id=$1 ORDER BY name`, []any{org},
			func(rows pgx.Rows) (any, error) {
				var id, name, tz string
				var email, phone *string
				var active bool
				err := rows.Scan(&id, &name, &email, &phone, &tz, &active)
				return map[string]any{"id": id, "name": name, "email": email, "phone": phone, "timezone": tz, "active": active}, err
			})
	case http.MethodPost:
		var in struct {
			Name     string  `json:"name"`
			Email    *string `json:"email"`
			Phone    *string `json:"phone"`
			Timezone string  `json:"timezone"`
		}
		if !Decode(w, r, &in) {
			return
		}
		if in.Name == "" {
			WriteErr(w, http.StatusBadRequest, "BAD_REQUEST", "name is required")
			return
		}
		if in.Timezone == "" {
			in.Timezone = "America/Sao_Paulo"
		}
		var id string
		err := s.Pool.QueryRow(r.Context(),
			`INSERT INTO professionals(org_id,name,email,phone,timezone) VALUES($1,$2,$3,$4,$5) RETURNING id`,
			org, in.Name, in.Email, in.Phone, in.Timezone).Scan(&id)
		if err != nil {
			WriteErr(w, http.StatusInternalServerError, "INTERNAL", "db error")
			return
		}
		WriteData(w, http.StatusCreated, map[string]any{"id": id, "name": in.Name, "email": in.Email, "phone": in.Phone, "timezone": in.Timezone, "active": true}, nil)
	}
}

func (s *Server) handleProfessional(w http.ResponseWriter, r *http.Request) {
	org, id := OrgID(r), chi.URLParam(r, "id")
	switch r.Method {
	case http.MethodGet:
		var name, tz string
		var email, phone *string
		var active bool
		var pid string
		err := s.Pool.QueryRow(r.Context(),
			`SELECT id,name,email,phone,timezone,active FROM professionals WHERE id=$1 AND org_id=$2`,
			id, org).Scan(&pid, &name, &email, &phone, &tz, &active)
		if err != nil {
			WriteErr(w, http.StatusNotFound, "NOT_FOUND", "professional not found")
			return
		}
		WriteData(w, http.StatusOK, map[string]any{"id": pid, "name": name, "email": email, "phone": phone, "timezone": tz, "active": active}, nil)
	case http.MethodPatch, http.MethodPut:
		var in struct {
			Name     *string `json:"name"`
			Email    *string `json:"email"`
			Phone    *string `json:"phone"`
			Timezone *string `json:"timezone"`
			Active   *bool   `json:"active"`
		}
		if !Decode(w, r, &in) {
			return
		}
		cmd, err := s.Pool.Exec(r.Context(),
			`UPDATE professionals SET name=COALESCE($3,name), email=COALESCE($4,email), phone=COALESCE($5,phone), timezone=COALESCE($6,timezone), active=COALESCE($7,active) WHERE id=$1 AND org_id=$2`,
			id, org, in.Name, in.Email, in.Phone, in.Timezone, in.Active)
		if err != nil || cmd.RowsAffected() == 0 {
			WriteErr(w, http.StatusNotFound, "NOT_FOUND", "professional not found")
			return
		}
		WriteData(w, http.StatusOK, map[string]any{"id": id}, nil)
	case http.MethodDelete:
		cmd, _ := s.Pool.Exec(r.Context(), `DELETE FROM professionals WHERE id=$1 AND org_id=$2`, id, org)
		if cmd.RowsAffected() == 0 {
			WriteErr(w, http.StatusNotFound, "NOT_FOUND", "professional not found")
			return
		}
		WriteData(w, http.StatusOK, map[string]any{"deleted": true}, nil)
	}
}

func (s *Server) handleServices(w http.ResponseWriter, r *http.Request) {
	org := OrgID(r)
	if r.Method == http.MethodGet {
		listRows(w, r, s.Pool, `SELECT id,name,duration_min,price_cents,active FROM services WHERE org_id=$1 ORDER BY name`, []any{org},
			func(rows pgx.Rows) (any, error) {
				var id, name string
				var dur, price int
				var active bool
				err := rows.Scan(&id, &name, &dur, &price, &active)
				return map[string]any{"id": id, "name": name, "duration_min": dur, "price_cents": price, "active": active}, err
			})
		return
	}
	var in struct {
		Name        string `json:"name"`
		DurationMin int    `json:"duration_min"`
		PriceCents  int    `json:"price_cents"`
	}
	if !Decode(w, r, &in) {
		return
	}
	if in.Name == "" || in.DurationMin <= 0 {
		WriteErr(w, http.StatusBadRequest, "BAD_REQUEST", "name and positive duration_min are required")
		return
	}
	var id string
	if err := s.Pool.QueryRow(r.Context(),
		`INSERT INTO services(org_id,name,duration_min,price_cents) VALUES($1,$2,$3,$4) RETURNING id`,
		org, in.Name, in.DurationMin, in.PriceCents).Scan(&id); err != nil {
		WriteErr(w, http.StatusInternalServerError, "INTERNAL", "db error")
		return
	}
	WriteData(w, http.StatusCreated, map[string]any{"id": id, "name": in.Name, "duration_min": in.DurationMin, "price_cents": in.PriceCents, "active": true}, nil)
}

func (s *Server) handleService(w http.ResponseWriter, r *http.Request) {
	org, id := OrgID(r), chi.URLParam(r, "id")
	switch r.Method {
	case http.MethodGet:
		var sid, name string
		var dur, price int
		var active bool
		err := s.Pool.QueryRow(r.Context(),
			`SELECT id,name,duration_min,price_cents,active FROM services WHERE id=$1 AND org_id=$2`, id, org).
			Scan(&sid, &name, &dur, &price, &active)
		if err != nil {
			WriteErr(w, http.StatusNotFound, "NOT_FOUND", "service not found")
			return
		}
		WriteData(w, http.StatusOK, map[string]any{"id": sid, "name": name, "duration_min": dur, "price_cents": price, "active": active}, nil)
	case http.MethodPatch, http.MethodPut:
		var in struct {
			Name        *string `json:"name"`
			DurationMin *int    `json:"duration_min"`
			PriceCents  *int    `json:"price_cents"`
			Active      *bool   `json:"active"`
		}
		if !Decode(w, r, &in) {
			return
		}
		cmd, err := s.Pool.Exec(r.Context(),
			`UPDATE services SET name=COALESCE($3,name), duration_min=COALESCE($4,duration_min), price_cents=COALESCE($5,price_cents), active=COALESCE($6,active) WHERE id=$1 AND org_id=$2`,
			id, org, in.Name, in.DurationMin, in.PriceCents, in.Active)
		if err != nil || cmd.RowsAffected() == 0 {
			WriteErr(w, http.StatusNotFound, "NOT_FOUND", "service not found")
			return
		}
		WriteData(w, http.StatusOK, map[string]any{"id": id}, nil)
	case http.MethodDelete:
		cmd, _ := s.Pool.Exec(r.Context(), `DELETE FROM services WHERE id=$1 AND org_id=$2`, id, org)
		if cmd.RowsAffected() == 0 {
			WriteErr(w, http.StatusNotFound, "NOT_FOUND", "service not found")
			return
		}
		WriteData(w, http.StatusOK, map[string]any{"deleted": true}, nil)
	}
}

func (s *Server) handleWorkingHours(w http.ResponseWriter, r *http.Request) {
	org := OrgID(r)
	if r.Method == http.MethodGet {
		pid := r.URL.Query().Get("professional_id")
		q := `SELECT id,professional_id,weekday,start_min,end_min FROM working_hours WHERE org_id=$1`
		args := []any{org}
		if pid != "" {
			q += ` AND professional_id=$2`
			args = append(args, pid)
		}
		q += ` ORDER BY weekday,start_min`
		listRows(w, r, s.Pool, q, args, func(rows pgx.Rows) (any, error) {
			var id, p string
			var wd, sm, em int
			err := rows.Scan(&id, &p, &wd, &sm, &em)
			return map[string]any{"id": id, "professional_id": p, "weekday": wd, "start_min": sm, "end_min": em}, err
		})
		return
	}
	var in struct {
		ProfessionalID string `json:"professional_id"`
		Weekday        int    `json:"weekday"`
		StartMin       int    `json:"start_min"`
		EndMin         int    `json:"end_min"`
	}
	if !Decode(w, r, &in) {
		return
	}
	if in.ProfessionalID == "" || in.Weekday < 0 || in.Weekday > 6 || in.StartMin < 0 || in.EndMin <= in.StartMin || in.EndMin > 1440 {
		WriteErr(w, http.StatusBadRequest, "BAD_REQUEST", "invalid working-hours payload")
		return
	}
	if !s.ownsProfessional(r, org, in.ProfessionalID) {
		WriteErr(w, http.StatusNotFound, "NOT_FOUND", "professional not found")
		return
	}
	var id string
	if err := s.Pool.QueryRow(r.Context(),
		`INSERT INTO working_hours(org_id,professional_id,weekday,start_min,end_min) VALUES($1,$2,$3,$4,$5) RETURNING id`,
		org, in.ProfessionalID, in.Weekday, in.StartMin, in.EndMin).Scan(&id); err != nil {
		WriteErr(w, http.StatusInternalServerError, "INTERNAL", "db error")
		return
	}
	WriteData(w, http.StatusCreated, map[string]any{"id": id, "professional_id": in.ProfessionalID, "weekday": in.Weekday, "start_min": in.StartMin, "end_min": in.EndMin}, nil)
}

func (s *Server) handleWorkingHourDelete(w http.ResponseWriter, r *http.Request) {
	cmd, _ := s.Pool.Exec(r.Context(), `DELETE FROM working_hours WHERE id=$1 AND org_id=$2`, chi.URLParam(r, "id"), OrgID(r))
	if cmd.RowsAffected() == 0 {
		WriteErr(w, http.StatusNotFound, "NOT_FOUND", "working-hours not found")
		return
	}
	WriteData(w, http.StatusOK, map[string]any{"deleted": true}, nil)
}

func (s *Server) handleBlocks(w http.ResponseWriter, r *http.Request) {
	org := OrgID(r)
	if r.Method == http.MethodGet {
		pid := r.URL.Query().Get("professional_id")
		q := `SELECT id,professional_id,starts_at::text,ends_at::text,reason FROM blocks WHERE org_id=$1`
		args := []any{org}
		if pid != "" {
			q += ` AND professional_id=$2`
			args = append(args, pid)
		}
		q += ` ORDER BY starts_at`
		listRows(w, r, s.Pool, q, args, func(rows pgx.Rows) (any, error) {
			var id, p string
			var st, en string
			var reason *string
			// scan timestamptz as string via ::text
			err := rows.Scan(&id, &p, &st, &en, &reason)
			return map[string]any{"id": id, "professional_id": p, "starts_at": st, "ends_at": en, "reason": reason}, err
		})
		return
	}
	var in struct {
		ProfessionalID string  `json:"professional_id"`
		StartsAt       string  `json:"starts_at"`
		EndsAt         string  `json:"ends_at"`
		Reason         *string `json:"reason"`
	}
	if !Decode(w, r, &in) {
		return
	}
	if in.ProfessionalID == "" || in.StartsAt == "" || in.EndsAt == "" || in.EndsAt <= in.StartsAt {
		WriteErr(w, http.StatusBadRequest, "BAD_REQUEST", "invalid block payload")
		return
	}
	if !s.ownsProfessional(r, org, in.ProfessionalID) {
		WriteErr(w, http.StatusNotFound, "NOT_FOUND", "professional not found")
		return
	}
	var id string
	if err := s.Pool.QueryRow(r.Context(),
		`INSERT INTO blocks(org_id,professional_id,starts_at,ends_at,reason) VALUES($1,$2,$3::timestamptz,$4::timestamptz,$5) RETURNING id`,
		org, in.ProfessionalID, in.StartsAt, in.EndsAt, in.Reason).Scan(&id); err != nil {
		WriteErr(w, http.StatusInternalServerError, "INTERNAL", "db error")
		return
	}
	WriteData(w, http.StatusCreated, map[string]any{"id": id}, nil)
}

func (s *Server) handleBlockDelete(w http.ResponseWriter, r *http.Request) {
	cmd, _ := s.Pool.Exec(r.Context(), `DELETE FROM blocks WHERE id=$1 AND org_id=$2`, chi.URLParam(r, "id"), OrgID(r))
	if cmd.RowsAffected() == 0 {
		WriteErr(w, http.StatusNotFound, "NOT_FOUND", "block not found")
		return
	}
	WriteData(w, http.StatusOK, map[string]any{"deleted": true}, nil)
}

func (s *Server) handleCustomers(w http.ResponseWriter, r *http.Request) {
	org := OrgID(r)
	if r.Method == http.MethodGet {
		listRows(w, r, s.Pool, `SELECT id,name,email,phone FROM customers WHERE org_id=$1 ORDER BY name`, []any{org},
			func(rows pgx.Rows) (any, error) {
				var id, name string
				var email, phone *string
				err := rows.Scan(&id, &name, &email, &phone)
				return map[string]any{"id": id, "name": name, "email": email, "phone": phone}, err
			})
		return
	}
	var in struct {
		Name  string  `json:"name"`
		Email *string `json:"email"`
		Phone *string `json:"phone"`
	}
	if !Decode(w, r, &in) {
		return
	}
	if in.Name == "" {
		WriteErr(w, http.StatusBadRequest, "BAD_REQUEST", "name is required")
		return
	}
	var id string
	if err := s.Pool.QueryRow(r.Context(),
		`INSERT INTO customers(org_id,name,email,phone) VALUES($1,$2,$3,$4) RETURNING id`,
		org, in.Name, in.Email, in.Phone).Scan(&id); err != nil {
		WriteErr(w, http.StatusInternalServerError, "INTERNAL", "db error")
		return
	}
	WriteData(w, http.StatusCreated, map[string]any{"id": id, "name": in.Name, "email": in.Email, "phone": in.Phone}, nil)
}

func (s *Server) handleCustomer(w http.ResponseWriter, r *http.Request) {
	org, id := OrgID(r), chi.URLParam(r, "id")
	if r.Method == http.MethodGet {
		var cid, name string
		var email, phone *string
		err := s.Pool.QueryRow(r.Context(), `SELECT id,name,email,phone FROM customers WHERE id=$1 AND org_id=$2`, id, org).
			Scan(&cid, &name, &email, &phone)
		if err != nil {
			WriteErr(w, http.StatusNotFound, "NOT_FOUND", "customer not found")
			return
		}
		WriteData(w, http.StatusOK, map[string]any{"id": cid, "name": name, "email": email, "phone": phone}, nil)
		return
	}
	var in struct {
		Name  *string `json:"name"`
		Email *string `json:"email"`
		Phone *string `json:"phone"`
	}
	if !Decode(w, r, &in) {
		return
	}
	cmd, err := s.Pool.Exec(r.Context(),
		`UPDATE customers SET name=COALESCE($3,name), email=COALESCE($4,email), phone=COALESCE($5,phone) WHERE id=$1 AND org_id=$2`,
		id, org, in.Name, in.Email, in.Phone)
	if err != nil || cmd.RowsAffected() == 0 {
		WriteErr(w, http.StatusNotFound, "NOT_FOUND", "customer not found")
		return
	}
	WriteData(w, http.StatusOK, map[string]any{"id": id}, nil)
}
