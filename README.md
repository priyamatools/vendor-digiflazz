# vendor-digiflazz

Addon vendor: API Seller Digiflazz → upstream Atrilinks (Query/PIN), PostgreSQL, webhook Digiflazz.

## Docker Hub (CI)

Set repository secrets di GitHub (`Settings` → `Secrets and variables` → `Actions`):

| Secret | Keterangan |
|--------|------------|
| `DOCKERHUB_USERNAME` | Username Docker Hub |
| `DOCKERHUB_TOKEN` | Access token Docker Hub (bukan password akun) |

Push ke branch `main` mem-build image `<DOCKERHUB_USERNAME>/vendor-digiflazz:latest` (username harus sama dengan secret login).

## Menjalankan lokal

```bash
cp .env.example .env
docker compose up -d
go run ./cmd/server
```

Endpoint: `POST /api/digiflazz/topup`, `POST /h2h/atrilinks/callback`, `GET /health`.
