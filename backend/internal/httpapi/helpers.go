package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
)

type ctxKey string

const (
	userKey ctxKey = "user_id"
	orgKey  ctxKey = "org_id"
)

func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userKey, id)
}

func WithOrgID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, orgKey, id)
}

func UserID(r *http.Request) string {
	v, _ := r.Context().Value(userKey).(string)
	return v
}

func OrgID(r *http.Request) string {
	v, _ := r.Context().Value(orgKey).(string)
	return v
}

func WriteData(w http.ResponseWriter, status int, data any, meta any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	out := map[string]any{"data": data}
	if meta != nil {
		out["meta"] = meta
	}
	_ = json.NewEncoder(w).Encode(out)
}

func WriteErr(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
}

func Decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		WriteErr(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body")
		return false
	}
	return true
}
