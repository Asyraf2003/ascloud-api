# Internal AI Docs

Dokumen internal untuk mengarahkan AI saat bantu perubahan repo.
Target: tidak halu, perubahan konsisten, minim efek berantai, audit gampang.

## Urutan baca
1) `AI_RULES.md`  
   Hard rules + boundaries + DoD. Ini paling tinggi.

2) `AI_PROMPT.md`  
   Template utama untuk semua task.

3) `EXAMPLE_PROMPT.md`  
   Contoh prompt yang sudah terisi.

4) `AI_TUNING.md`  
   Mode hemat message/token.

5) `DESCRIPTION.md`  
   Ringkasan alasan template + apa yang wajib diisi.

## Update policy
Kalau struktur repo berubah, update:
- snapshot list di `AI_PROMPT.md` dan `AI_TUNING.md`
- link di `docs/README.md` bila path berubah
