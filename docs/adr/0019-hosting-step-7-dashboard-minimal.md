# ADR 0019: Hosting Step 7 — Dashboard Minimal (Control-plane API) untuk Login → Upload ZIP → Status/History → Rollback (Pointer KVS)

Tanggal: 2026-03-03  
Status: Accepted

## Context (Keadaan awal)
- Step 4 (Upload pipeline) sudah tersedia:
  - API generate pre-signed PUT URL ke S3, lalu `complete` melakukan `Head`, enforce size, enqueue message ke SQS.
- Step 5 (Worker Deploy Engine) sudah tersedia:
  - Worker consume SQS, ZIP security + extract, publish release immutable ke S3:
    - `sites/{site_id}/releases/{release_id}/...`
  - Worker update metadata release/site/upload di DynamoDB.
- Step 6 (Edge routing) sudah di-accept (ADR 0018):
  - CloudFront Function (viewer-request) rewrite berdasarkan CloudFront KeyValueStore (KVS):
    - Host → `{site_id, current_release_id, suspended}`
  - Rollback instan via pointer update `current_release_id` di KVS.
- Prinsip arsitektur yang harus dipertahankan:
  - AWS-first MVP: DynamoDB, S3, SQS, CloudFront (+ KVS).
  - Event-driven: Upload → enqueue → worker deploy (bukan deploy synchronous di API).
  - Immutable releases: tidak overwrite konten live.
  - Pointer-based rollback: ubah pointer, bukan pindah file.
  - Edge is king: routing di edge (CloudFront Function + KVS), bukan query DB per request.
  - Boundary hexagonal ketat dan kontrak publik stabil (router register + presenter error handling).
- Fakta repo:
  - Repo ini **tidak berisi UI**; web/dashboard berada di repo lain sebagai client yang consume API (lihat `docs/core/ARCHITECTURE.md`).

## Problem (Masalah)
Kita butuh “dashboard minimal” yang bisa dipakai orang awam end-to-end tanpa SSH:
1) Login aman.
2) Create/list site.
3) Upload ZIP (initiate → PUT → complete).
4) Lihat status deploy + history release.
5) Rollback/activate instan (pointer update KVS).

Hambatan yang muncul saat implementasi & verifikasi E2E:
- **Trust enforcement terlalu global**: seluruh protected routes ditahan oleh `trust.Enforce`/step-up, sehingga alur MVP (me + hosting) tidak bisa dipakai untuk AAL1.
- **Local E2E testing worker**: worker default jalan sebagai Lambda runtime; untuk local invoke perlu build-tag khusus (`localinvoke`).
- **Konfigurasi environment**: banyak error berasal dari shell yang tidak load `.env` (AWS_REGION kosong → endpoint `https://*.amazonaws.com` invalid; worker boot gagal karena `HOSTING_S3_BUCKET` tidak ada).
- **OAuth**: invalid_client jika client_id/secret tidak sesuai atau proses API belum restart.
- **Model data dashboard**: tabel DynamoDB hosting menggunakan PK tunggal; untuk list sites dan list releases per site tidak ada GSI (MVP butuh pendekatan yang tidak merusak skema).

## Decision (Keputusan)
Menerapkan Step 7 sebagai **control-plane API** yang dipakai UI repo lain, dengan komponen berikut:

### 1) API surface (v1/hosting) untuk dashboard minimal
Endpoint baru/yang distabilkan (semua di `/v1/hosting/*`):
- Sites:
  - `GET  /v1/hosting/sites`
  - `POST /v1/hosting/sites` (custom `site_id` atau auto-generate)
  - `GET  /v1/hosting/sites/:site_id`
- Upload flow (existing):
  - `POST /v1/hosting/sites/:site_id/uploads` (initiate)
  - `POST /v1/hosting/sites/:site_id/uploads/:upload_id/complete` (complete + enqueue)
- Releases:
  - `GET /v1/hosting/sites/:site_id/releases` (history)
  - `GET /v1/hosting/sites/:site_id/releases/:release_id` (status detail)
