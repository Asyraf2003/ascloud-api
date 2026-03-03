# ADR 0018: Hosting Step 6 — Edge Routing (CloudFront Function + KeyValueStore) untuk Host→Release, Suspend, dan Rollback Instan

Tanggal: 2026-03-02
Status: Accepted

## Context (Keadaan awal)
- Step 5 (Worker Deploy Engine) sudah menghasilkan release immutable di S3 dengan struktur:
  - `sites/{site_id}/releases/{release_id}/index.html`
- Goal Step 6 (baseline arsitektur):
  - Routing publik berbasis `Host` tanpa hit origin DB.
  - Support suspend instan.
  - Rollback instan hanya dengan mengubah pointer `current_release_id` (tanpa rebuild/move file).
- Domain publik yang dipakai:
  - `asyrafcloud.my.id` (DNS dikelola di Hostinger / nameserver `dns-parking.com`).
- CDN & Edge compute yang disepakati:
  - CloudFront + CloudFront Function + CloudFront KeyValueStore (KVS).

## Problem (Masalah)
Kita butuh mekanisme edge routing yang:
1) Membaca header `Host` dari request viewer.
2) Menentukan `site_id` dan `current_release_id` tanpa query DynamoDB (zero origin DB hit).
3) Melakukan rewrite path agar request menuju prefix release di S3:
   - `/{path}` → `/sites/{site_id}/releases/{current_release_id}/{path}`
4) Mendukung:
   - Unknown host → 404 bersih.
   - Suspended host → response maintenance konsisten.
5) Rollback instan:
   - cukup update pointer `current_release_id` saja (tanpa publish ulang function / redeploy konten).

## Constraints (Batasan)
- Origin memakai S3 REST endpoint (bukan S3 website endpoint).
- Bucket harus private; akses ke origin via CloudFront menggunakan OAC.
- Routing harus bekerja untuk wildcard host `*.asyrafcloud.my.id`.

## Options (Opsi)
### Opsi 1 — CloudFront Flat-rate plan Pro (pakai KVS)
- Tetap di distribution flat-rate, upgrade plan ke Pro agar edge KVS bisa di-associate.
Kelebihan:
- Tidak perlu migrasi distribution/DNS.
Kekurangan:
- Biaya fixed bulanan.
- Pada kondisi awal, Free tier tidak mengizinkan association CloudFront Function yang memakai KVS.

### Opsi 2 — CloudFront Pay-as-you-go distribution (pakai KVS)  ✅ dipilih
- Buat distribution baru pay-as-you-go, gunakan CloudFront Function runtime 2.0 + KVS, cutover DNS wildcard.
Kelebihan:
- Tidak ada biaya fixed; bayar berbasis usage.
- Fitur KVS tetap dapat dipakai untuk memenuhi DoD rollback/suspend instan.
Kekurangan:
- Perlu migrasi distribution dan update bucket policy/OAC untuk distribution baru.
- DNS cutover memerlukan verifikasi resolusi.

### Opsi 3 — Tanpa KVS (hardcode mapping di code function)
- Mapping host→release disimpan di source code function.
Kelebihan:
- Tidak perlu KVS.
Kekurangan:
- Gagal memenuhi DoD rollback instan tanpa redeploy (perubahan pointer butuh publish ulang function).

### Opsi 4 — Per-site distribution atau update config distribution untuk rollback
- Tiap site punya distribution sendiri atau rollback dilakukan via perubahan Origin Path/behavior.
Kelebihan:
- Konfigurasi eksplisit.
Kekurangan:
- Operasional berat, scaling buruk.
- Rollback bukan “pointer-only” instan (butuh update distribusi dan propagasi).

## Decision (Keputusan)
Memilih **Opsi 2 (Pay-as-you-go distribution)** dengan desain:
- Wildcard DNS:
  - `*.asyrafcloud.my.id` (dan minimal `site-1.asyrafcloud.my.id`) CNAME → `dxxxxx.cloudfront.net` (domain distribution pay-as-you-go).
- Origin S3 REST endpoint + **OAC** (bucket private).
- **CloudFront KeyValueStore** menyimpan mapping per host.
- **CloudFront Function (viewer-request, runtime 2.0)**:
  - Read Host → lookup KVS → rewrite URI → return request.
  - Unknown host → return 404 text/plain + no-store.
  - Suspended → return 403 text/plain + no-store.
- Rollback instan dilakukan dengan update value di KVS saja.

## Design (Rancangan)
### KVS schema (1 lookup per request)
Key:
- `host` (lowercase, tanpa port, tanpa trailing dot)
  - contoh: `site-1.asyrafcloud.my.id`

Value (JSON):
- `site_id` (string)
- `current_release_id` (string UUID)
- `suspended` (bool)

Contoh:
- `site-1.asyrafcloud.my.id` →
  - `{"site_id":"site-1","current_release_id":"<release_id>","suspended":false}`

### CloudFront Function behavior
Event:
- viewer-request

Algoritma:
1) Normalize host:
   - toLowerCase
   - strip `:port`
   - strip trailing `.`
