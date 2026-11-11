import Link from "next/link";
import { ArrowLeft, Cloud } from "lucide-react";

export function AuthFrame({ title, copy, children }: { title: string; copy: string; children: React.ReactNode }) {
  return <main className="auth-screen"><Link href="/" className="auth-brand"><Cloud size={19}/> CLOUDX</Link><div className="auth-card"><div className="auth-eyebrow">CLOUDX / ACCOUNT</div><h1>{title}</h1><p>{copy}</p>{children}</div><Link href="/" className="auth-back"><ArrowLeft size={14}/> Back to CloudX</Link></main>;
}
