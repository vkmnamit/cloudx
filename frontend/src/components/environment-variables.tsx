"use client";

import { FormEvent, useEffect, useState } from "react";
import { Eye, EyeOff, KeyRound, Plus, Save, ShieldCheck, Trash2 } from "lucide-react";
import { api } from "@/lib/api";

type Variable = { id?: string; key: string; value: string; is_secret: boolean };

export function EnvironmentVariables({ serviceId }: { serviceId: string }) {
  const [variables, setVariables] = useState<Variable[]>([]);
  const [visible, setVisible] = useState<Record<string, boolean>>({});
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    api<Variable[]>(`/services/${serviceId}/env`).then((items) => setVariables(items || []))
      .catch((cause) => setError(cause instanceof Error ? cause.message : "Could not load environment variables."));
  }, [serviceId]);

  function update(index: number, changes: Partial<Variable>) {
    setVariables((current) => current.map((item, itemIndex) => itemIndex === index ? { ...item, ...changes } : item));
  }

  async function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); setBusy(true); setError(""); setSaved(false);
    try {
      await api(`/services/${serviceId}/env`, { method: "PUT", body: JSON.stringify({ vars: variables.filter((item) => item.key.trim()).map(({ key, value, is_secret }) => ({ key: key.trim(), value, is_secret })) }) });
      setSaved(true); setTimeout(() => setSaved(false), 2200);
      setVariables(await api<Variable[]>(`/services/${serviceId}/env`) || []);
    } catch (cause) { setError(cause instanceof Error ? cause.message : "Could not save environment variables."); }
    finally { setBusy(false); }
  }

  return <section className="panel env-settings-panel">
    <div className="panel-heading"><div><span className="panel-eyebrow">RUNTIME CONFIGURATION</span><h2>Environment variables</h2></div><KeyRound size={16}/></div>
    <p className="env-settings-intro">Variables are available to this service at runtime. Secret values are encrypted and never shown again after saving.</p>
    <form onSubmit={save}>
      {variables.length ? <div className="env-variable-list service-env-list">{variables.map((variable, index) => <div className="env-variable-row" key={variable.id || index}>
        <input className="env-key-input" value={variable.key} onChange={(event) => update(index, { key: event.target.value })} placeholder="VARIABLE_NAME" aria-label="Variable name" autoComplete="off" spellCheck={false}/>
        <div className="env-value-wrap"><input value={variable.value} type={variable.is_secret && !visible[String(index)] ? "password" : "text"} onChange={(event) => update(index, { value: event.target.value })} placeholder={variable.is_secret ? "Saved secret · leave blank to keep" : "Value"} aria-label={`Value for ${variable.key || "environment variable"}`} autoComplete="new-password" spellCheck={false}/><button type="button" aria-label={visible[String(index)] ? "Hide value" : "Reveal value"} onClick={() => setVisible((current) => ({ ...current, [String(index)]: !current[String(index)] }))}>{visible[String(index)] ? <EyeOff size={14}/> : <Eye size={14}/>}</button></div>
        <label className="secret-toggle"><input type="checkbox" checked={variable.is_secret} onChange={(event) => update(index, { is_secret: event.target.checked })}/><ShieldCheck size={13}/> Secret</label>
        <button className="icon-button" type="button" aria-label="Remove variable" onClick={() => setVariables((current) => current.filter((_, itemIndex) => itemIndex !== index))}><Trash2 size={14}/></button>
      </div>)}</div> : <div className="env-empty"><ShieldCheck size={16}/><span><b>No variables configured</b><small>Add credentials and settings your application reads at runtime.</small></span></div>}
      <div className="env-settings-actions"><button type="button" className="secondary-button" onClick={() => setVariables((current) => [...current, { key: "", value: "", is_secret: true }])}><Plus size={13}/> Add variable</button><span>{saved && <><ShieldCheck size={13}/> Saved</>}</span><button className="primary-button" disabled={busy}><Save size={13}/>{busy ? "Saving…" : "Save variables"}</button></div>
      {error && <p className="form-error">{error}</p>}
    </form>
  </section>;
}
