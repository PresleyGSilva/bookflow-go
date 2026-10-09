import type { Metadata } from "next";
import Link from "next/link";
import "./globals.css";

export const metadata: Metadata = { title: "BookFlow", description: "Agendamentos simples" };

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="pt-BR">
      <body>
        <header className="bg-white border-b border-slate-200">
          <div className="max-w-6xl mx-auto px-4 py-3 flex items-center gap-4">
            <Link href="/" className="font-bold text-lg text-primary-600">BookFlow</Link>
            <nav className="flex gap-1">
              <Link className="navlink" href="/book">Agendar</Link>
              <Link className="navlink" href="/admin">Admin</Link>
              <Link className="navlink" href="/login">Entrar</Link>
            </nav>
          </div>
        </header>
        <main className="max-w-6xl mx-auto px-4 py-6">{children}</main>
      </body>
    </html>
  );
}
