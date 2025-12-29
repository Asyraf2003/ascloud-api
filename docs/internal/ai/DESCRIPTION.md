# DESCRIPTION

Dokumen ini menjelaskan **kenapa** template AI di repo ini dibuat seperti itu, dan **bagian mana** yang wajib kamu sesuaikan tiap task.

Target akhirnya sederhana:
- AI tidak mengarang path/kontrak.
- Perubahan minim efek berantai.
- Hasilnya enak diaudit (bukan “percaya aja bro”).

---

## Kenapa template ini efektif

### 1) Snapshot dulu = anti halu
AI sering “yakin” pada struktur repo yang sebenarnya sudah berubah.
Snapshot memaksa AI kerja dari fakta:
- struktur folder
- kontrak router/presenter
- boundary rules yang berlaku saat ini

### 2) Decision list = beda pilihan, beda dunia
Beberapa keputusan kelihatan kecil, tapi dampaknya besar:
- cookie vs body untuk refresh token
- web browser vs API-to-API (CSRF/CORS beda total)
- expiry & rotation model
- step-up / trust score

Kalau decision tidak eksplisit, hasil implementasi bisa “benar” tapi salah konteks.

### 3) Blueprint dulu = cegah perbaikan berantai
Tanpa blueprint, AI cenderung:
- nambah file random,
- ubah struktur seenaknya,
- “fix” satu bug tapi ngerusak kontrak lain.

Blueprint memaksa:
- flow endpoint jelas
- ownership data jelas
- ports & dependency jelas
- plan test jelas

### 4) DoD jelas = audit gampang
Kalau tidak ada DoD, ujungnya debat perasaan.
Minimal DoD:
- gofmt
- go test
- go vet
- sanity curl (kalau HTTP)
- make audit (kalau tersedia)

---

## Bagian yang wajib kamu sesuaikan (tiap task)

### `[MODULE_PATH]`
Harus sama dengan `go.mod` (contoh: `example.com/your-api`).

### `[TARGET_MODULE]`
Modul yang lagi dikerjakan (auth, hosting, domains, trust, dll).

### `[EXTRA_SNAPSHOT_FILES]`
Tambahkan file yang relevan dengan task.
Contoh:
- auth: `internal/platform/google/*`, `internal/platform/token/*`, `internal/modules/auth/*`
- hosting: `internal/platform/objectstore/*`, `internal/platform/edge/*`, `internal/modules/hosting/*`
- queue/worker: `internal/platform/queue/*`, `cmd/worker/*`

---

## Kesalahan umum (biar gak mengulang dosa)

- AI langsung nulis code tanpa baca router/presenter kontrak.
- Menyebut “integration test” tapi tidak menyentuh dependency real.
- Domain/usecase diam-diam import vendor/platform (melanggar boundaries).
- Mengubah format response sendiri (bypass presenter).
- Menambah debug endpoint tanpa gating `DEBUG_ROUTES=1`.

---

## Checklist cepat sebelum mulai task
- [ ] Snapshot ditempel lengkap (tree + kontrak inti).
- [ ] Decisions penting sudah dijawab (client type, token model, expiry, dll).
- [ ] Blueprint sudah ditulis sebelum eksekusi.
- [ ] DoD sudah disepakati (commands + sanity).
