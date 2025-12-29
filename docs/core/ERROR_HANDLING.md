# Error Handling & Audit Policy

Dokumen ini menetapkan kebijakan error dan log untuk repo ini.
Tujuan: response aman untuk user, log berguna untuk dev/ops, audit trail rapi dan tidak jadi tempat buang sampah data sensitif.

---

## Tiga Jalur Output (wajib dipisah)

### 1) Ke User (HTTP response)
Response error **wajib** berbentuk JSON envelope via `presenter.HTTPErrorHandler`.

Minimal field (kontrak publik):
- `code` (stabil, untuk client & analytics)
- `message` (aman untuk user)
- `request_id` (untuk tracing)

Larangan keras:
- stack trace
- SQL/vendor error mentah
- payload request mentah
- token/cookie/secret

### 2) Ke Dev/Ops (log)
Log boleh detail, tapi wajib redaction:
- `Authorization`, `Cookie`, `Set-Cookie`, `ApiKey`, token apa pun → **mask/redact**
- jangan log body raw kalau berisi credential/PII (kecuali memang disanitasi)

Minimal yang disarankan ada di log:
- `request_id`
- route/method/status
- error `code`
- error cause (internal)
- latency

### 3) Ke Audit (append-only event)
Audit event adalah catatan security/business yang append-only.

Aturan `meta` (JSONB) kalau dipakai:
- whitelist key yang relevan (jangan dump object)
- buang token/cookie/secret
- batasi ukuran (anti jadi tempat dump request)
- meta yang dikirim ke audit **bukan** untuk dikembalikan ke client

---

## Kontrak Error (AppError)

Gunakan `internal/shared/apperr.AppError` untuk error yang keluar dari usecase.

Tujuan field:
- `Code`: stabil untuk client + analytics (mis. `auth.invalid_state`)
- `PublicMessage`: aman untuk user
- `Cause`: detail internal (log only)
- `HTTPStatus`: status yang benar untuk response

Aturan:
- Usecase boleh return `AppError` langsung.
- Presenter bertanggung jawab memetakan `AppError` → JSON envelope dan **sanitize** error lain.

---

## Mapping yang disarankan (contoh)

| Kondisi | HTTP | code | message (public) |
|---|---:|---|---|
| Bad request / invalid input | 400 | `request.invalid` | "Permintaan tidak valid." |
| Unauthorized (no/invalid auth) | 401 | `auth.unauthorized` | "Unauthorized." |
| Forbidden (access denied) | 403 | `auth.forbidden` | "Akses ditolak." |
| Not found | 404 | `resource.not_found` | "Resource tidak ditemukan." |
| Conflict | 409 | `resource.conflict` | "Terjadi konflik data." |
| Rate limited | 429 | `rate_limited` | "Terlalu banyak permintaan." |
| Internal / unknown | 500 | `internal_error` | "Terjadi kesalahan. Coba lagi." |

Catatan:
- `code` harus konsisten lintas modul, tidak berubah tanpa ADR kalau sudah dipakai client.
- Hindari message yang “mengaku” hal spesifik (mis. “email tidak terdaftar”) kalau itu bisa jadi enumerasi.

---

## Redaction Policy (minimum)
Header/field yang wajib disensor di log:
- `Authorization`
- `Cookie`, `Set-Cookie`
- `X-Api-Key` / `ApiKey`
- `access_token`, `refresh_token`, `id_token`
- password/secret apa pun

Kalau ragu: redact.

---

## Audit Event Guidelines (minimum)
Event yang biasanya wajib masuk audit (auth/security):
- login_success / login_failed
- refresh_rotated / refresh_reused_detected
- logout
- stepup_success / stepup_failed
- suspicious_activity (rate-limit/trust)

Setiap event minimal punya:
- timestamp
- actor (account_id/tenant_id kalau ada)
- request_id
- type
- meta sanitized

---

## Checklist cepat sebelum merge
- [ ] Error dari usecase pakai `AppError` (bukan fmt.Errorf random).
- [ ] Handler tidak membentuk response error sendiri (semua lewat presenter).
- [ ] Log sudah redact token/cookie.
- [ ] Tidak ada data sensitif di audit meta.
