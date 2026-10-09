package httpapi

import (
	"reflect"
	"testing"
	"time"
)

func TestGenerateSlotsBasic(t *testing.T) {
	work := []interval{{480, 720}} // 08:00-12:00
	got := GenerateSlots(work, nil, 30, 30)
	want := []string{"08:00", "08:30", "09:00", "09:30", "10:00", "10:30", "11:00", "11:30"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestGenerateSlotsSkipsBusy(t *testing.T) {
	work := []interval{{480, 720}}
	busy := []interval{{540, 570}} // 09:00-09:30 taken
	got := GenerateSlots(work, busy, 30, 30)
	for _, s := range got {
		if s == "09:00" {
			t.Fatalf("slot 09:00 should be blocked, got %v", got)
		}
	}
	found := false
	for _, s := range got {
		if s == "09:30" {
			found = true
		}
	}
	if !found {
		t.Fatalf("slot 09:30 should be free, got %v", got)
	}
}

func TestGenerateSlotsDurationFit(t *testing.T) {
	work := []interval{{480, 540}} // 1h window
	got := GenerateSlots(work, nil, 60, 30)
	if !reflect.DeepEqual(got, []string{"08:00"}) {
		t.Fatalf("got %v", got)
	}
	if n := len(GenerateSlots(work, nil, 90, 30)); n != 0 {
		t.Fatalf("90min service should yield no slots, got %d", n)
	}
}

func TestGenerateSlotsStep(t *testing.T) {
	work := []interval{{480, 600}}
	got := GenerateSlots(work, nil, 30, 15)
	if len(got) != 7 || got[0] != "08:00" || got[1] != "08:15" {
		t.Fatalf("got %v", got)
	}
}

func TestOverlaps(t *testing.T) {
	base := time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC)
	cases := []struct {
		name       string
		bStart     time.Time
		bEnd       time.Time
		want       bool
	}{
		{"exact", base, base.Add(time.Hour), true},
		{"inside", base.Add(15 * time.Minute), base.Add(30 * time.Minute), true},
		{"partial-end", base.Add(30 * time.Minute), base.Add(90 * time.Minute), true},
		{"partial-start", base.Add(-30 * time.Minute), base.Add(30 * time.Minute), true},
		{"adjacent-after", base.Add(time.Hour), base.Add(2 * time.Hour), false},
		{"adjacent-before", base.Add(-time.Hour), base, false},
		{"far", base.Add(2 * time.Hour), base.Add(3 * time.Hour), false},
	}
	for _, c := range cases {
		if got := Overlaps(base, base.Add(time.Hour), c.bStart, c.bEnd); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}

func TestBeforeHours(t *testing.T) {
	now := time.Date(2026, 10, 10, 7, 0, 0, 0, time.UTC)
	if !BeforeHours(now, now.Add(3*time.Hour), 2) {
		t.Error("3h ahead should allow 2h cancel")
	}
	if BeforeHours(now, now.Add(90*time.Minute), 2) {
		t.Error("1.5h ahead should block 2h cancel")
	}
	if !BeforeHours(now, now.Add(5*time.Hour), 4) {
		t.Error("5h ahead should allow 4h reschedule")
	}
	if BeforeHours(now, now.Add(3*time.Hour), 4) {
		t.Error("3h ahead should block 4h reschedule")
	}
}
