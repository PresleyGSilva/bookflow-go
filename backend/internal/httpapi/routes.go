package httpapi

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func (s *Server) Routes(openapiPath string) http.Handler {
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Org-ID", "X-User"},
		AllowCredentials: true,
	}))
	r.Get("/healthz", func(w http.ResponseWriter, req *http.Request) {
		WriteData(w, http.StatusOK, map[string]any{"status": "ok"}, nil)
	})
	r.Get("/openapi.yaml", func(w http.ResponseWriter, req *http.Request) {
		b, err := os.ReadFile(openapiPath)
		if err != nil {
			WriteErr(w, http.StatusNotFound, "NOT_FOUND", "openapi spec not found")
			return
		}
		w.Header().Set("Content-Type", "text/yaml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", s.handleRegister)
		r.Post("/auth/login", s.handleLogin)
		r.Post("/auth/refresh", s.handleRefresh)

		// public booking access (no auth)
		r.Get("/public/orgs", s.handlePublicOrgs)
		r.Get("/public/services", s.handlePublicServices)
		r.Get("/public/professionals", s.handlePublicProfessionals)
		r.Get("/public/availability", s.handlePublicAvailability)
		r.Post("/bookings/public", s.handlePublicCreate)
		r.Get("/bookings/public/{code}", s.handlePublicGet)
		r.Post("/bookings/public/{code}/reschedule", s.handlePublicReschedule)
		r.Post("/bookings/public/{code}/cancel", s.handlePublicCancel)

		r.Group(func(r chi.Router) {
			r.Use(JWT(s.JWTSecret))
			r.Get("/auth/me", s.handleMe)
			r.With(RequireOrg(s.Pool)).Group(func(r chi.Router) {
				r.Get("/professionals", s.handleProfessionals)
				r.Post("/professionals", s.handleProfessionals)
				r.Get("/professionals/{id}", s.handleProfessional)
				r.Patch("/professionals/{id}", s.handleProfessional)
				r.Delete("/professionals/{id}", s.handleProfessional)

				r.Get("/services", s.handleServices)
				r.Post("/services", s.handleServices)
				r.Get("/services/{id}", s.handleService)
				r.Patch("/services/{id}", s.handleService)
				r.Delete("/services/{id}", s.handleService)

				r.Get("/working-hours", s.handleWorkingHours)
				r.Post("/working-hours", s.handleWorkingHours)
				r.Delete("/working-hours/{id}", s.handleWorkingHourDelete)

				r.Get("/blocks", s.handleBlocks)
				r.Post("/blocks", s.handleBlocks)
				r.Delete("/blocks/{id}", s.handleBlockDelete)

				r.Get("/customers", s.handleCustomers)
				r.Post("/customers", s.handleCustomers)
				r.Get("/customers/{id}", s.handleCustomer)
				r.Patch("/customers/{id}", s.handleCustomer)

				r.Get("/availability", s.handleAvailability)

				r.Get("/appointments", s.handleAppointments)
				r.Post("/appointments", s.handleAppointments)
				r.Post("/appointments/{id}/reschedule", s.handleApptReschedule)
				r.Post("/appointments/{id}/cancel", s.handleApptCancel)

				r.Get("/admin/calendar", s.handleCalendar)
				r.Get("/admin/reports", s.handleReports)
				r.Get("/admin/settings", s.handleSettings)
				r.Patch("/admin/settings", s.handleSettings)
			})
		})
	})
	return r
}
