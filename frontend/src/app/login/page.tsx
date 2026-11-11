"use client";
import Link from "next/link";
import { FormEvent, useState } from "react";
import { ArrowUpRight } from "lucide-react";
import { api } from "@/lib/api";
import { AuthFrame } from "@/components/auth-frame";

export default function LoginPage() {
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); setError(""); setBusy(true);
    const data = new FormData(event.currentTarget);
    try {
      const result = await api<{access_token?: string; token?: string}>("/login", { method: "POST", body: JSON.stringify({ email: data.get("email"), password: data.get("password") }) });
      const token = result.access_token || result.token;
      if (!token) throw new Error("The API did not return an access token.");
      localStorage.setItem("cloudx_token", token); localStorage.setItem("cloudx_email", String(data.get("email")));
      localStorage.setItem("cloudx_token_expires", String(Date.now() + 15 * 60 * 1000));
      window.location.href = "/dashboard";
    } catch (e) { setError(e instanceof Error ? e.message : "Unable to sign in."); setBusy(false); }
  }
  return <AuthFrame title="Welcome back." copy="Your deployments are right where you left them."><form className="auth-form" onSubmit={submit}><label>Email address<input name="email" type="email" autoComplete="email" placeholder="you@company.com" required/></label><label>Password<input name="password" type="password" autoComplete="current-password" placeholder="Enter your password" required/></label>{error && <p className="form-error">{error}</p>}<button className="button button-light auth-submit" disabled={busy}>{busy ? "Signing in…" : "Sign in"}<ArrowUpRight size={15}/></button><p className="auth-switch">New to CloudX? <Link href="/signup">Create an account</Link></p></form></AuthFrame>;
}
