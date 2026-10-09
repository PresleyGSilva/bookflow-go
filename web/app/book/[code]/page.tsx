"use client";
import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { apiFetch } from "../../../lib/api";
import type { Appointment } from "../../../lib/types";

export default function ManageBookingPage() {
  const params = useParams();
  const code = Array.isArray(params.code) ? params.code[0] : (params.code as string);
  const [appt, setAppt] = useState<Appointment | null>(null);
  const [error, setError] = useState("");
  const [ok, setOk] = useState("");
  const [date, setDate] = useState("");
  const [time, setTime] = useState("");
  const [loading, setLoading] = useState(true);

  async function load() {
    setLoading(true); setError("");
    try {
      setAppt(await apiFetch<Appointment>(`/api/v1/bookings/public/${code}`));
    } catch (e) { setError(e instanceof Error ? e.message : String(e)); } finally { setLoading(false); }
  }

  useEffect(() => { load(); }, [code]);

  async function reschedule() {
    setError(""); setOk("");
    try {
      const updated = await apiFetch<Appointment>(`/api/v1/bookings/public/${code}/reschedule`, { method: "POST", body: { date, time } });
      setAppt(updated); setOk("Reserva reagendada com sucesso.");
    } catch (e) { setError(e instanceof Error ? e.message : String(e)); }
  }

  async function cancel() {
    setError(""); setOk("");
    try {
      const updated = await apiFetch<Appointment>(`/api/v1/bookings/public/${code}/cancel`, { method: "POST" });
      setAppt(updated); setOk("Reserva cancelada.");
    } catch (e) { setError(e instanceof Error ? e.message : String(e)); }
  }

  if (loading) return <p>Carregando...</p>;
  if (!appt) return <p className="text-red-600">{error || "Reserva não encontrada."}</p>;

  return (
    <div className="max-w-md mx-auto grid gap-4">
      <h1 className="text-2xl font-bold">Reserva {appt.code}</h1>
      <div className="card grid gap-1 text-sm">
        <p><b>Status:</b> {appt.status}</p>
        <p><b>Início:</b> {new Date(appt.starts_at).toLocaleString("pt-BR")}</p>
        <p><b>Fim:</b> {new Date(appt.ends_at).toLocaleString("pt-BR")}</p>
        {appt.customer_name && <p><b>Cliente:</b> {appt.customer_name}</p>}
      </div>
      {appt.status !== "cancelled" && (
        <div className="card grid gap-3">
          <h2 className="font-semibold">Reagendar (até 4h antes)</h2>
          <div className="grid grid-cols-2 gap-3">
            <div><label className="label">Data</label><input className="input" type="date" value={date} onChange={(e) => setDate(e.target.value)} /></div>
            <div><label className="label">Hora</label><input className="input" value={time} onChange={(e) => setTime(e.target.value)} placeholder="14:00" /></div>
          </div>
          <button className="btn-ghost" onClick={reschedule}>Reagendar</button>
          <h2 className="font-semibold mt-2">Cancelar (até 2h antes)</h2>
          <button className="btn bg-red-600 hover:bg-red-700" onClick={cancel}>Cancelar reserva</button>
        </div>
      )}
      {ok && <p className="text-sm text-green-700">{ok}</p>}
      {error && <p className="text-sm text-red-600">{error}</p>}
    </div>
  );
}
