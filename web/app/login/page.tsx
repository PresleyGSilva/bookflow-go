"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { apiFetch } from "../../lib/api";

type LoginResp = { access_token: string; user: { id: string; name: string; email: string }; orgs: { id: string; name: string }[] };

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError(""); setLoading(true);
    try {
      const data = await apiFetch<LoginResp>("/api/v1/auth/login", { method: "POST", body: { email, password } });
      localStorage.setItem("bookflow_token", data.access_token);
      localStorage.setItem("bookflow_user_id", data.user.id);
      if (data.orgs?.[0]) localStorage.setItem("bookflow_org_id", data.orgs[0].id);
      router.push("/admin");
    } catch (e) { setError(e instanceof Error ? e.message : String(e)); } finally { setLoading(false); }
  }

  return (
    <div className="max-w-md mx-auto card">
      <h1 className="text-xl font-bold mb-4">Entrar</h1>
      <form onSubmit={submit} className="grid gap-3">
        <div><label className="label">Email</label><input className="input" value={email} onChange={(e) => setEmail(e.target.value)} /></div>
        <div><label className="label">Senha</label><input className="input" type="password" value={password} onChange={(e) => setPassword(e.target.value)} /></div>
        {error && <p className="text-sm text-red-600">{error}</p>}
        <button className="btn" disabled={loading}>{loading ? "Entrando..." : "Entrar"}</button>
      </form>
    </div>
  );
}
