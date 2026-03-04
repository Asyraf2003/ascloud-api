# ADR 0022: Hosting Step 9 — Abuse & Safety Baseline (Rate Limit, Upload Content-Type, Multi-Violation ZIP, Extracted File Sanitization, Retry Semantics)

Tanggal: 2026-03-05  
Status: Accepted

## Context (Keadaan awal)
- Step 4–8 sudah tersedia dan terbukti E2E:
  - Upload pipeline: API presigned PUT (S3) + complete enqueue (SQS).
  - Worker deploy engine: ZIP security + extract + publish immutable release ke S3 (`sites/{site_id}/releases/{release_id}/...`).
  - Edge routing: CloudFront Function rewrite berdasarkan CloudFront KeyValueStore (KVS) host → `{site_id,current_release_id,suspended}`.
  - Dashboard control-plane API: sites/releases/rollback.
  - Observability: request_id mengalir API→SQS→worker + audit trail + EMF metrics + DLQ/alarm minimal.
- Repo constraints:
  - AWS-first MVP (CloudFront + S3 + Lambda + SQS + DynamoDB), provider lain inactive.
  - Event-driven deploy, immutable releases, pointer-based rollback, edge is king.
  - Hexagonal boundary ketat; kontrak publik stabil:
    - `internal/transport/http/router.Register(*echo.Echo)`
    - `internal/transport/http/router/v1.Register(*echo.Echo)`
    - `internal/transport/http/presenter.HTTPErrorHandler(error, echo.Context)`
  - Error response ke user wajib envelope via presenter; logging redact header sensitif; tidak ada raw leak.

## Problem (Masalah)
Sebelum public launch, diperlukan baseline defense agar layanan tidak ambruk oleh serangan “bodoh” dan input berbahaya:
1) Spam request (brute refresh / spam upload initiate+complete / spam rollback).
2) Upload non-zip / salah Content-Type yang memicu error/biaya.
3) ZIP berbahaya:
   - path traversal (zip slip)
   - symlink/special file
   - terlalu banyak file / terlalu deep
   - over quota / zip bomb style (expand melebihi kuota)
4) File extracted berbahaya (executable/script) yang seharusnya tidak dilayani oleh static hosting.
5) UX debugging deploy error:
   - sebelumnya error ZIP cenderung fail-fast (hanya 1 code),
   - padahal UI butuh `violations[]` (multi-violation) untuk pesan human-readable (ID/EN) di client.

## Decision (Keputusan)

### A) Rate limiting: per-endpoint pada hosting routes (Balanced)
- Memakai middleware existing `trust.RateLimit(name, limit, window)` (key: account_id, fallback IP).
- Dipasang per-route pada `/v1/hosting/*` agar tiap endpoint bisa beda limit.

Balanced limits (per acc, fallback IP):
- Read (GET sites/releases): 120/min
- Create site (POST /hosting/sites): 30/min
- Initiate upload (POST /uploads): 30/min
- Complete upload (POST /complete): 20/min
- Rollback (POST /rollback): 10/min

Catatan:
- `trust.RateLimit` adalah in-memory per process (guardrail). Untuk public-grade, layer WAF/CloudFront rate-based rule direkomendasikan sebagai follow-up.

### B) Validasi tipe konten upload: enforce di S3 presigned PUT + hint ke client
- Adapter AWS `PresignPutZip` mengunci `ContentType: application/zip` pada presigned PUT.
- API `initiate upload` menambahkan field hint agar client wajib set header yang benar saat PUT:
  - `required_content_type: "application/zip"`

Hasil:
- Upload dengan `Content-Type` selain `application/zip` akan ditolak oleh S3 (signature mismatch/403), tanpa membebani API.

### C) Multi-violation ZIP: collect violations (bukan fail-fast), simpan dan expose
- ZIP validation diubah untuk mengumpulkan pelanggaran (violations) dan mengembalikan:
  - single violation: sentinel error langsung (compatibility test/behavior),
  - multiple violations: `ViolationsError` yang `Unwrap() []error` sehingga `errors.Is()` tetap bekerja.
- Deployer memetakan violations menjadi:
  - `error_code` (primary) = violation paling “severe” berdasarkan urutan severity,
  - `violations[]` = daftar semua code violation yang terdeteksi.
- Persist:
  - `Release` di DynamoDB menyimpan attribute `violations` (list string).
- Expose:
  - API release response menambahkan `violations` (omitempty).

Severity order (primary lebih dulu):
1) `hosting.zip_slip`
2) `hosting.zip_symlink`
3) `hosting.file_disallowed`
4) `hosting.zip_too_many_files`
5) `hosting.zip_too_deep`
6) `hosting.extract_over_quota`

### D) Sanitasi extracted files: denylist minimal executable/script
- Pada extraction, file dengan ekstensi denylist dianggap violation:
  - `.exe .dll .so .dylib .bat .cmd .ps1 .sh`
- Violation code:
  - `hosting.file_disallowed`

### E) Retry/DLQ semantics: violations dianggap permanent (tidak retry)
- Worker sudah punya `isPermanent(code)` untuk menentukan retry.
- Semua code violation (zip/sanitize/quota) termasuk permanent:
  - worker akan ACK message (tidak retry, tidak dorong ke DLQ).
- Error transient (S3/DDB/network) tetap retry → DLQ bila melewati `maxReceiveCount`.

## Options (Opsi yang dipertimbangkan)

