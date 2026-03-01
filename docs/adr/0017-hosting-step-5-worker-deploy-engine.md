# ADR 0017: Hosting Step 5 — Worker Deploy Engine (SQS → ZIP Security → Immutable Release Publish)

Tanggal: 2026-03-01  
Status: Accepted

## Context (Keadaan awal)
- Step 4 (Upload Pipeline) sudah selesai dan lolos `go test ./...`, `go vet ./...`, `make audit`.
- Baseline MVP (AWS-first, event-driven) sudah disepakati:
  - Queue: SQS
  - Object storage: S3
  - Metadata DB: DynamoDB
  - Release model: immutable releases + pointer rollback
  - Quota: HARD FAIL jika extracted total > 20MB
- Kontrak publik & boundary ketat (hexagonal) wajib dipertahankan:
  - Domain/usecase tidak boleh bergantung vendor
  - Vendor implementation hanya di adapters/platform
  - Semua error user-facing lewat presenter (tidak tersentuh di Step 5)
- Repo memiliki quality gate `make audit` yang memaksa:
  - `go test` + tag component tertentu
  - audit boundaries/testtags/docs/content
  - *hard fail* bila `.go` file > 100 lines (sehingga implementasi harus dipecah per tanggung jawab)

## Problem (Masalah)
Kita perlu mengimplementasikan “deploy engine” di worker yang:
1) Mengkonsumsi job `DeployMessage` (SQS) dan memproses ZIP dari S3 (uploads).
2) Menjalankan ZIP security non-negotiable:
   - anti zip-slip (path traversal)
   - reject symlink/hardlink/special files
   - anti zip-bomb (limit total extracted bytes, jumlah file, depth)
   - hard fail jika extracted total > 20MB
3) Mempublish release immutable ke S3:
   - `sites/{site_id}/releases/{release_id}/...`
   - tidak overwrite konten live
4) Mengupdate metadata di DynamoDB:
   - Upload status: `deployed` / `failed`
   - Release status: `success` / `failed` + `error_code`
   - Site pointer: `current_release_id = release_id` (untuk Step 6)
5) Memiliki retry semantics yang benar:
   - permanent errors: tidak perlu retry
   - transient errors: minta retry SQS/Lambda (batch item failure)
6) Bisa dibuktikan E2E (sanity) tanpa “debug endpoint”.

## Options (Opsi)
### Opsi 1 — Worker Lambda SQS handler + deploy engine di `hosting/usecase` (event-driven)
- SQS event → handler → call `Deployer.Deploy(msg)`
- Download ZIP dari S3 ke `/tmp` dengan max byte limit
- Extract aman ke `/tmp/release-{release_id}` memakai library internal `zipsec`
- Upload setiap file ke S3 prefix release immutable
- Update DDB: Upload/Release/Site
- Best-effort delete ZIP upload
- Retry semantics: batch item failures untuk transient only

**Kelebihan**
- Konsisten dengan baseline event-driven dan serverless (cost idle ~0).
- IO-heavy ditangani worker, bukan API.
- Boundary tetap bersih: usecase tidak import AWS SDK.
- Bisa dites unit + sanity.

**Kekurangan**
- Upload publish per-file ke S3 butuh banyak PutObject (acceptable untuk MVP).
- Idempotency perlu disiplin status+keys (ditangani melalui status & immutable prefix).
- Per-item retry butuh infra setting “ReportBatchItemFailures” (config Lambda/SQS).

### Opsi 2 — Deploy synchronous di API (proxy ZIP / extract / publish)
**Kelebihan**
- Lebih mudah bagi client: satu request deploy selesai.

**Kekurangan**
- Melanggar baseline event-driven (upload → queue → worker).
- Mahal (bandwidth/CPU/latency) dan memperbesar blast radius API.
- Risiko timeout dan resource exhaustion lebih besar.
- Tidak serverless-friendly.

### Opsi 3 — Worker non-Lambda (polling SQS / ECS/EC2)
**Kelebihan**
- Kontrol runtime lebih bebas.

**Kekurangan**
- Tidak sesuai MVP baseline (EC2 dilarang, lebih mahal/operasional).
- Bertentangan dengan target “Lambda (API+Worker)”.

## Decision (Keputusan)
Memilih **Opsi 1**.

Implementasi dipecah menjadi 4 lapisan (sesuai struktur repo):
1) **Ports (`internal/modules/hosting/ports`)**
   - Extend `ObjectStore` untuk kebutuhan Step 5:
     - `Get(objectKey)`, `Put(objectKey, body, contentType, cacheControl)`, `Delete(objectKey)`
   - Extend store ports untuk pointer & status:
     - `SiteStore.UpdateCurrentRelease(siteID, releaseID)`
     - `UploadStore.UpdateStatusSizeAndReleaseID(uploadID, status, sizeBytes, releaseID)`
     - `ReleaseStore.Put(release)`, `ReleaseStore.UpdateStatus(releaseID, status, sizeBytes, errorCode)`
   - `DeployMessage` membawa `site_id`, `upload_id`, `release_id`, `object_key`, `size_bytes`, `queued_at_unix`.

