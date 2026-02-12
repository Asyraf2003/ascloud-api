# ADR 0011: Centralize Quality Targets ke make/quality.mk

Tanggal: 2026-02-11  
Status: Accepted

## Context (Keadaan awal)
- Target `audit` membutuhkan quality gates: `fmt-check`, `test`, `vet`, `check`.
- Terjadi kondisi di mana `fmt-check` muncul di database Make, tetapi tidak memiliki rule/recipe yang dapat dieksekusi, sehingga `make audit` gagal dengan:
  - `No rule to make target 'fmt-check', needed by 'audit'.`

## Problem (Masalah)
- `make audit` tidak deterministik karena definisi target quality bisa “hilang”/tidak terbaca dengan benar oleh Make.
- Audit harus stabil dan konsisten sebagai Definition of Done.

## Options (Opsi)
1) Memperbaiki `make/go.mk` agar semua target quality selalu terdefinisi dengan benar
2) Membuat satu file khusus (`make/quality.mk`) sebagai sumber kebenaran untuk target quality dan di-include dari Makefile

## Decision (Keputusan)
- Memilih opsi (2): menambahkan `make/quality.mk` yang mendefinisikan:
  - `fmt`, `fmt-check`
  - `test-unit`, `test-component`, `test-integration`, `test`
  - `vet`
  - `check`
- `Makefile` meng-include `make/quality.mk` agar quality gates selalu tersedia untuk `make audit`.

## Consequences (Dampak)
Positif:
- `make audit` kembali deterministik dan stabil.
- CI/dev environment tidak tergantung pada kondisi format/parse file make lain.
- Tidak mempengaruhi arsitektur hexagonal aplikasi (hanya build hygiene).

Negatif/Risiko:
- Ada duplikasi konsep quality gates jika file lain masih mendefinisikan target serupa (perlu disiplin agar `make/quality.mk` jadi sumber kebenaran).
