# ADR 0010: Decouple Audit dari Docker + Legacy Postgres Tooling

Tanggal: 2026-02-11  
Status: Accepted

## Context (Keadaan awal)
- `make audit` gagal di environment tanpa Docker karena `prereq` mewajibkan `docker`.
- Repo masih punya dev workflow Postgres (docker compose), psql helper, dan migrasi SQL.
- Blueprint produk: AWS-first (DynamoDB, S3, SQS, CloudFront), sehingga Postgres diperlakukan sebagai artefak transisi/masa lalu.

## Problem (Masalah)
- Audit/CI seharusnya bisa berjalan tanpa dependency runtime dev database (Docker/Postgres).
- Workflow Makefile memberi sinyal seolah Postgres masih jalur utama, bertentangan dengan blueprint.

## Options (Opsi)
1) Tetap mewajibkan Docker untuk semua target
2) Docker hanya wajib untuk target yang benar-benar memakai docker compose / Postgres tooling

## Decision (Keputusan)
- Memecah prereq menjadi:
  - `prereq-core`: bash, rg, go (tanpa Docker)
  - `prereq-docker`: docker
  - `prereq`: gabungan keduanya (untuk target yang memang butuh Docker)
- Mengubah `make audit` agar bergantung pada `prereq-core` (audit tetap jalan tanpa Docker).
- Menandai tooling Postgres sebagai legacy di UI Make (`make help`) dan menyediakan alias deprecated untuk kompatibilitas.

## Consequences (Dampak)
Positif:
- `make audit` bisa berjalan di mesin/CI tanpa Docker.
- Dependency Docker tidak “menyandera” quality gates.
- Repo lebih jujur terhadap arah AWS-first (Postgres diposisikan legacy).

Negatif/Risiko:
- Ada tambahan kompleksitas pada Make targets (legacy vs deprecated alias).
- Postgres masih bisa digunakan via legacy targets sampai migrasi ke DynamoDB selesai (potensi kebingungan jika dokumentasi tidak dibaca).
