"use client";
import { useEffect, useState } from "react";
import { apiFetch } from "../../../lib/api";
import type { Professional } from "../../../lib/types";

type WH = { id: string; professional_id: string; weekday: number; start_min: number; end_min: number };
type Block = { id: string; professional_id: string; starts_at: string; ends_at: string; reason?: string };
const DAYS = ["Dom", "Seg", "Ter", "Qua", "Qui", "Sex", "Sáb"];

export default function ProsPage() {
  const [items, setItems] = useState<Professional[]>([]);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [selected, setSelected] = useState("");
  const [hours, setHours] = useState<WH[]>([]);
  const [blocks, setBlocks] = useState<Block[]>([]);
  const [weekday, setWeekday] = useState("1");
  const [start, setStart] = useState("480");
  const [end, setEnd] = useState("1080");
  const [blockStart, setBlockStart] = useState("");
  const [blockEnd, setBlockEnd] = useState("");
  const [reason, setReason] = useState("");
  const [error, setError] = useState("");

  async function load() {
    try { setItems(await apiFetch<Professional[]>("/api/v1/professionals")); }
    catch (e) { setError(e instanceof Error ? e.message : String(e)); }
  }
  useEffect(() => { load(); }, []);

  async function loadDetail(id: string) {
    setSelected(id);
    try {
      const [h, b] = await Promise.all([
        apiFetch<WH[]>(`/api/v1/working-hours?professional_id=${id}`),
        apiFetch<Block[]>(`/api/v1/blocks?professional_id=${id}`),
      ]);
      setHours(h); setBlocks(b);
    } catch (e) { setError(e instanceof Error ? e.message : String(e)); }
  }

  async function create(e: React.FormEvent) {
    e.preventDefault(); setError("");
    try {
      await apiFetch("/api/v1/professionals", { method: "POST", body: { name, email, phone, timezone: "America/Sao_Paulo", active: true } });
      setName(""); setEmail(""); setPhone(""); load();
    } catch (e) { setError(e instanceof Error ? e.message : String(e)); }
  }

  async function addHours(e: React.FormEvent) {
    e.preventDefault(); setError("");
    try {
      await apiFetch("/api/v1/working-hours", { method: "POST", body: { professional_id: selected, weekday: Number(weekday), start_min: Number(start), end_min: Number(end) } });
      loadDetail(selected);
    } catch (e) { setError(e instanceof Error ? e.message : String(e)); }
  }

  async function addBlock(e: React.FormEvent) {
    e.preventDefault(); setError("");
    try {
      await apiFetch("/api/v1/blocks", { method: "POST", body: { professional_id: selected, starts_at: blockStart, end_at: blockEnd, reason } });
      setBlockStart(""); setBlockEnd(""); setReason(""); loadDetail(selected);
    } catch (e) { setError(e instanceof Error ? e.message : String(e)); }
  }

  return (
    <div className="grid gap-4">
      <h1 className="text-2xl font-bold">Profissionais</h1>
      {error && <p className="text-sm text-red-600">{error}</p>}
      <form onSubmit={create} className="card flex flex-wrap gap-3 items-end">
        <div><label className="label">Nome</label><input className="input" value={name} onChange={(e) => setName(e.target.value)} required /></div>
        <div><label className="label">Email</label><input className="input" value={email} onChange={(e) => setEmail(e.target.value)} /></div>
        <div><label className="label">Telefone</label><input className="input" value={phone} onChange={(e) => setPhone(e.target.value)} /></div>
        <button className="btn">Adicionar</button>
      </form>
      <div className="card">
        <table className="grid">
          <thead><tr><th>Nome</th><th>Email</th><th></th></tr></thead>
          <tbody>{items.map((p) => <tr key={p.id}><td>{p.name}</td><td>{p.email ?? "—"}</td><td><button className="text-primary-600 text-xs underline" onClick={() => loadDetail(p.id)}>horários / bloqueios</button></td></tr>)}</tbody>
        </table>
      </div>
      {selected && (
        <div className="grid md:grid-cols-2 gap-4">
          <div className="card grid gap-3">
            <h2 className="font-semibold">Horários de trabalho</h2>
            <ul className="text-sm grid gap-1">
              {hours.map((h) => (
                <li key={h.id} className="flex justify-between border-b border-slate-100 py-1">
                  <span>{DAYS[h.weekday]} — {Math.floor(h.start_min / 60)}:{String(h.start_min % 60).padStart(2, "0")} às {Math.floor(h.end_min / 60)}:{String(h.end_min % 60).padStart(2, "0")}</span>
                  <button className="text-red-600 text-xs underline" onClick={() => apiFetch(`/api/v1/working-hours/${h.id}`, { method: "DELETE" }).then(() => loadDetail(selected)).catch((e) => setError(e instanceof Error ? e.message : String(e)))}>excluir</button>
                </li>
              ))}
            </ul>
            <form onSubmit={addHours} className="flex flex-wrap gap-2 items-end">
              <div><label className="label">Dia</label><select className="input" value={weekday} onChange={(e) => setWeekday(e.target.value)}>{DAYS.map((d, i) => <option key={d} value={i}>{d}</option>)}</select></div>
              <div><label className="label">Início (min)</label><input className="input" type="number" value={start} onChange={(e) => setStart(e.target.value)} /></div>
              <div><label className="label">Fim (min)</label><input className="input" type="number" value={end} onChange={(e) => setEnd(e.target.value)} /></div>
              <button className="btn-ghost">Adicionar</button>
            </form>
          </div>
          <div className="card grid gap-3">
            <h2 className="font-semibold">Bloqueios</h2>
            <ul className="text-sm grid gap-1">
              {blocks.map((b) => (
                <li key={b.id} className="flex justify-between border-b border-slate-100 py-1">
                  <span>{new Date(b.starts_at).toLocaleString("pt-BR")} → {new Date(b.ends_at).toLocaleString("pt-BR")} {b.reason ? `(${b.reason})` : ""}</span>
                  <button className="text-red-600 text-xs underline" onClick={() => apiFetch(`/api/v1/blocks/${b.id}`, { method: "DELETE" }).then(() => loadDetail(selected)).catch((e) => setError(e instanceof Error ? e.message : String(e)))}>excluir</button>
                </li>
              ))}
            </ul>
            <form onSubmit={addBlock} className="grid gap-2">
              <div className="grid grid-cols-2 gap-2">
                <div><label className="label">Início</label><input className="input" type="datetime-local" value={blockStart} onChange={(e) => setBlockStart(e.target.value)} /></div>
                <div><label className="label">Fim</label><input className="input" type="datetime-local" value={blockEnd} onChange={(e) => setBlockEnd(e.target.value)} /></div>
              </div>
              <div><label className="label">Motivo</label><input className="input" value={reason} onChange={(e) => setReason(e.target.value)} /></div>
              <button className="btn-ghost">Bloquear</button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
