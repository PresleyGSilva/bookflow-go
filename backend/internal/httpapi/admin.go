package httpapi

import (
	"fmt"
	"net/http"
	"time"
)

func (s *Server) handleCalendar(w http.ResponseWriter, r *http.Request) {
	org := OrgID(r)
	q := `SELECT a.id,a.code,a.org_id,a.professional_id,a.service_id,a.customer_id,
		a.customer_name,a.customer_email,a.customer_phone,a.starts_at,a.ends_at,a.status,a.notes,
		p.name,s.name FROM appointments a
		JOIN professionals p ON p.id=a.professional_id
		JOIN services s ON s.id=a.service_id
		WHERE a.org_id=$1`
	args := []any{org}
	i := 2
	if v := r.URL.Query().Get("from"); v != "" {
		q += fmt.Sprintf(` AND a.starts_at >= $%d::timestamptz`, i)
		args = append(args, v)
		i++
	}
	if v := r.URL.Query().Get("to"); v != "" {
		q += fmt.Sprintf(` AND a.starts_at < $%d::timestamptz`, i)
		args = append(args, v)
		i++
	}
	if v := r.URL.Query().Get("professional_id"); v != "" {
		q += fmt.Sprintf(` AND a.professional_id=$%d`, i)
		args = append(args, v)
		i++
	}
	q += ` ORDER BY a.starts_at`
	rows, err := s.Pool.Query(r.Context(), q, args...)
	if err != nil {
		WriteErr(w, http.StatusInternalServerError, "INTERNAL", "db error")
		return
	}
	defer rows.Close()
	out := []any{}
	for rows.Next() {
		var id, code, oid, pid, sid, status, pname, sname string
		var cid *string
		var cn, ce, cp, notes *string
		var st, en time.Time
		_ = rows.Scan(&id, &code, &oid, &pid, &sid, &cid, &cn, &ce, &cp, &st, &en, &status, &notes, &pname, &sname)
		str := func(p *string) string {
			if p == nil {
				return ""
			}
			return *p
		}
		out = append(out, map[string]any{
			"id": id, "code": code, "org_id": oid, "professional_id": pid, "service_id": sid,
			"customer_id": cid, "customer_name": str(cn), "customer_email": str(ce), "customer_phone": str(cp),
			"starts_at": st.Format(time.RFC3339), "ends_at": en.Format(time.RFC3339), "status": status, "notes": str(notes),
			"professional_name": pname, "service_name": sname,
		})
	}
	WriteData(w, http.StatusOK, out, nil)
}

