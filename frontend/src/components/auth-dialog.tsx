"use client";

import { FormEvent, useEffect, useState } from "react";
import { ArrowUpRight, Check, Cloud, X } from "lucide-react";
import { api } from "@/lib/api";

type Mode = "login" | "signup";

export function AuthTrigger({ mode, className, children }: { mode: Mode; className?: string; children: React.ReactNode }) {
  return <button type="button" className={className} onClick={() => window.dispatchEvent(new CustomEvent("cloudx:auth", { detail: mode }))}>{children}</button>;
}

export function AuthDialog() {
  const [mode, setMode] = useState<Mode>("signup");
  const [open, setOpen] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    const onOpen = (event: Event) => {
      const requestedMode = (event as CustomEvent<Mode>).detail;
      setMode(requestedMode === "login" ? "login" : "signup");
      setError("");
      setOpen(true);
    };
    const onKeyDown = (event: KeyboardEvent) => { if (event.key === "Escape") setOpen(false); };
    window.addEventListener("cloudx:auth", onOpen);
    window.addEventListener("keydown", onKeyDown);
    return () => { window.removeEventListener("cloudx:auth", onOpen); window.removeEventListener("keydown", onKeyDown); };
  }, []);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    setError("");
    const form = new FormData(event.currentTarget);
    const email = String(form.get("email"));
    const password = String(form.get("password"));
    try {
      if (mode === "signup") await api("/signup", { method: "POST", body: JSON.stringify({ name: form.get("name"), email, password }) });
      const result = await api<{ access_token?: string; token?: string }>("/login", { method: "POST", body: JSON.stringify({ email, password }) });
      const token = result.access_token || result.token;
      if (!token) throw new Error("CloudX did not return a sign-in token.");
      localStorage.setItem("cloudx_token", token);
      localStorage.setItem("cloudx_email", email);
      localStorage.setItem("cloudx_token_expires", String(Date.now() + 15 * 60 * 1000));
      window.location.assign("/dashboard");
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Unable to continue. Please try again.");
      setBusy(false);
    }
  }

  if (!open) return null;

  return <div className="auth-modal-backdrop" onMouseDown={(event) => { if (event.target === event.currentTarget) setOpen(false); }}>
    <section className="auth-modal" role="dialog" aria-modal="true" aria-labelledby="auth-modal-title">
      <button className="auth-modal-close" type="button" aria-label="Close dialog" onClick={() => setOpen(false)}><X size={17}/></button>
      <div className="auth-modal-visual"><div className="auth-modal-grid"/><span className="auth-modal-brand"><Cloud size={19}/> CLOUDX</span><div className="auth-phone-mark"><span><Cloud size={29}/></span><i/><i/><i/></div><div className="auth-visual-caption"><b>YOUR NEXT DEPLOY</b><span>IS CLOSER THAN YOU THINK.</span></div><div className="auth-visual-foot"><span><i/> READY WHEN YOU ARE</span><span>01 / 03</span></div></div>
      <div className="auth-modal-content"><div className="auth-modal-kicker"><span/> CLOUDX / ACCOUNT</div><h2 id="auth-modal-title">{mode === "login" ? "Welcome back." : "Let’s get building."}</h2><p className="auth-modal-subtitle">{mode === "login" ? "Sign in to pick up where you left off." : "Create your account and ship your first project."}</p>
        <div className="auth-tabs" role="tablist" aria-label="Choose account action"><button type="button" role="tab" aria-selected={mode === "signup"} className={mode === "signup" ? "selected" : ""} onClick={() => { setMode("signup"); setError(""); }}>Create account</button><button type="button" role="tab" aria-selected={mode === "login"} className={mode === "login" ? "selected" : ""} onClick={() => { setMode("login"); setError(""); }}>Sign in</button></div>
        <form className="auth-modal-form" onSubmit={submit} key={mode}>
          {mode === "signup" && <label>YOUR NAME<input name="name" autoComplete="name" placeholder="Your name" required/></label>}
          <label>EMAIL ADDRESS<input name="email" type="email" autoComplete="email" placeholder="you@company.com" required/></label>
          <label>PASSWORD<input name="password" type="password" autoComplete={mode === "login" ? "current-password" : "new-password"} minLength={mode === "signup" ? 8 : undefined} placeholder={mode === "login" ? "Your password" : "At least 8 characters"} required/></label>
          {mode === "login" && <div className="auth-remember"><span><i/> Your workspace is waiting</span><Check size={13}/></div>}
          {error && <p className="form-error">{error}</p>}
          <button className="auth-modal-submit" disabled={busy}>{busy ? "One moment…" : mode === "login" ? "Sign in to CloudX" : "Create your account"}<ArrowUpRight size={15}/></button>
        </form>
        <p className="auth-legal">By continuing, you agree to use CloudX to build great things.</p>
      </div>
    </section>
  </div>;
}
