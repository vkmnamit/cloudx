"use client";

import Link from "next/link";
import { FormEvent, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import {
  ArrowLeft, ArrowRight, Box, Check, ChevronRight, Eye, EyeOff, GitBranch,
  Github, Globe2, LoaderCircle, Plus, Rocket, ShieldCheck, Trash2,
} from "lucide-react";
import { AppShell } from "@/components/sidebar";
import { api } from "@/lib/api";

type Project = { id: string; name: string; slug: string; organization_id: string };
type Org = { id: string; name: string };
type Environment = { id: string; project_id: string; name: string };
type Service = { id: string; name: string };
type Deployment = { id: string; status: string };
type GitHubRepo = {
  id: number; full_name: string; html_url: string; clone_url: string;
  default_branch: string; description: string | null; private: boolean;
};
type GitHubInstallation = { installation_id: number; account_name: string; organization_id: string };
type Variable = { id: string; key: string; value: string; secret: boolean; reveal: boolean };

const steps = ["SOURCE", "CONFIGURE", "REVIEW"];

export default function DeployPage() {
  const router = useRouter();
  const [step, setStep] = useState(0);
  const [username, setUsername] = useState("");
  const [repos, setRepos] = useState<GitHubRepo[]>([]);
  const [repoSearch, setRepoSearch] = useState("");
  const [selectedRepo, setSelectedRepo] = useState<GitHubRepo | null>(null);
  const [repoBusy, setRepoBusy] = useState(false);
  const [repoError, setRepoError] = useState("");
  const [installations, setInstallations] = useState<GitHubInstallation[]>([]);
  const [installationBusy, setInstallationBusy] = useState(false);
  const [serviceName, setServiceName] = useState("");
  const [branch, setBranch] = useState("");
  const [runtime, setRuntime] = useState("Node.js");
  const [port, setPort] = useState("3000");
  const [projects, setProjects] = useState<Project[]>([]);
  const [orgs, setOrgs] = useState<Org[]>([]);
  const [environments, setEnvironments] = useState<Environment[]>([]);
  const [projectChoice, setProjectChoice] = useState("");
  const [newProjectName, setNewProjectName] = useState("");
  const [variables, setVariables] = useState<Variable[]>([]);
  const [pasteOpen, setPasteOpen] = useState(false);
  const [pasteValue, setPasteValue] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!localStorage.getItem("cloudx_token")) {
      location.href = "/login";
      return;
    }
    Promise.all([
      api<Project[]>("/projects"), api<Org[]>("/organizations"), api<Environment[]>("/environments"),
    ]).then(([projectData, orgData, environmentData]) => {
      setProjects(projectData || []);
      setOrgs(orgData || []);
      setEnvironments(environmentData || []);
      if (projectData?.length) setProjectChoice(projectData[0].id);
    }).catch((cause) => setError(cause instanceof Error ? cause.message : "Could not load your workspace."));
    api<GitHubInstallation[]>("/github/installations").then(async (connected) => {
      setInstallations(connected || []);
      if (connected?.length) {
        setRepoBusy(true);
        try {
          const response = await api<GitHubRepo[]>(`/github/repositories?installation_id=${connected[0].installation_id}`);
          setRepos(response || []);
        } catch (cause) {
          setRepoError(cause instanceof Error ? cause.message : "Could not load connected repositories.");
        } finally {
          setRepoBusy(false);
        }
      }
    }).catch(() => undefined);
  }, []);

  const filteredRepos = useMemo(() => repos.filter((repo) =>
    repo.full_name.toLowerCase().includes(repoSearch.toLowerCase())), [repos, repoSearch]);
  const selectedProject = projects.find((project) => project.id === projectChoice);
  const selectedEnvironment = environments.find((environment) => environment.project_id === projectChoice);
  const configuredVariables = variables.filter((variable) => variable.key.trim() || variable.value.trim());

  async function findRepos() {
    const account = username.trim();
    if (!account) return;
    setRepoBusy(true);
    setRepoError("");
    setRepos([]);
    setSelectedRepo(null);
    try {
      const response = await fetch(`https://api.github.com/users/${encodeURIComponent(account)}/repos?sort=updated&per_page=100`, {
        headers: { Accept: "application/vnd.github+json" },
      });
      const payload = await response.json();
      if (!response.ok) throw new Error(payload.message || "Could not load public repositories.");
      const publicRepos = (payload as GitHubRepo[]).filter((repo) => !repo.private);
      setRepos(publicRepos);
      if (!publicRepos.length) setRepoError("No public repositories found for that account.");
    } catch (cause) {
      setRepoError(cause instanceof Error ? cause.message : "GitHub lookup failed.");
    } finally {
      setRepoBusy(false);
    }
  }

  async function connectGitHub() {
    const orgID = selectedProject?.organization_id || orgs[0]?.id;
    if (!orgID) {
      setError("Create a project first so CloudX knows which workspace should receive the GitHub connection.");
      return;
    }
    setInstallationBusy(true);
    setError("");
    try {
      const result = await api<{ installation_url: string }>(`/github/connect?organization_id=${encodeURIComponent(orgID)}`);
      window.location.assign(result.installation_url);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Could not connect GitHub.");
      setInstallationBusy(false);
    }
  }

  function chooseRepo(repo: GitHubRepo) {
    setSelectedRepo(repo);
    setBranch(repo.default_branch || "main");
    setServiceName(repo.full_name.split("/").at(-1)?.toLowerCase().replace(/[^a-z0-9-]+/g, "-") || "my-service");
    setPort(repo.full_name.toLowerCase().includes("next") ? "3000" : "8080");
    setError("");
  }

  function updateVariable(id: string, update: Partial<Variable>) {
    setVariables((current) => current.map((variable) => variable.id === id ? { ...variable, ...update } : variable));
  }

  function addVariable() {
    setVariables((current) => [...current, { id: crypto.randomUUID(), key: "", value: "", secret: true, reveal: false }]);
  }

  function importDotEnv() {
    const parsed = pasteValue.split(/\r?\n/).flatMap((line) => {
      const trimmed = line.trim();
      if (!trimmed || trimmed.startsWith("#")) return [];
      const separator = trimmed.indexOf("=");
      if (separator < 1) return [];
      const key = trimmed.slice(0, separator).trim();
      let value = trimmed.slice(separator + 1).trim();
      if ((value.startsWith('"') && value.endsWith('"')) || (value.startsWith("'") && value.endsWith("'"))) value = value.slice(1, -1);
      return [{ id: crypto.randomUUID(), key, value, secret: true, reveal: false }];
    });
    setVariables((current) => {
      const byKey = new Map(current.filter((item) => item.key.trim()).map((item) => [item.key, item]));
      parsed.forEach((item) => byKey.set(item.key, item));
      return [...byKey.values()];
    });
    setPasteValue("");
    setPasteOpen(false);
  }

  function continueFromSource(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    if (!selectedRepo) { setError("Choose a public GitHub repository to continue."); return; }
    if (!serviceName.trim()) { setError("Enter a service name."); return; }
    if (!branch.trim()) { setError("Enter a branch name."); return; }
    setStep(1);
  }

  function continueFromConfig() {
    setError("");
    const keys = new Set<string>();
    for (const variable of configuredVariables) {
      if (!variable.key.trim() || !/^[A-Za-z_][A-Za-z0-9_]*$/.test(variable.key.trim())) {
        setError(`“${variable.key || "Environment variable"}” needs a valid key (letters, numbers, and underscores).`);
        return;
      }
      if (!variable.value) { setError(`${variable.key} needs a value, or remove the empty row.`); return; }
      if (keys.has(variable.key.trim())) { setError(`${variable.key} is listed more than once.`); return; }
      keys.add(variable.key.trim());
    }
    if (!projectChoice && !newProjectName.trim()) { setError("Give the new project a name."); return; }
    setStep(2);
  }

  async function launchDeployment() {
    if (!selectedRepo) return;
    setBusy(true);
    setError("");
    try {
      let organization = orgs[0];
      if (!organization) {
        organization = await api<Org>("/organizations", {
          method: "POST",
          body: JSON.stringify({ name: "Personal workspace", slug: `personal-${Date.now()}` }),
        });
      }

      let project = selectedProject;
      if (!project) {
        const name = newProjectName.trim();
        project = await api<Project>("/projects", {
          method: "POST",
          body: JSON.stringify({
            name,
            slug: name.toLowerCase().trim().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, ""),
            organization_id: organization.id,
            visibility: "PRIVATE",
          }),
        });
      }

      let environment = environments.find((item) => item.project_id === project!.id);
      if (!environment) {
        environment = await api<Environment>("/environments", {
          method: "POST",
          body: JSON.stringify({ project_id: project.id, name: "production" }),
        });
      }

      const service = await api<Service>("/services", {
        method: "POST",
        body: JSON.stringify({
          environment_id: environment.id,
          name: serviceName.trim(),
          runtime,
          repository_url: selectedRepo.clone_url,
          branch: branch.trim(),
          build_context: ".",
          dockerfile_path: "Dockerfile",
          port: Number(port),
          replicas: 1,
          auto_deploy: true,
        }),
      });

      const deployment = await api<Deployment>("/deployments/", {
        method: "POST",
        body: JSON.stringify({
          service_id: service.id,
          env_vars: configuredVariables.map((variable) => ({
            key: variable.key.trim(), value: variable.value, is_secret: variable.secret,
          })),
        }),
      });
      router.push(`/deployments/${deployment.id}`);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "CloudX could not start this deployment.");
    } finally {
      setBusy(false);
    }
  }

  return <AppShell title="New deployment" description="Connect a repository, add runtime configuration, and ship a release.">
    <Link href="/services" className="back-link"><ArrowLeft size={14}/> Back to services</Link>
    <div className="deploy-wizard">
      <div className="wizard-progress" aria-label={`Step ${step + 1} of ${steps.length}`}>
        {steps.map((label, index) => <div className={`wizard-step ${index === step ? "current" : index < step ? "complete" : ""}`} key={label}>
          <span>{index < step ? <Check size={13}/> : `0${index + 1}`}</span><b>{label}</b>{index < steps.length - 1 && <i/>}
        </div>)}
      </div>

      <section className="panel wizard-panel">
        {step === 0 && <form onSubmit={continueFromSource}>
          <span className="panel-eyebrow">01 / SOURCE</span>
          <h2>Choose what to ship.</h2>
          <p className="wizard-intro">Connect a GitHub repository. CloudX will clone the selected branch and detect the build.</p>
          <div className="github-picker wizard-repo-picker">
            <div className="github-picker-title"><span><Github size={15}/> GITHUB REPOSITORY</span>{selectedRepo && <small>CONNECTED</small>}</div>
            {selectedRepo ? <div className="selected-repo">
              <span><Github size={15}/><b>{selectedRepo.full_name}</b><a href={selectedRepo.html_url} target="_blank" rel="noreferrer" aria-label="Open repository"><ChevronRight size={14}/></a></span>
              <small>{selectedRepo.description || "No repository description"}</small>
              <button type="button" className="repo-change" onClick={() => setSelectedRepo(null)}>Choose a different repository</button>
            </div> : <>
              {installations.length ? <>
                <div className="connected-github-banner"><span><i/> GitHub connected as <b>{installations[0].account_name}</b></span><button type="button" onClick={() => void connectGitHub()} disabled={installationBusy}>{installationBusy ? "Opening GitHub…" : "Add account"}</button></div>
                <div className="github-hint">Choose from repositories authorized for this CloudX workspace.</div>
              </> : <>
                <div className="github-search"><input value={username} onChange={(event) => setUsername(event.target.value)} onKeyDown={(event) => { if (event.key === "Enter") { event.preventDefault(); void findRepos(); } }} placeholder="GitHub username" aria-label="GitHub username"/><button type="button" className="secondary-button" disabled={repoBusy || !username.trim()} onClick={() => void findRepos()}>{repoBusy ? "Loading…" : "Find public repos"}</button></div>
                <div className="github-hint">Use public repos now, or connect GitHub to browse private repositories you authorize.</div>
                <button type="button" className="github-connect-button" onClick={() => void connectGitHub()} disabled={installationBusy}><Github size={14}/>{installationBusy ? "Opening GitHub…" : "Connect GitHub App"}</button>
              </>}
              {repoError && <p className="form-error">{repoError}</p>}
              {repos.length > 0 && <>
                <input className="repo-filter" value={repoSearch} onChange={(event) => setRepoSearch(event.target.value)} placeholder="Filter repositories" aria-label="Filter repositories"/>
                <div className="repo-results wizard-repo-results">{filteredRepos.slice(0, 8).map((repo) => <button type="button" className="repo-result" key={repo.id} onClick={() => chooseRepo(repo)}><span><b>{repo.full_name}</b><small>{repo.description || "No description"}</small></span><span><GitBranch size={12}/> {repo.default_branch}</span></button>)}</div>
              </>}
            </>}
          </div>
          <div className="wizard-fields">
            <label>Service name<input value={serviceName} onChange={(event) => setServiceName(event.target.value)} placeholder="web-service" required/></label>
            <label>Branch<input value={branch} onChange={(event) => setBranch(event.target.value)} placeholder="main" required/></label>
            <label>Runtime<select value={runtime} onChange={(event) => setRuntime(event.target.value)}><option>Node.js</option><option>Go</option><option>Python</option><option>Rust</option></select></label>
            <label>Container port<input type="number" value={port} min="1" max="65535" onChange={(event) => setPort(event.target.value)} required/></label>
          </div>
          <div className="wizard-note"><Globe2 size={14}/><span>{installations.length ? "The selected repository uses a short-lived GitHub installation token." : "Public repositories work without setup. Connect GitHub to deploy private repositories."}</span></div>
          {error && <p className="form-error">{error}</p>}
          <div className="wizard-actions"><span>STEP 01 OF 03</span><button className="primary-button" type="submit">Configure service <ArrowRight size={14}/></button></div>
        </form>}

        {step === 1 && <div>
          <span className="panel-eyebrow">02 / CONFIGURE</span>
          <h2>Set up your environment.</h2>
          <p className="wizard-intro">Choose a project and add the variables your app needs at runtime.</p>
          <div className="wizard-fields wizard-project-fields">
            <label>Project<select value={projectChoice} onChange={(event) => setProjectChoice(event.target.value)}><option value="">Create a new project</option>{projects.map((project) => <option value={project.id} key={project.id}>{project.name}</option>)}</select></label>
            {!projectChoice && <label>New project name<input value={newProjectName} onChange={(event) => setNewProjectName(event.target.value)} placeholder={`${serviceName || "My service"} project`}/></label>}
          </div>
          <div className="wizard-section-heading"><div><span className="panel-eyebrow">RUNTIME VARIABLES</span><h3>Environment variables</h3></div><div><button type="button" className="text-button" onClick={() => { setPasteOpen((open) => !open); setError(""); }}>Paste .env</button><button type="button" className="secondary-button" onClick={addVariable}><Plus size={13}/> Add variable</button></div></div>
          {pasteOpen && <div className="env-paste-panel"><label htmlFor="env-paste">Paste KEY=value lines</label><textarea id="env-paste" value={pasteValue} onChange={(event) => setPasteValue(event.target.value)} placeholder={'DATABASE_URL=postgres://…\nAPI_KEY=…'} spellCheck={false}/><div><span>Values are sent only to CloudX over HTTPS.</span><button type="button" className="primary-button" onClick={importDotEnv}>Import variables</button></div></div>}
          <div className="env-variable-list">
            {variables.length ? variables.map((variable) => <div className="env-variable-row" key={variable.id}>
              <input className="env-key-input" value={variable.key} onChange={(event) => updateVariable(variable.id, { key: event.target.value })} placeholder="VARIABLE_NAME" aria-label="Variable name" autoComplete="off" spellCheck={false}/>
              <div className="env-value-wrap"><input value={variable.value} type={variable.secret && !variable.reveal ? "password" : "text"} onChange={(event) => updateVariable(variable.id, { value: event.target.value })} placeholder="Value" aria-label={`Value for ${variable.key || "environment variable"}`} autoComplete="new-password" spellCheck={false}/><button type="button" aria-label={variable.reveal ? "Hide value" : "Reveal value"} onClick={() => updateVariable(variable.id, { reveal: !variable.reveal })}>{variable.reveal ? <EyeOff size={14}/> : <Eye size={14}/>}</button></div>
              <label className="secret-toggle"><input type="checkbox" checked={variable.secret} onChange={(event) => updateVariable(variable.id, { secret: event.target.checked, reveal: false })}/><ShieldCheck size={13}/> Secret</label>
              <button className="icon-button" type="button" aria-label="Remove variable" onClick={() => setVariables((current) => current.filter((item) => item.id !== variable.id))}><Trash2 size={14}/></button>
            </div>) : <div className="env-empty"><ShieldCheck size={16}/><span><b>No variables added</b><small>Add app configuration now, or continue without it.</small></span><button type="button" className="secondary-button" onClick={addVariable}><Plus size={13}/> Add variable</button></div>}
          </div>
          <div className="wizard-note"><ShieldCheck size={14}/><span>Secret values are encrypted before database storage and masked after submission.</span></div>
          {error && <p className="form-error">{error}</p>}
          <div className="wizard-actions"><button type="button" className="secondary-button" onClick={() => { setError(""); setStep(0); }}><ArrowLeft size={13}/> Back</button><span>STEP 02 OF 03</span><button className="primary-button" type="button" onClick={continueFromConfig}>Review deployment <ArrowRight size={14}/></button></div>
        </div>}

        {step === 2 && <div>
          <span className="panel-eyebrow">03 / REVIEW</span>
          <h2>Ready for takeoff.</h2>
          <p className="wizard-intro">Review your source and runtime settings before CloudX starts the build.</p>
          <div className="deploy-review-card">
            <div><span className="review-icon"><Github size={16}/></span><span><small>REPOSITORY</small><b>{selectedRepo?.full_name}</b><i><GitBranch size={12}/> {branch}</i></span></div>
            <div><span className="review-icon"><Box size={16}/></span><span><small>SERVICE</small><b>{serviceName}</b><i>{runtime} · port {port}</i></span></div>
            <div><span className="review-icon"><Rocket size={16}/></span><span><small>PROJECT / ENVIRONMENT</small><b>{selectedProject?.name || newProjectName}</b><i>{selectedEnvironment?.name || "production"}</i></span></div>
            <div><span className="review-icon"><ShieldCheck size={16}/></span><span><small>ENVIRONMENT VARIABLES</small><b>{configuredVariables.length} configured</b><i>{configuredVariables.filter((variable) => variable.secret).length} secret values</i></span></div>
          </div>
          <div className="wizard-note"><Globe2 size={14}/><span>CloudX will build and start the container. A public app URL is not available until service routing is configured.</span></div>
          {error && <p className="form-error">{error}</p>}
          <div className="wizard-actions"><button type="button" className="secondary-button" disabled={busy} onClick={() => { setError(""); setStep(1); }}><ArrowLeft size={13}/> Back</button><span>{busy ? <><LoaderCircle className="spin" size={13}/> BUILD IN PROGRESS</> : "READY TO DEPLOY"}</span><button className="primary-button" type="button" disabled={busy} onClick={() => void launchDeployment()}>{busy ? <><LoaderCircle className="spin" size={14}/> Deploying…</> : <>Deploy service <Rocket size={14}/></>}</button></div>
          {busy && <p className="deploy-wait-copy">CloudX is cloning the repository, building the image, and starting the service. This can take a few minutes.</p>}
        </div>}
      </section>
      <div className="wizard-footer"><span><ShieldCheck size={13}/> ENVIRONMENT VALUES ARE ENCRYPTED AT REST</span><Link href="/docs">Deployment docs <ChevronRight size={13}/></Link></div>
    </div>
  </AppShell>;
}
