"use client";
import { useEffect, useState } from "react";
import { apiFetch } from "../../../lib/api";
import type { Customer } from "../../../lib/types";

export default function ClientsPage() {
  const [items, setItems] = useState<Customer[]>([]);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [error, setError] = useState("");

  async function load() {
    try { setItems(await apiFetch<Customer[]>("/api/v1/customers")); }
    catch (e) { setError(e instanceof Error ? e.message : String(e)); }
  }
  useEffect(() => { load(); }, []);

  async function create(e: React.FormEvent) {
    e.preventDefault(); setError("");
    try {
      await apiFetch("/api/v1/customers", { method: "POST", body: { name, email, phone } });
      setName(""); setEmail(""); setPhone(""); load();
    } catch (e) { setError(e instanceof Error ? e.message : String(e)); }
  }

  return (
    <div className="grid gap-4">
      <h1 className="text-2xl font-bold">Clientes</h1>
      {error && <p className="text-sm text-red-600">{error}</p>}
      <form onSubmit={create} className="card flex flex-wrap gap-3 items-end">
        <div><label className="label">Nome</label><input className="input" value={name} onChange={(e) => setName(e.target.value)} required /></div>
        <div><label className="label">Email</label><input className="input" value={email} onChange={(e) => setEmail(e.target.value)} /></div>
        <div><label className="label">Telefone</label><input className="input" value={phone} onChange={(e) => setPhone(e.target.value)} /></div>
        <button className="btn">Adicionar</button>
      </form>
      <div className="card">
        <table className="grid">
          <thead><tr><th>Nome</th><th>Email</th><th>Telefone</th></tr></thead>
          <tbody>{items.map((c) => <tr key={c.id}><td>{c.name}</td><td>{c.email ?? "—"}</td><td>{c.phone ?? "—"}</td></tr>)}</tbody>
        </table>
      </div>
    </div>
  );
}
