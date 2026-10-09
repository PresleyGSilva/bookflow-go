export type Service = { id: string; name: string; duration_min: number; price_cents: number; active: boolean };
export type Professional = { id: string; name: string; email?: string; phone?: string; timezone?: string; active: boolean };
export type Customer = { id: string; name: string; email?: string; phone?: string };
export type Appointment = {
  id: string; code: string; professional_id: string; service_id: string;
  starts_at: string; ends_at: string; status: string; notes?: string;
  customer_name?: string; customer_email?: string; service_name?: string; professional_name?: string;
};
export type Availability = { slots: string[]; timezone: string };
