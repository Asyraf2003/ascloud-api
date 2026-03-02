Roadmap “10 step sampai public-ready” (struktur detail)

Di bawah ini versi yang lebih “operasional”: tiap step ada tujuan, output/deliverables, komponen utama, dan cek lulus (acceptance).

1) Skeleton MVP E2E

Tujuan: repo bisa jalan end-to-end tanpa fitur lengkap, tapi wiring rapi, boundary aman, CI nyaman.
Deliverables:

cmd/api, cmd/worker, cmd/migrate bisa build/run.

Router dasar (health, version) jalan.

Konfigurasi env terstruktur (tanpa hardcode).

Audit boundaries & test dasar lolos.
Cek lulus:

go test ./...

make audit

Minimal sanity curl endpoint yang ada.

2) Auth Google OIDC + refresh rotation

Tujuan: login aman untuk dashboard/API, refresh token rotation, cookie + CSRF + origin policy beres.
Deliverables:

Google OIDC callback flow.

Session/JWT untuk protected endpoints.

Refresh token rotation + anti-reuse + revoke.

Cookie HttpOnly + mitigasi CSRF (header token + cookie).

CORS allowlist ketat.
Cek lulus:

Component tests auth (tags component) lulus.

Tidak ada leak token/secret di log.

Semua error lewat presenter.HTTPErrorHandler.

3) Model data inti

Tujuan: fondasi data yang dipakai step 4–10 stabil dan vendor-first sesuai MVP (DynamoDB), tanpa ngerusak boundary.
Deliverables:

Model domain & ports “final-ish” untuk:

account/identity/session/audit

hosting: site/upload/release (minimal)

Storage utama sesuai keputusan MVP: DynamoDB.

Wiring default runtime tidak tergantung legacy postgres.

ADR terkait keputusan storage & wiring (contoh: ADR 0015 yang kamu tulis).
Cek lulus:

Default build/run tanpa tag legacy.

make audit + go test ./... lulus.

4) Upload pipeline + SQS

Tujuan: user bisa initiate upload → upload ZIP ke S3 → complete upload → enqueue deploy job ke SQS.
Deliverables:

Initiate upload:

buat record upload (DDB)

return presigned PUT URL (S3)

Complete upload:

HeadObject untuk verifikasi ZIP ada & ambil size

enforce ZIP size limit (kalau ada)

status upload jadi queued

enqueue DeployMessage ke SQS
Cek lulus:

Endpoint initiate/complete jalan.

Message format stabil (kontrak job jelas).

Error/status upload konsisten (gagal enqueue = status failed).

5) Worker deploy engine + ZIP security + quota

Tujuan: worker drain SQS, proses ZIP aman, publish release immutable ke S3, update metadata.
Deliverables inti:

Worker handler:

parse DeployMessage (upload_id, queued_at_unix)

load upload metadata (DDB)

download ZIP dari S3

ZIP security wajib:

anti zip slip (path traversal)

anti zip bomb (limit total extracted bytes, jumlah file, depth, timeout)

tolak symlink/hardlink

limit ukuran file per-entry (kalau perlu)

Quota:

hard fail deploy kalau extracted total > 20MB (sesuai keputusan MVP)

Release publish:

extract ke prefix immutable: sites/{site_id}/releases/{release_id}/...

update current_release_id (di data store) untuk nanti dipakai Step 6

status upload: deployed / failed
Cek lulus:

Satu job berhasil menghasilkan release prefix baru.

Job berbahaya (zip slip/bomb/symlink) ditolak dengan aman.

Over-quota extracted -> deploy gagal (hard fail) dan status konsisten.

Logging redacted (tidak bocor token/cookie).

6) Edge routing CloudFront Function + KVS

Tujuan: request Host diarahkan ke site & release tanpa hit origin DB; support suspend & rollback instan.
Deliverables:

KVS mapping: host -> site_id -> current_release_id + suspended flag

CloudFront Function:

baca host

lookup KVS

rewrite path ke S3 origin prefix release

kalau suspended -> response 403/maintenance

Rollback = update pointer current_release_id saja.
Cek lulus:

Perubahan pointer langsung ngaruh tanpa redeploy konten.

Host unknown -> 404 yang bersih.

Suspended -> respons konsisten.

7) Dashboard minimal

Tujuan: UI paling minimal untuk orang awam: login, upload ZIP, lihat status, rollback.
Deliverables:

List sites / create site (kalau scope MVP)

Upload flow (initiate + PUT + complete)

Status deploy + history release

Tombol rollback (pointer update)
Cek lulus:

Happy path bisa dipakai end-to-end tanpa “klik ini lalu SSH ke server”.

