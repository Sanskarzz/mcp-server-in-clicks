MCP SaaS Frontend (React + Vite + Tailwind)

Quick start

1) Set API base URL (match your running backend port)

Create a `.env.local` in this `frontend/` folder:

```
VITE_API_BASE_URL=http://localhost:8989
```

2) Install deps and run the dev server

```
cd frontend
npm i
npm run dev
```

Open `http://localhost:5173` in your browser.

Useful routes

- `/login`: demo login card (Quick Demo Login navigates to dashboard)
- `/dashboard`: server cards + “+ New Server” CTA
- `/servers/new`: multi-step wizard (Basic → Tools → Prompts → Resources → Review)
- `/servers/:id`: server detail placeholder

Build and preview (production)

```
npm run build
npm run preview
```

Tech stack

- React 18 + TypeScript + Vite
- Tailwind CSS (with brand tokens and gradient hero)
- React Router v6, TanStack Query, Zustand, Axios, Zod (wired for future API integration)

Notes

- Ensure `VITE_API_BASE_URL` matches your backend (e.g., `./bin/backend server --port 8989`).
- Current UI is scaffolded; forms/validation and API wiring will be completed next.

