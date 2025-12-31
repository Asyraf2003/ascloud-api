# Note: Router/middleware tidak boleh load config/env langsung

Tanggal: 2025-12-31  
Status: Active

## Context
- Sebelumnya ada pola router/middleware yang memanggil loader config langsung (mis: `config.LoadAuth()` atau `os.Getenv()`).
- Transport layer butuh beberapa keputusan runtime (contoh: `require_https`, allowed origins, csrf cookie name) untuk trust/cookie route.

## Problem
- Melanggar Separation of Concerns: router jadi tahu “cara config diambil”.
- Jika mekanisme config berubah (env -> file -> remote config), router ikut kebongkar.
- Sulit dites secara deterministik (router bisa “diam-diam” membaca env di runtime).

## Notes / Findings
- Pola yang aman: bootstrap (cmd/api) load config sekali, lalu inject *nilai yang dibutuhkan* ke transport melalui `InitPolicy(...)`.
- Router cukup membaca policy runtime via fungsi kecil (`RequireHTTPS()`, `AllowedOrigins()`, `CSRFCookieName()`), tanpa tahu sumber config.
- Versi kamu sekarang sudah mengarah benar:
  - `v1/router` tidak load config, hanya `authPkg.RequireHTTPS()`.
  - `policy_state` menyimpan state via `atomic.Value`.

## Options (kalau ada)
1) Router memanggil `config.LoadAuth()` / `os.Getenv()` langsung
   - (+) cepat
   - (-) coupling tinggi, ripple besar, testing makin rapuh

2) Pass config ke `router.Register/v1.Register` via parameter
   - (+) eksplisit
   - (-) kontrak router jadi sensitif, ripple ke banyak callsite

3) `InitPolicy(...)` + state internal (atomic) dan router baca via getter kecil
   - (+) kontrak router stabil, source config bisa berubah tanpa bongkar router
   - (-) ada global state, butuh disiplin init order

## Next Actions
- [ ] Pastikan tidak ada lagi `config.LoadAuth()`/`os.Getenv()` di router/middleware auth selain di bootstrap.
- [ ] Pertimbangkan ubah signature `InitPolicy(cfg config.AuthConfig)` menjadi `InitPolicy(input struct lokal)` agar router tidak import package `config` sama sekali.
- [ ] Tambah component test: `COOKIE_SECURE=true` menolak HTTP (403), `COOKIE_SECURE=false` mengizinkan HTTP.
