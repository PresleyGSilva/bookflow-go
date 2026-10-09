import Link from "next/link";

export default function Home() {
  return (
    <div className="grid gap-6">
      <div className="card text-center py-12">
        <h1 className="text-3xl font-bold mb-2">BookFlow — agendamentos sem fricção</h1>
        <p className="text-slate-600 mb-6">Escolha o serviço, o profissional e o horário. Receba um código para gerenciar sua reserva.</p>
        <div className="flex gap-3 justify-center">
          <Link href="/book" className="btn">Agendar horário</Link>
          <Link href="/admin" className="btn-ghost">Painel admin</Link>
        </div>
      </div>
      <div className="grid md:grid-cols-3 gap-4">
        {[["1. Escolha", "Serviço e profissional de sua preferência."],["2. Reserve", "Selecione data e hora disponíveis."],["3. Gerencie", "Use o código BF-XXXXXX para reagendar ou cancelar."]].map(([t, d]) => (
          <div key={t} className="card"><h3 className="font-semibold mb-1">{t}</h3><p className="text-sm text-slate-600">{d}</p></div>
        ))}
      </div>
    </div>
  );
}
