"use client";
import { useEffect, useState } from "react";
import { apiFetch } from "../../../lib/api";

type Settings = { cancel_min_hours: number; reschedule_min_hours: number; slot_step_min: number };

export default function SettingsPage() {
  const [s, setS] = useState<Settings>({ cancel_min_hours: 2, reschedule_min_hours: 4, slot_step_min: 30 });
  const [msg, setMsg] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    apiFetch<Settings>("/api/v1/admin/settings").then(setS).catch((e) => setError(e instanceof Error ? e.message : String(e)));
  }, []);

  async function save(e: React.FormEvent) {
    e.preventDefault(); setMsg(""); setError("");
    try {
      setS(await apiFetch<Settings>("/api/v1/admin/settings", { method: "PATCH", body: s }));
      setMsg("Configurações salvas.");
    } catch (e) { setError(e instanceof Error ? e.message : String(e)); }
  }

  return (
    <div className="grid gap-4 max-w-lg">
      <h1 className="text-2xl font-bold">Configurações</h1>
      <form onSubmit={save} className="card grid gap-3">
        <div><label className="label">Cancelamento mínimo (horas)</label><input className="input" type="number" value={s.cancel_min_hours} onChange={(e) => setS({ ...s, cancel_min_hours: Number(e.target.value) })} /></div>
        <div><label className="label">Reagendamento mínimo (horas)</label><input className="input" type="number" value={s.reschedule_min_hours} onChange={(e) => setS({ ...s, reschedule_min_hours: Number(e.target.value) })} /></div>
        <div><label className="label">Passo do slot (min)</label><input className="input" type="number" value={s.slot_step_min} onChange={(e) => setS({ ...s, slot_step_min: Number(e.target.value) })} /></div>
        {msg && <p className="text-sm text-green-700">{msg}</p>}
        {error && <p className="text-sm text-red-600">{error}</p>}
        <button className="btn">Salvar</button>
      </form>
    </div>
  );
}
