"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Activity, Box, ChevronDown, CircleHelp, Cloud, Command, Gauge, Layers3, LifeBuoy, LogOut, Plus, Settings2, ShieldCheck, Github, GitBranch } from "lucide-react";
import { useEffect, useState } from "react";

const nav = [
  { label: "Projects", href: "/projects", icon: Layers3 },
  { label: "Services", href: "/services", icon: Box },
  { label: "Deployments", href: "/deployments", icon: Activity },
];

export function Sidebar() {
  const pathname = usePathname();
  const [email, setEmail] = useState("your workspace");
  useEffect(() => { setEmail(localStorage.getItem("cloudx_email") || "your workspace"); }, []);
  return <aside className="sidebar">
    <Link className="sidebar-brand" href="/"><span className="brand-icon"><Cloud size={19} strokeWidth={1.7}/></span><span>CLOUDX</span><span className="brand-tag">CLOUD</span></Link>
    <button className="workspace-picker"><span className="workspace-avatar">{email.slice(0,1).toUpperCase()}</span><span className="workspace-label"><b>{email.split("@")[0]}</b><small>Free workspace</small></span><ChevronDown size={14}/></button>
    <div className="side-label">WORKSPACE <button aria-label="Add project" onClick={() => window.location.assign("/projects")}><Plus size={13}/></button></div>
    <Link className={`side-overview ${pathname === "/dashboard" ? "active" : ""}`} href="/dashboard"><Gauge size={16}/>Overview</Link>
    <nav className="side-nav">{nav.map(({ label, href, icon: Icon }) => <Link className={pathname === href || pathname.startsWith(`${href}/`) ? "active" : ""} href={href} key={href}><Icon size={16}/>{label}</Link>)}</nav>
    <div className="side-label environments-label">SOURCE</div>
    <Link className={`environment-link ${pathname === "/services" ? "source-active" : ""}`} href="/services"><Github size={14}/>GitHub repositories</Link>
    <Link className="environment-link" href="/projects"><GitBranch size={14}/>Environments</Link>
    <Link className={`environment-link ${pathname === "/deploy" ? "source-active" : ""}`} href="/deploy"><Plus size={14}/>New deployment</Link>
    <div className="sidebar-bottom"><div className="side-label">WORKSPACE</div><Link href="/settings"><Settings2 size={16}/>Settings</Link><Link href="/docs"><CircleHelp size={16}/>Documentation<ExternalArrow/></Link><button className="account-row" onClick={() => { localStorage.removeItem("cloudx_token"); localStorage.removeItem("cloudx_email"); localStorage.removeItem("cloudx_token_expires"); window.location.href = "/login"; }}><span className="account-avatar">{email.slice(0,1).toUpperCase()}</span><span><b>{email.split("@")[0]}</b><small>{email}</small></span><LogOut size={14}/></button></div>
    <div className="sidebar-shortcut"><Command size={13}/> Quick search <kbd>⌘ K</kbd></div>
  </aside>;
}

function ExternalArrow() { return <span className="external-arrow">↗</span>; }

export function AppShell({ children, title, description, action }: { children: React.ReactNode; title: string; description: string; action?: React.ReactNode }) {
  return <div className="app-frame"><Sidebar/><main className="app-main"><header className="topbar"><div className="breadcrumbs"><span>Workspace</span><b>/</b><strong>{title}</strong></div><div className="topbar-right"><button className="console-search"><span>⌕</span> Search <kbd>⌘ K</kbd></button><Link className="top-new" href="/deploy"><Plus size={15}/> New deploy</Link><Link className="help-button" href="/docs"><LifeBuoy size={15}/> Help</Link><span className="system-status"><i/> Operational</span><button className="top-avatar" aria-label="Account">C</button></div></header><div className="page-content"><div className="page-heading"><div><div className="page-eyebrow"><span className="eyebrow-square"/> CLOUDX <span>/</span> WORKSPACE</div><h1>{title}</h1><p>{description}</p></div>{action && <div className="heading-action">{action}</div>}</div>{children}</div><footer className="app-footer"><span><ShieldCheck size={13}/> CloudX infrastructure</span><span>STATUS <i/> OPERATIONAL</span><span>MADE FOR DEVELOPERS</span></footer></main></div>;
}

export function IconGlyph({ icon: Icon }: { icon: typeof Cloud }) { return <Icon size={17} strokeWidth={1.7}/>; }
