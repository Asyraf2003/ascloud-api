# ADR 0001: Auth via Google OIDC (Login/Register) + Refresh Rotation + Step-up

Tanggal: 2025-12-14
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

## Consequences
- Ada tambahan beban storage session dan check per request protected.
- Provider baru (Apple/Microsoft) tinggal tambah adapter platform; usecase tetap.
- Audit trail wajib disimpan dan meta harus allowlist/redact sebelum log.
