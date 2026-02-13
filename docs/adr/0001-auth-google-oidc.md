# ADR 0001: Auth via Google OIDC (Login/Register) + Refresh Rotation + Step-up

Tanggal: 2025-12-29 09:42 WITA
Status: Accepted

## Context
Kita butuh fitur login/register user via Google saja (tanpa metode lain), dengan keamanan setara finance dan konsep zero trust.
Repo punya aturan:
- Error response wajib JSON envelope lewat presenter.HTTPErrorHandler.
- Boundary: HTTP parse/validate ringan → panggil usecase → presenter.
- Vendor/IO adapter hanya di internal/platform/*.
- File <=100 baris dan 1 folder = 1 package.

Keputusan produk:
- Target client: web
- Access token: JWT exp 30 menit
- Refresh token: HttpOnly cookie, rotation + anti-reuse
- Single-session: 1 user 1 session aktif; login baru revoke session lama
- Step-up auth wajib
- Audit events: semua event auth dicatat

## Decision
1) Google OIDC diperlakukan sebagai provider lewat interface ports.OIDCProvider (provider-ready).
2) Flow login:
   - /v1/auth/google/start: state+nonce+PKCE disimpan (TTL pendek) lalu redirect.
   - /v1/auth/google/callback: exchange+verify id_token; resolve identity (provider,sub); create account jika baru; revoke session lama; create session baru; issue access JWT + set refresh cookie.
3) Refresh:
   - /v1/auth/refresh menggunakan CSRF double-submit + Origin allowlist.
   - Refresh rotation + reuse detection: refresh lama dipakai ulang → revoke session + audit token_reuse_detected.
4) Logout:
   - /v1/auth/logout revoke session + clear cookies.
5) Step-up:
   - /v1/auth/google/stepup/start & callback; upgrade trust_level (aal2).
6) JWT revoke “instant”:
   - Protected routes wajib lakukan session-check (sid harus aktif), karena JWT stateless.

## Alternatives considered
- Opaque access token + introspection (lebih mudah revoke, tapi butuh infra dan latensi).
- JWT access TTL sangat pendek (5-10m) tanpa session-check (lebih ringan, tapi keputusan produk minta 30m).
Dipilih: JWT 30m + session-check agar single-session dan revoke tetap kuat.

## Clarifications (Implementation Policy)

Bagian ini mengunci detail operasional agar audit security deterministik. Tidak mengubah keputusan inti, hanya memperjelas implementasi.

### A. Cookie Policy (Refresh + CSRF)

**Refresh Cookie**
- Name: `refresh_token`
- HttpOnly: `true`
- Secure: mengikuti config `CookieSecure` (prod: true, dev: false)
- SameSite: `Strict` (karena app.domain.com dan api.domain.com masih satu “site”/eTLD+1)
- Path: `/v1/auth/` (dipersempit, bukan `/`)
- Domain: **api host only** (tanpa Domain attribute) agar cookie tidak nyasar ke subdomain lain.
- Max-Age: sesuai TTL refresh (mis. 7–30 hari, tergantung policy)
- Rotasi: setiap refresh sukses -> set cookie baru (token baru), token lama invalid.

**CSRF Cookie (double-submit)**
- Name: `csrf_token`
- HttpOnly: `false` (dibaca JS)
- Secure: mengikuti config `CookieSecure`
- SameSite: `Strict`
- Path: `/`
- Domain: `.domain.com` (agar bisa dibaca oleh dashboard `app.domain.com` dan ikut terkirim ke API `api.domain.com`)
- Nilai token: random 128-bit+ (base64url), rotate minimal per-session (boleh per-refresh jika mau lebih ketat).

Catatan:
- Refresh cookie **tetap host-only + HttpOnly** (paling penting).
- CSRF cookie boleh parent-domain karena bukan credential, hanya nonce pembanding.

### B. CSRF Double-Submit Rule

Endpoint cookie-based (minimal: `/v1/auth/refresh`, `/v1/auth/logout`) wajib memenuhi:
- Method: `POST` (bukan GET)
- Header: `X-CSRF-Token` wajib ada
- Cookie: `csrf_token` wajib ada
- Validasi: `X-CSRF-Token` harus sama persis dengan cookie `csrf_token`
- Jika mismatch/missing -> reject 403 + audit event

### C. Origin Allowlist + CORS Rules

Untuk endpoint cookie-based:
- Wajib ada header `Origin`
- `Origin` harus match allowlist (contoh: `https://app.domain.com` dan localhost dev yang di-allow)
- Jika `Origin` tidak valid -> reject (tanpa proses lebih lanjut)
- CORS:
  - `Access-Control-Allow-Origin` = origin yang valid (spesifik, bukan `*`)
  - `Access-Control-Allow-Credentials: true`
  - `Vary: Origin`

(Opsional tapi bagus) Enforce Fetch Metadata:
- `Sec-Fetch-Site` harus `same-site` atau `same-origin` untuk endpoint cookie-based.

### D. OIDC ID Token Verification Checklist

Pada callback, verifikasi minimal:
- Signature valid via JWK (cache + rotate aware)
- `iss` valid
- `aud` cocok (client id)
- `exp/iat` valid (dengan clock skew kecil)
- `nonce` cocok dengan yang disimpan saat start
- (Jika ada) `hd`/hosted domain rules sesuai policy

### E. JWT Claims + Session Check

Access JWT minimal memuat:
- `sub` (account id)
- `sid` (session id)
- `exp`, `iat`, `iss`, `aud`

Protected routes **wajib**:
- Verifikasi JWT signature & standard claims
- Session-check: `sid` harus masih aktif (single-session + instant revoke)
- (Jika endpoint butuh AAL2) check `trust_level >= aal2`

### F. Step-up (AAL2) Policy

- Step-up menaikkan `trust_level` pada session menjadi `aal2`
- AAL2 memiliki TTL (rekomendasi: 10 menit) untuk operasi sensitif
- Setelah TTL habis, `trust_level` turun kembali ke baseline (aal1) tanpa perlu logout
- Revoke session otomatis menghapus status step-up

### G. Rate Limit Layering (minimum viable)

- `/v1/auth/google/start`: rate limit per-IP (mencegah abuse redirect)
- `/v1/auth/google/callback`: per-IP + per-identity/provider-sub (mencegah brute)
- `/v1/auth/refresh`: per-session + per-IP (mencegah hammer refresh)
- Reuse detection: jika refresh token lama dipakai lagi -> revoke session + throttle tambahan

### H. Audit Event Minimal

Semua event auth dicatat dengan allowlist fields (no raw token):
- `auth_login_start`
- `auth_callback_success` / `auth_callback_failed`
- `auth_refresh_success`
- `auth_refresh_reuse_detected`
- `auth_logout`
- `auth_stepup_start`
- `auth_stepup_success` / `auth_stepup_failed`

## Consequences
- Ada tambahan beban storage session dan check per request protected.
- Provider baru (Apple/Microsoft) tinggal tambah adapter platform; usecase tetap.
- Audit trail wajib disimpan dan meta harus allowlist/redact sebelum log.