8) Observability + audit trail

Tujuan: produksi tanpa observability itu cuma “tebak-tebakan mahal”.
Deliverables:

Structured logging + request id

Metrics penting: deploy duration, fail reasons, extracted bytes, queue lag

Audit trail untuk action penting (deploy, rollback, suspend, auth events)

Alarm minimal (SQS DLQ, error rate, worker failures)
Cek lulus:

Bisa jawab: “kenapa deploy gagal?” tanpa membuka kode 2 jam.

9) Abuse & safety baseline

Tujuan: sebelum publik, minimal defense sudah ada.
Deliverables:

Rate limit (terutama auth & upload endpoints)

Validasi ukuran & tipe konten

DLQ + retry policy yang masuk akal

Sanitasi file extracted (no executable policies kalau diperlukan)

Proteksi data sensitif (redaction log)
Cek lulus:

Serangan bodoh (spam upload, zip bomb, brute refresh) tidak bikin layanan ambruk.

10) Launch checklist

Tujuan: rilis publik itu operasi, bukan sekadar “merge ke main”.
Deliverables:

Domain + TLS + CORS policy final

Runbook (deploy, rollback, incident handling)

IAM least-privilege (S3/SQS/DDB/CF)

Budget guardrails + alerts

Docs onboarding user
Cek lulus:

Ada checklist yang benar-benar bisa diikuti manusia saat panic.

Aturan & cara berpikir yang dipakai di project ini (detail)

Ini “konstitusi” kerja kita. Bukan karena suka aturan, tapi karena tanpa aturan repo akan jadi museum kompromi.

A) Prinsip arsitektur (yang tidak boleh kebalik)

AWS-first untuk MVP: DynamoDB, S3, SQS, CloudFront (provider lain dianggap inactive untuk sekarang).

Event-driven: Upload → enqueue → worker deploy (bukan synchronous deploy di API).

Immutable releases: deploy = prefix release baru; dilarang overwrite konten live.

Pointer-based rollback: rollback cuma ubah current_release_id, bukan copy file.

Edge is king: routing di CloudFront Function + KVS, bukan query DB per request.

B) Boundary hexagonal ketat

Domain/usecase tidak boleh bergantung vendor/platform.

Adapter vendor (AWS) hanya implement ports.

Struktur folder = struktur import (1 folder = 1 package).

Harus lolos boundary enforcement (scripts/audit_boundaries.sh, make audit).

C) Kontrak publik “dilarang rusak”

Kontrak yang dianggap publik dan harus stabil:

Router registration contract (router/v1 register)

Presenter error contract: semua error response lewat presenter.HTTPErrorHandler

Endpoint auth yang sudah ada: jangan berubah tanpa ADR + alasan kuat.

D) Error handling & logging (keamanan)

Tidak boleh bocorin: stacktrace vendor, SQL, token, cookie, header sensitif.

Logging wajib redaction untuk Authorization, Cookie, ApiKey, token.

Error response harus bersih, konsisten, dan tidak “ngobrol kebanyakan”.

E) Debug endpoints

Dilarang nambah debug endpoint kecuali diminta user.

Kalau ada, wajib gated: DEBUG_ROUTES=1.

F) JSONB policy

JSONB hanya internal storage.

Tidak boleh “lempar raw JSONB” ke API response.

Harus allowlist fields + redact yang sensitif.

G) Aturan ukuran file dan kebersihan paket

Target file <= 100 lines. Kalau lewat, boleh, tapi harus ada alasan (#TODO justify) dan tetap rapi.

Hindari “god file” yang isinya semua hal.

H) Workflow kerja (yang kita ikuti biar nggak ngarang)

Snapshot dulu sebelum blueprint/perubahan besar (biar ada baseline).

Verifikasi berbasis data repo untuk hal kritikal: wiring, kontrak publik, flow security.

Kalau belum pasti, cari dulu (mis. rg), baru kasih instruksi spesifik.

Kalau butuh keputusan, tidak tanya kosong.

Harus kasih beberapa opsi + plus/minus + rekomendasi yang nyambung ke blueprint & kondisi repo.

Step-by-step mode untuk audit/debug:

Kasih 1 langkah berikutnya saja, tunggu hasil user, baru lanjut.

I) Definition of Done (DoD) setiap perubahan

Minimal:

gofmt (file terkait)

go test ./...

go vet (kalau dipakai di pipeline)

make audit

Component tests jika menyentuh transport/auth

Sanity curl untuk endpoint yang disentuh

Laporkan: daftar file berubah + isi penuh (kalau lagi review intens)
