package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Skipf("no test db: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		t.Skipf("no test db: %v", err)
	}
	return pool
}

func seedOrg(t *testing.T, pool *pgxpool.Pool) (org, pid, sid string) {
	t.Helper()
	ctx := context.Background()
	_ = pool.QueryRow(ctx, `INSERT INTO orgs(name) VALUES('test org') RETURNING id`).Scan(&org)
	_ = pool.QueryRow(ctx, `INSERT INTO users(name,email,password_hash) VALUES('t',$1,'x') RETURNING id`, fmt.Sprintf("t%d@x.com", time.Now().UnixNano())).Scan(new(string))
	_ = pool.QueryRow(ctx, `INSERT INTO professionals(org_id,name,timezone) VALUES($1,'Pro','America/Sao_Paulo') RETURNING id`, org).Scan(&pid)
	_ = pool.QueryRow(ctx, `INSERT INTO services(org_id,name,duration_min,price_cents) VALUES($1,'Cut',30,1000) RETURNING id`, org).Scan(&sid)
	_, _ = pool.Exec(ctx, `INSERT INTO settings(org_id) VALUES($1) ON CONFLICT DO NOTHING`, org)
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM orgs WHERE id=$1`, org) })
	return org, pid, sid
}

// Two concurrent bookings on the same slot: exactly one must win.
func TestConcurrentSameSlot(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	s := New(pool, "test-secret")
	org, pid, sid := seedOrg(t, pool)
	starts := time.Now().Add(48 * time.Hour).Truncate(time.Minute).UTC()

	var won atomic.Int64
	var wg sync.WaitGroup
	for i := range 2 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _, _, err := s.createAppointmentTx(context.Background(), org, pid, sid, 30, starts,
				nil, fmt.Sprintf("c%d", i), "", "", "")
			if err == nil {
				won.Add(1)
			} else if err.Error() != "SLOT_TAKEN" {
				t.Errorf("unexpected error: %v", err)
			}
		}(i)
	}
	wg.Wait()
	if won.Load() != 1 {
		t.Fatalf("expected exactly 1 winner, got %d", won.Load())
	}
}

func TestAvailabilityRules(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	s := New(pool, "test-secret")
	org, pid, sid := seedOrg(t, pool)
	ctx := context.Background()
	// Monday working hours 08:00-09:00
	_, _ = pool.Exec(ctx, `INSERT INTO working_hours(org_id,professional_id,weekday,start_min,end_min) VALUES($1,$2,1,480,540)`, org, pid)
	// next Monday
	now := time.Now()
	days := (8 - int(now.Weekday())) % 7
	if days == 0 {
		days = 7
	}
	monday := now.AddDate(0, 0, days).Format("2006-01-02")

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/availability?professional_id="+pid+"&service_id="+sid+"&date="+monday, nil)
	req = req.WithContext(WithOrgID(WithUserID(req.Context(), "x"), org))
	rec := httptest.NewRecorder()
	s.handleAvailability(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); !contains(body, "08:00") || !contains(body, "08:30") {
		t.Fatalf("expected 08:00 and 08:30 slots, got %s", body)
	}
}

func TestAuthPermissions(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()


	// no token -> 401
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	rec := httptest.NewRecorder()
	mw := JWT("test-secret")
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	// other org -> 403
	orgA, _, _ := seedOrg(t, pool)
	orgB, _, _ := seedOrg(t, pool)
	var otherUser string
	_ = pool.QueryRow(context.Background(),
		`INSERT INTO users(name,email,password_hash) VALUES('o',$1,'x') RETURNING id`,
		fmt.Sprintf("o%d@x.com", time.Now().UnixNano())).Scan(&otherUser)
	_, _ = pool.Exec(context.Background(), `INSERT INTO org_members(org_id,user_id) VALUES($1,$2)`, orgB, otherUser)

	// user IS member of orgB -> passes; use orgA for forbidden
	req3 := httptest.NewRequest(http.MethodGet, "/api/v1/professionals", nil)
	req3.Header.Set("X-Org-ID", orgA)
	req3 = req3.WithContext(WithUserID(req3.Context(), otherUser))
	rec3 := httptest.NewRecorder()
	RequireOrg(pool)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for foreign org, got %d", rec3.Code)
	}
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}
