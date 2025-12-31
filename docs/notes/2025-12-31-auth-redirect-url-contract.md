# Note: Kontrak `redirect_url` pada Google OIDC belum konsisten

Tanggal: 2025-12-31  
Status: Active

## Context
- Endpoint auth menyediakan:
  - `GET /v1/auth/google/start?purpose=...&redirect_url=...`
  - `GET /v1/auth/google/callback?...`
- Usecase `GoogleStart` dan `GoogleCallback` memiliki field input `RedirectURL`.

## Problem
- Handler HTTP saat ini mengisi `RedirectURL` dari config (`h.cfg.Google.RedirectURL`), bukan dari query `redirect_url`.
- Akibatnya, parameter `redirect_url` di request start menjadi tidak berguna (kontrak membingungkan).
- Callback request terlihat “berhasil lolos validasi redirect_url” walau tidak mengirim redirect_url, karena handler mengisi dari config.

## Notes / Findings
- Perilaku saat ini:
  - `redirect_url` query diabaikan.
  - `RedirectURL` untuk start/callback berasal dari config (statik).
- Dampak:
  - (+) Mengurangi risiko open redirect karena client tidak bisa menyuntik return URL.
  - (-) Kontrak API tidak jujur (parameter ada tapi tidak dipakai).
  - (-) Blueprint multi-client (web/dashboard berbeda origin) tidak bisa berkembang tanpa perubahan desain.

## Options (kalau ada)
1) Jadikan redirect statik (single client)
   - Hapus `redirect_url` dari contract start/callback
   - Rename field agar tidak membingungkan (mis. OIDCCallbackURL vs ReturnTo)
2) Dukung redirect dinamis tapi aman
   - Terima `redirect_url` di start
   - Validasi ketat (allowlist origins/paths)
   - Simpan `return_to` di AuthState (state store)
   - Callback tidak menerima `redirect_url` dari request (ambil dari state)

## Next Actions
- [ ] Putuskan: single-client statik vs multi-client dinamis (dengan allowlist + state).
- [ ] Rapikan penamaan: bedakan OIDC `redirect_uri` (callback URL) vs app `return_to` (dashboard URL).
- [ ] Update docs + tests sesuai keputusan.
