"use client";
import { useEffect, useState } from "react";
import Link from "next/link";
import { apiFetch } from "../../lib/api";
import type { Service, Professional, Availability, Appointment } from "../../lib/types";

export default function BookPage() {
  const [orgs, setOrgs] = useState<{ id: string; name: string }[]>([]);
  const [orgId, setOrgId] = useState("");
  const [services, setServices] = useState<Service[]>([]);
  const [pros, setPros] = useState<Professional[]>([]);
  const [serviceId, setServiceId] = useState("");
  const [proId, setProId] = useState("");
  const [date, setDate] = useState(new Date().toISOString().slice(0, 10));
  const [slots, setSlots] = useState<string[]>([]);
  const [timezone, setTimezone] = useState("");
  const [time, setTime] = useState("");
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [done, setDone] = useState<Appointment | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    (async () => {
      try {
        const o = await apiFetch<{ id: string; name: string }[]>("/api/v1/public/orgs");
        setOrgs(o);
        if (o.length > 0) setOrgId(o[0].id);
      } catch (e) { setError(e instanceof Error ? e.message : String(e)); }
    })();
  }, []);

  useEffect(() => {
    if (!orgId) return;
    (async () => {
      try {
        const [s, p] = await Promise.all([
          apiFetch<Service[]>(`/api/v1/public/services?org_id=${orgId}`),
          apiFetch<Professional[]>(`/api/v1/public/professionals?org_id=${orgId}`),
        ]);
        setServices(s.filter((x) => x.active !== false));
        setPros(p.filter((x) => x.active !== false));
      } catch (e) { setError(e instanceof Error ? e.message : String(e)); }
    })();
  }, [orgId]);

  async function loadSlots() {
    setError("");
    if (!proId || !serviceId || !date) { setError("Escolha serviço, profissional e data."); return; }
    setLoading(true);
    try {
      const a = await apiFetch<Availability>(`/api/v1/public/availability?professional_id=${proId}&service_id=${serviceId}&date=${date}`);
      setSlots(a.slots);
      setTimezone(a.timezone);
      if (a.slots.length === 0) setError("Nenhum horário livre neste dia.");
    } catch (e) { setError(e instanceof Error ? e.message : String(e)); } finally { setLoading(false); }
  }

  async function confirm(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    if (!time) { setError("Escolha um horário."); return; }
    setLoading(true);
    try {
      const ap = await apiFetch<Appointment>("/api/v1/bookings/public", {
        method: "POST",
        body: { service_id: serviceId, professional_id: proId, date, time, customer: { name, email, phone } },
      });
      setDone(ap);
    } catch (e) { setError(e instanceof Error ? e.message : String(e)); } finally { setLoading(false); }
  }

  if (done) {
    return (
      <div className="max-w-md mx-auto card text-center">
        <h1 className="text-xl font-bold mb-2">Reserva confirmada!</h1>
        <p className="text-3xl font-mono font-bold text-primary-600 my-4">{done.code}</p>
        <p className="text-sm text-slate-600 mb-4">Guarde este código para reagendar ou cancelar.</p>
        <Link href={`/book/${done.code}`} className="btn">Gerenciar reserva</Link>
      </div>
    );
  }

  return (
    <div className="max-w-xl mx-auto grid gap-4">
      <h1 className="text-2xl font-bold">Agendar horário</h1>
      {orgs.length > 1 && (
        <div><label className="label">Estabelecimento</label>
          <select className="input" value={orgId} onChange={(e) => setOrgId(e.target.value)}>
            {orgs.map((o) => <option key={o.id} value={o.id}>{o.name}</option>)}
          </select></div>
      )}
      <div className="card grid gap-3">
        <h2 className="font-semibold">1. Serviço e profissional</h2>
        {services.length > 0 ? (
          <div><label className="label">Serviço</label>
            <select className="input" value={serviceId} onChange={(e) => setServiceId(e.target.value)}>
              <option value="">Selecione...</option>
              {services.map((s) => <option key={s.id} value={s.id}>{s.name} — {(s.price_cents / 100).toLocaleString("pt-BR", { style: "currency", currency: "BRL" })}</option>)}
            </select></div>
        ) : (
          <div><label className="label">ID do serviço</label><input className="input" value={serviceId} onChange={(e) => setServiceId(e.target.value)} placeholder="service_id" /></div>
        )}
        {pros.length > 0 ? (
          <div><label className="label">Profissional</label>
            <select className="input" value={proId} onChange={(e) => setProId(e.target.value)}>
              <option value="">Selecione...</option>
              {pros.map((p) => <option key={p.id} value={p.id}>{p.name}</option>)}
            </select></div>
        ) : (
          <div><label className="label">ID do profissional</label><input className="input" value={proId} onChange={(e) => setProId(e.target.value)} placeholder="professional_id" /></div>
        )}
      </div>
      <div className="card grid gap-3">
        <h2 className="font-semibold">2. Data e hora {timezone && <span className="text-xs font-normal text-slate-500">({timezone})</span>}</h2>
        <div><label className="label">Data</label><input className="input" type="date" value={date} onChange={(e) => setDate(e.target.value)} /></div>
        <button className="btn-ghost" onClick={loadSlots} disabled={loading}>Ver horários</button>
        {slots.length > 0 && (
          <div className="flex flex-wrap gap-2">
            {slots.map((s) => (
              <button key={s} onClick={() => setTime(s)} className={`px-3 py-1.5 rounded-lg border text-sm ${time === s ? "bg-primary-500 text-white border-primary-500" : "border-slate-300 hover:bg-slate-100"}`}>{s}</button>
            ))}
          </div>
        )}
      </div>
      <form onSubmit={confirm} className="card grid gap-3">
        <h2 className="font-semibold">3. Seus dados</h2>
        <div><label className="label">Nome</label><input className="input" value={name} onChange={(e) => setName(e.target.value)} required /></div>
        <div className="grid grid-cols-2 gap-3">
          <div><label className="label">Email</label><input className="input" value={email} onChange={(e) => setEmail(e.target.value)} /></div>
          <div><label className="label">Telefone</label><input className="input" value={phone} onChange={(e) => setPhone(e.target.value)} /></div>
        </div>
        {error && <p className="text-sm text-red-600">{error}</p>}
        <button className="btn" disabled={loading}>{loading ? "Confirmando..." : `Confirmar ${time ? `às ${time}` : "reserva"}`}</button>
      </form>
    </div>
  );
}
