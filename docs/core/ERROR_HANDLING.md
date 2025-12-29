# Error Handling & Audit Policy

Tiga jalur yang wajib dipisah:

## 1) Ke User (HTTP response)
- Hanya: `code`, `message`, `request_id`
- Tidak boleh: stack trace, SQL error, payload mentah, token/cookie.

## 2) Ke Dev/Ops (log)
- Boleh detail, tapi wajib redaction:
  - Authorization, Cookie, ApiKey, token apapun => mask/redact
- Idealnya log terstruktur (JSON) dan include request_id.

## 3) Ke Audit (append-only event)
- `meta` boleh JSONB tapi:
  - whitelist key yang relevan
  - buang token/cookie/secret
  - batasi ukuran (anti jadi tempat dump request)

## Kontrak Error
Gunakan `internal/shared/apperr.AppError`:
- `Code` stabil untuk client + analytics
- `PublicMessage` aman untuk user
- `Cause` untuk dev (log only)
