"use client";

import { useEffect, useState } from "react";
import { ExternalLink, Github, ShieldCheck } from "lucide-react";
import { api } from "@/lib/api";

type Org = { id: string; name: string };
type Installation = { installation_id: number; account_name: string; organization_id: string };

export function GitHubConnection() {
  const [organizations, setOrganizations] = useState<Org[]>([]);
  const [installations, setInstallations] = useState<Installation[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  async function load() {
    try { const [orgs, connected] = await Promise.all([api<Org[]>("/organizations"), api<Installation[]>("/github/installations")]); setOrganizations(orgs || []); setInstallations(connected || []); }
    catch (cause) { setError(cause instanceof Error ? cause.message : "Could not load GitHub connection."); }
  }
  useEffect(() => { void load(); }, []);
  async function connect(organizationID: string) {
    setBusy(true); setError("");
    try { const result = await api<{ installation_url: string }>(`/github/connect?organization_id=${encodeURIComponent(organizationID)}`); window.location.assign(result.installation_url); }
    catch (cause) { setError(cause instanceof Error ? cause.message : "Could not connect GitHub."); setBusy(false); }
  }
  return <section className="panel settings-panel github-settings-panel">
    <div className="settings-section-heading"><span className="settings-icon"><Github size={16}/></span><div><h2>GitHub</h2><p>Install CloudX on an account or organization to browse private repositories.</p></div></div>
    {installations.length ? <div className="github-installation-list">{installations.map((item) => <div className="github-installation-row" key={item.installation_id}><span><i/><b>{item.account_name}</b><small>{organizations.find((org) => org.id === item.organization_id)?.name || "CloudX workspace"}</small></span><ShieldCheck size={15}/></div>)}</div> : <p className="github-setup-hint">No GitHub App is connected yet. The CloudX server administrator must configure the GitHub App slug, app ID, private key, and callback URL before installation is available.</p>}
    {organizations.length > 0 && <div className="github-connect-actions">{organizations.map((org) => <button className="secondary-button" key={org.id} onClick={() => void connect(org.id)} disabled={busy}><Github size={13}/>{busy ? "Opening GitHub…" : `Connect ${org.name}`}<ExternalLink size={12}/></button>)}</div>}
    {error && <p className="form-error">{error}</p>}
  </section>;
}
