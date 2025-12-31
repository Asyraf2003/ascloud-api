# Note: Decouple Router dari Config LoadAuth (policy snapshot)

Tanggal: 2025-12-31  
Status: Active

## Context
- Router v1 perlu keputusan runtime seperti "require HTTPS atau tidak" untuk group auth/protected.
- Sebelumnya ada kecenderungan router membaca config langsung (mis. `config.LoadAuth()`), yang membuat transport layer jadi “kepo” ke config storage.

## Problem
- Melanggar separation of concerns: router transport jadi tergantung cara config disimpan/di-load.
- Jika mekanisme config berubah (env → file → secret manager, dll), router ikut terdampak.
- Risiko ripple changes meningkat (transport layer ikut terbawa refactor config/auth).

## Notes / Findings
- Solusi lebih bersih: bootstrap memuat config sekali, lalu mengisi “policy snapshot” yang dibaca router/middleware.
- Implementasi sekarang:
  - `cmd/api` memanggil `authPkg.InitPolicy(authCfg)`
  - `v1/router` hanya bertanya `authPkg.RequireHTTPS()` (berdasarkan policy snapshot), tanpa membaca config langsung.

## Options (kalau ada)
1) Router memanggil `config.LoadAuth()` langsung
   - (+) simpel
   - (-) coupling tinggi, ripple besar

2) Policy snapshot (InitPolicy + atomic store) dan router hanya membaca policy
   - (+) router stabil, config bisa berubah tanpa bongkar router
   - (-) ada state global, perlu bootstrap order yang disiplin

3) Inject policy melalui parameter `v1.Register(e, policy)`
   - (+) dependency eksplisit, tanpa global state
   - (-) kontrak router kembali sensitif (efek berantai seperti kasus verifier injection)

## Next Actions
- [ ] Pastikan semua keputusan router/middleware yang berbasis config lewat policy snapshot, bukan LoadAuth.
- [ ] Tambahkan test kecil untuk memastikan policy default aman ketika InitPolicy belum dipanggil (fail-fast atau default conservative).
