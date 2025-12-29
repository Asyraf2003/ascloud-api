# Testing

Dokumen ini menjelaskan kategori test di repo ini, aturan build tags, dan cara menjalankannya.
Tujuan: test suite bisa diaudit, deterministic, dan tidak “nyolok” ikut jalan di pipeline yang salah.

## Prinsip
- Default test harus aman dijalankan kapan pun (lokal/CI) tanpa butuh service eksternal.
- Test yang butuh dependency real (DB/network) wajib dipisahkan dengan build tag.
- Jangan menyebut sesuatu “integration” kalau tidak benar-benar menyentuh dependency real.

---

## Kategori Test

### 1) Unit Tests (default)
**Definisi**
- Deterministic.
- Tidak ada network, tidak ada DB real, tidak butuh port listen.
- Dependency di-mock/fake via interface (`ports`).

**Build tag**
- Tidak pakai build tag (default).

**Run**
- `make test-unit`
- (setara) `go test ./... -count=1`

**Contoh cocok**
- Domain logic, hashing/crypto helper, policy evaluator, JWT codec/verifier, usecase dengan fake store.

---

### 2) Component Tests (HTTP in-memory)
**Definisi**
- Menguji HTTP layer (router/middleware/handler) memakai `httptest` + Echo in-memory.
- Tidak menyentuh DB real dan tidak listen port.
- Fokus: wiring route + middleware enforcement + response envelope.

**Build tag**
- Wajib: `//go:build component`
- Alasan: supaya component test tidak ikut “default unit test” kalau jumlahnya besar dan butuh setup HTTP.

**Run**
- `make test-component`
- (setara) contoh:
  - `go test -tags=component ./internal/transport/http/... -count=1`
  - `go test -tags=component ./internal/modules/auth/transport/http/... -count=1`

**Contoh cocok**
- CSRF cookie/header enforcement → 403
- Origin not allowed → 403
- OPTIONS passthrough → 204
- /health returns 200
- JWTAuth middleware → 401/200

---

### 3) Integration Tests (real dependencies)
**Definisi**
- Menyentuh dependency real: Postgres, migrations, queue, object storage, dsb.
- Biasanya menggunakan `docker compose` untuk bring-up deps.

**Build tag**
- Wajib: `//go:build integration`
- Integration test tidak boleh jalan di default `make test`.

**Run**
- `make test-integration`
- (setara) `go test -tags=integration ./... -count=1`

**Catatan**
- Integration tests harus fail fast dengan error yang jelas kalau dependency belum siap.

---

## Aturan Penamaan & Lokasi

### File naming
- Unit: `*_test.go`
- Component: `*_component_test.go` + build tag `component` (recommended)
- Integration: `*_integration_test.go` + build tag `integration` (recommended)

### Lokasi
- Unit test: dekat code yang dites (co-located).
- Component test: co-located di package HTTP terkait.
- Integration test: co-located atau dikumpulkan di `internal/integration/...` atau `tests/integration/...` (pilih salah satu, konsisten).

---

## Template Build Tags

### Component test template
---go
//go:build component

package yourpkg_test

import "testing"

func TestSomethingComponent(t *testing.T) {}
---

### Integration test template
---go
//go:build integration

package yourpkg_test

import "testing"

func TestSomethingIntegration(t *testing.T) {}
---

## Testing Transactions

Usecase menerima dependency `ports.Transactor`:
- Unit test menggunakan `ports.NoopTransactor` agar flow bisa dites tanpa DB.
- Integration test (tag `integration`) digunakan untuk validasi transaksi nyata di Postgres.

Aturan:
- Repo/store yang butuh transaksi harus selalu query via `postgres.GetExecutor(ctx, db)`.
- Test yang mengecek atomicity (rollback/commit) wajib integration test.
