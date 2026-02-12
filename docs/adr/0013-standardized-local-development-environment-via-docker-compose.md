# ADR 0013: Standardized Local Development Environment via Docker Compose (Legacy Postgres)

Tanggal: 2026-02-12
Status: Accepted

## Context (Keadaan Awal)
Project your-api memiliki arah jangka panjang:
* **AWS-first**: Serverless runtime (Lambda).
* **DynamoDB**: Sebagai primary data store (MVP keputusan A).
* **CloudFront + S3**: Untuk hosting.
* **Postgres**: Hanya bagian dari legacy path.

**Saat ini:**
* Modul auth legacy masih menggunakan Postgres.
* Build default men-disable legacy Postgres via build tag.
* Development membutuhkan database yang konsisten, mudah di-bootstrap, dan tidak tergantung environment OS developer.

**Resource yang sudah ada:**
* `deploy/docker/docker-compose.dev.yml`
* `make/legacy_dev_postgres.mk`
* Target alias `dev-*` dan Guard `prereq-docker`.

## Problem (Masalah)
1.  Dev environment harus reproducible tanpa mencemari domain/usecase layer.
2.  Legacy Postgres tidak boleh menjadi default runtime.
3.  Docker tidak boleh mengubah boundary arsitektur atau menjadi runtime strategy di production.
4.  Butuh standarisasi karena tidak semua developer memiliki Postgres native.
5.  Blueprint AWS-first membutuhkan environment yang bisa berkembang (seperti simulasi LocalStack).

## Architectural Constraints
* **Domain Isolation**: Domain layer tidak boleh tahu keberadaan Docker.
* **Platform Abstraction**: Tetap melalui `internal/platform`.
* **Production Fidelity**: Production runtime harus tetap serverless (AWS).
* **Zero Coupling**: Tidak ada keterkaitan antara business logic dan dev tooling.

## Options

### 1) Native Postgres (Tanpa Docker)
* **Kelebihan**: Simpel untuk satu developer; tidak butuh container runtime.
* **Kekurangan**: Tidak reproducible; setup antar mesin berbeda; sulit diskalakan ke dependency lain (LocalStack/Redis).

### 2) Docker untuk Seluruh Runtime (Containerized API)
* **Kelebihan**: Uniform deployment model.
* **Kekurangan**: Menyimpang dari AWS Lambda blueprint; menggeser arah arsitektur. Ditolak karena bertentangan dengan AWS-first decision.

### 3) Docker hanya untuk Development Dependencies (Dipilih)
* **Kelebihan**: API tetap dijalankan native (`go run`), database via Docker Compose.
* **Detail**: Build tag memisahkan legacy dan default; Docker hanya untuk infra lokal.

## Decision (Keputusan)
Project menetapkan:
1.  **Docker Compose** digunakan eksklusif hanya untuk development dependencies.
2.  **Postgres Legacy** dijalankan melalui perintah:
    ~~~bash
    make legacy-dev-up-postgres
    # Alias: make dev-up
    ~~~
3.  **Runtime Production** tetap serverless (AWS).
4.  **Default Build** tetap men-disable legacy Postgres.
5.  Struktur ini menjadi fondasi untuk penambahan LocalStack (SQS, DynamoDB) dan pengujian AWS event-driven path di masa depan.

## Consequences (Dampak)

### Positif
* Dev environment menjadi deterministic dan memudahkan onboarding.
* Konsisten dengan hexagonal architecture.
* Siap untuk evolusi simulasi AWS tanpa mengganggu migrasi DynamoDB.

### Negatif / Risiko
* Developer wajib menginstal Docker.
* Dibutuhkan disiplin tinggi agar Docker tidak secara tidak sengaja dianggap sebagai default runtime production.
* Legacy Postgres bisa menciptakan ilusi stabilitas jika migrasi ke DynamoDB tidak diprioritaskan.