2) `mapping = kvs.get(host, {format:"json"})`
   - jika tidak ada → 404 bersih.
3) Jika `mapping.suspended == true` → 403 maintenance (no-store).
4) Rewrite URI:
   - Jika `uri == "/"` → `/index.html`
   - Jika `uri` berakhiran `/` → append `index.html`
   - Prefix:
     - `/sites/{site_id}/releases/{current_release_id}`
   - Final:
     - `request.uri = prefix + uri`
5) Basic hardening:
   - reject traversal `..` dan backslash `\` (return 404) untuk menghindari path ambiguity.
6) Return request (forward ke S3 origin via CloudFront caching layer).

Catatan perilaku S3:
- Untuk bucket private + OAC, object “tidak ada” bisa tampak sebagai `403 AccessDenied` (S3 menghindari bocor eksistensi objek). Ini dianggap normal untuk mismatch release_id, dan diselesaikan dengan pointer KVS yang valid.

## Implementation Notes (Implementasi)
- KVS ARN yang dipakai:
  - `arn:aws:cloudfront::219673121952:key-value-store/170fcc97-8ab8-46de-b053-2aae144ce81d`
- Function harus runtime 2.0 agar dapat memakai `cf.kvs()` helper.
- Function diasosiasikan ke KVS via CloudFront console (bukan hardcode ARN di code).
- Distribution pay-as-you-go memakai:
  - S3 REST origin + OAC
  - Default behavior: viewer-request function association
  - Viewer protocol policy: redirect HTTP → HTTPS (recommended)
- DNS:
  - Tambah CNAME `*` dan `site-1` ke domain distribution baru.
  - Validasi resolusi dengan `dig` ke authoritative nameserver.

## Verification (Acceptance / DoD)
### 1) DNS wildcard resolve ke CloudFront
- `dig +short CNAME site-1.asyrafcloud.my.id @ns1.dns-parking.com` → `dxxxxx.cloudfront.net.`
- `dig +short CNAME anything.asyrafcloud.my.id @ns1.dns-parking.com` → `dxxxxx.cloudfront.net.`

### 2) Origin connectivity (CloudFront→S3 via OAC) untuk object release yang ada
- `curl -I https://site-1.asyrafcloud.my.id/sites/site-1/releases/<known_release_id>/index.html` → `200`

### 3) Routing aktif: `/` menjadi release `index.html`
- `curl -I https://site-1.asyrafcloud.my.id/` → `200`

### 4) Unknown host → 404 bersih (tanpa hit DB)
- `curl -I https://unknown.asyrafcloud.my.id/` → `404`
- header:
  - `content-type: text/plain; charset=utf-8`
  - `cache-control: no-store, max-age=0`

### 5) Suspended → respons konsisten
- Set KVS `suspended:true` untuk host valid.
- `curl -I https://site-1.asyrafcloud.my.id/` → `403`
- header:
  - `content-type: text/plain; charset=utf-8`
  - `cache-control: no-store, max-age=0`

### 6) Rollback pointer instan (tanpa redeploy konten/function)
- Siapkan 2 release valid (konten berbeda, contoh “ok” vs “v2”).
- Update KVS `current_release_id` ke release lain:
  - `curl https://site-1.asyrafcloud.my.id/ | head` berubah dari `<h1>ok</h1>` → `<h1>v2</h1>` tanpa publish ulang function/distribution.
- Update kembali ke release lama:
  - output kembali ke `<h1>ok</h1>`.

Bukti observasi:
- Switching pointer berhasil dengan delay beberapa detik sesuai propagasi update KVS dan caching edge, tanpa redeploy function maupun konten.

## Consequences (Dampak)
### Positif
- Step 6 goals tercapai:
  - Host→site/release dilakukan di edge tanpa query DB.
  - Suspend instan via flag KVS.
  - Rollback instan via update pointer `current_release_id` di KVS.
- Immutable releases tetap jadi source of truth untuk konten; edge hanya pointer.
- Unknown host ditangani dengan response bersih di edge.

### Negatif / Risiko
- Migrasi distribution (flat-rate → pay-as-you-go) membutuhkan:
  - update DNS CNAME
  - update OAC/bucket policy agar distribution baru bisa akses S3
- Jika KVS menunjuk release_id yang tidak ada, viewer dapat melihat `403 AccessDenied` dari origin (S3 private behavior).
  - Opsional mitigasi: Custom error responses (map 403→404) di CloudFront, bila ingin UX lebih “not found”.

## Follow-ups
- Integrasi control-plane/worker:
  - Saat deploy sukses (Step 5), update KVS untuk host yang terkait (atau melalui API rollback endpoint) agar edge pointer sinkron.
  - Saat suspend/un-suspend site, update KVS flag.
- Observability:
  - Logging/metrics untuk rate unknown host, suspend hits, dan mismatch release pointer.
- Dokumentasi:
  - Tambah doc runbook: cara update KVS untuk rollback/suspend + verifikasi curl/dig.
