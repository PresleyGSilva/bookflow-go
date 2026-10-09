"use client";
import { useEffect, useState } from "react";
import { apiFetch } from "../../lib/api";
import type { Appointment, Professional } from "../../lib/types";

export default function AdminCalendar() {
  const [from, setFrom] = useState(new Date().toISOString().slice(0, 10));
  const [to, setTo] = useState(new Date(Date.now() + 6 * 864e5).toISOString().slice(0, 10));
  const [proId, setProId] = useState("");
  const [pros, setPros] = useState<Professional[]>([]);
  const [items, setItems] = useState<Appointment[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    apiFetch<Professional[]>("/api/v1/professionals").then(setPros).catch(() => {});
  }, []);

  async function load() {
    setError("");
    try {
      const q = new URLSearchParams({ from, to, ...(proId ? { professional_id: proId } : {}) });
      setItems(await apiFetch<Appointment[]>(`/api/v1/admin/calendar?${q}`));
    } catch (e) { setError(e instanceof Error ? e.message : String(e)); }
  }

  useEffect(() => { load(); }, []);

  async function cancel(id: string) {
    try {
      await apiFetch(`/api/v1/appointments/${id}/cancel`, { method: "POST" });
      load();
    } catch (e) { setError(e instanceof Error ? e.message : String(e)); }
  }

  const days: Record<string, Appointment[]> = {};
  for (const a of items) {
    const d = new Date(a.starts_at).toLocaleDateString("pt-BR");
    (days[d] ||= []).push(a);
  }

  return (
    <div className="grid gap-4">
      <h1 className="text-2xl font-bold">Calendário</h1>
      <div className="card flex flex-wrap gap-3 items-end">
        <div><label className="label">De</label><input className="input" type="date" value={from} onChange={(e) => setFrom(e.target.value)} /></div>
        <div><label className="label">Até</label><input className="input" type="date" value={to} onChange={(e) => setTo(e.target.value)} /></div>
        <div><label className="label">Profissional</label>
          <select className="input" value={proId} onChange={(e) => setProId(e.target.value)}>
            <option value="">Todos</option>
            {pros.map((p) => <option key={p.id} value={p.id}>{p.name}</option>)}
          </select></div>
        <button className="btn" onClick={load}>Atualizar</button>
      </div>
      {error && <p className="text-sm text-red-600">{error}</p>}
      {Object.keys(days).sort().map((d) => (
        <div key={d} className="card">
          <h2 className="font-semibold mb-2">{d}</h2>
          <table className="grid">
            <thead><tr><th>Hora</th><th>Serviço</th><th>Profissional</th><th>Cliente</th><th>Status</th><th></th></tr></thead>
            <tbody>
              {days[d].sort((a, b) => a.starts_at.localeCompare(b.starts_at)).map((a) => (
                <tr key={a.id}>
                  <td>{new Date(a.starts_at).toLocaleTimeString("pt-BR", { hour: "2-digit", minute: "2-digit" })}</td>
                  <td>{a.service_name ?? a.service_id}</td>
                  <td>{a.professional_name ?? a.professional_id}</td>
                  <td>{a.customer_name ?? "—"}</td>
                  <td>{a.status}</td>
                  <td>{a.status === "confirmed" && <button className="text-red-600 text-xs underline" onClick={() => cancel(a.id)}>cancelar</button>}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ))}
      {items.length === 0 && <p className="text-slate-500">Sem agendamentos no período.</p>}
    </div>
  );
}
