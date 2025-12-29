# ADR 0003: Domain boleh import internal/shared (pure)

Tanggal: 2025-12-29 10:xx WITA
Status: Accepted

## Context (Keadaan awal)
- Repo menargetkan hexagonal/modular boundaries yang ketat.
- Audit boundary sebelumnya memaksa `domain` hanya boleh standard library.
- Sementara itu, dokumen error handling mengarahkan penggunaan `internal/shared/apperr` sebagai error type standar.

## Problem (Masalah)
- Aturan “domain stdlib only” membuat domain tidak bisa memakai tipe/error/utilitas yang netral (`internal/shared/*`).
- Akibatnya:
  - error semantics cenderung pindah ke usecase/transport,
  - domain jadi miskin expressiveness,
  - boundary terlihat rapi di folder tapi tidak natural dipakai.

## Options (Opsi)
1) Tetap domain stdlib only
   - Pro: sangat ketat
   - Kontra: memaksa pola desain tidak natural dan mendorong logic/error keluar domain

2) Izinkan domain import `internal/shared/*` yang pure
   - Pro: domain tetap bersih dari IO/vendor tapi bisa pakai utilitas netral
   - Kontra: perlu disiplin memastikan `internal/shared` tetap pure (tanpa IO)

## Decision (Keputusan)
- Memilih opsi (2):
  - `domain` boleh import `internal/shared/*` (pure utilities: error types, clock, redact helpers yang netral, dll).
  - `domain` tetap dilarang import:
    - `internal/platform/*`
    - `internal/transport/http/*`
    - module lain (`internal/modules/*`)
    - third-party packages
- Enforcement dilakukan lewat `scripts/audit_boundaries.sh`.

## Consequences (Dampak)
Positif:
- Domain bisa menyimpan semantic error/utility tanpa bocor ke usecase/transport.
- Desain lebih natural dan scalable.

Negatif/Risiko:
- `internal/shared` harus dijaga tetap pure; kalau ada IO/vendor nyelip, boundary jadi rusak.
- Perlu review berkala pada `internal/shared/*`.
