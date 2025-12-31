# Note: Auth OIDC state store masih memory (single-instance only)

Tanggal: 2025-12-31  
Status: Active

## Context
- Flow Google OIDC menyimpan `state -> {nonce, code_verifier, purpose, created_at}` di `internal/platform/state/memory/AuthStateStore`.
- State dipakai sekali (GetDel) dan punya TTL (`AUTH_STATE_TTL_MIN`).

## Problem
- Skala >1 instance / restart server membuat state tidak konsisten:
  - Callback bisa masuk ke instance lain yang tidak punya state.
  - Restart sebelum callback membuat state hilang.
  - Hasilnya: user kena `state invalid`, login terlihat “random gagal”.
- Penyimpanan memory tanpa janitor berpotensi menumpuk:
  - Item expired akan tetap ada jika tidak pernah di-GetDel (TTL hanya dicek saat read).
  - Risk ini kecil karena ada rate limit, tapi tetap ada secara teori.

## Notes / Findings
- Untuk fase dev/single instance: aman dan cepat.
- Untuk fase production multi-instance (atau HA) dan rolling deploy: akan jadi sumber error intermiten.
- Store ideal untuk state adalah storage yang:
  - mendukung TTL native,
  - konsisten lintas instance,
  - cepat (low latency),
  - dapat “Get+Delete” (consume-once).

## Options (kalau ada)
1) Tetap memory + janitor + hard cap
   - (+) tanpa dependency eksternal
   - (-) tetap tidak bisa multi-instance; hanya mengurangi memory growth
2) Redis/KeyDB (TTL native)
   - (+) cepat, TTL native, cocok untuk ephemeral auth state
   - (-) tambah dependency runtime
3) DynamoDB dengan TTL attribute
   - (+) managed, TTL native, cocok untuk cloud-native
   - (-) latency lebih tinggi dari Redis; butuh desain key yang rapih
4) Postgres table untuk state + cleanup job
   - (+) tidak tambah infra baru
   - (-) kurang ideal untuk ephemeral state, cleanup jadi beban

## Next Actions
- [ ] Treat memory store sebagai “single-instance mode” secara eksplisit (docs/core atau docs/notes).
- [ ] Saat masuk fase multi-instance: migrasi state store ke Redis atau DynamoDB TTL.
- [ ] (Opsional) Tambah janitor sederhana untuk purge expired + limit size untuk safety.
