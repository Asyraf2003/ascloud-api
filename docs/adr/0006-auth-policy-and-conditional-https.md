# ADR 0002: Auth policy initialization and conditional HTTPS enforcement

Tanggal: 2025-12-30 11:00 WITA
Status: Accepted

## Context (Keadaan awal)
- Endpoint auth cookie (refresh/logout) memakai Origin + CSRF + RateLimit.
- mwAuthCookieStrict sebelumnya membaca config lewat LoadAuth() (env) di runtime.
- Route group v1/auth dan v1/protected memakai RequireHTTPS tanpa mempertimbangkan mode dev.

## Problem (Masalah)
- Membaca env di runtime membuat policy bisa berubah tidak terkontrol.
- RequireHTTPS memblokir dev flow (localhost HTTP) meski CookieSecure=false.
- Policy seharusnya dibekukan saat bootstrap untuk audit & determinisme.

## Options (Opsi)
1) Tetap LoadAuth() di middleware dan hard RequireHTTPS
   - (+) sederhana
   - (-) nondeterministic, dev experience buruk, audit tidak rapi
2) Init policy sekali saat bootstrap, gunakan CookieSecure untuk memutuskan RequireHTTPS
   - (+) deterministic, sesuai security posture, dev tetap bisa HTTP
   - (-) perlu policy store dan inisialisasi wajib

## Decision (Keputusan)
- Tambah policy store v1/auth: allowedOrigins, csrfCookie, requireHTTPS.
- InitPolicy(authCfg) dipanggil saat bootstrap.
- mwAuthCookieStrict memakai policy store (tanpa LoadAuth()).
- v1 router memasang trust.RequireHTTPS hanya jika policy.requireHTTPS=true.

## Clarifications: HTTPS detection behind proxies

`RequireHTTPS` harus mempertimbangkan deployment di belakang reverse proxy / load balancer.

Policy:
- Jika server menerima TLS langsung (`req.TLS != nil`) -> dianggap HTTPS.
- Jika di belakang proxy, HTTPS ditentukan dari forwarded headers yang **trusted** (contoh: `X-Forwarded-Proto: https`).
- Trust terhadap forwarded headers hanya boleh aktif jika server berada di lingkungan yang mengontrol proxy (mis. ALB/API Gateway), untuk menghindari spoofing dari client.

Checklist implementasi:
- Pastikan middleware/echo config yang membaca forwarded headers di-enable dengan mode trust yang benar.
- Tambahkan test (component/integration) untuk:
  - dev HTTP (CookieSecure=false) -> tidak memaksa HTTPS
  - prod (CookieSecure=true) + `X-Forwarded-Proto=https` -> lolos
  - prod (CookieSecure=true) tanpa TLS/forwarded -> ditolak

## Consequences (Dampak)
Positif:
- Policy konsisten per proses (audit-friendly).
- Dev localhost HTTP berjalan tanpa melemahkan prod.
- CookieSecure menjadi sumber kebenaran untuk kebutuhan HTTPS.

Negatif/Risiko:
- Harus memastikan InitPolicy dipanggil sebelum router.Register.
- Jika prod di belakang reverse proxy, RequireHTTPS perlu memperhitungkan forwarded headers (jika trust.RequireHTTPS belum menanganinya).
