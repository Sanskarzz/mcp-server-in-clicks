### MCP Servers No‑Code Platform — One‑Pager

#### Vision
Enable any team to safely expose internal APIs and workflows to AI assistants (Claude, Cursor, VS Code) via Model Context Protocol (MCP) with zero custom code and enterprise‑grade governance.

#### Problem
- Teams want agents to take action, but exposing internal systems is risky and slow.
- Every app builds custom glue: auth, validation, logging, audits, deployments.
- Inconsistent quality and security; hard to operate at scale.

#### Solution
A config‑driven MCP server template and deployment system:
- Define tools/prompts/resources in JSON (validated), no code changes.
- Secure HTTP execution with auth, headers, templated bodies, timeouts/retries.
- Ship as Docker + Helm for Kubernetes; works locally with a Go binary.
- Enterprise add‑ons: RBAC, audit trail, secret management, SSO, observability.

#### Product (initial scope)
- Core OSS: Go template using `mark3labs/mcp-go`, JSON config, validator, structured logging, health/metrics, Dockerfile, Helm chart, Makefile, examples, tests.
- Cloud/Enterprise: config RBAC, audit logs, usage metering, SSO/SAML, secrets (Vault/External Secrets/KMS), request/response redaction, allow/deny lists, GitOps flows.

#### Key Capabilities
- Dynamic tool registration from JSON (methods, headers, query/body templates, auth).
- Input/output validation (`go-playground/validator` + custom rules).
- JSON‑RPC 2.0 over HTTP with CORS + preflight; graceful shutdown.
- Observability: logrus, health, metrics; retries, timeouts, backoff.
- K8s‑first: Helm values for per‑tenant config, HPA, Ingress, ServiceAccount.

#### High‑Level Architecture
- Config → Go MCP server (SDK) → JSON‑RPC over HTTP.
- Tool execution → HTTP client with templating, auth injection, validation.
- Deploy via Docker/Helm; store configs in Git/Mongo (future backend).

#### Ideal Customer Profile (ICP) & Use Cases
- Platform/DevEx teams in mid‑to‑large orgs enabling governed agent access.
- SecOps/DataOps automating playbooks (tickets, filters, dashboards, reports).
- Consultancies needing repeatable, production‑grade MCP deployments.

#### Differentiation
- No‑code config + strong validation and governance out of the box.
- Kubernetes‑native deployment and scaling.
- Opinionated DX: examples, tests, Makefile, conformance checks.

#### Pricing & Packaging (initial hypothesis)
- OSS Core: free.
- Cloud: per server + usage (requests/stream minutes).
- Enterprise: annual license (on‑prem operator), SSO, audit, RBAC, support SLAs.
- Services: fixed‑fee accelerators (MCP enablement, hardening, tool onboarding).

#### Go‑to‑Market
- Open source + content (guides, templates), Helm repo, quickstarts for Cursor/VS Code/Claude.
- Design partners (5–10 platform teams) to co‑develop governance features.
- Community templates: Jira, GitHub, PagerDuty, ServiceNow, Slack, Datadog, Snowflake.

#### KPIs
- Time‑to‑first successful tool call (<15 minutes).
- Error rate, P50/P95 request latency, % deployments via Helm vs cloud.
- Secrets provider adoption, SSO enablement, audit queries executed.

#### Risks & Mitigations
- Spec/SDK changes → compatibility matrix, fast upgrade cadence, tests.
- Security exposure → allow/deny lists, scoped tokens, redaction, audits, egress policies.
- Commoditization → governance + enterprise integrations + great DX.

---

### 30–60–90 Day Execution Plan

#### Day 0 Prereqs
- Public GitHub repo with OSS license and contribution guide.
- CI for build/test/lint; basic conformance test for MCP flows.

#### Days 1–30 — MVP Hardening and Launch
- Features
  - Finalize JSON config schema (tools/prompts/resources/security/runtime).
  - Robust validation (custom rules: semver, endpoint, duration, json, enum).
  - HTTP client: retries/backoff, timeouts, templated bodies, default headers, auth injection.
  - JSON‑RPC over HTTP handler with CORS + OPTIONS preflight.
  - Health endpoint, basic metrics, structured logs, graceful shutdown.
  - Examples: simple, weather, filter (with working curls).
  - Tests: unit + golden tests for config loader/validator and HTTP client; example conformance.
- DevEx
  - Makefile targets (run, test, docker, helm, lint).
  - Multi‑stage Dockerfile; publish `sanskardevops/mcp-server-template:0.0.1`.
  - Helm chart with `templates/` (Deployment, Service, Ingress, HPA, ConfigMap, ServiceAccount).
- Docs
  - README quickstart: local (Go binary), Docker, Helm; Cursor/VS Code/Claude examples.
  - ARCHITECTURE.md; configuration reference; troubleshooting guide.
- Acceptance Criteria
  - Green CI; docker image published; helm install works with an example config.
  - MCP clients can: initialize, list tools, call a tool, get prompts/resources.

#### Days 31–60 — Governance & Enterprise Readiness
- Features
  - RBAC per server/workspace/tool; rate limits per tenant/tool.
  - Audit logging (who/what/when, inputs/outputs with redaction policy).
  - Secret management integrations (Vault, External Secrets, KMS env injection).
  - Request/response redaction and allow/deny list for outbound hosts/paths.
  - Request tracing (trace IDs, OpenTelemetry hooks) and error taxonomy.
- Ops
  - Per‑tool dashboards (metrics/latency/error rate), SLOs for availability/latency.
  - GitOps flow: signed configs, versioning, staged rollouts, rollback.
- Docs & Samples
  - Enterprise templates: Jira, ServiceNow, Snowflake.
  - Conformance test suite runnable post‑deploy.
