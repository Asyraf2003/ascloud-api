# ADR 0020: Hosting Step 7 — Host Claim (Custom Subdomain) + Activation Policy (UI Auto-Activate) + Error Messaging Catalog (ID/EN)

Tanggal: 2026-03-03
Status: Accepted

## Context (Keadaan awal)
- Baseline arsitektur (AWS-first, event-driven, immutable releases, pointer rollback, edge routing via CloudFront Function + KVS) sudah ditetapkan di `docs/core/ARCHITECTURE.md` dan ADR-0018 (Edge Routing).
- Step 4–6 sudah terbukti E2E:
  - API: presigned upload (S3) + complete enqueue (SQS)
  - Worker: ZIP security + publish immutable release ke S3 prefix `sites/{site_id}/releases/{release_id}/...`
  - Edge: CloudFront Function rewrite berdasarkan KVS mapping host → `{site_id,current_release_id,suspended}`
  - Rollback/activate: update pointer `current_release_id` (tanpa memindahkan file) + update KVS
- Step 7 “dashboard minimal” (ADR 0019) menambahkan API surface:
  - Sites (list/create/get)
  - Upload flow (initiate/complete)
  - Releases (list/get)
  - Rollback/activate (pointer update KVS)
- Repo `your-api` tidak berisi UI; UI berada di repo lain sebagai client yang consume API.

## Problem (Masalah)
Kebutuhan UX terbaru untuk user awam:
1) User masuk landing page.
2) Login Google.
3) Upload ZIP.
4) ZIP diproses (async). Jika gagal: user dapat penjelasan “human readable”.
5) Jika sukses: user (opsional) mengisi custom subdomain; jika tidak, sistem memberi default.
6) Jika user refresh: tampil dashboard minimal (site + status + history + rollback).
7) Aktivasi/go-live setelah success harus terasa otomatis untuk user awam, tapi tetap sesuai prinsip immutable + pointer rollback.

Tantangan teknis yang harus dijawab secara eksplisit:
- Siapa yang memicu aktivasi pointer (update KVS) setelah release sukses?
- Bagaimana cara “claim” custom subdomain agar unik dan aman (tidak bisa dipakai 2 site)?
- Bagaimana menampilkan error deploy dengan bahasa manusia, dan bisa diganti bahasa (ID/EN)?
- Mana yang harus selesai di Step 7 (MVP) vs hutang teknis yang dirapikan di Step 9?

## Options (Opsi)
### Opsi A — UI auto-activate (dipilih)
UI melakukan:
- polling `GET /v1/hosting/sites/{site_id}/releases`
- jika ada release `status=success` yang ingin dipublish:
  - UI call `POST /v1/hosting/sites/{site_id}/rollback` dengan `release_id` yang dipilih

Kelebihan:
- Minim perubahan worker; aktivasi jadi bagian orchestration di client.
- Kegagalan aktivasi KVS bisa ditampilkan sebagai aksi yang bisa diulang (retry) tanpa mengubah semantic deploy success.
- Selaras dengan prinsip event-driven: worker fokus deploy artefak + update status; client/control-plane yang mengatur kapan go-live.

Kekurangan:
- Ada 1 call tambahan (success ≠ otomatis live sampai UI mengeksekusi aktivasi).
- Jika user menutup tab sebelum aktivasi, release sukses tapi belum live (butuh UX “resume”).

### Opsi B — Worker auto-activate
Worker saat deploy success langsung update KVS (go-live otomatis).
Kelebihan:
- UX “magis”: deploy sukses langsung live.
Kekurangan:
- Menambah failure mode (publish success tapi KVS gagal).
- Membebani worker dengan edge-control-plane side effect yang perlu observability lebih matang.
Status: tidak dipilih untuk Step 7.

## Decision (Keputusan)
Memilih **Opsi A (UI auto-activate)** untuk Step 7.

Tambahan keputusan terkait Step 7:
1) **Host claim (custom subdomain) harus unik** dan dilakukan via transaksi/conditional write di DynamoDB (bukan hanya KVS).
2) **Error deploy ditampilkan melalui catalog message di client (ID/EN)** berdasarkan `error_code` stabil dari API. Multi-violation (banyak error sekaligus) ditunda sebagai hutang teknis untuk Step 9.
3) Subdomain default dibuat otomatis oleh sistem (misal berdasarkan `site_id`), tetapi subdomain custom dapat diklaim setelah deploy success (sesuai UX).
4) Rename (mengganti subdomain) diperbolehkan, namun kebijakan “nama lama hilang” punya risiko takeover; mitigasi tahap lanjut akan dikerjakan kemudian (lihat Follow-ups).

## Design (Rancangan)

### A) Activation flow (UI-driven)
1) User upload → `complete` → status `queued`.
2) UI polling releases sampai ada `status=success`.
3) UI melakukan aktivasi:
   - `POST /v1/hosting/sites/{site_id}/rollback` body `{ "release_id": "<rid>" }`
4) API:
   - update `sites.current_release_id` di DynamoDB
   - update mapping host di KVS: `{site_id,current_release_id,suspended:false}`
5) Visitor publik langsung mengikuti pointer KVS.

Catatan:
- Aktivasi bisa dipakai untuk “go-live pertama” maupun rollback ke release sebelumnya.
- UI wajib menampilkan state yang jelas:
  - `queued` / `processing` (polling)
  - `success (ready to publish)` jika belum di-activate
  - `live` jika `current_release_id` == release_id aktif

### B) Host claim (custom subdomain) — Source of truth di DynamoDB
Konsep:
- `site_id` adalah ID internal yang stabil (tidak berubah).
- `host` adalah binding yang bisa diklaim user, dan harus unik.

