const BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8085";

export class ApiError extends Error {
  code: string;
  status: number;
  constructor(status: number, code: string, message: string) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

type Opts = {
  method?: string;
  body?: unknown;
  auth?: boolean;
  headers?: Record<string, string>;
};

function tenantHeaders(): Record<string, string> {
  if (typeof window === "undefined") return {};
  const h: Record<string, string> = {};
  const org = window.localStorage.getItem("bookflow_org_id");
  if (org) h["X-Org-ID"] = org;
  const uid = window.localStorage.getItem("bookflow_user_id");
  if (uid) h["X-User"] = uid;
  const tok = window.localStorage.getItem("bookflow_token");
  if (tok) h["Authorization"] = `Bearer ${tok}`;
  return h;
}

export async function apiFetch<T>(path: string, opts: Opts = {}): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method: opts.method ?? "GET",
    headers: {
      "Content-Type": "application/json",
      ...tenantHeaders(),
      ...(opts.headers ?? {}),
    },
    body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
  });
  const json = await res.json().catch(() => ({}));
  if (!res.ok) {
    const e = json?.error ?? {};
    throw new ApiError(res.status, e.code ?? "UNKNOWN", e.message ?? `HTTP ${res.status}`);
  }
  return (json?.data ?? json) as T;
}

export function publicPost<T>(path: string, body: unknown): Promise<T> {
  return apiFetch<T>(path, { method: "POST", body });
}

export const API_BASE = BASE;
