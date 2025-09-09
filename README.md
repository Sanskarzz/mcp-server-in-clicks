# MCP Server in Clicks

No‑code SaaS to deploy custom Model Context Protocol (MCP) servers from JSON config. This monorepo includes:

- a production‑ready Go MCP server template (config‑driven, no code changes)
- a Go backend to orchestrate multi‑tenant deployments via Helm on Kubernetes
- a React + Tailwind frontend (SaaS UI)


## Highlights

- Config‑driven MCP server using `mark3labs/mcp-go`
- Dynamic tools: map any HTTP API endpoint (method, headers, query, body)
- Secure auth: Bearer/basic, env var substitution, OAuth 2.0 discovery + challenges
- Input/output validation, robust errors to prevent LLM hallucination
- JSON‑RPC over HTTP for streamable MCP clients (Cursor, Claude, VS Code)
- Dockerized; Helm charts for one‑click per‑tenant deploys
- Backend with JWT auth, Google login (OIDC), MongoDB persistence, Helm Go SDK


## Monorepo layout

```
.
├─ mcp-server-template/     # Go MCP server template (SDK: mark3labs/mcp-go)
│  ├─ cmd/server/           # Main entrypoint
│  ├─ internal/             # config, handlers, server, validation
│  ├─ examples/             # JSON configs (tools/prompts/resources)
│  ├─ deploy/helm/          # Helm chart for MCP server instances
│  ├─ Dockerfile            # Multi‑stage Docker build
│  └─ Makefile              # Build/test/lint/docker/helm
├─ backend/                 # Go backend API + CLI (Cobra, chi, viper)
│  ├─ cmd/                  # CLI entrypoints
│  ├─ internal/             # api, auth, config, helm svc, storage
│  ├─ Dockerfile            # Backend Docker image
│  ├─ k8s.yaml              # Backend Kubernetes manifests
│  └─ Makefile              # Build/test/docker/k8s/helm helpers
└─ frontend/                # React + TypeScript + Tailwind SaaS UI
   └─ (Vite, route guard, auth store, pages)
```


## Prerequisites

- Go 1.23+
- Node 18+/pnpm or npm
- Docker 24+
- kubectl + access to a Kubernetes cluster
- Helm 3.14+


## Quick start: run the MCP server locally (binary)

1) Build the server

```bash
cd mcp-server-template
go build -o mcp-server cmd/server/main.go
```

2) Run with an example configuration

```bash
./mcp-server --config examples/simple-server.json --port 9090 --log-level debug
```

3) Smoke test JSON‑RPC

```bash
curl -s -X POST http://127.0.0.1:9090/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | jq
```

You should see your configured tools. To call a tool, pass parameters via `tools/call`:

```bash
curl -s -X POST http://127.0.0.1:9090/mcp \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc":"2.0",
    "id":2,
    "method":"tools/call",
    "params":{
      "name":"get_quote",
      "arguments":{}
    }
  }' | jq
```

Tip: To use this server from Cursor/Claude locally, point your client to `http://127.0.0.1:9090/mcp` (e.g. Cursor’s `~/.cursor/mcp.json`).


## MCP server template: concepts

- Config first: define tools, prompts, resources entirely in JSON
- Tools map to HTTP endpoints with support for
  - method (GET/POST/PUT/PATCH/DELETE/HEAD/OPTIONS)
  - path/query templating from input params
  - headers (static or env‑var substituted)
  - auth (bearer/basic/env var)
  - timeouts and retries
- Validation: request inputs and response shapes
- OAuth 2.0 resource hints
  - `/.well-known/oauth-protected-resource` for discovery
  - `WWW-Authenticate` challenges on `/mcp` when enabled


## Example configuration (excerpt)

```json
{
  "server": { "name": "example", "version": "0.1.0" },
  "runtime": { "default_timeout": "10s" },
  "security": {
    "oauth": {
      "enabled": true,
      "accepted_audiences": ["your-client-id"],
      "required_scopes": ["read"]
    }
  },
  "tools": [
    {
      "name": "get_quote",
      "description": "Get a random quote",
      "endpoint": "https://zenquotes.io/api/random",
      "method": "GET",
      "timeout": "10s"
    }
  ]
}
```


## Build a Docker image (MCP server)

```bash
cd mcp-server-template
docker build -t sanskardevops/mcp-server-template:0.0.1 .
docker push sanskardevops/mcp-server-template:0.0.1
```


## Deploy via Helm (MCP server)

The Helm chart lives in `mcp-server-template/deploy/helm`.

1) Create a values override file with your config JSON embedded:

```yaml
# my-values.yaml
image:
  repository: sanskardevops/mcp-server-template
  tag: 0.0.1

server:
  port: 9090
  logLevel: info

config: |
  {
    "server": { "name": "example", "version": "0.1.0" },
    "tools": [
      {"name": "get_quote", "description": "Get a random quote", "endpoint": "https://zenquotes.io/api/random", "method": "GET"}
    ]
  }
```

2) Install to a namespace (e.g. `agents`):

```bash
cd mcp-server-template/deploy/helm
helm upgrade --install mcp-example . \
  -n agents --create-namespace \
  -f values.yaml -f my-values.yaml
```


## Backend (Go) – run locally

Environment variables (common):

- `PORT` (default 6000)
- `MONGO_URI` (e.g. `mongodb://localhost:27017/mcp`) and `MONGO_DB`
- `JWT_SECRET`
- Google OAuth: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`
- `KUBECONFIG` if running Helm against out‑of‑cluster

Run the server:

```bash
cd backend
go run ./cmd/backend server --port 6000
# or
make run
```

The backend includes:

- JWT auth + Google login endpoints
- MongoDB persistence for users, workspaces, server configs
- Helm Go SDK service for install/upgrade status tracking


## Frontend (React + Tailwind) – run locally

```bash
cd frontend
npm i
npm run dev
```

Configure `.env.local` as needed (API base URL, OAuth IDs). The UI includes login/register, dashboard, server wizard, and server details.


## Makefiles

- `mcp-server-template/Makefile`: build, test, lint, docker, helm helpers
- `backend/Makefile`: build, docker, k8s/helm deploy helpers

Common flows:

```bash
# MCP server
cd mcp-server-template
make build
make test
make docker-build docker-push IMG=sanskardevops/mcp-server-template:0.0.1

# Backend
cd backend
make build
make docker-build
```


## Security notes

- For bearer tokens in tool configs, prefer `EnvVar` so the server reads from process env (e.g. `AUTH_TOKEN`, `ACCUKNOX_TOKEN`).
- OAuth discovery and `WWW-Authenticate` on `/mcp` help clients obtain tokens. Future work: full automatic token passing for configured tools.
- CORS enabled for MCP JSON‑RPC; configure ingress accordingly in Helm for public access.


## Roadmap (selected)

- Backend: multi‑tenant workspaces, RBAC, job status for Helm ops
- MCP server: richer schema validation, response shape contracts, pagination helpers
- Frontend: polished UX, server creation wizard with schema hints and live preview
- Observability: structured logs, metrics, traces; audit trails per tenant
