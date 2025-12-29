# Notes

Folder ini untuk catatan sementara: draft, ide, log debugging, atau potongan diskusi yang belum layak jadi ADR / Core docs.

> Rule: kalau sudah jadi keputusan arsitektur, pindahkan ke `docs/adr/`.
> Kalau sudah jadi “aturan main repo”, pindahkan ke `docs/core/` atau `docs/internal/ai/`.

---

## Kapan pakai Notes
- Draft keputusan (belum final).
- Hasil investigasi/debugging.
- Checklist kerja / TODO sementara.
- Catatan refactor yang belum dieksekusi.

## Kapan harus keluar dari Notes
- Keputusan final yang mempengaruhi arsitektur/perilaku sistem → `docs/adr/`
- Dokumen permanen yang jadi pedoman repo → `docs/core/`
- Aturan kerja AI / template → `docs/internal/ai/`

---

## Format penamaan file
Gunakan format:
`YYYY-MM-DD-judul-singkat.md`

Contoh:
- `2025-12-29-strukturnulis.md`
- `2025-12-30-auth-refresh-debug.md`

---

## Template isi notes (minimal)
Setiap notes idealnya punya:

- **Context**: keadaan awal
- **Problem**: apa yang rusak / bingung
- **Findings**: temuan penting (fakta)
- **Next actions**: langkah lanjut
- **Status**: draft / done / migrated-to-ADR / obsolete
