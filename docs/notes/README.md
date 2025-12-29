# Notes

Folder ini isinya catatan sementara: ide, draft, investigasi, atau brainstorming.
Bukan kontrak arsitektur, bukan keputusan final.

Kalau sesuatu sudah “resmi” dan harus jadi pegangan tim/masa depan:
- Keputusan besar → pindahkan ke `docs/adr/`
- Kontrak/peta sistem → pindahkan ke `docs/core/`
- Aturan kerja AI → pindahkan ke `docs/internal/ai/`

---

## Aturan Nama File
Format:
- `YYYY-MM-DD-judul-singkat.md`

Contoh:
- `2025-12-29-strukturnulis.md`
- `2025-12-30-debug-refresh-rotation.md`

---

## Template Catatan (wajib dipakai)
Minimal ada:
- Context: lagi ngerjain apa, kondisi awal
- Problem: apa yang bikin macet/risiko
- Notes: temuan, opsi, link internal (file/commit/ADR terkait)
- Next Actions: langkah konkret (cek file apa / ubah apa / test apa)

Kalau catatan sudah selesai dan tidak relevan:
- boleh dipindah ke `archive/` atau dihapus (kalau benar-benar noise).
