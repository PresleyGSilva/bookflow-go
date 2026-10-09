"use client";
import { useEffect, useState } from "react";
import { apiFetch } from "../../../lib/api";
import { formatBRL } from "../../../lib/format";

type Report = {
  totals: { appointments: number; cancelled: number; revenue_cents: number };
  per_professional: { professional_id: string; name?: string; count: number; revenue_cents: number }[];
  per_service: { service_id: string; name?: string; count: number; revenue_cents: number }[];
  revenue_cents: number;
};

export default function ReportsPage() {
  const [from, setFrom] = useState(new Date(Date.now() - 30 * 864e5).toISOString().slice(0, 10));
  const [to, setTo] = useState(new Date().toISOString().slice(0, 10));
  const [rep, setRep] = useState<Report | null>(null);
  const [error, setError] = useState("");

  async function load() {
    setError("");
    try { setRep(await apiFetch<Report>(`/api/v1/admin/reports?from=${from}&to=${to}`)); }
    catch (e) { setError(e instanceof Error ? e.message : String(e)); }
  }
  useEffect(() => { load(); }, []);

  const revenue = rep ? (rep.revenue_cents || rep.totals.revenue_cents || 0) : 0;

  return (
    <div className="grid gap-4">
      <h1 className="text-2xl font-bold">Relatórios</h1>
      <div className="card flex flex-wrap gap-3 items-end">
        <div><label className="label">De</label><input className="input" type="date" value={from} onChange={(e) => setFrom(e.target.value)} /></div>
        <div><label className="label">Até</label><input className="input" type="date" value={to} onChange={(e) => setTo(e.target.value)} /></div>
        <button className="btn" onClick={load}>Atualizar</button>
      </div>
      {error && <p className="text-sm text-red-600">{error}</p>}
      {rep && (
        <>
          <div className="grid grid-cols-3 gap-4">
            <div className="card text-center"><p className="text-2xl font-bold">{rep.totals.appointments}</p><p className="text-sm text-slate-500">Agendamentos</p></div>
            <div className="card text-center"><p className="text-2xl font-bold">{rep.totals.cancelled}</p><p className="text-sm text-slate-500">Cancelados</p></div>
            <div className="card text-center"><p className="text-2xl font-bold text-primary-600">{formatBRL(revenue)}</p><p className="text-sm text-slate-500">Receita</p></div>
          </div>
          <div className="grid md:grid-cols-2 gap-4">
            <div className="card">
              <h2 className="font-semibold mb-2">Por profissional</h2>
              <table className="grid"><thead><tr><th>Nome</th><th>Qtd</th><th>Receita</th></tr></thead>
              <tbody>{rep.per_professional.map((p) => <tr key={p.professional_id}><td>{p.name ?? p.professional_id}</td><td>{p.count}</td><td>{formatBRL(p.revenue_cents)}</td></tr>)}</tbody></table>
            </div>
            <div className="card">
              <h2 className="font-semibold mb-2">Por serviço</h2>
              <table className="grid"><thead><tr><th>Nome</th><th>Qtd</th><th>Receita</th></tr></thead>
              <tbody>{rep.per_service.map((s) => <tr key={s.service_id}><td>{s.name ?? s.service_id}</td><td>{s.count}</td><td>{formatBRL(s.revenue_cents)}</td></tr>)}</tbody></table>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
