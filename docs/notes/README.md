# Notes

Folder ini untuk **catatan/draft**: ide, investigasi, coretan debugging, rencana kasar.
Notes **bukan kontrak** dan **bukan keputusan final**.

Kalau sesuatu sudah jadi keputusan yang mengikat arsitektur/kontrak/perilaku:
➡️ pindahkan jadi ADR di `docs/adr/`.

---

## Kapan pakai Notes
- Draft rencana perbaikan
- Hasil investigasi bug
- Checklist implementasi sementara
- Catatan “kenapa dulu begini” tapi belum layak ADR

## Kapan HARUS jadi ADR (bukan Notes)
- Ubah boundary rules
- Ubah auth model / token model
- Ubah data ownership / schema besar
- Ubah strategi hosting (statis/dinamis), queue, provisioning
- Ubah policy security (CORS/CSRF/trust/rate limit)

---

## Format nama file
Gunakan format tanggal supaya gampang audit:
`YYYY-MM-DD-<topik-singkat>.md`

Contoh:
- `2025-12-29-strukturnulis.md`
- `2025-12-30-auth-refresh-reuse-ideas.md`

---

## Template notes (optional)
Minimal:
- Context
- Temuan/masalah
- Rencana / opsi
- Next action
