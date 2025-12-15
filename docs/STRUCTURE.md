# Struktur Repo (Pegangan)

Target: awal ringan (1–50 user) tapi gampang naik kelas tanpa bongkar pondasi.

Aturan:
- Usecase cuma kenal interface di `internal/modules/*/ports`
- Implementasi vendor/IO hanya di `internal/platform/*` (atau `internal/adapters/*` kalau belum rename)
- Handler HTTP cuma mapping request/response, tidak taruh logika bisnis

Fase 0:
- Repo compile
- API jalan + /health + /ready
- Logging JSON + request_id
- DB optional (buat nanti debugging)
