package httpapi

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"bookflow-go/internal/auth"
)

func setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/v1/auth",
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Password string `json:"password"`
		OrgName string `json:"org_name"`
	}
	if !Decode(w, r, &in) {
		return
	}
	if in.Name == "" || in.Email == "" || in.Password == "" || in.OrgName == "" {
		WriteErr(w, http.StatusBadRequest, "BAD_REQUEST", "name, email, password and org_name are required")
		return
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		WriteErr(w, http.StatusInternalServerError, "INTERNAL", "could not hash password")
		return
	}
	ctx := r.Context()
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		WriteErr(w, http.StatusInternalServerError, "INTERNAL", "db error")
		return
	}
	defer tx.Rollback(ctx)
	var userID, orgID string
	err = tx.QueryRow(ctx,
		`INSERT INTO users(name,email,password_hash) VALUES($1,$2,$3) RETURNING id`,
		in.Name, in.Email, hash).Scan(&userID)
	if err != nil {
		WriteErr(w, http.StatusConflict, "EMAIL_TAKEN", "email already registered")
		return
	}
	err = tx.QueryRow(ctx, `INSERT INTO orgs(name) VALUES($1) RETURNING id`, in.OrgName).Scan(&orgID)
	if err != nil {
		WriteErr(w, http.StatusInternalServerError, "INTERNAL", "db error")
		return
	}
	if _, err = tx.Exec(ctx, `INSERT INTO org_members(org_id,user_id,role) VALUES($1,$2,'admin')`, orgID, userID); err != nil {
		WriteErr(w, http.StatusInternalServerError, "INTERNAL", "db error")
		return
	}
	if _, err = tx.Exec(ctx, `INSERT INTO settings(org_id) VALUES($1) ON CONFLICT DO NOTHING`, orgID); err != nil {
		WriteErr(w, http.StatusInternalServerError, "INTERNAL", "db error")
		return
	}
	if err = tx.Commit(ctx); err != nil {
		WriteErr(w, http.StatusInternalServerError, "INTERNAL", "db error")
		return
	}
	access, _ := auth.Sign(s.JWTSecret, userID, "access", auth.AccessTTL)
	refresh, _ := auth.Sign(s.JWTSecret, userID, "refresh", auth.RefreshTTL)
	setRefreshCookie(w, refresh)
	WriteData(w, http.StatusCreated, map[string]any{
		"access_token": access,
		"user":         map[string]any{"id": userID, "name": in.Name, "email": in.Email},
		"orgs":         []any{map[string]any{"id": orgID, "name": in.OrgName}},
	}, nil)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !Decode(w, r, &in) {
		return
	}
	var id, name, email, hash string
	err := s.Pool.QueryRow(r.Context(),
		`SELECT id,name,email,password_hash FROM users WHERE email=$1`, in.Email).Scan(&id, &name, &email, &hash)
	if err == pgx.ErrNoRows || !auth.CheckPassword(hash, in.Password) {
		WriteErr(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password")
		return
	}
	if err != nil {
		WriteErr(w, http.StatusInternalServerError, "INTERNAL", "db error")
		return
	}
	rows, err := s.Pool.Query(r.Context(),
		`SELECT o.id,o.name FROM orgs o JOIN org_members m ON m.org_id=o.id WHERE m.user_id=$1`, id)
	if err != nil {
		WriteErr(w, http.StatusInternalServerError, "INTERNAL", "db error")
		return
	}
	orgs := []any{}
	for rows.Next() {
		var oid, oname string
		_ = rows.Scan(&oid, &oname)
		orgs = append(orgs, map[string]any{"id": oid, "name": oname})
	}
	rows.Close()
	access, _ := auth.Sign(s.JWTSecret, id, "access", auth.AccessTTL)
	refresh, _ := auth.Sign(s.JWTSecret, id, "refresh", auth.RefreshTTL)
	setRefreshCookie(w, refresh)
	WriteData(w, http.StatusOK, map[string]any{
		"access_token": access,
		"user":         map[string]any{"id": id, "name": name, "email": email},
		"orgs":         orgs,
	}, nil)
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("refresh_token")
	if err != nil {
		WriteErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing refresh token")
		return
	}
	uid, err := auth.Parse(s.JWTSecret, c.Value, "refresh")
	if err != nil {
		WriteErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid refresh token")
		return
	}
	access, _ := auth.Sign(s.JWTSecret, uid, "access", auth.AccessTTL)
	WriteData(w, http.StatusOK, map[string]any{"access_token": access}, nil)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	var id, name, email string
	err := s.Pool.QueryRow(r.Context(),
		`SELECT id,name,email FROM users WHERE id=$1`, UserID(r)).Scan(&id, &name, &email)
	if err != nil {
		WriteErr(w, http.StatusNotFound, "NOT_FOUND", "user not found")
		return
	}
	rows, _ := s.Pool.Query(r.Context(),
		`SELECT o.id,o.name FROM orgs o JOIN org_members m ON m.org_id=o.id WHERE m.user_id=$1`, id)
	orgs := []any{}
	if rows != nil {
		for rows.Next() {
			var oid, oname string
			_ = rows.Scan(&oid, &oname)
			orgs = append(orgs, map[string]any{"id": oid, "name": oname})
		}
		rows.Close()
	}
	WriteData(w, http.StatusOK, map[string]any{
		"id": id, "name": name, "email": email, "orgs": orgs,
	}, nil)
}