### 1) Rate limiting: group-level vs per-endpoint
- Group-level:
  - (+) mudah
  - (-) semua endpoint kena limit sama; tidak cocok untuk upload vs read
- Per-endpoint ✅:
  - (+) granular; upload/rollback lebih ketat
  - (-) lebih verbose

### 2) Rate limiting: in-process vs WAF
- In-process (`trust.RateLimit`) ✅ (dipilih untuk baseline code)
  - (+) cepat, konsisten 429 + envelope
  - (-) tidak global pada serverless
- WAF rate-based (follow-up)
  - (+) global di edge, paling efektif
  - (-) butuh provisioning infra

### 3) Content-Type validation: API check vs presign enforcement
- API check:
  - (-) tidak memvalidasi PUT langsung ke S3; hanya memvalidasi request API
- Presign enforcement ✅:
  - (+) validasi terjadi di S3, paling dekat ke sumber data
  - (+) mengurangi biaya/kerja API

### 4) ZIP errors: fail-fast vs collect violations
- Fail-fast:
  - (+) sederhana
  - (-) UI tidak dapat banyak konteks, debugging lama
- Collect violations ✅:
  - (+) UI bisa tampilkan banyak pelanggaran sekaligus
  - (+) `errors.Is()` tetap bekerja dengan multi-unwrapped errors
  - (-) perlu desain agar tidak jadi DoS (quota/too-many-files dijadikan terminal)

### 5) Sanitasi extracted: allowlist vs denylist
- Allowlist:
  - (+) security tinggi
  - (-) rawan false positive (fonts/wasm/dll scenario)
- Denylist minimal ✅:
  - (+) MVP-friendly; blok yang paling berbahaya dulu
  - (-) lebih longgar daripada allowlist

### 6) Retry semantics: retry violations vs permanent
- Retry violations:
  - (-) membuang biaya; poison-pill masuk loop hingga DLQ
- Permanent ✅:
  - (+) cepat selesai; tidak mengganggu queue
  - (+) status deploy gagal jelas di release metadata

## Implementation Summary (Ringkasan implementasi)
Perubahan utama:
- Router hosting:
  - Tambah `trust.RateLimit(...)` per endpoint untuk `/v1/hosting/*`.
- Upload initiate response:
  - Tambah `required_content_type` di response agar client set `Content-Type` benar.
- ZIP validation:
  - `zipsec` menambah `ViolationsError` (multi unwrap).
  - Extraction collect violations + denylist ext → `ErrDisallowedFile`.
- Mapping violations:
  - `zipErrCodes(err)` menghasilkan `[]string` codes dari multi errors.
  - Deployer menulis primary `error_code` + `violations[]`.
- Persistence:
  - `domain.Release` menambah `Violations []string`.
  - DynamoDB release item menambah attribute `violations` (list string).
- Worker:
  - `isPermanent` menambah `hosting.file_disallowed`.

Boundary/kontrak:
- Tidak mengubah kontrak publik router/presenter.
- Tidak menambah debug routes.
- Error response tetap via `presenter.HTTPErrorHandler` (no raw leak).

## Verification (Acceptance / DoD)
### Commands (wajib)
- `gofmt -w .`
- `go test ./... -count=1`
- `go vet ./...`
- `make audit`

### Sanity checks
1) Rate limit hosting:
   - Burst request ke `POST /v1/hosting/sites/:id/uploads` → 429 envelope `RATE_LIMITED`.
2) Content-Type enforcement:
   - Presigned PUT tanpa `Content-Type: application/zip` → gagal (403) dari S3.
3) Multi-violation:
   - Deploy ZIP yang memicu >1 violation → release `status=failed`, `error_code` primary, dan `violations[]` terisi (GET release detail).
4) Permanent vs transient:
   - Error violation ZIP/sanitize tidak retry (worker batch tidak menandai failure item).
   - Error transient (mis. S3 get failed) retry (worker batch menandai failure item) dan akhirnya DLQ bila melebihi maxReceiveCount.
5) No leaks:
   - Error response ke client tetap JSON envelope bersih; log redact header sensitif.

Cek lulus milestone:
- Serangan bodoh (spam upload/brute refresh/zip berbahaya) tidak bikin layanan ambruk; kegagalan bersifat deterministik dan tercatat.

## Consequences (Dampak)

### Positif
- Baseline defense tercapai sebelum publik:
  - rate limit granular,
  - enforcement content-type di S3,
  - multi-violation ZIP + sanitasi file,
  - retry semantics menjaga queue tetap sehat.
- UX debugging meningkat:
  - UI bisa memetakan `error_code` + `violations[]` ke pesan human-readable bilingual.
- Biaya lebih terkendali:
  - input berbahaya ditolak lebih awal dan tidak masuk retry loop.

### Negatif / Risiko
- `trust.RateLimit` in-memory bukan global; skala serverless besar masih perlu WAF.
- Denylist minimal bisa perlu tuning (false positive/negative) seiring fitur bertambah.
- Urutan severity (primary code) perlu konsistensi bila katalog violation berkembang.

## Follow-ups (Setelah Step 9)
- Tambah WAF rate-based rules di CloudFront untuk `/v1/hosting/*` dan `/v1/auth/*` (hybrid defense).
- Pertimbangkan allowlist/extended policy untuk file types (fonts/wasm) bila dibutuhkan.
- Dokumentasi client:
  - wajib set `Content-Type` dari `required_content_type` pada presigned PUT.
- Hardening policy host rename (cooldown/alias/grace) dan launching ops (Milestone 10).
