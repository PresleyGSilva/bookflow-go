"use client";
import { useEffect, useState } from "react";
import { apiFetch } from "../../../lib/api";
import { formatBRL } from "../../../lib/format";
import type { Service } from "../../../lib/types";

export default function ServicesPage() {
  const [items, setItems] = useState<Service[]>([]);
  const [name, setName] = useState("");
  const [duration, setDuration] = useState("30");
  const [price, setPrice] = useState("0");
  const [error, setError] = useState("");

  async function load() {
    try { setItems(await apiFetch<Service[]>("/api/v1/services")); }
    catch (e) { setError(e instanceof Error ? e.message : String(e)); }
  }
  useEffect(() => { load(); }, []);

  async function create(e: React.FormEvent) {
    e.preventDefault(); setError("");
    try {
      await apiFetch("/api/v1/services", { method: "POST", body: { name, duration_min: Number(duration), price_cents: Math.round(Number(price) * 100), active: true } });
      setName(""); setDuration("30"); setPrice("0"); load();
    } catch (e) { setError(e instanceof Error ? e.message : String(e)); }
  }

  async function remove(id: string) {
    try { await apiFetch(`/api/v1/services/${id}`, { method: "DELETE" }); load(); }
    catch (e) { setError(e instanceof Error ? e.message : String(e)); }
  }

  return (
    <div className="grid gap-4">
      <h1 className="text-2xl font-bold">Serviços</h1>
      {error && <p className="text-sm text-red-600">{error}</p>}
      <form onSubmit={create} className="card flex flex-wrap gap-3 items-end">
        <div><label className="label">Nome</label><input className="input" value={name} onChange={(e) => setName(e.target.value)} required /></div>
        <div><label className="label">Duração (min)</label><input className="input" type="number" value={duration} onChange={(e) => setDuration(e.target.value)} /></div>
        <div><label className="label">Preço (R$)</label><input className="input" type="number" step="0.01" value={price} onChange={(e) => setPrice(e.target.value)} /></div>
        <button className="btn">Adicionar</button>
      </form>
      <div className="card">
        <table className="grid">
          <thead><tr><th>Nome</th><th>Duração</th><th>Preço</th><th></th></tr></thead>
          <tbody>{items.map((s) => <tr key={s.id}><td>{s.name}</td><td>{s.duration_min} min</td><td>{formatBRL(s.price_cents)}</td><td><button className="text-red-600 text-xs underline" onClick={() => remove(s.id)}>excluir</button></td></tr>)}</tbody>
        </table>
      </div>
    </div>
  );
}
