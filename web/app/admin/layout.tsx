import Link from "next/link";

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const links = [["/admin", "Calendário"], ["/admin/clients", "Clientes"], ["/admin/services", "Serviços"], ["/admin/pros", "Profissionais"], ["/admin/reports", "Relatórios"], ["/admin/settings", "Config"]];
  return (
    <div className="grid md:grid-cols-[180px_1fr] gap-6">
      <aside className="card !p-3 h-fit flex md:flex-col flex-row flex-wrap gap-1">
        {links.map(([href, label]) => <Link key={href} className="navlink" href={href}>{label}</Link>)}
      </aside>
      <div>{children}</div>
    </div>
  );
}
