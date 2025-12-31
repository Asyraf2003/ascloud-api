# ADR 0008: Konsolidasi ADR Verifier Injection (0005 vs 0007)

Tanggal: 2025-12-31  
Status: Accepted

## Context (Keadaan awal)
- Terdapat ADR 0005 dan ADR 0007 yang membahas keputusan yang sama: stabilisasi kontrak router dengan setter/verifier slot (router.SetAccessTokenVerifier → middleware).
- Keduanya berstatus Accepted dan berpotensi menimbulkan kebingungan audit.

## Problem (Masalah)
- Duplikasi keputusan: auditor/pengembang baru tidak tahu ADR mana yang jadi sumber kebenaran.
- Risiko perubahan di satu dokumen tidak tersinkron dengan dokumen lain.
- Reasoning historis menjadi kabur karena dua dokumen “mengklaim” keputusan yang sama.

## Options (Opsi)
1) Biarkan keduanya tetap Accepted
   - (+) tidak perlu perubahan
   - (-) audit rancu, duplikasi permanen

2) Deprecate salah satu ADR dan jadikan satu dokumen sebagai canonical
   - (+) sumber kebenaran tunggal, audit jelas
   - (-) perlu update status dan pointer “superseded by”

3) Merge konten, delete salah satu file
   - (+) rapi
   - (-) histori commit/penomoran bisa jadi membingungkan

## Decision (Keputusan)
- Pilih opsi (2):
  - Jadikan **ADR 0007** sebagai dokumen canonical (lebih lengkap dan ada mitigasi eksplisit).
  - Ubah **ADR 0005** menjadi **Deprecated** dan tambahkan catatan: “Superseded by ADR 0007”.

## Consequences (Dampak)
Positif:
- Sumber kebenaran tunggal: ADR 0007.
- Audit dan onboarding lebih jelas.

Negatif/Risiko:
- Perlu disiplin update status ADR 0005 agar tidak ada “dua keputusan aktif”.