- Rollback/activate (pointer update):
  - `POST /v1/hosting/sites/:site_id/rollback` body: `{"release_id":"..."}`  
    Efek: update pointer KVS + update `sites.current_release_id` di DynamoDB.

### 2) Trust policy (Opsi C): enforcement per-subgroup
- Membuat `protectedBase` (common middleware: init/https/ua/jwt/aal score).
- Untuk low-risk MVP flows (`/v1/me` dan `/v1/hosting/*`):
  - **Tidak memakai `trust.Enforce`** (menghindari step-up requirement untuk AAL1).
- Untuk area lebih sensitif (`account`, `trust`, `billing`, `domains`):
  - Tetap memakai `trust.Enforce` dengan thresholds lebih ketat.
Tujuan: MVP hosting usable tanpa melemahkan kontrol di area high-risk.

### 3) Data model & ports (dashboard metadata)
- Menambah field `Host` dan `UpdatedAt` pada `domain.Site`.
- Menambah kemampuan `Put/List` pada `SiteStore`.
- Menambah `Get/ListBySite` pada `ReleaseStore`.
- Menambah port `EdgeStore` untuk update mapping host→(site_id,current_release_id,suspended) di KVS.
- Implementasi AWS KVS adapter memakai:
  - `DescribeKeyValueStore` untuk ETag
  - `PutKey` dengan `IfMatch` (ETag) untuk update value JSON.

### 4) Listing strategy (MVP): Scan + filter
Karena DynamoDB hosting tables saat ini hanya memiliki HASH key `pk` (tanpa GSI), maka:
- `ListSites` dilakukan via `Scan` dengan filter `begins_with(pk, "site#")` pada table sites.
- `ListBySite` releases dilakukan via `Scan` filter `site_id = :sid` pada table releases.
Catatan: ini MVP; akan dievaluasi untuk index/GSI pada fase berikut.

### 5) UX/Client placement: UI tetap di repo lain
Step 7 untuk repo ini fokus pada API + wiring + ports/adapters agar client UI bisa:
- Landing → login → pilih ZIP → upload → status → rollback,
tanpa menaruh UI code di repo `your-api`.

## Options (Opsi yang dipertimbangkan)

### A) UI di repo ini (embedded)
- (+) cepat demo
- (-) mencampur concern, merusak boundary, mengaburkan kontrak API, menyulitkan deploy terpisah  
Status: ditolak.

### B) UI di repo lain (client) + API minimal di sini ✅
- (+) sesuai `docs/core/ARCHITECTURE.md`
- (+) boundary jelas; API contract stabil
- (+) mudah evolusi UI tanpa risiko core infra
Status: dipilih.

### C) Trust enforcement global di semua protected routes
- (+) konsisten ketat
- (-) MVP hosting/me tidak usable untuk AAL1, memaksa step-up sebelum ada UX step-up yang siap
Status: ditolak.

### D) Trust enforcement per-subgroup (Opsi C) ✅
- (+) low-risk MVP flows usable
- (+) high-risk tetap ketat
- (+) struktur lebih audit-friendly
Status: dipilih.

### E) Local worker test: gunakan Lambda runtime env vars palsu
- (-) masuk loop runtime, tidak memproses file event
Status: ditolak.

### F) Local worker test: build-tag `localinvoke` ✅
- (+) sesuai struktur repo (build tag sudah ada)
- (+) bisa replay SQSEvent lokal deterministik
Status: dipilih.

### G) History query: tambah GSI sekarang vs Scan MVP
- Tambah GSI sekarang:
  - (+) scalable query
  - (-) perubahan infra + migrasi, memperbesar scope Step 7
- Scan MVP ✅:
  - (+) cepat, cukup untuk MVP/dev
  - (-) biaya dan performa akan memburuk saat data besar
Status: Scan MVP dipilih; GSI sebagai follow-up.

