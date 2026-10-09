// Package models documents the core BookFlow domain types.
// Money is always integer cents; timestamps are UTC (timestamptz).
package models

import "time"

type Org struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Professional struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Email    *string `json:"email,omitempty"`
	Phone    *string `json:"phone,omitempty"`
	Timezone string  `json:"timezone"`
	Active   bool    `json:"active"`
}

type Service struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DurationMin int    `json:"duration_min"`
	PriceCents  int    `json:"price_cents"`
	Active      bool   `json:"active"`
}

type WorkingHours struct {
	ID             string `json:"id"`
	ProfessionalID string `json:"professional_id"`
	Weekday        int    `json:"weekday"` // 0=Sunday .. 6=Saturday
	StartMin       int    `json:"start_min"`
	EndMin         int    `json:"end_min"`
}

type Block struct {
	ID             string    `json:"id"`
	ProfessionalID string    `json:"professional_id"`
	StartsAt       time.Time `json:"starts_at"`
	EndsAt         time.Time `json:"ends_at"`
	Reason         string    `json:"reason,omitempty"`
}

type Customer struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Email *string `json:"email,omitempty"`
	Phone *string `json:"phone,omitempty"`
}

// Appointment statuses: confirmed | cancelled | completed | rescheduled.
type Appointment struct {
	ID             string    `json:"id"`
	Code           string    `json:"code"`
	OrgID          string    `json:"org_id"`
	ProfessionalID string    `json:"professional_id"`
	ServiceID      string    `json:"service_id"`
	CustomerID     *string   `json:"customer_id,omitempty"`
	CustomerName   string    `json:"customer_name"`
	CustomerEmail  string    `json:"customer_email"`
	CustomerPhone  string    `json:"customer_phone"`
	StartsAt       time.Time `json:"starts_at"`
	EndsAt         time.Time `json:"ends_at"`
	Status         string    `json:"status"`
	Notes          string    `json:"notes,omitempty"`
}

type Settings struct {
	CancelMinHours     float64 `json:"cancel_min_hours"`
	RescheduleMinHours float64 `json:"reschedule_min_hours"`
	SlotStepMin        int     `json:"slot_step_min"`
}