- Acceptance Criteria
  - RBAC + audit + secrets demoed with at least one enterprise template.

#### Days 61–90 — Hosted Beta & Monetization
- Features
  - Multi‑tenant control plane (basic): workspace auth, usage metering, billing events.
  - Hosted deployments (K8s fleet), tenant‑scoped secrets, per‑tenant Helm values.
  - SSO/SAML integration; SCIM optional.
- GTM
  - Design partner onboard (5–10), weekly cadence, backlog from real use.
  - Content: blog series, template gallery, conference CFPs.
- Acceptance Criteria
  - 3+ design partners running production‑like workflows; usage data captured.

#### Parallel Tracks (continuous)
- Content & Community: guides, video walkthroughs, templates, Discord.
- Sales Readiness: one‑pager, pricing sheet, ROI calculator, case studies.
- Security Review: threat model, pen test plan, dependency audit, SBOM.

#### Hiring (minimalistic)
- Contract designer for docs/UX polish.
- Part‑time DevOps/SRE for hosted beta and observability.

#### MVP Definition of Done (DoD)
- JSON‑configured server runs locally, via Docker, and via Helm.
- Tools work with auth, headers, templated bodies; validated inputs/outputs.
- Health/metrics/logs available; tests pass; example clients succeed.

---

### SaaS Platform Blueprint (Frontend + Backend)

Reference UI inspiration: [mcpcraft.preview.emergentagent.com](https://mcpcraft.preview.emergentagent.com/)

#### Product Overview
- Multi‑tenant SaaS where users create MCP servers “in clicks.”
- Workspace‑scoped servers with tools, prompts, and resources defined via UI → JSON config → deployed via Helm.
- Managed Kubernetes fleet or BYO‑cluster; each server runs isolated (namespace) with per‑tenant quotas and rate limits.

#### Frontend (React/TypeScript)
- Pages: Dashboard, Create/Edit Server (tabs: Basic, Tools, Prompts, Resources), Deployments, Logs, Settings, Billing.
- UX features: schema‑driven forms with validation, body/header templating helpers, live JSON preview, test runner (dry‑run + curl generator), copy‑to‑clipboard MCP client snippets.
- Auth: Email+OTP (MVP) → OAuth/OIDC (Google/GitHub/Microsoft). Workspace invites/roles.
- State: React Query + Zod for schema validation; optimistic updates; toasts; deep‑linkable routes.

#### Backend (Go)
- REST API (JSON) and Webhooks.
- Core services:
  - Config Service: validate, version, diff; env‑var substitution; secret references (not values) in configs.
  - Provisioner: render Helm values, create namespace, apply chart, watch rollout; supports upgrades/rollbacks.
  - Secrets: integrate Vault/External Secrets; never persist raw secrets.
  - Usage + Audit: store request counters, durations, errors; append‑only audit events per action.
  - Billing hooks: emit metering events (server‑minutes, requests) to billing.
- Observability: OpenTelemetry, structured logs, request IDs, per‑tool metrics.
- API surface (v1):
  - POST /workspaces, POST /invites
  - CRUD /servers, /servers/{id}/deploy, /servers/{id}/rollback, /servers/{id}/status
  - CRUD /servers/{id}/tools|prompts|resources
  - GET /deployments, /events, /usage
  - POST /secrets (stores references), GET /secrets/{id}

#### Data Model (MongoDB)
- User, Workspace(role: owner, admin, editor, viewer)
- Server {name, version, configRef, status}
- ConfigVersion {serverId, json, checksum, createdBy}
- Tool/Prompt/Resource documents or embedded in ConfigVersion
- Deployment {serverId, configVersionId, status, startedAt, completedAt, logsRef}
- SecretRef {workspaceId, name, provider, path/key}
- AuditEvent, UsageRecord

#### Provisioning Flow
1) User creates/edits server in UI → validated JSON generated in real time.
2) Save → ConfigVersion persisted; dry‑run renders Helm values.
3) Deploy → Provisioner applies chart; status streamed back; health check gated.
4) On success → endpoint + client snippet shown; usage metering starts.
5) Rollback → select previous ConfigVersion; Provisioner downgrades release.

#### Security & Governance
- Network egress allow/deny lists per workspace; per‑tool rate limits.
- RBAC by workspace and by server; signed configs; audit on every action.
- Secrets via provider; redaction of request/response logs; PII scrubber hooks.

#### Scalability
- One namespace per server or per workspace; HPA enabled; resource quotas.
- Queue for deployments (minimize control‑plane churn); backpressure and retries.

---

### Execution Plan Addendum (SaaS Tracks)

#### Days 1–30 (UI + API MVP)
- Frontend: Dashboard + Create/Edit Server wizard (Basic, Tools, Prompts, Resources), JSON preview, client snippet, dry‑run button.
- Backend: CRUD for workspaces/servers/configs; validation; dry‑run renderer; simple auth (JWT) + roles; audit log.
- DevOps: Helm chart published; staging cluster; CI/CD for UI + API.

#### Days 31–60 (Provisioning + Observability)
- Backend: Provisioner with Helm ops (install/upgrade/rollback), rollout watcher, status endpoints, logs streaming.
- Frontend: Deployments page, live status toasts, per‑tool test runner; secrets UI (references only).
- Security: allow/deny egress, per‑tool rate limits; request IDs; initial metrics and dashboards.

#### Days 61–90 (Multi‑tenant Beta + Billing)
- Workspaces: invites, roles, usage dashboards.
- Billing: usage events pipeline; basic plans; limits enforcement.
- Enterprise: SSO/OIDC, External Secrets/Vault integration; audit export.

Acceptance Gate for Hosted Beta
- Create → Deploy → Call tool from Cursor/VS Code in <15 minutes.
- Rollback works; audit/usage recorded; secrets never stored in plain text.


