# Testing (AWS-First)

Dokumen ini menjelaskan kategori test di repo ini, aturan build tags, dan cara menjalankannya.
Tujuan: test suite deterministic, audit-friendly, dan tidak “nyolok” ikut jalan di pipeline yang salah.

---

## Prinsip
- Default test aman dijalankan kapan pun (lokal/CI) tanpa dependency eksternal.
- Test yang butuh dependency real (AWS services / network / DB) wajib dipisahkan via build tag.
- Jangan menyebut sesuatu “integration” kalau tidak benar-benar menyentuh dependency real.

---

## Kategori Test

### 1) Unit Tests (default)
- Deterministic, tanpa network/DB real.
- Dependency di-fake/mock via interface `ports`.

Run:
- `make test-unit`
- `go test ./... -count=1`

---

### 2) Component Tests (HTTP in-memory)
- Uji router/middleware/handler via `httptest` + Echo in-memory.
- Tanpa DB real dan tidak listen port.

Run:
- `make test-component`
- `go test -tags=component ./internal/transport/http/... -count=1`

---

### 3) Integration Tests (real dependencies: AWS)
Menyentuh dependency real, contoh:
- DynamoDB (AWS sandbox atau local emulator)
- S3 (AWS sandbox atau localstack)
- SQS (AWS sandbox atau localstack)
- (future) CloudFront-related tests lebih cocok via contract tests / staging.

Build tag:
- `//go:build integration`

Run:
- `make test-integration`
- `go test -tags=integration ./... -count=1`

Catatan:
- Integration tests harus fail fast dengan error jelas jika dependency belum tersedia.
- Prefer environment “sandbox” yang tidak pakai data produksi.

---

## Template Build Tags
(Component / Integration templates tetap seperti sekarang.)

---

## Konsistensi & Idempotency (wajib diuji)
Untuk flow event-driven (queue → worker):
- Job harus idempotent (replay SQS / retry worker tidak bikin state rusak).
- Update metadata harus pakai mekanisme yang aman (conditional update / versioning / transact-write bila dipakai).
- Error vendor harus dimapping ke `AppError` atau error_code internal yang aman.

---

## Legacy Tests (historical)
Jika masih ada integration test legacy Postgres:
- harus diisolasi dengan build tag `legacy_postgres` (dan/atau `integration`),
- tidak boleh jalan pada default pipeline AWS-first.
