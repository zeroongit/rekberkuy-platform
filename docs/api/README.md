# API Collection — Rekberkuy Core Service

Postman collection untuk Rekberkuy Core Service.

## File

- [`rekberkuy-core-service.postman_collection.json`](./rekberkuy-core-service.postman_collection.json)

## Cara Import

1. Buka Postman → **Import** → **File** → pilih JSON di atas.
2. Buat environment baru dengan variabel berikut:

| Variable | Contoh Nilai | Keterangan |
|---|---|---|
| `base_url` | `http://localhost:8080` | Base URL service |
| `token` | `eyJ...` | JWT hasil login |

3. Pilih environment di kanan atas Postman.

## Autentikasi

Semua endpoint (kecuali `/health`, `/login`) butuh header: Authorization: Bearer {{token}}


## Alur Testing yang Disarankan

1. `POST /login` → copy `token` ke environment
2. `GET /me` → verifikasi token valid
3. ... (lanjutkan sesuai endpoint yang tersedia)

## Referensi

- Alur lengkap: [Application Flow & Module Guide](../application-flow-and-module-guide.md)
- Keputusan auth: [ADR-0001](../adr/0001-local-jwt-credentials-instead-of-supabase-auth.md)