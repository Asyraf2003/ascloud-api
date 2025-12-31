# ADR 0001: Access token verifier injection via router -> middleware

Tanggal: 2025-12-30 10:00 WITA
Status: Accepted

## Context (Keadaan awal)
- Router v1 membutuhkan verifier untuk memasang JWT middleware.
- Awalnya verifier di-pass lewat signature router.Register/v1.Register.
- Perubahan auth/token berpotensi memaksa perubahan signature router (efek berantai ke cmd/test).

## Problem (Masalah)
- Public contract router tidak stabil.
- Test & wiring jadi rapuh setiap kali auth token berubah.
- Risiko “transport layer ikut kebawa domain/token refactor”.

## Options (Opsi)
1) Pass verifier lewat parameter router.Register/v1.Register
   - (+) eksplisit dependency
   - (-) contract router tidak stabil, efek berantai ke call sites
2) Buat global provider verifier (set sekali saat bootstrap), middleware mengambil dari provider
   - (+) router contract stabil, perubahan auth minim ripple
   - (-) ada global state, perlu fail-fast jika belum di-init

## Decision (Keputusan)
- Pakai provider verifier: bootstrap memanggil router.SetAccessTokenVerifier(verifier).
- middleware.JWTAuth() mengambil verifier dari provider dan fail-fast bila belum diset.

## Consequences (Dampak)
Positif:
- Signature router tetap stabil.
- Wiring auth/token tidak memaksa perubahan transport contract.
- Test lebih sederhana (set verifier sekali).

Negatif/Risiko:
- Global state: perlu disiplin bootstrap order dan fail-fast untuk menghindari silent insecure behavior.
