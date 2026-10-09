package httpapi

import (
	"net/http"
	"sort"
	"time"
)

// interval in minutes since midnight.
type interval [2]int

// GenerateSlots returns "HH:MM" slot starts that fit durationMin inside work
// intervals (step grid), skipping any that overlap busy intervals.
// Pure function — unit tested without a DB.
func GenerateSlots(work []interval, busy []interval, durationMin, stepMin int) []string {
	if stepMin <= 0 {
		stepMin = 30
	}
	slots := []string{}
	fits := func(s, e int) bool {
		for _, b := range busy {
			if s < b[1] && b[0] < e {
				return false
			}
		}
		return true
	}
	for _, w := range work {
		for s := w[0]; s+durationMin <= w[1]; s += stepMin {
			if fits(s, s+durationMin) {
				slots = append(slots, minutesToStr(s))
			}
		}
	}
	sort.Strings(slots)
	return slots
}

func minutesToStr(m int) string {
	return time.Date(0, 1, 1, m/60, m%60, 0, 0, time.UTC).Format("15:04")
}

// Overlaps reports whether [aStart,aEnd) overlaps [bStart,bEnd).
func Overlaps(aStart, aEnd, bStart, bEnd time.Time) bool {
	return aStart.Before(bEnd) && bStart.Before(aEnd)
}

// BeforeHours returns true when now is at least minHours before start.
func BeforeHours(now, start time.Time, minHours float64) bool {
	return start.Sub(now).Hours() >= minHours
}

func (s *Server) handleAvailability(w http.ResponseWriter, r *http.Request) {
	s.serveAvailability(w, r, OrgID(r))
}

func (s *Server) serveAvailability(w http.ResponseWriter, r *http.Request, org string) {
	pid := r.URL.Query().Get("professional_id")
	sid := r.URL.Query().Get("service_id")
	dateStr := r.URL.Query().Get("date")
	if pid == "" || sid == "" || dateStr == "" {
		WriteErr(w, http.StatusBadRequest, "BAD_REQUEST", "professional_id, service_id and date are required")
		return
	}
	ctx := r.Context()
	var tzName string
	var active bool
	err := s.Pool.QueryRow(ctx,
		`SELECT timezone, active FROM professionals WHERE id=$1 AND org_id=$2`, pid, org).Scan(&tzName, &active)
	if err != nil {
		WriteErr(w, http.StatusNotFound, "NOT_FOUND", "professional not found")
		return
	}
	var durationMin int
	err = s.Pool.QueryRow(ctx,
		`SELECT duration_min FROM services WHERE id=$1 AND org_id=$2`, sid, org).Scan(&durationMin)
	if err != nil {
		WriteErr(w, http.StatusNotFound, "NOT_FOUND", "service not found")
		return
	}
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		loc = time.UTC
	}
	day, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		WriteErr(w, http.StatusBadRequest, "BAD_REQUEST", "date must be YYYY-MM-DD")
		return
	}
	weekday := int(day.Weekday())
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, loc)
	dayEnd := dayStart.Add(24 * time.Hour)

	rows, err := s.Pool.Query(ctx,
		`SELECT start_min,end_min FROM working_hours WHERE org_id=$1 AND professional_id=$2 AND weekday=$3`,
		org, pid, weekday)
	if err != nil {
		WriteErr(w, http.StatusInternalServerError, "INTERNAL", "db error")
		return
	}
	var work []interval
	for rows.Next() {
		var sm, em int
		_ = rows.Scan(&sm, &em)
		work = append(work, interval{sm, em})
	}
	rows.Close()

	var busy []interval
	toMin := func(t time.Time) int {
		lt := t.In(loc)
		return lt.Hour()*60 + lt.Minute()
	}
	// blocks overlapping the day
	brows, err := s.Pool.Query(ctx,
		`SELECT starts_at,ends_at FROM blocks WHERE org_id=$1 AND professional_id=$2 AND starts_at < $3 AND ends_at > $4`,
		org, pid, dayEnd.UTC(), dayStart.UTC())
	if err == nil {
		for brows.Next() {
			var st, en time.Time
			_ = brows.Scan(&st, &en)
			if st.Before(dayStart.UTC()) {
				st = dayStart.UTC()
			}
			if en.After(dayEnd.UTC()) {
				en = dayEnd.UTC()
			}
			busy = append(busy, interval{toMin(st), toMin(en)})
		}
		brows.Close()
	}
	// confirmed/rescheduled appointments overlapping the day
	arows, err := s.Pool.Query(ctx,
		`SELECT starts_at,ends_at FROM appointments WHERE org_id=$1 AND professional_id=$2 AND status IN ('confirmed','rescheduled') AND starts_at < $3 AND ends_at > $4`,
		org, pid, dayEnd.UTC(), dayStart.UTC())
	if err == nil {
		for arows.Next() {
			var st, en time.Time
			_ = arows.Scan(&st, &en)
			busy = append(busy, interval{toMin(st), toMin(en)})
		}
		arows.Close()
	}

	step := s.slotStep(ctx, org)
	slots := GenerateSlots(work, busy, durationMin, step)
	if slots == nil {
		slots = []string{}
	}
	// hide past slots when the requested day is today in the professional's timezone
	nowLoc := time.Now().In(loc)
	if day.Year() == nowLoc.Year() && day.YearDay() == nowLoc.YearDay() {
		nowMin := nowLoc.Hour()*60 + nowLoc.Minute()
		kept := slots[:0]
		for _, s := range slots {
			h, m := parseSlot(s)
			if h*60+m > nowMin {
				kept = append(kept, s)
			}
		}
		slots = kept
	}

	WriteData(w, http.StatusOK, map[string]any{"slots": slots, "timezone": tzName}, nil)
}
func parseSlot(s string) (int, int) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return -1, 0
	}
	return t.Hour(), t.Minute()
}
