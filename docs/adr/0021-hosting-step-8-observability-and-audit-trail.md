# ADR 0021: Hosting Step 8 — Observability + Audit Trail (Request-ID, Audit Events, EMF Metrics, Alarms, DLQ)

Tanggal: 2026-03-04  
Status: Accepted

## Context (Keadaan awal)
- Repo `your-api` menjalankan control-plane API (Echo) + worker deploy engine (SQS consumer).
- Step 4–7 sudah tersedia dan terbukti E2E:
  - Upload presigned S3 + `complete` enqueue ke SQS
  - Worker extract ZIP + publish immutable release
  - Dashboard API (sites/releases/rollback)
- Konstitusi arsitektur:
  - AWS-first MVP (S3, SQS, DynamoDB, CloudFront+KVS)
  - Event-driven deploy, immutable releases, pointer-based rollback
  - Hexagonal boundary ketat (domain/usecase tidak import vendor)
  - Error contract via `presenter.HTTPErrorHandler`, redaction wajib, no raw leaks
  - Debug routes dilarang kecuali gated `DEBUG_ROUTES=1`
- Kondisi observability sebelum Step 8:
  - `request_id` ada di envelope dan middleware, namun belum konsisten mengalir ke SQS/worker.
  - Audit sink sudah ada untuk auth (DynamoDB) dengan schema: `pk`, `sk`, `event`, `at`, `meta_json`.
  - Alarm & DLQ belum terpasang secara nyata di AWS account/region yang benar.
  - Ada risiko kebingungan region (hosting resources berada di `ap-southeast-3`, sementara ada Lambda lain di `ap-southeast-1`).

## Problem (Masalah)
Target Step 8:
1) **Structured logging + request id** agar tracing lintas API→Queue→Worker deterministik.
2) **Metrics penting** untuk jawaban cepat:
   - deploy duration
   - fail reasons
   - queue lag
3) **Audit trail** untuk action penting:
   - deploy enqueue / deploy success/fail
   - rollback/activate
   - create site
4) **Alarm minimal**:
   - backlog/lag
   - DLQ
   - worker failures

Hambatan yang ditemukan selama implementasi:
- SQS queue berada di region `ap-southeast-3`, namun CLI default region user `ap-southeast-1` → `list-queues` terlihat kosong bila region salah.
- Default table name di code untuk audit adalah fallback `"audit_events"`, sedangkan nama table aktual di AWS adalah `your-api-dev-audit-events`.
- Endpoint hosting dilindungi middleware `JWTAuth()` (wajib `Authorization: Bearer <token>`), sehingga trigger audit hosting tidak akan muncul bila request tidak authenticated.
- Tanpa IaC untuk CloudWatch alarms/DLQ, provisioning perlu dilakukan via CLI (manual) agar milestone bisa selesai.

## Decision (Keputusan)

### A) Correlation ID end-to-end: API → SQS → Worker
- Menetapkan `request_id` sebagai correlation key utama.
- Middleware `RequestID`:
  - Jika `X-Request-ID` kosong, generate baru.
  - Set ke request header dan response header.
  - Inject ke context request (shared `requestid` package).
- Pada `CompleteUpload`:
  - Ambil `request_id` dari context dan masukkan ke `ports.DeployMessage.RequestID`.
- Pada worker handler:
  - Baca `msg.RequestID` dari SQS body.
  - Inject ke context dengan `requestid.With(ctx, msg.RequestID)`.
  - Log worker selalu include `lambda_request_id` + `request_id` + `site_id/upload_id/release_id`.

**Alasan:**
- Ini membuat tracing deterministik tanpa bergantung tracing vendor/OTEL di MVP.

### B) Audit Trail hosting: reuse table audit yang sudah ada (schema pk/sk)
- Menambahkan hosting audit sink dengan schema item yang sama:
  - `pk`, `sk`, `event`, `at`, `meta_json`
- Partition key hosting:
  - `pk = site#<site_id>` untuk hosting events.
- Partition key auth tetap:
  - `pk = acc#<account_id>` atau `acc#anon`.
- `meta_json` diserialisasi dari `meta map` setelah redaction (`redact.RedactMap`).

**Alasan:**
- Tidak membuat table baru (lebih kecil scope).
- Query audit per-site menjadi sederhana (`pk=site#...`).
- Tidak bentrok dengan auth karena prefix `site#` vs `acc#`.

### C) Event catalog (minimal, P0) untuk hosting audit
Events yang di-record:
- API:
  - `hosting_upload_complete_enqueued` (CompleteUpload sukses enqueue)
  - `hosting_site_created` (CreateSite sukses)
  - `hosting_rollback_requested`
  - `hosting_rollback_succeeded`
  - `hosting_rollback_failed` (+ stage + error_code)
