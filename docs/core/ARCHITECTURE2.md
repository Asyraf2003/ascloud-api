# Blueprint Enterprise — Static Site Hosting via ZIP (AWS-First, Final)

**Status:** FINAL BASELINE  
**Tujuan Dokumen:** Menjadi pegangan arsitektur jangka panjang.  
> Jika ada konflik dengan diskusi lain, dokumen ini yang menang.

---

## 0) Keputusan Non-Negotiable (MVP Pegangan)
* **Provider aktif:** AWS only
* **CDN & Edge:** CloudFront + CloudFront Function + KVS
* **Storage:** S3
* **Compute:** Lambda (API + Worker)
* **Queue:** SQS
* **Metadata DB:** DynamoDB (on-demand)
* **Storage quota:** 20MB extracted size, **HARD FAIL** deploy
* **Pageview quota:** HTML-only, batch dari CloudFront logs, **SOFT WARNING + GRACE**
* **Auth browser:** Dashboard & API 1 eTLD+1, HttpOnly refresh rotation + CSRF
* **Release model:** Immutable release + pointer rollback
* **EC2:** **DILARANG** untuk MVP

---

## 1) Tujuan Sistem

### Target Produk (User-Oriented)
* **Target user:** UMKM / orang awam
* **Flow user:**
    1. Login dengan Google
    2. Upload ZIP static site
    3. Sistem memproses otomatis
    4. User langsung dapat URL (subdomain default)
* **Fitur utama:**
    * Versioning otomatis
    * Rollback instan (1 klik)
* **Output:** Public website, performa tinggi, tanpa konfigurasi teknis.

### Target Operasional (Business & Engineering)
* Biaya idle = 0
* Skalabilitas: 100 → 1.000 → 10.000 user tanpa redesign besar
* Arsitektur modular/hexagonal: Domain logic tidak terkunci vendor/runtime

---

## 2) Prinsip Arsitektur
* **Event-driven:** Upload → Queue → Worker → Publish
* **Immutable Releases:** Setiap deploy = folder release baru. Tidak pernah overwrite file live.
* **Pointer-based Rollback:** Rollback = ganti `current_release_id`. Tidak ada copy data, tidak rebuild.
* **Serverless Compute:** API & worker berjalan di Lambda.
* **Edge-first Delivery:** CloudFront sebagai titik serving utama.

---

## 3) Teknologi Baseline (Dipaku)

| Layer | Teknologi |
| :--- | :--- |
| **CDN** | CloudFront |
| **Edge Logic** | CloudFront Function + KeyValueStore |
| **API** | AWS Lambda (Go) + Function URL / API Gateway |
| **Worker** | AWS Lambda (Go, SQS triggered) |
| **Queue** | AWS SQS (Standard) |
| **Storage** | AWS S3 |
| **Database** | DynamoDB (on-demand) |
| **Region** | ap-southeast-3 (Jakarta) |
| **Upgrade Path** | ECS Fargate (run-per-job) |

*Provider lain di repo tidak aktif dan tidak digunakan tanpa keputusan eksplisit baru.*

---

## 4) Arsitektur Tingkat Tinggi

### Control Plane
Dashboard → API (Lambda) → DynamoDB + SQS

### Data Plane
Upload ZIP → S3 uploads → Worker (Lambda) → S3 releases

### Edge Plane
Visitor → CloudFront → Function (rewrite) → S3 release folder

---

## 5) Komponen & Tanggung Jawab

### 5.1 API Service (Go, Lambda)
* Google OIDC login
* Session via HttpOnly cookie (Refresh token rotation + anti-reuse)
* Generate pre-signed S3 upload URL
* Create site, release, job
* Enqueue SQS job
* List release, rollback (update pointer)
* Quota & status site (ACTIVE / SUSPENDED)

### 5.2 Worker (Go, Lambda + SQS)
* **Wajib aman. Tidak ada kompromi.**
* Download ZIP dari S3
* **ZIP safety:** Anti zip-slip (`../`), Anti zip-bomb (size, file count, depth), Reject symlink, Timeout.
* Extract ke: `sites/{site_id}/releases/{release_id}/`
* Hitung size hasil extract. **Jika > 20MB → FAIL DEPLOY**.
* Set content-type & cache headers.
* Update job & release status di DynamoDB.
* Hapus ZIP.

### 5.3 Storage (S3)
* `uploads/{job_id}.zip` (sementara)
* `sites/{site_id}/releases/{release_id}/...` (immutable)
* Tidak ada “live folder”.

### 5.4 Edge Routing (CloudFront)
* CloudFront Function + KVS mapping:
    * `host` → `site_id`
    * `site_id` → `current_release_id`
    * `site_id` → `suspended flag`
* **Rewrite internal:** `/sites/{site_id}/releases/{current_release_id}{path}`

---

## 6) Model Data (DynamoDB-Friendly)

* **users:** `user_id` (PK), `google_sub` (unique), `email`, `created_at`
* **sites:** `site_id` (PK), `owner_user_id`, `slug` (unique), `current_release_id`, `status` (ACTIVE / SUSPENDED), `created_at`
* **releases:** `release_id` (PK), `site_id`, `status` (PENDING / SUCCESS / FAILED), `size_bytes`, `error_reason`, `created_at`
* **jobs:** `job_id` (PK), `site_id`, `release_id`, `zip_key`, `status` (QUEUED / RUNNING / SUCCESS / FAILED), `attempts`, `created_at`
* **usage_monthly:** `site_id + yyyymm` (PK), `storage_bytes`, `pageviews_html`, `updated_at`

---

## 7) Alur Utama

### Deploy
1. API create release + job, return pre-signed upload URL.
2. Client upload ZIP ke S3.
3. API enqueue SQS.
4. Worker validate + extract.
5. Update status SUCCESS.
6. API set `current_release_id`.
7. CloudFront serve release terbaru.

### Rollback
1. API update `current_release_id`.
2. Tidak ada file movement, tidak ada rebuild.

---

## 8) Quota & Billing

* **Storage (Hard Limit):** Free 20MB extracted size per site. Jika > 20MB → **DEPLOY FAILED**.
* **Pageviews (Soft Limit):** Dihitung dari HTML requests saja via CloudFront access logs (batch).
* **Enforcement:** Warning → Grace period → Suspend site (Update DB & Sync flag ke KVS).

---

## 9) Keamanan
* ZIP safety ketat (non-negotiable).
* Header: `Content-Type` benar, `X-Content-Type-Options: nosniff`.
* Cache policy: HTML (TTL pendek), Asset (TTL panjang).
* Rate limit endpoint sensitif & Abuse takedown workflow.

---

## 10) Observability & Audit
* Structured logs (`job_id`, `site_id`, `release_id`).
* **Metrics:** Job success rate, Deploy duration, Queue depth.
* **Alarms:** Failure spike, Queue backlog, Lambda error rate.
* **Audit:** Deploy & Rollback (by siapa & kapan).

---

## 11) Roadmap
* Phase 1: MVP static hosting
* Phase 2: Pageview quota + suspension
* Phase 3: Custom subdomain
* Phase 4: BYOD custom domain
* Phase 5: Domain reseller
* Phase 6: Dynamic deploy (Fargate)

---

## 12) Ringkasan Filosofi
1. **Simple for user.**
2. **Strict for system.**
3. **Cheap at idle.**
4. **Scalable by design.**