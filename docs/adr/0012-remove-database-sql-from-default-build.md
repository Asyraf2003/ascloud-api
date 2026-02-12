# ADR 0012: Remove database/sql from default build (AWS-first)

Tanggal: 2026-02-12  
Status: Accepted

## Context (Keadaan awal)
- Default build masih menarik dependency `database/sql` walaupun Postgres sudah dilabeli `legacy_*`.
- Penyebab: transport HTTP (`internal/transport/http/server`) dan system readiness handler (`internal/modules/system/transport/http`) mengimpor `database/sql`.

## Problem (Masalah)
- Blueprint AWS-first tidak menggunakan SQL database pada jalur default (DynamoDB sebagai DB utama).
- Dependency `database/sql` di transport layer membuat default build tidak “bersih” dan mengikat lapisan HTTP ke detail vendor DB.

## Options (Opsi)
1) Refactor transport untuk tidak mengimpor `database/sql` (pakai `any` + structural ping interface pada readiness).
2) Split file dengan build tags (default vs postgres), berpotensi duplikasi.
3) Biarkan saja `database/sql` di default build.

## Decision (Keputusan)
- Pilih opsi (1):
  - `server.New` menerima `db any` dan menyimpan ke context tanpa dependency `database/sql`.
  - Readiness handler menggunakan structural interface `PingContext(context.Context)` dan deteksi nil dengan `reflect`, tanpa import `database/sql`.
- Hasil: default build tidak menarik `database/sql`, namun legacy Postgres tetap bisa berjalan via build tag.

## Consequences (Dampak)
Positif:
- Default build bersih dan selaras dengan AWS-first blueprint.
- Transport HTTP tidak bergantung pada detail SQL/driver.
- Legacy Postgres masih bisa diaktifkan saat dibutuhkan.

Negatif/Risiko:
- Penggunaan `any` untuk context DB adalah kompatibilitas sementara; harus digantikan dengan wiring ports/usecase yang eksplisit saat AWS runtime sudah siap.
