# ADR 0003: Domain boleh import internal/shared (pure)

Tanggal: 2025-12-29
Status: Accepted

## Context (Keadaan awal)
- Repo menargetkan boundaries hexagonal/modular yang ketat.
- Aturan boundary awal memaksa `domain` hanya boleh standard library.
- Di sisi lain, repo butuh tipe/utilitas netral yang dipakai lintas modul, misalnya error type standar dan helper pure.

## Problem (Masalah)
Aturan “domain stdlib only” membuat domain tidak bisa memakai utilitas netral dari `internal/shared/*`.
Efek sampingnya:
- semantic error cenderung pindah ke usecase/transport,
- domain jadi miskin expressiveness,
- kontrak domain terasa tidak natural dipakai, walau struktur folder terlihat rapi.

## Options (Opsi)
1) Tetap domain stdlib only  
   - Pro: paling ketat  
   - Kontra: memaksa desain tidak natural dan mendorong logic/error keluar domain

2) Izinkan domain import `internal/shared/*` yang pure  
   - Pro: domain tetap bersih dari IO/vendor tapi bisa pakai utilitas netral  
   - Kontra: perlu disiplin memastikan `internal/shared/*` tetap pure (tanpa IO/vendor)

## Decision (Keputusan)
Memilih opsi (2):
- `domain` boleh import `internal/shared/*` **yang pure** (contoh: error types, clock interface, helpers yang tidak menyentuh IO).
- `domain` tetap dilarang import:
  - `internal/platform/*`
  - `internal/transport/http/*`
  - module lain (`internal/modules/*`)
  - third-party packages

Enforcement dilakukan lewat `scripts/audit_boundaries.sh`.

## Enforcement Guidance

Untuk menjaga `internal/shared/*` tetap pure:
- `internal/shared/*` dilarang import:
  - `internal/platform/*`
  - `internal/transport/http/*`
  - third-party IO/network/db client libraries
- Disarankan membuat allowlist package shared yang boleh dipakai domain (contoh: `internal/shared/errors`, `internal/shared/clock`, `internal/shared/validate`) agar review lebih mudah.

## Consequences (Dampak)
Positif:
- Domain bisa menyimpan semantic error/utility tanpa bocor ke usecase/transport.
- Desain lebih natural dan scalable tanpa mengorbankan boundary anti-IO.

Negatif/Risiko:
- `internal/shared/*` harus dijaga tetap pure; kalau ada IO/vendor nyelip, boundary jadi busuk dari dalam.
- Perlu review berkala pada `internal/shared/*` (terutama saat nambah util baru).