2) **ZIP Security Library (`internal/modules/hosting/usecase/zipsec`)**
   - `zipsec.Extract(ctx, zipPath, destDir, options)`:
     - Path normalization + reject traversal (`../`, absolute, backslash windows)
     - Reject symlink/special file
     - Enforce:
       - `MaxTotalBytes` (20MB default)
       - `MaxFiles`
       - `MaxDepth`
       - `MaxFileBytes`
     - Error types stabil:
       - `ErrZipSlip`, `ErrZipSymlink`, `ErrTooManyFiles`, `ErrTooDeep`, `ErrOverQuota`
   - Semua file dipaksa <100 LOC dengan pemecahan file.

3) **Deploy Engine Usecase (`internal/modules/hosting/usecase`)**
   - `Deployer.Deploy(ctx, ports.DeployMessage)`:
     - Validate message minimal (`site_id/upload_id/release_id/object_key` wajib)
     - Load upload record; enforce `status == queued` dan `site_id` match
     - Download ZIP dari S3 → local temp file (max bytes)
     - Extract via `zipsec` → local extract dir
     - Publish ke S3 immutable prefix:
       - key: `sites/{site_id}/releases/{release_id}/{relativePath}`
       - content-type via `mime.TypeByExtension` (fallback octet-stream)
       - cache-control default: `public, max-age=31536000, immutable`
     - Update DDB:
       - Release: `success` + `size_bytes`
       - Site: `current_release_id=release_id`
       - Upload: `deployed` + `size_bytes` + `release_id`
     - Best-effort delete ZIP upload (`ObjectStore.Delete(object_key)`).
   - Error codes untuk permanent classification:
     - `hosting.zip_slip`, `hosting.zip_symlink`, `hosting.zip_too_many_files`, `hosting.zip_too_deep`, `hosting.extract_over_quota`, `hosting.zip_too_large`, dll.
   - Error dibungkus `internal/shared/apperr.AppError` untuk code yang stabil.

4) **Worker Runtime (`cmd/worker`)**
   - Lambda SQS handler:
     - Parse `events.SQSEvent`
     - Call deployer per record
     - Retry policy:
       - permanent → tidak masuk `batchItemFailures`
       - transient/unknown → masuk `batchItemFailures`
   - Karena versi `aws-lambda-go` di environment tidak menyediakan `events.SQSBatchResponse`, response diimplementasikan sebagai type internal:
     - JSON shape `{"batchItemFailures":[{"itemIdentifier":"..."}]}`
   - Tambahan dev-only tool untuk sanity tanpa deploy Lambda:
     - `go run -tags=localinvoke ./cmd/worker -event ./event.json`
     - diproteksi build tag supaya tidak mengganggu binary produksi.

## AWS Resources (Dev sanity)
Untuk verifikasi E2E (dev):
- DynamoDB tables dibuat via CloudFormation stack `your-api-dev-ddb`:
  - `your-api-dev-hosting-uploads`
  - `your-api-dev-hosting-sites`
  - `your-api-dev-hosting-releases`
- S3 bucket hosting dibuat (unik) dan diverifikasi exist:
  - contoh: `your-api-dev-hosting-<account>-<suffix>`
- SQS queue deploy dibuat:
  - `your-api-dev-hosting-deploy`
- Env non-secret dipakai di runtime worker:
  - `HOSTING_S3_BUCKET`
  - `DDB_HOSTING_UPLOADS_TABLE`
  - `DDB_HOSTING_SITES_TABLE`
  - `DDB_HOSTING_RELEASES_TABLE`
  - (SQS URL dipakai untuk Step 4 enqueue; sanity Step 5 bisa invoke langsung via localinvoke)

## Verification (Acceptance / DoD)
### Quality gates
- `gofmt -w .`
- `go test ./... -count=1`
- `go vet ./...`
- `make audit` (harus hijau, termasuk content/docs/testtags, dan no `.go` >100 LOC)

### Sanity E2E (dibuktikan)
1) **Happy path sukses**
   - Input: ZIP berisi `index.html`
   - Result:
     - Upload status `deployed`
     - Release status `success`
     - Site `current_release_id` ter-update
     - S3 object ada di:
       - `sites/{site_id}/releases/{release_id}/index.html`
2) **Security rejection: zip-slip**
   - Input: ZIP entry literal `../evil.txt` (dibuat via Python `zipfile.writestr`)
   - Result:
     - Upload status `failed`
     - Release status `failed` dengan `error_code=hosting.zip_slip`
     - S3 release prefix kosong (tidak publish)

## Consequences (Dampak)
### Positif
- Milestone 5 deliverables tercapai sesuai baseline:
  - worker deploy engine async, secure, immutable releases, pointer update
- Boundary tetap bersih:
  - deploy logic ada di usecase, AWS SDK hanya di adapter aws
- Error code stabil mendukung retry policy dan observability di Step 8.
- Ada jalur sanity tanpa deploy Lambda (localinvoke build tag).

### Negatif / Risiko
- Publish per-file ke S3 bisa menjadi bottleneck untuk ZIP besar (acceptable untuk MVP).
- Retry semantics per-item butuh konfigurasi event source mapping (SQS→Lambda) agar “ReportBatchItemFailures” efektif.
- Idempotency lanjutan (double-processing) masih bergantung pada status DDB + immutable prefix; dapat diperkuat di ADR/Step observability.

## Follow-ups
- Step 6: Edge routing (CloudFront Function + KVS) akan memakai `site_id` dan `current_release_id` yang sudah ditulis Step 5.
- Step 8: tambah metrics, audit trail, dan alarm (DLQ, worker failures, deploy duration, fail reasons).
- Infra: pastikan SQS→Lambda enable per-item failure reporting dan DLQ policy.