Tambahan data model (DynamoDB):
- Table baru (atau struktur baru) untuk host claim:
  - `hosting-hosts` (nama final akan ditetapkan saat implementasi)
  - PK: `host#{fqdn}` (contoh `host#tokoku.asyrafcloud.my.id`)
  - Attr: `site_id`, `created_at`, `updated_at`

Aturan claim:
- Claim dilakukan dengan conditional write:
  - `ConditionExpression: attribute_not_exists(pk)`
- Jika host sudah ada → return error `HOST_TAKEN` (409).

Prosedur claim:
1) UI meminta host custom (mis `tokoku`):
   - API normalize + bentuk fqdn: `tokoku.<base_domain>`
2) API melakukan claim (conditional put).
3) Jika sukses:
   - update `sites.host` (metadata) agar dashboard menampilkan host utama site.
   - update KVS mapping untuk host baru agar mengarah ke site yang sama:
     - jika site sudah punya `current_release_id`, mapping host baru langsung live
     - jika belum, host baru tetap `suspended=true` atau `current_release_id=""` sesuai kebijakan.
4) (Opsional, fase lanjut) melepas host lama / membuat alias.

Kebijakan rename (MVP):
- MVP Step 7: boleh claim host custom sebagai “primary host”.
- Pelepasan host lama / cooldown alias ditunda sebagai follow-up policy (lihat Consequences & Follow-ups).

### C) Error messaging (human readable + bilingual)
Kontrak API:
- API tetap mengembalikan `error_code` (string stabil).
- API tidak wajib menyediakan message terlokalisasi untuk Step 7.

Catalog di client:
- UI menyimpan mapping `error_code -> message_id -> {id,en}`.
- UI menampilkan message “human readable” sesuai bahasa pilihan user.
- Bahasa default: Indonesia (ID), dapat diganti ke EN di UI.

Contoh (non-exhaustive):
- `hosting.zip_slip` → ID: "ZIP berbahaya: ada path tidak valid." / EN: "Unsafe ZIP: invalid path traversal."
- `hosting.zip_bomb` → ID: "ZIP terlalu besar setelah diekstrak." / EN: "ZIP expands beyond allowed size."
- `hosting.no_index_html` → ID: "Tidak ditemukan index.html di root." / EN: "index.html not found at root."
- `hosting.upload_too_large` → ID: "File upload melebihi batas." / EN: "Upload exceeds size limit."

Multi-error (multiple violations):
- Ditunda: saat ini `error_code` tunggal tetap dipakai.
- Upgrade ke `violations[]` akan masuk Step 9 (Abuse & safety baseline), karena butuh perubahan ZIP validation agar tidak fail-fast.

## Verification (Acceptance / DoD)
### 1) Aktivasi UI-driven (Opsi A)
- Setelah release `success`, UI dapat memanggil rollback endpoint:
  - `POST /v1/hosting/sites/{site_id}/rollback` dengan `release_id`
- Verifikasi:
  - `sites.current_release_id` berubah
  - KVS mapping host mengarah ke release baru
  - `curl -i https://{host}/` → 200 dan konten sesuai release

### 2) Host claim unik
- Claim host yang belum ada:
  - success (201/200)
  - host tampil sebagai primary host pada site
- Claim host yang sudah dipakai site lain:
  - gagal 409 `HOST_TAKEN`
- Verifikasi edge:
  - host baru resolve ke site/release yang sama (jika sudah activated)

### 3) Error messaging bilingual
- Jika deploy gagal dan API mengembalikan `error_code`, UI menampilkan message bahasa ID.
- Jika user switch bahasa ke EN, message berubah tanpa mengubah `error_code`.

## Consequences (Dampak)
### Positif
- UX awam tercapai tanpa menambah kompleksitas worker: aktivasi dilakukan UI.
- Host claim aman karena uniqueness dijamin via DynamoDB conditional write.
- Error message dapat dibuat “human readable” dan bilingual tanpa mengubah kontrak API.

### Negatif / Risiko
- Release sukses tidak otomatis live jika UI tidak melakukan aktivasi (tab ditutup, crash, dsb).
  - Mitigasi: UI “resume” dengan polling dan tombol publish/retry; simpan state minimal (site_id/release_id) di local storage.
- Rename/subdomain “nama lama hilang” berisiko takeover host lama.
  - Mitigasi (fase lanjut): cooldown/alias period sebelum host lama dilepas, atau larang rename setelah first activation tanpa cooldown.
- Listing releases/sites berbasis Scan (MVP) memiliki risiko scaling, namun bukan fokus ADR ini.

## Follow-ups (Rencana)
### Milestone 8 (Observability + audit trail)
- Pastikan setiap:
  - enqueue deploy, deploy success/fail, rollback/activate, host claim
  dicatat sebagai audit event + structured logs dengan `site_id`, `upload_id`, `release_id`, `host`.
- Tambahkan metric minimal agar bisa jawab “kenapa deploy gagal?” tanpa debugging lama.

### Milestone 9 (Abuse & safety baseline) — hutang teknis
- Multi-violation errors:
  - ubah ZIP validation dari fail-fast menjadi collect violations
  - expose `violations[]` (allowlisted) untuk UI
- Rate limit upload/auth dan retry/DLQ policy yang lebih ketat.
- Kebijakan rename host yang aman:
  - cooldown alias / grace period
  - aturan kapan host lama dilepas

### Dokumentasi
- `docs/core/ARCHITECTURE.md` tetap high-level.
- Detail kebijakan host claim, aktivasi, dan error catalog dirujuk dari ADR ini.
