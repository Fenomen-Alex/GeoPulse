# GeoPulse

Spatial intelligence SaaS for multi-modal reachability analysis, POI density indexing, and spatial workspace management.

Drop a pin → Configure parameters → Run spatial calculation → Render reachability contours and accessibility estimates.

## Quick Start

```bash
# Install dependencies
cd client && npm ci
cd ../server && go mod download

# Start the frontend and API together
cd client && npm run dev:all
```

Local dev: Frontend at `http://localhost:5173`, API at `http://localhost:8080`.

## Docker

```bash
docker compose up --build
```

The app runs at `http://localhost:8080`.

## Architecture

| Layer | Tech |
|---|---|
| **Frontend** | SolidJS + Vite + Tailwind CSS v4 |
| **Backend** | Go 1.26 + Chi router |
| **Map** | MapLibre GL JS |
| **Spatial API** | OpenRouteService (proxied) |
| **Database** | Turso (LibSQL) via `database/sql` |
| **Email** | Resend API for contact/quota extensions |
| **Auth** | Google OIDC + JWT cookies |

## Features

- Multi-modal reachability estimates (walk, bike, drive) with 5–60 minute synthetic bands
- Street-level turn-by-turn route navigation (OpenRouteService Directions) drawn as road-network polylines
- Simulated POI and accessibility metrics, ready for replacement with a live POI provider
- Token-bucket rate limiting with 15 runs/day default quota for authenticated users
- Authenticated quota-status synchronization across page reloads via `GET /api/v1/quota`
- Quota extension requests via contact form with Resend email dispatch

## Acceptance Criteria

1. Unauthenticated users see the landing page; calling `/api/v1/analysis` or `/api/v1/routes` without a session cookie returns `401`.
2. Google OIDC auth gates the Spatial Workbench; quota counter shows `15 / 15 remaining`.
3. Contact form submits to `/api/v1/contact`, dispatches email via Resend API, stores record in DB.
4. 16th combined spatial request (isochrone + route) returns HTTP `429` and triggers the Quota Extension Modal.
5. Route tool: map click sets start (green marker), second click sets destination (red marker), "Get Directions" draws a street-network polyline with distance/duration and turn-by-turn steps.
6. No upstream API keys (`ORS_API_KEY`, `RESEND_API_KEY`) appear in client bundle or network inspector.

## Project Structure

```
├── client/          # SolidJS SPA (source)
│   ├── src/
│   │   ├── components/    # UI components
│   │   ├── store/         # Reactive state (analysisStore)
│   │   ├── views/         # Page views
│   │   └── services/      # API client
│   └── dist/          # Built assets (embedded by Go)
├── server/          # Go backend
│   ├── cmd/server/  # Entrypoint with go:embed
│   ├── internal/
│   │   ├── handler/ # HTTP endpoints
│   │   ├── middleware/ # Auth, CORS, rate limiting
│   │   ├── db/      # Turso connection & queries
│   │   └── contact/ # Resend email integration
│   └── migrations/  # SQL schema
├── Dockerfile       # Multi-stage build
└── docker-compose.yml
```

## Environment Variables

See `example.env` at the repo root. Key variables:

- `TURSO_DATABASE_URL` + `TURSO_AUTH_TOKEN` for edge persistence
- `ORS_API_KEY` for routing (sealed in backend proxy)
- `RESEND_API_KEY` for transactional emails
- `GOOGLE_CLIENT_SECRET` for OIDC (frontend uses `VITE_GOOGLE_CLIENT_ID`)

`VITE_TEST_MODE=true` enables the development-only test-user flow. Production builds compile the client-side bypass out.