func (s *Server) handleReports(w http.ResponseWriter, r *http.Request) {
	org := OrgID(r)
	from, to := r.URL.Query().Get("from"), r.URL.Query().Get("to")
	args := []any{org}
	filter := `a.org_id=$1 AND a.status IN ('confirmed','rescheduled','completed')`
	if from != "" {
		filter += fmt.Sprintf(` AND starts_at >= $%d::timestamptz`, len(args)+1)
		args = append(args, from)
	}
	if to != "" {
		filter += fmt.Sprintf(` AND starts_at < $%d::timestamptz`, len(args)+1)
		args = append(args, to)
	}
	ctx := r.Context()
	var total int
	var revenue int
	_ = s.Pool.QueryRow(ctx,
		`SELECT COUNT(*), COALESCE(SUM(s.price_cents),0) FROM appointments a JOIN services s ON s.id=a.service_id WHERE `+filter,
		args...).Scan(&total, &revenue)
	var cancelled int
	cancelFilter := `a.org_id=$1 AND a.status='cancelled'`
	cancelArgs := []any{org}
	if from != "" {
		cancelFilter += fmt.Sprintf(` AND a.starts_at >= $%d::timestamptz`, len(cancelArgs)+1)
		cancelArgs = append(cancelArgs, from)
	}
	if to != "" {
		cancelFilter += fmt.Sprintf(` AND a.starts_at < $%d::timestamptz`, len(cancelArgs)+1)
		cancelArgs = append(cancelArgs, to)
	}
	_ = s.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM appointments a WHERE `+cancelFilter, cancelArgs...).Scan(&cancelled)

	perArgs := func() (string, []any) {
		f := ""
		a := []any{org}
		if from != "" {
			f += fmt.Sprintf(` AND a.starts_at >= $%d::timestamptz`, len(a)+1)
			a = append(a, from)
		}
		if to != "" {
			f += fmt.Sprintf(` AND a.starts_at < $%d::timestamptz`, len(a)+1)
			a = append(a, to)
		}
		return f, a
	}
	perProf := []any{}
	if df, pa := perArgs(); true {
		rows, err := s.Pool.Query(ctx,
			`SELECT p.id,p.name,COUNT(a.id),COALESCE(SUM(s.price_cents),0) FROM professionals p
			 LEFT JOIN appointments a ON a.professional_id=p.id AND a.status IN ('confirmed','rescheduled','completed')`+
				df+
				` LEFT JOIN services s ON s.id=a.service_id WHERE p.org_id=$1 GROUP BY p.id,p.name ORDER BY p.name`, pa...)
		if err == nil {
			for rows.Next() {
				var id, name string
				var c, rev int
				_ = rows.Scan(&id, &name, &c, &rev)
				perProf = append(perProf, map[string]any{"professional_id": id, "name": name, "count": c, "revenue_cents": rev})
			}
			rows.Close()
		}
	}
	perSvc := []any{}
	if df, pa := perArgs(); true {
		rows2, err := s.Pool.Query(ctx,
			`SELECT s.id,s.name,COUNT(a.id),COALESCE(SUM(CASE WHEN a.id IS NULL THEN 0 ELSE s.price_cents END),0) FROM services s
			 LEFT JOIN appointments a ON a.service_id=s.id AND a.status IN ('confirmed','rescheduled','completed')`+
				df+
				` WHERE s.org_id=$1 GROUP BY s.id,s.name ORDER BY s.name`, pa...)
		if err == nil {
			for rows2.Next() {
				var id, name string
				var c, rev int
				_ = rows2.Scan(&id, &name, &c, &rev)
				perSvc = append(perSvc, map[string]any{"service_id": id, "name": name, "count": c, "revenue_cents": rev})
			}
			rows2.Close()
		}
	}
	WriteData(w, http.StatusOK, map[string]any{
		"totals":           map[string]any{"appointments": total, "cancelled": cancelled, "revenue_cents": revenue},
		"per_professional": perProf,
		"per_service":      perSvc,
	}, nil)
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	org := OrgID(r)
	if r.Method == http.MethodGet {
		st := s.getSettings(r.Context(), org)
		WriteData(w, http.StatusOK, map[string]any{
			"cancel_min_hours": st.cancelH, "reschedule_min_hours": st.reschedH, "slot_step_min": st.stepMin,
		}, nil)
		return
	}
	var in struct {
		CancelMinHours     *float64 `json:"cancel_min_hours"`
		RescheduleMinHours *float64 `json:"reschedule_min_hours"`
		SlotStepMin        *int     `json:"slot_step_min"`
	}
	if !Decode(w, r, &in) {
		return
	}
	_, err := s.Pool.Exec(r.Context(),
		`INSERT INTO settings(org_id,cancel_min_hours,reschedule_min_hours,slot_step_min) VALUES($1,2,4,30)
		 ON CONFLICT (org_id) DO UPDATE SET
		 cancel_min_hours=COALESCE($2,cancel_min_hours),
		 reschedule_min_hours=COALESCE($3,reschedule_min_hours),
		 slot_step_min=COALESCE($4,slot_step_min)`,
		org, in.CancelMinHours, in.RescheduleMinHours, in.SlotStepMin)
	if err != nil {
		WriteErr(w, http.StatusInternalServerError, "INTERNAL", "db error")
		return
	}
	st := s.getSettings(r.Context(), org)
	WriteData(w, http.StatusOK, map[string]any{
		"cancel_min_hours": st.cancelH, "reschedule_min_hours": st.reschedH, "slot_step_min": st.stepMin,
	}, nil)
}
