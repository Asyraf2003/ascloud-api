# ADR 0007: Inject Access Token Verifier via Router/Middleware Setter (stabilkan kontrak router)

Tanggal: 2025-12-31 08:00 WITA
Status: Accepted

## Context (Keadaan awal)
- HTTP router (`internal/transport/http/router`) melakukan wiring route v1 protected dengan middleware JWT.
- Sebelumnya, `router.Register` / `v1.Register` membawa dependency verifier lewat parameter (mis: `Register(e, jwtv)`).
- Perubahan signature router memicu ripple ke banyak tempat (cmd/api wiring, component tests, router tests).
- Target arsitektur: transport layer tetap tipis, kontrak publik router stabil, dan dependency auth/token tidak “nyebar” liar.

## Problem (Masalah)
- Signature `Register(..., verifier)` membuat perubahan auth/token mudah menyebabkan efek berantai (tests & wiring ikut rusak).
- Router menjadi “DI surface” yang terlalu sensitif, padahal router seharusnya kontrak stabil.
- Middleware JWT butuh akses ke verifier, tapi dependency passing per-call tidak skalabel ketika jumlah router/group bertambah.

## Options (Opsi yang dipertimbangkan)
1) Tetap pass verifier sebagai argumen ke `router.Register` dan `v1.Register`
   - (+) eksplisit, tidak ada global state
   - (-) ripple besar setiap kontrak berubah, boilerplate wiring meningkat

2) Simpan verifier di `echo.Context`/container global dan middleware ambil dari situ
   - (+) signature router tetap
   - (-) coupling ke cara bootstrap server, rawan runtime error bila key tidak konsisten

3) Gunakan setter terkontrol: `router.SetAccessTokenVerifier(v)` → `middleware.SetAccessTokenVerifier(v)` dan `JWTAuth()` mengambil dari `MustAccessTokenVerifier()`
   - (+) kontrak router stabil, perubahan terlokalisir, tests lebih sederhana
   - (-) ada global state + perlu disiplin urutan init

## Decision (Keputusan)
- Memilih opsi (3):
  - `cmd/api` menginisialisasi verifier sekali, lalu memanggil `router.SetAccessTokenVerifier(jwtv)`
  - `router.Register(e)` tidak lagi menerima verifier sebagai parameter
  - `middleware.JWTAuth()` tidak menerima parameter, dan mengambil verifier dari `MustAccessTokenVerifier()`
  - Jika verifier belum diset, aplikasi fail-fast saat startup/test (panic/guard) untuk mencegah silent insecure behavior

## Consequences (Dampak)
Positif:
- Kontrak router stabil, perubahan auth/token tidak memicu ripple ke semua callsite.
- Transport/router tetap tipis dan tidak jadi “tempat DI berjalan”.
- Component tests lebih tahan perubahan karena hanya butuh set verifier di 1 titik.

Negatif/Risiko:
- Global state membutuhkan urutan init yang benar (setter harus dipanggil sebelum route protected dipakai).
- Risiko “hidden dependency” bila ada package lain memanggil `JWTAuth()` tanpa men-setup verifier.

Mitigasi:
- `MustAccessTokenVerifier()` fail-fast saat verifier belum diset.
- Component tests memastikan route protected benar-benar pakai verifier yang di-set.
- Disiplin bootstrap: `SetAccessTokenVerifier()` dipanggil sebelum `router.Register()`.

## Catatan:
- ADR 0005 dideprecate karena duplikat; ADR 0007 adalah canonical.