- Worker/Deployer:
  - `hosting_deploy_succeeded`
  - `hosting_deploy_failed` (+ error_code)

Catatan:
- Error code untuk audit diambil dari `apperr` jika tersedia; fallback `hosting.internal_error`.
- Audit bersifat **best-effort** (audit failure tidak boleh menggagalkan deploy/rollback).

### D) Metrics minimal via CloudWatch EMF (log-based)
- Menggunakan EMF (Embedded Metric Format) lewat log JSON.
- Menambahkan helper `internal/shared/obs/emf.go`:
  - `obs.EmitEMF(log, namespace, dims, metrics)` menghasilkan event log `metric` dengan payload `_aws` + fields dimensi + nilai metric.
- Worker emit metric per deploy message:
  - `deploy_duration_ms`
  - `queue_lag_ms` (dari `QueuedAtUnix` pada message)
  - `deploy_attempt_total`
  - `deploy_success_total`
  - `deploy_fail_total`
- Dimensions (stabil):
  - `service=worker`, `op=deploy`, `result=success|fail`, `code=<error_code>|ok`

**Alasan:**
- Tidak perlu agent/collector.
- Tidak perlu `PutMetricData` call terpisah.
- Cukup untuk alarm minimal dan analisa cepat.

### E) Alarm minimal + DLQ di AWS (tanpa IaC)
Karena repo tidak memuat definisi infra untuk alarm, diputuskan:
- Provisioning alarm via AWS CLI.

Resources (fakta saat implementasi):
- SQS main queue (region `ap-southeast-3`):
  - `your-api-dev-hosting-deploy`
- DLQ dibuat:
  - `your-api-dev-hosting-deploy-dlq`
- RedrivePolicy main queue:
  - `deadLetterTargetArn=<DLQ ARN>`
  - `maxReceiveCount=5`

Alarms dibuat:
- SQS (ap-southeast-3):
  - `your-api-dev-hosting-deploy-backlog-visible`
    - Metric: `AWS/SQS ApproximateNumberOfMessagesVisible > 0` (2 periode)
  - `your-api-dev-hosting-deploy-queue-lag-high`
    - Metric: `AWS/SQS ApproximateAgeOfOldestMessage > 300s` (2 periode)
  - `your-api-dev-hosting-deploy-dlq-not-empty`
    - Metric: `AWS/SQS ApproximateNumberOfMessagesVisible > 0` (1 periode)
- Lambda (ap-southeast-1):
  - `AsyrafCloud-Unzipper-errors`
    - Metric: `AWS/Lambda Errors > 0` (1 periode)

Catatan:
- Alarm actions (SNS) tidak diaktifkan pada milestone ini (belum ada topic ARN disepakati).
- Alarm dibuat untuk visibility & readiness; notifikasi dapat ditambahkan kemudian.

## Options (Opsi yang dipertimbangkan)

### 1) Metrics: EMF vs OTel vs PutMetricData
- OTel:
  - (+) tracing/metrics terpadu
  - (-) scope besar untuk MVP; butuh collector/exporter; lebih banyak moving parts
- PutMetricData:
  - (+) metric eksplisit (tanpa parsing log)
  - (-) tambahan API call & permission; lebih banyak failure mode
- EMF (dipilih) ✅:
  - (+) sederhana, log-based, cepat jadi
  - (-) butuh CloudWatch log ingestion; naming/dims harus stabil sejak awal

### 2) Audit: table baru hosting vs reuse table audit existing
- Table baru:
  - (+) isolasi domain
  - (-) overhead provisioning + naming + perms; scope lebih besar
- Reuse table existing (dipilih) ✅:
  - (+) schema sudah terbukti (pk/sk/event/at/meta_json)
  - (+) cepat dan audit-friendly
  - (-) perlu disiplin prefix pk (`site#` vs `acc#`)

### 3) Alarms: IaC vs CLI manual
- IaC:
  - (+) reproducible
  - (-) repo ini belum punya IaC; butuh repo/stack baru
- CLI manual (dipilih untuk milestone) ✅:
  - (+) cepat memenuhi deliverable
  - (-) risk drift; perlu follow-up untuk migrasi ke IaC

## Risks (Risiko) & Mitigations (Mitigasi)

### R1: Salah region → resource terlihat “kosong”
- Risiko: queue/lambda/log group tidak ditemukan, debugging tersesat.
- Mitigasi:
  - Pin region explicit dalam command (`--region`).
  - Dokumentasikan split region:
    - hosting infra `ap-southeast-3`
    - Lambda unzipper `ap-southeast-1` (sementara)

