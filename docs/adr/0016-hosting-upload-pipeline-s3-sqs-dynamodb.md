# ADR XXXX: Hosting Step 4 - Upload Pipeline (S3 Presign + DynamoDB UploadStore + SQS Deploy Queue)

Tanggal: 2026-02-15  
Status: Acceptedb

## Context (Keadaan awal)
- Produk: static site hosting via ZIP (AWS-first) dengan arsitektur event-driven: Upload → Queue → Worker.
- Keputusan MVP yang sudah disepakati:
  - DB: DynamoDB
  - Object storage: S3
  - Queue: SQS
  - Upload quota: hard fail jika melebihi 20MB (extract quota ditegakkan di tahap berikutnya).
- Di repo, modul hosting sebelumnya masih placeholder di layer transport/wire.

## Problem (Masalah)
Kita butuh pondasi “upload pipeline” yang:
- Bisa membuat upload record dan menghasilkan presigned URL untuk upload ZIP ke S3.
- Bisa memvalidasi ukuran file yang di-upload (minimal via HEAD) dan menolak jika melebihi batas.
- Bisa menandai upload “queued” dan mengirim pesan ke SQS untuk diproses worker (deploy engine) di tahap berikutnya.
- Tetap mengikuti batas arsitektur (ports/usecase/adapters/store) dan pola repo yang sudah ada.

## Options (Opsi)
1) **S3 presign + DynamoDB metadata + SQS enqueue (event-driven)**  
   - Presign untuk PUT ZIP ke S3  
   - Store upload metadata ke DynamoDB  
   - Setelah upload selesai, lakukan HEAD untuk size check  
   - Jika lolos, enqueue message ke SQS  
   Kelebihan: sesuai blueprint, scalable, cost idle $0.  
   Kekurangan: enforce ukuran presign tidak “hard” tanpa policy; butuh HEAD check.

2) **Upload lewat API server (proxy upload) + simpan langsung**  
   Kelebihan: bisa enforce size di server.  
   Kekurangan: mahal (bandwidth/CPU), tidak serverless-friendly, memperburuk latency dan biaya.

## Decision (Keputusan)
Memilih **Opsi (1)**: event-driven upload pipeline (S3 + DynamoDB + SQS) karena ini paling konsisten dengan blueprint AWS-first dan target operasional serverless.

Implementasi yang diambil:
- `hosting/usecase`:
  - `InitiateUpload(siteID)`:
    - buat `UploadID`, bentuk `objectKey` `sites/<site_id>/uploads/<upload_id>.zip`
    - minta presigned PUT URL ke `ObjectStore`
    - simpan record upload di `UploadStore` status `initiated`
    - return `PutURL`, `ExpiresAtUnix`, `MaxBytes`
  - `CompleteUpload(siteID, uploadID)`:
    - ambil record upload
    - HEAD S3 untuk `sizeBytes`
    - jika > `MaxUploadBytes`: set status `failed` dan return `ErrUploadTooLarge`
    - set status `queued` dan simpan `size_bytes`
    - enqueue `DeployMessage` ke `DeployQueue` (SQS) dengan `QueuedAtUnix`
- `hosting/ports`:
  - `UploadStore`, `ObjectStore`, `DeployQueue` dan payload `DeployMessage`.
- `hosting/adapters/aws`:
  - `ObjectStore` (S3): `PresignPutZip`, `Head` (HeadObject ContentLength).
  - `DeployQueue` (SQS): `EnqueueDeploy` (SendMessage JSON).
- `hosting/store/dynamodb`:
  - `UploadStore` untuk Put/Get/Update status dan size.
- Wiring:
  - `auth` sudah dibuktikan pattern wiring via holder + `WireAuthGoogle(...)`.
  - `hosting/wire/hosting_wire_dynamodb.go` disiapkan sebagai tempat wiring hosting (masih placeholder pada tahap ini).

Konvensi waktu:
- **Internal**: `time.Time` untuk `CreatedAt` dan expiry math.
- **Boundary** (HTTP payload / queue message / numeric field DDB): Unix seconds `int64` (`ExpiresAtUnix`, `QueuedAtUnix`).

## Consequences (Dampak)
Positif:
- Upload pipeline siap untuk MVP: presign upload → metadata → verifikasi ukuran → enqueue deploy.
- Selaras dengan blueprint event-driven dan keputusan AWS-first (S3/SQS/DynamoDB).
- Data boundary lebih stabil: payload waktu diekspor sebagai `int64` Unix untuk JSON/queue.

Negatif/Risiko:
- Presigned URL tidak otomatis enforce `MaxUploadBytes` tanpa policy/conditions; mitigasi dilakukan dengan HEAD check saat `CompleteUpload`.
- Tahap ini belum mencakup deploy engine (extract ZIP, release immutable, rollback pointer, CloudFront routing); itu masuk Step 5/6.
- Upload status model masih sederhana (initiated/uploaded/queued/failed). Jika butuh retry/visibility lebih detail, akan berkembang di tahap berikutnya.
