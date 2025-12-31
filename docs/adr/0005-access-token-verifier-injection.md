# ADR 0005: Access token verifier injection via router -> middleware

Tanggal: 2025-12-30 10:00 WITA  
Status: Deprecated (Superseded by ADR 0007)

## Context (Keadaan awal)
- Router v1 membutuhkan verifier untuk memasang JWT middleware.
- Pada iterasi awal, verifier sempat di-pass lewat signature `router.Register/v1.Register`.
- Konsep “verifier injection” kemudian distabilkan lewat mekanisme setter agar kontrak router tidak mudah berubah.

## Problem (Masalah)
- Dokumen ini menduplikasi keputusan yang sama dengan ADR 0007.
- Duplikasi membuat audit rancu: seolah ada 2 keputusan berbeda, padahal satu keputusan yang sama (stabilisasi kontrak router + verifier provider).

## Options (Opsi)
1) Biarkan dua ADR tetap Accepted
   - (+) tidak perlu edit
   - (-) audit makin membingungkan, risk salah rujuk
2) Deprecate ADR 0005 dan jadikan ADR 0007 sebagai canonical
   - (+) audit jelas, satu sumber kebenaran
   - (-) perlu update status dan cross-reference

## Decision (Keputusan)
- Memilih opsi (2): ADR 0005 dideprecate.
- Keputusan final dan canonical ada di **ADR 0007: Inject Access Token Verifier via Router/Middleware Setter (stabilkan kontrak router)**.

## Consequences (Dampak)
Positif:
- Satu sumber kebenaran untuk keputusan verifier injection.
- Audit dan reasoning lebih rapi.

Negatif/Risiko:
- Reader harus diarahkan ke ADR 0007 (mitigasi: link/catatan eksplisit di dokumen ini).
