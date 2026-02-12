# ADR 0009: Kontrak Redirect pada Google OIDC (bedakan callback URL vs return_to)

Tanggal: 2025-12-31  
Status: Proposed

## Context (Keadaan awal)
- Flow Google OIDC butuh `redirect_uri` (callback endpoint) yang terdaftar di Google Console.
- API juga ingin punya konsep `return_to` (kembali ke web/dashboard setelah login).
- Implementasi sekarang mengisi `RedirectURL` dari config (`h.cfg.Google.RedirectURL`), bukan dari query `redirect_url`.

## Problem (Masalah)
- Nama `RedirectURL` ambigu: bisa berarti OIDC callback URL atau return_to dashboard.
- Kontrak API membingungkan karena parameter `redirect_url` ada, tapi tidak dipakai.
- Jika suatu saat butuh multi-client (lebih dari satu origin), desain statik menghambat.
- Jika menerima return_to dinamis tanpa validasi ketat, risiko open redirect tinggi.

## Options (Opsi)
1) Single-client statik (return_to tidak dinamis)
   - `redirect_url` dihapus dari request
   - Config memegang:
     - `OIDC_REDIRECT_URI` (callback URL yang terdaftar)
     - Opsional: `DEFAULT_RETURN_TO` (kalau memang butuh)
   - (+) sederhana, aman
   - (-) tidak fleksibel untuk multi-client

2) Multi-client dinamis dengan allowlist + state
   - Start menerima `redirect_url` sebagai `return_to`
   - Validasi `return_to` (allowlist origin + path policy)
   - Simpan `return_to` di AuthState (state store)
   - Callback tidak menerima `redirect_url` dari request (ambil dari state)
   - (+) fleksibel, tetap aman kalau allowlist benar
   - (-) butuh policy jelas dan test coverage

## Decision (Keputusan)
- (Belum diputuskan) Target default blueprint: **Option 2** jika ingin multi-client,
  atau **Option 1** jika target produk awal single dashboard saja.

## Consequences (Dampak)
Positif:
- Kontrak jelas: tidak ada parameter “pajangan”.
- Penamaan tegas: tidak rancu antara callback URL dan return_to.
- Risiko open redirect bisa dikendalikan (khusus Option 2).

Negatif/Risiko:
- Option 2 butuh allowlist + policy yang disiplin dan test yang memadai.
- Option 1 bisa jadi menghambat evolusi produk kalau kebutuhan multi-client datang cepat.
