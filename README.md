# vendor-digiflazz

Addon **Seller API** untuk [Digiflazz](https://developer.digiflazz.com/api/seller/prabayar/topup/): menerima request topup/cek status, memvalidasi signature, menyimpan transaksi, meneruskan order ke upstream **[Atrilinks H2H](https://h2h.atrilinks.com/docs)**, lalu mengembalikan response dan webhook sesuai kontrak Digiflazz.

## Fitur

- **Topup & cek status** — idempoten berdasarkan `ref_id` (topup ulang dengan `ref_id` sama mengembalikan state dari database, tanpa order ulang ke Atrilinks).
- **Validasi Digiflazz** — verifikasi `sign` = `md5(username + apiKey + ref_id)` dan `username` dari environment.
- **Integrasi Atrilinks (Query/PIN)** — transaksi prabayar fixed (`jenistrx=1`) via `GET /h2h/trx/query`, dengan `urlcallback` ke addon.
- **Callback Atrilinks** — update status transaksi dari notifikasi upstream; webhook ke Digiflazz saat status final (sukses/gagal).
- **Mapping status** — Atrilinks (`success` / `pending` / `failed`) → Digiflazz (`status` `0`/`1`/`2` + `rc`); RC dan status numerik hanya format Digiflazz, bukan dari upstream.
- **PostgreSQL** — persistensi transaksi, audit response upstream (JSON), pelacakan webhook.

## Alur singkat

```text
Digiflazz  →  POST topup (addon)
Addon      →  validasi sign → DB (ref_id) → Atrilinks query (+ urlcallback)
Atrilinks  →  callback status → addon → update DB
Addon      →  webhook Digiflazz (jika sukses/gagal)
Digiflazz  ←  response { "data": { ... } }
```

## Endpoint HTTP (addon)

| Method | Path | Keterangan |
|--------|------|------------|
| `POST` | `/api/digiflazz/topup` | Request dari Digiflazz (JSON) |
| `POST` | `/h2h/atrilinks/callback` | `urlcallback` dari Atrilinks (sesuaikan dengan `ADDON_ATRILINKS_CALLBACK_URL`) |
| `GET` | `/health` | Health check |

Response ke Digiflazz selalu dibungkus `{ "data": { ... } }`.

## Dokumentasi

| Sumber | Isi |
|--------|-----|
| [Digiflazz — Topup / Cek Status](https://developer.digiflazz.com/api/seller/prabayar/topup/) | Request, signature, field response |
| [Digiflazz — Status](https://developer.digiflazz.com/api/seller/prabayar/status/) | `0` pending, `1` sukses, `2` gagal |
| [Digiflazz — Response Code](https://developer.digiflazz.com/api/seller/prabayar/response-code/) | `rc` (`00`, `39`, `07`, `204`, …) |
| [Digiflazz — Webhooks](https://developer.digiflazz.com/api/seller/prabayar/webhook/) | Callback ke Digiflazz saat status berubah |
| [Atrilinks H2H](https://h2h.atrilinks.com/docs) | Docs resmi upstream |
| [`docs-atrilinks.txt`](./docs-atrilinks.txt) | Rangkuman API Atrilinks di repo (Query, parameter, contoh response) |

Konvensi implementasi dan mapping detail juga ada di `.cursor/rules/digiflazz-vendor-atrilinks.mdc` (untuk pengembangan dengan Cursor).

## Konfigurasi

Salin [`.env.example`](./.env.example) ke `.env`. Variabel utama:

- **Server:** `HTTP_PORT`, `DATABASE_URL`
- **Digiflazz:** `DIGIFLAZZ_API_USERNAME`, `DIGIFLAZZ_API_KEY`, `DIGIFLAZZ_CALLBACK_URL`
- **Atrilinks:** `ATRILINKS_BASE_URL`, `ATRILINKS_USER_ID`, `ATRILINKS_PIN`, `ATRILINKS_PASS` (opsional), `ADDON_ATRILINKS_CALLBACK_URL`

Jangan commit file `.env`. Rahasia hanya di environment runtime.

**Mapping produk:** `pulsa_code` Digiflazz saat ini dipakai sebagai `kodeproduk` Atrilinks (huruf kecil). Tabel mapping terpisah dapat ditambahkan jika diperlukan.

## Tech stack

- Go 1.22+
- PostgreSQL 16
- Docker / Docker Compose (deploy opsional)

## Struktur proyek

```text
cmd/server/           # entrypoint HTTP
internal/config/      # env
internal/handler/     # Digiflazz + callback Atrilinks
internal/service/     # orkestrasi transaksi
internal/digiflazz/   # sign, mapper, webhook client
internal/atrilinks/   # client H2H
internal/repository/  # PostgreSQL
internal/database/    # koneksi DB (retry startup)
migrations/           # skema SQL
```

## Menjalankan

**Docker Compose (disarankan):**

```bash
cp .env.example .env
# isi kredensial
docker compose up -d
```

**Lokal (Go):** Postgres harus reachable dari `DATABASE_URL`; lihat komentar di `.env.example` dan `docker-compose.yml`.

**CI:** workflow GitHub Actions mem-build dan push image ke Docker Hub (secrets `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN`).

## Lisensi

Belum ditetapkan — sesuaikan kebijakan organisasi Anda.
