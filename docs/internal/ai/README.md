# Internal AI Docs

Folder ini isinya **aturan kerja untuk AI** saat bantu perubahan repo.
Targetnya simpel: AI tidak halu, tidak bikin efek berantai, perubahan konsisten, dan audit gampang.

> Ini dokumen internal engineering. Bukan dokumentasi produk.

---

## Urutan baca (kalau baru masuk repo)

1) **AI_RULES.md**  
   Hard rules + boundaries + DoD. Ini “konstitusi” repo. Kalau ada konflik, file ini menang.

2) **AI_PROMPT.md**  
   Template prompt untuk task apa pun. Dipakai tiap kali minta AI ngerjain sesuatu.

3) **EXAMPLE_PROMPT.md**  
   Contoh prompt yang sudah terisi (biar kelihatan bentuk finalnya).

4) **AI_TUNING.md**  
   Mode “budget” (message terbatas) dan snapshot wajib minimal.

5) **DESCRIPTION.md**  
   Penjelasan kenapa template ini efektif + bagian yang harus diganti.

---

## Kontrak minimal yang harus selalu dijaga
- AI **wajib minta snapshot** sebelum usul blueprint/ubah kode.
- AI **tidak boleh** mengubah signature publik yang dilindungi (lihat AI_RULES).
- Perubahan harus **minim efek berantai** dan punya DoD jelas (gofmt/test/vet + sanity curl).
- Hard rules > preferensi gaya.

---

## Update policy (biar tidak membusuk)
Kalau ada perubahan besar:
- **Rule/kontrak berubah** → update `AI_RULES.md` + buat ADR bila relevan.
- **Cara kerja berubah** → update `AI_PROMPT.md` + `EXAMPLE_PROMPT.md`.
- **Path/struktur repo berubah** → update snapshot list (jangan ada path ngaco).

---

## Checklist cepat sebelum start task dengan AI
- [ ] Repo header sudah benar (module path, contracts).
- [ ] Snapshot sudah ditempel user (tree + cat file inti).
- [ ] Keputusan produk yang krusial sudah jelas (token model, client type, expiry, dll).
- [ ] Blueprint sudah ditulis sebelum eksekusi.
