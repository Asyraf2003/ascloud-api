# ADR (Architecture Decision Record)

ADR dipakai untuk mencatat keputusan penting agar audit & reasoning jelas.
Setiap perubahan besar yang mempengaruhi arsitektur/kontrak/perilaku sistem wajib punya ADR.

## Kapan bikin ADR
- Perubahan auth model, token, session
- Perubahan penyimpanan (Postgres/Dynamo), schema besar
- Perubahan boundary module / data ownership
- Perubahan strategi hosting (statis/dinamis), provisioning, queue
- Perubahan kebijakan security (CORS/CSRF/trust/rate limit)

## Format wajib
- Context: keadaan awal
- Problem: masalah/risiko
- Options: opsi yang dipertimbangkan (+ tradeoff singkat)
- Decision: keputusan final
- Consequences: dampak + risiko lanjutan

## Template
Gunakan: `docs/adr/TEMPLATE.md`