### R2: Table audit name mismatch (`audit_events` vs `your-api-dev-audit-events`)
- Risiko: audit event ditulis ke table yang tidak ada / salah table.
- Mitigasi:
  - Standarisasi `.env`: `DDB_AUDIT_EVENTS_TABLE=your-api-dev-audit-events`.
  - Verifikasi runtime env via `make env-print`.
  - Uji query faktual ke table yang benar.

### R3: Hosting audit tidak muncul karena request tidak authenticated
- Risiko: auditor mengira audit tidak bekerja, padahal request gagal 401.
- Mitigasi:
  - Gunakan `Authorization: Bearer <access_token>` sesuai `JWTAuth()` untuk endpoint `/v1/hosting/*`.
  - Verifikasi dengan create site authenticated → item audit muncul.

### R4: Alarm tanpa notif (ActionsEnabled true tapi tanpa SNS)
- Risiko: alarm tidak “membangunkan manusia”.
- Mitigasi:
  - Alarm tetap berguna sebagai dashboard readiness.
  - Follow-up: add SNS topic + alarm actions (Milestone 10 / ops hardening).

### R5: Log group Lambda belum muncul bila function belum invoke
- Risiko: kesan observability “kosong”.
- Mitigasi:
  - Invoke function minimal sekali untuk memicu log group.
  - Fokus milestone pada metric/alarm dan audit DDB yang deterministik.

## Implementation Summary (Ringkasan implementasi)
Perubahan code utama:
- `internal/shared/requestid`: context helper
- `internal/shared/obs/emf.go`: EMF metric emitter
- `internal/modules/hosting/ports`: tambah `DeployMessage.RequestID`, `AuditSink` hosting
- `internal/modules/hosting/usecase`:
  - `CompleteUpload`: inject request_id ke DeployMessage + audit enqueue
  - `Deployer`: audit deploy success/fail
  - `Dashboard`: audit site_created + rollback request/success/fail
- `cmd/worker/handler.go`:
  - structured logs dengan dims stabil
  - emit EMF metrics per record
  - TODO justify untuk file >100 LOC
- Wiring:
  - `WireHostingUploadPipeline`, `WireHostingDeployer`, `WireHostingDashboard` inject `DDB_AUDIT_EVENTS_TABLE`

Infra provisioning (CLI):
- SQS alarms (backlog, age of oldest, DLQ not empty)
- DLQ creation + RedrivePolicy set (maxReceiveCount=5)
- Lambda error alarm untuk `AsyrafCloud-Unzipper`

## Verification (Acceptance / DoD)
Quality gates:
- `make audit` ✅
- `go test ./... -count=1` ✅
- `go vet ./...` ✅
- component tests tetap hijau ✅

Bukti cek lulus “kenapa deploy gagal?” tanpa buka kode:
- Alarm SQS menunjukkan backlog/lag + DLQ status.
- Audit table query membuktikan hosting audit masuk:
  - CreateSite authenticated (`POST /v1/hosting/sites`) menghasilkan:
    - DDB item: `pk=site#audit-sample-1`, `event=hosting_site_created`, `meta_json` include `request_id`.
- Worker metrics tersedia via EMF log events (namespace `your-api/hosting`, dims `service/op/result/code`).

## Consequences (Dampak)

### Positif
- Observability MVP terpenuhi:
  - request_id konsisten lintas komponen
  - audit trail untuk deploy + rollback + create site
  - metric minimal untuk duration/fail reasons/queue lag
  - alarm minimal + DLQ siap
- Debugging lebih cepat dan deterministik.

### Negatif / Hutang teknis
- Alarm & DLQ dibuat manual (drift risk) → perlu migrasi ke IaC.
- Log group Lambda tidak dijamin ada tanpa invoke.
- File >100 LOC sementara masih ada (ditandai TODO justify) → target pecah di Milestone 9.
- Tidak ada notifikasi SNS untuk alarm (baru visibility).

## Follow-ups (Rencana lanjut)
- Migrasi alarm/DLQ ke IaC (CloudFormation/Terraform/CDK) agar reproducible.
- Tambahkan alarm actions ke SNS (Milestone 10).
- Tambahkan audit event untuk:
  - `hosting_upload_complete_enqueued` + deploy fail reason drilldown (sudah ada, tinggal operasionalisasi query/report).
- Perbaiki region consistency (pastikan worker consumer SQS berada di region hosting).
- Audit menyeluruh:
  - jalur testing, production readiness, ketahanan, konsistensi desain & struktur.
