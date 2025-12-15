Prompt Template (untuk tiap task)

Ini template prompt yang kamu kirim ke GPT setiap kali mau ngerjain fitur. Ada placeholder yang kamu ganti.

KONTEKS REPO
- Module: [MODULE_PATH]
- Aturan repo: ikuti docs/AI_RULES.md (hard rules + boundaries + DoD).
- Struktur router: internal/http/router/* (router induk + v1 modular).
- Presenter: internal/http/presenter/* (success/auth/billing/error).
- Error response wajib JSON envelope via presenter.HTTPErrorHandler. JSONB hanya storage internal.

TASK
- [TULIS TASK DI SINI]
  Contoh: Buat fitur login/register user via Google (tanpa metode lain), tapi struktur siap untuk provider lain di masa depan.

SEBELUM MULAI (WAJIB)
1) Minta snapshot repo ini dulu (jangan asumsi):
   - tree -L 6 internal/http
   - tree -L 6 internal/modules/[TARGET_MODULE]
   - cat internal/http/router/router.go
   - cat internal/http/router/v1/router.go
   - cat internal/http/presenter/error.go
   - rg -n "^package " internal/http/router
   - [EXTRA SNAPSHOT FILES, kalau relevan]
2) Tanyakan keputusan yang dibutuhkan (kalau belum jelas):
   - Target client: web / mobile / keduanya?
   - Token delivery: refresh via HttpOnly cookie atau via response body?
   - Expiry access token & refresh token?
   - Apakah butuh “trust score” / step-up auth?
   - Audit events apa saja yang wajib dicatat?

WORKFLOW
- Setelah snapshot + jawaban keputusan terkumpul:
  1) Ringkas requirement final (bullet list).
  2) Kritik pilihan yang berbahaya dan tawarkan alternatif best practice.
  3) Baru buat blueprint (struktur folder + kontrak ports + flow endpoint).
  4) Minta saya audit blueprint.
  5) Setelah saya setuju, eksekusi implementasi.

DELIVERABLES (SAAT EKSEKUSI)
- Daftar file baru/berubah.
- Isi final tiap file (bukan potongan).
- Command terminal untuk test + expected output.
- Curl sanity tests + expected response.
- Jika ada bug, perbaiki hanya file terkait (minim efek berantai).
- Tidak boleh ada file >100 baris (split sesuai tanggung jawab).

Bagian yang kamu ganti tiap task

[TULIS TASK DI SINI]

[TARGET_MODULE] (misal auth, hosting)

[EXTRA SNAPSHOT FILES] (misal internal/platform/google/idtoken_verifier.go)


------------------------------------------------------------------------------------------------------------------------------------------------------------------------

<!-- header prompt untuk chat terbaru -->

Header chat versi ringkas yang cukup “mengunci”

Ini yang lu paste di awal chat baru (singkat tapi mematikan):

REPO HEADER (ringkas)
- Module: [MODULE_PATH]
- Ikuti docs/AI_RULES.md (hard rules).
- Kontrak stabil: router.Register, v1.Register, presenter.HTTPErrorHandler.
- Error response: JSON envelope via HTTPErrorHandler (no secrets).
- JSONB hanya storage/meta internal, tidak boleh mentah ke client.
- Debug routes wajib gated DEBUG_ROUTES=1.
- 1 folder = 1 package. File <=100 baris (split by responsibility).
- DoD: gofmt, go test ./..., go vet ./..., curl sanity.

SNAPSHOT WAJIB (paste output):
- tree -L 6 internal/http
- tree -L 6 internal/modules/[TARGET]
- cat internal/http/router/v1/router.go
- cat internal/http/presenter/error.go
- rg -n "^package " internal/http/router