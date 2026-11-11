export const API_BASE = "";

export async function api<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = typeof window !== "undefined" ? localStorage.getItem("cloudx_token") : null;
  const response = await fetch(`${API_BASE}/api${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  });
  const payload = await response.json().catch(() => ({}));
  if (response.status === 401 && typeof window !== "undefined") {
    localStorage.removeItem("cloudx_token");
    localStorage.removeItem("cloudx_token_expires");
    window.location.assign("/login");
  }
  if (!response.ok) throw new Error(payload.error || `Request failed (${response.status})`);
  return payload as T;
}
