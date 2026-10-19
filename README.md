# vendor-digiflazz

Addon **Seller API** untuk [Digiflazz](https://developer.digiflazz.com/api/seller/prabayar/topup/): menerima request topup/cek status, memvalidasi signature, menyimpan transaksi, meneruskan order ke upstream **[Atrilinks H2H](https://h2h.atrilinks.com/docs)**, lalu mengembalikan response dan webhook sesuai kontrak Digiflazz.

## Fitur

- **Topup & cek status** — idempoten berdasarkan `ref_id` (topup ulang dengan `ref_id` sama mengembalikan state dari database, tanpa order ulang ke Atrilinks).
- **Validasi Digiflazz** — verifikasi `sign` = `md5(username + apiKey + ref_id)` dan `username` dari environment.
- **Integrasi Atrilinks (metode Atrilinks + BLAKE3)** — transaksi prabayar fixed (`jenistrx=1`) via `GET /h2h/trx/atrilinks`, dengan `sign` dan `urlcallback`.
- **Callback Atrilinks** — update status dari notifikasi upstream (`msg` / JSON); jawab `200 OK` cepat, webhook Digiflazz di background.
- **Rekonsiliasi pending** — saat Digiflazz cek status (`ref_id` sama) atau worker latar belakang memanggil Atrilinks `checktoday` (transaksi hari ini) bila callback hilang/timeout.
- **Mapping status** — Atrilinks (`success` / `pending` / `failed`) → Digiflazz (`status` `0`/`1`/`2` + `rc`).
- **PostgreSQL** — persistensi transaksi, audit response upstream (JSON), pelacakan webhook.

## Tanda tangan Atrilinks (upstream)

Metode yang dipakai addon (disarankan dokumentasi Atrilinks):

```text
combined = reqid|kodeproduk|tujuan:userid:password:pin:Atrilinks
sign     = blake3_256(combined)   → hex lowercase, 64 karakter
```

- `password` = `ATRILINKS_PASS`, `pin` = `ATRILINKS_PIN`, `userid` = `ATRILINKS_USER_ID`
- `:Atrilinks` di akhir string adalah salt tetap
- Implementasi: `internal/atrilinks/sign.go`

## Alur singkat

```text
Digiflazz  →  POST topup (addon)
Addon      →  validasi sign → DB (ref_id) → Atrilinks /h2h/trx/atrilinks (+ sign, urlcallback)
Atrilinks  →  callback status → addon → update DB → OK cepat → webhook Digiflazz (async)
Addon      →  (fallback) checktoday saat replay pending / worker
Digiflazz  ←  response { "data": { ... } }
```

## Endpoint HTTP (addon)

| Method | Path | Keterangan |
|--------|------|------------|
| `POST` | `/api/digiflazz/topup` | Request dari Digiflazz (JSON) |
| `GET` / `POST` | `/h2h/atrilinks/callback` | `urlcallback` dari Atrilinks |
| `GET` | `/health` | Health check |

Response ke Digiflazz selalu dibungkus `{ "data": { ... } }`. Webhook ke Digiflazz memakai format yang sama.

## Dokumentasi

| Sumber | Isi |
|--------|-----|
| [Digiflazz — Topup / Cek Status](https://developer.digiflazz.com/api/seller/prabayar/topup/) | Request, signature, field response |
| [Digiflazz — Webhooks](https://developer.digiflazz.com/api/seller/prabayar/webhook/) | Callback ke Digiflazz |
| [Atrilinks H2H](https://h2h.atrilinks.com/docs) | Docs resmi upstream |
| [`docs-atrilinks.txt`](./docs-atrilinks.txt) | Rangkuman lokal (BLAKE3, callback `msg`, dll.) |

## Konfigurasi

Salin [`.env.example`](./.env.example) ke `.env`.

| Variabel | Keterangan |
|----------|------------|
| `HTTP_PORT`, `DATABASE_URL` | Server & PostgreSQL |
| `APP_ENV=development` | Log debug request/response |
| `DIGIFLAZZ_API_USERNAME`, `DIGIFLAZZ_API_KEY`, `DIGIFLAZZ_CALLBACK_URL` | Seller Digiflazz |
| `ATRILINKS_BASE_URL` | Default `http://api.atrilinks.com` |
| `ATRILINKS_USER_ID` | `userid` member |
| `ATRILINKS_PASS` | Password untuk bahan `combined` sign |
| `ATRILINKS_PIN` | PIN untuk bahan `combined` sign |
| `ADDON_ATRILINKS_CALLBACK_URL` | URL publik callback (path `/h2h/atrilinks/callback`) |
| `PENDING_WORKER_*`, `PENDING_RECONCILE_*`, `RECONCILE_TIMEZONE` | Worker & throttle poll `checktoday` (lihat `.env.example`) |

Jangan commit `.env`. Jangan log `sign`, `pin`, atau `password`.

**Mapping produk:** `pulsa_code` Digiflazz → `kodeproduk` Atrilinks (lowercase).

## Tech stack

- Go 1.22+ (BLAKE3: `github.com/zeebo/blake3`)
- PostgreSQL 16
- Docker / Docker Compose (opsional)

## Struktur proyek

```text
cmd/server/
internal/atrilinks/   # sign BLAKE3, client H2H, callback payload
internal/digiflazz/   # sign MD5 Digiflazz, mapper, webhook
internal/handler/
internal/service/
internal/repository/
internal/worker/      # poll pending → checktoday
migrations/
```

## Menjalankan

```bash
cp .env.example .env
docker compose up -d
```

## CI / Release

| Workflow | Trigger | Hasil |
|----------|---------|--------|
| `release.yml` | Tag `v*` | Binary multi-OS + `migrations/` |
| `docker-hub.yml` | `main` / tag `v*` | Image Docker Hub |

```bash
git tag v1.0.0 && git push origin v1.0.0
```

Docker Hub: `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN`.

## Lisensi

Belum ditetapkan — sesuaikan kebijakan organisasi Anda.
