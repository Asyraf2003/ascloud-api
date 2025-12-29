# DESCRIPTION

Dokumen ini menjelaskan:
- Kenapa paket dokumen AI ini efektif (anti-halu + anti-efek-berantai)
- Cara pakai yang benar
- Bagian yang wajib disesuaikan tiap task

---

## Kenapa ini efektif
1) **Memaksa snapshot**
   AI tidak boleh ngarang path/kontrak yang sudah berubah. Snapshot membuat perubahan selalu grounded.

2) **Ada boundary rules**
   Mencegah “benerin A malah nyenggol B sampai C rusak” (efek berantai).

3) **Ada Definition of Done (DoD)**
   AI wajib kasih command dan output yang bisa kamu audit cepat: gofmt/test/vet + sanity curl.

4) **Keputusan eksplisit**
   Hal seperti token delivery (cookie vs body), client type (browser vs api-to-api), expiry, step-up, itu beda dunia security-nya.
   Kalau keputusan belum jelas, AI harus berhenti dan tanya.

---

## Cara pakai (urutan yang benar)
1) Buka `AI_RULES.md` dan pastikan kontrak repo masih valid.
2) Copy template dari `AI_PROMPT.md`.
3) Isi TASK secara spesifik.
4) Paste output snapshot yang diminta.
5) Jawab decision list (kalau ada).
6) Baru minta AI bikin blueprint, lalu eksekusi.

---

## Bagian yang WAJIB kamu sesuaikan tiap task

### 1) Module path
Isi sesuai `go.mod`:
- `example.com/your-api`

### 2) Target module
Contoh:
- `auth`, `hosting`, `domains`, `billing` (kalau ada)

### 3) Extra snapshot files
Kalau task menyentuh area tertentu, tambahkan file snapshot yang relevan.
Contoh:

**auth**
- `internal/platform/google/*`
- `internal/platform/token/*`
- `internal/modules/auth/*`
- `internal/transport/http/middleware/*` (CSRF/Origin/JWT)

**hosting**
- `internal/platform/objectstore/*`
- `internal/platform/edge/*`
- `internal/platform/queue/*`
- `internal/modules/hosting/*`

**datastore**
- `internal/platform/datastore/postgres/*`
- migrations terkait

### 4) Decision list
Beda task, beda keputusan.
Tapi selalu minimal punya:
- target client (browser vs api-to-api)
- token delivery
- expiry
- auth/trust assumptions
- audit events minimal

---

## Pitfalls yang sering bikin repo busuk
- Snapshot list tidak di-update setelah refactor (akhirnya AI “halu”).
- Hard rules ditulis tapi tidak ada enforcement (pastikan `make audit` tetap relevan).
- Doc jadi terlalu panjang dan tidak operasional (hindari esai, fokus checklist + kontrak).
- Menyimpan draft/ide liar di ADR (draft taruh di `docs/notes/`, ADR hanya keputusan final).

---

## Kapan perlu update dokumen ini
- Struktur repo berubah (path/kontrak/router/presenter berubah)
- Policy security berubah (CSRF/CORS/session/jwt model berubah)
- Audit/DoD berubah (command, tag test, sanity curl)