## Implementation Summary (Ringkasan implementasi)
- Router:
  - Menambah routes hosting (sites/releases/rollback) dan memisahkan subgroup `low/def/high`.
- Domain:
  - `Site.Host`, `Site.UpdatedAt`.
- Ports:
  - `SiteStore.Put/List`, `ReleaseStore.Get/ListBySite`, `EdgeStore.PutHostMapping`.
- DynamoDB store:
  - `SiteStore.Put/List` (scan prefix).
  - `ReleaseStore.Get/ListBySite` (scan by site_id).
- Usecase:
  - `Dashboard` (create/list/get site, list/get release, rollback).
- Adapters:
  - `CloudFrontKVS` implement `EdgeStore` (ETag + PutKey).
- Wiring:
  - `WireHostingDashboard` untuk construct dashboard usecase + handlers.
- Verifikasi E2E:
  - Login → create site → initiate → PUT S3 → complete → message masuk SQS → worker localinvoke memproses → release success → rollback update pointer → curl host publik menampilkan HTML.

## Verification (Acceptance / DoD)
Happy path end-to-end tanpa SSH:
1) Auth: login berhasil (OIDC callback 200).
2) Create site:
   - `POST /v1/hosting/sites` → 201 (atau 409 jika sudah ada).
3) Upload initiate:
   - `POST /v1/hosting/sites/{id}/uploads` → presigned PUT URL.
4) PUT ZIP ke S3 via presigned URL → 200/204.
5) Complete:
   - `POST /v1/hosting/sites/{id}/uploads/{upload_id}/complete` → status `queued`.
6) SQS:
   - Message terlihat di queue dengan `site_id/upload_id/release_id/object_key`.
7) Worker:
   - `go run -tags=localinvoke ./cmd/worker -event <SQSEvent>` → batch success.
8) Releases:
   - `GET /v1/hosting/sites/{id}/releases` → ada item `status=success`.
9) Activate/rollback:
   - `POST /v1/hosting/sites/{id}/rollback` → `current_release_id` berubah, `suspended=false`.
10) Public:
   - `curl -i https://{site}.asyrafcloud.my.id/` → 200 dan body sesuai release.

## Consequences (Dampak)

### Positif
- Step 7 goal tercapai: user awam bisa end-to-end tanpa SSH (dibuktikan dengan curl/flow dan CloudFront).
- Boundary tetap terjaga: domain/usecase tidak import vendor; vendor adapter via ports.
- Rollback instan sesuai Step 6: pointer-only update via KVS.
- Trust model lebih realistis: low-risk MVP usable, high-risk tetap ketat.

### Negatif / Risiko
- Listing via DynamoDB Scan tidak scalable untuk data besar (butuh follow-up GSI).
- Local testing worker perlu build tag `localinvoke` dan file SQSEvent; rawan kebingungan jika tidak didokumentasikan.
- Update KVS memakai ETag (IfMatch) → perlu strategi retry bila concurrent updates meningkat.
- Audit/observability belum lengkap → pindah ke Step 8.

## Follow-ups (Setelah Step 7)
- Step 8 (Observability + audit trail):
  - structured logs (job_id/site_id/release_id), metrics deploy duration/fail reasons/queue lag, audit events deploy+rollback.
- Step 9 (Abuse & safety):
  - rate limit upload/auth, DLQ policy, retry semantics, policy on extracted files (bila diperlukan).
- Dashboard UI (repo client):
  - landing + login + upload wizard + status/history + rollback button.
- Data model scalability:
  - rancang GSI untuk releases-by-site dan sites-by-owner (jika multi-tenant).
- Runbook:
  - dokumentasi localinvoke worker + debugging minimal.

## Notes (Catatan penting)
- Step 7 di repo ini secara scope adalah **control-plane API**; UI tetap di repo client.
- Trust enforcement dipisah per-subgroup agar MVP hosting usable tanpa memaksa step-up sebelum UX step-up tersedia.
- Semua error tetap lewat `presenter.HTTPErrorHandler` dan logging wajib redact (token/cookie/secret).
