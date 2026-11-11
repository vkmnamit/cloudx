"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { Activity, ArrowRight, ArrowUpRight, Box, CircleHelp, GitBranch, Layers3, Plus, Rocket, Search } from "lucide-react";
import { AppShell } from "@/components/sidebar";
import { api } from "@/lib/api";

type Project = { id: string; name: string; slug: string; organization_id: string; visibility?: string };
type Environment = { id: string; project_id: string; name: string };
type Service = { id: string; name: string; runtime: string; environment_id: string; repository_url: string; branch?: string };
type Deployment = { id: string; service_id: string; status: string; created_at: string; commit_sha?: string; commit_message?: string };

export default function DashboardPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [environments, setEnvironments] = useState<Environment[]>([]);
  const [services, setServices] = useState<Service[]>([]);
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [search, setSearch] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    if (!localStorage.getItem("cloudx_token")) { location.href = "/login"; return; }
    Promise.all([api<Project[]>("/projects"), api<Environment[]>("/environments"), api<Service[]>("/services"), api<Deployment[]>("/deployments/")])
      .then(([projectData, environmentData, serviceData, deploymentData]) => {
        setProjects(projectData || []); setEnvironments(environmentData || []); setServices(serviceData || []); setDeployments(deploymentData || []);
      }).catch((cause) => setError(cause instanceof Error ? cause.message : "Could not load your workspace."));
  }, []);

  const visibleServices = services.filter((service) => `${service.name} ${service.repository_url}`.toLowerCase().includes(search.toLowerCase()));
  const serviceName = (id: string) => services.find((service) => service.id === id)?.name || id.slice(0, 8);
  const projectServices = (projectId: string) => {
    const environmentIds = environments.filter((environment) => environment.project_id === projectId).map((environment) => environment.id);
    return services.filter((service) => environmentIds.includes(service.environment_id));
  };

  return <AppShell title="Overview" description="Projects, services, and deployments in one place." action={<Link className="primary-button" href="/projects"><Plus size={15}/> Create project</Link>}>
    <section className="overview-section"><div className="overview-section-heading"><div><span className="panel-eyebrow">YOUR WORKSPACE</span><h2>Projects</h2></div><Link href="/projects">All projects <ArrowRight size={14}/></Link></div>
      {error && <div className="inline-error">{error}</div>}
      <div className="project-card-grid">{projects.slice(0, 3).map((project) => { const count = projectServices(project.id).length; return <Link href={`/projects/${project.id}`} className="overview-project-card" key={project.id}><div className="overview-project-icon"><Layers3 size={18}/></div><div className="overview-project-title"><b>{project.name}</b><span>{project.slug}</span></div><div className="overview-project-meta"><span className="overview-status-dot"/>{count ? `${count} service${count === 1 ? "" : "s"}` : "No services yet"}<ArrowUpRight size={14}/></div></Link>; })}
        <Link href="/projects" className="overview-create-card"><Plus size={17}/><span>Create new project</span></Link>
      </div>
    </section>

    <section className="overview-section services-overview"><div className="overview-section-heading"><div><span className="panel-eyebrow">DEPLOYABLE RESOURCES</span><h2>Services <span className="heading-count">{services.length}</span></h2></div><Link href="/services">Manage services <ArrowRight size={14}/></Link></div>
      <div className="overview-toolbar"><div className="overview-tabs"><span className="selected">All services <b>{services.length}</b></span><Link href="/deployments">Deployments <b>{deployments.length}</b></Link></div><label className="search-box"><Search size={14}/><input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search services"/></label></div>
      {visibleServices.length ? <div className="overview-service-table"><div className="overview-table-head"><span>SERVICE NAME</span><span>STATUS</span><span>RUNTIME</span><span>UPDATED</span><span/></div>{visibleServices.map((service) => { const latest = deployments.find((deployment) => deployment.service_id === service.id); const running = latest?.status === "RUNNING"; return <Link className="overview-service-row" href={`/services/${service.id}`} key={service.id}><span className="overview-service-name"><i><Box size={15}/></i><b>{service.name}<small>{service.repository_url || "Repository not connected"}</small></b></span><span><i className={`service-state-dot ${running ? "is-running" : ""}`}/>{running ? "Running" : latest?.status || "Ready to deploy"}</span><span><i className="runtime-mark">{service.runtime.slice(0, 1)}</i>{service.runtime}</span><span>{latest ? new Date(latest.created_at).toLocaleDateString() : "—"}</span><ArrowUpRight size={14}/></Link>; })}</div> : <div className="overview-empty"><span><Box size={20}/></span><b>{search ? "No matching services" : "No services yet"}</b><p>{search ? "Try another search." : "Connect a repository to configure your first service."}</p>{!search && <Link href="/services">Create a service <ArrowRight size={13}/></Link>}</div>}
    </section>

    <section className="overview-bottom-grid"><div className="overview-mini-panel"><div className="overview-section-heading"><div><span className="panel-eyebrow">LATEST ACTIVITY</span><h2>Recent deployments</h2></div><Link href="/deployments">View all <ArrowRight size={13}/></Link></div>{deployments.length ? deployments.slice(0, 4).map((deployment) => <Link href={`/deployments/${deployment.id}`} className="overview-deployment-row" key={deployment.id}><span className={`deployment-mark ${deployment.status.toLowerCase()}`}><Activity size={13}/></span><span><b>{serviceName(deployment.service_id)}</b><small>{deployment.commit_message || deployment.commit_sha?.slice(0, 8) || deployment.status}</small></span><time>{new Date(deployment.created_at).toLocaleDateString()}</time><ArrowUpRight size={13}/></Link>) : <div className="overview-empty compact"><span><Rocket size={17}/></span><b>No deployments yet</b><p>Launch a service to see releases here.</p></div>}</div>
      <div className="overview-tip"><span className="panel-eyebrow">NEXT UP</span><h2>Get your code in the cloud.</h2><p>Connect a GitHub repository, configure environment variables, and launch a build.</p><Link className="secondary-button" href="/deploy"><GitBranch size={14}/> Start a deployment <ArrowUpRight size={13}/></Link><Link className="overview-doc-link" href="/docs"><CircleHelp size={14}/> Read the quickstart</Link></div>
    </section>
  </AppShell>;
}
