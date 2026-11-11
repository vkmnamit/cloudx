"use client";
import Link from "next/link";
import { FormEvent, useState } from "react";
import { ArrowUpRight } from "lucide-react";
import { api } from "@/lib/api";
import { AuthFrame } from "@/components/auth-frame";

export default function SignupPage() {
  const [error, setError] = useState(""); const [busy,setBusy]=useState(false);
  async function submit(event:FormEvent<HTMLFormElement>){event.preventDefault();setError("");setBusy(true);const form=new FormData(event.currentTarget);const email=String(form.get("email"));try{await api("/signup",{method:"POST",body:JSON.stringify({name:form.get("name"),email,password:form.get("password")})});const result=await api<{access_token?:string;token?:string}>("/login",{method:"POST",body:JSON.stringify({email,password:form.get("password")})});const token=result.access_token||result.token;if(!token)throw new Error("Account created. Please sign in to continue.");localStorage.setItem("cloudx_token",token);localStorage.setItem("cloudx_email",email);localStorage.setItem("cloudx_token_expires",String(Date.now()+15*60*1000));window.location.href="/dashboard";}catch(e){setError(e instanceof Error?e.message:"Unable to create account.");setBusy(false);}}
  return <AuthFrame title="Start shipping." copy="Create your workspace and launch your first project."><form className="auth-form" onSubmit={submit}><label>Your name<input name="name" autoComplete="name" placeholder="Ada Lovelace" required/></label><label>Email address<input name="email" type="email" autoComplete="email" placeholder="you@company.com" required/></label><label>Password<input name="password" type="password" autoComplete="new-password" minLength={8} placeholder="At least 8 characters" required/></label>{error&&<p className="form-error">{error}</p>}<button className="button button-light auth-submit" disabled={busy}>{busy?"Creating account…":"Create account"}<ArrowUpRight size={15}/></button><p className="auth-switch">Already have an account? <Link href="/login">Sign in</Link></p></form></AuthFrame>;
}
