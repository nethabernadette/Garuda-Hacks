# Jalin Frontend

Vite + React conversion of the original `initial.html` mobile prototype.

## Run

```bash
cd Frontend
npm install
npm run dev
```

Set the backend URL:

```bash
cp .env.example .env
```

```env
VITE_API_BASE_URL=http://localhost:8080
```

## Build

```bash
npm run build
```

Untuk membuka langsung dari File Explorer tanpa localhost:

```bash
npm run build:single
```

Lalu buka:

```text
Frontend/dist/index.html
```

## Backend Integration

The API client is centralized in `src/services/api.js`. It uses `Authorization: Bearer <token>` when an access token is available from login/register.

Connected endpoints include auth, profile, posts, notifications, agreements, RFQ document generation, and contact reveal. Match/chat UI still uses centralized demo data where backend endpoints require match IDs that are not discoverable from the current documented API.
