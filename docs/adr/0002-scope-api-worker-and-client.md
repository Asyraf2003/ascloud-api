# ADR 0002: Scope API + Worker, web sebagai client

Tanggal: 29 December 2025 9:42
Status: Accepted

## Context (Keadaan awal)
- Repo berisi cmd/api, cmd/worker, cmd/migrate
- Target produk: layanan hosting statis dulu
- Web/dashboard hanya client yang consume API

## Problem (Masalah)
- Istilah “microservice vs monolith” bikin rancu arah
- Doc sebelumnya tidak konsisten dan sulit audit

## Options (Opsi yang dipertimbangkan)
1) Banyak microservice sejak awal
2) Modular monolith + 2 binary (api & worker), siap diextract nanti

## Decision (Keputusan)
- Memilih opsi (2):
  - 1 repo, boundary module ketat
  - deployable: API + Worker terpisah
  - web/dashboard berada di repo lain, consume API

## Consequences (Dampak)
Positif:
- Debugging & audit jelas
- Delivery cepat untuk fase statis
- Fondasi siap scale fitur

Negatif/Risiko:
- Perlu disiplin boundary (audit rules wajib)
- Perlu rencana event/job untuk hosting provisioning
