# AI_TUNING

Mode “budget”: dipakai saat message terbatas, biar AI tetap patuh aturan repo tanpa banyak basa-basi.

Referensi:
- Aturan utama: `docs/internal/ai/AI_RULES.md`
- Template lengkap: `docs/internal/ai/AI_PROMPT.md`

---

## Kapan dipakai
- Kamu cuma mau hasil cepat tapi tetap **zero-assumption**.
- Task kecil-menengah yang tetap butuh blueprint dan DoD.
- Lagi males paste 1 dokumen panjang, tapi tetap mau rapi.

## Prinsip
- AI **wajib minta snapshot** dulu (biar tidak halu).
- AI **wajib bikin blueprint singkat** sebelum ngedit file.
- AI **wajib DoD** (gofmt/test/vet + sanity curl kalau HTTP).

---

## Budget Prompt (copy-paste)

### REPO HEADER
- Module: `example.com/your-api`
- Ikuti `docs/internal/ai/AI_RULES.md` (hard rules + boundaries + DoD).
- Kontrak stabil: router.Register, v1.Register, presenter.HTTPErrorHandler.
- Error response: JSON envelope (no secrets).
- Debug routes gated: `DEBUG_ROUTES=1`.
- 1 folder = 1 package. Target file Go ≤ 100 baris.

### TASK
- [TULIS TASK DI SINI]  
  (jelaskan tujuan + scope + constraint)

### SNAPSHOT WAJIB (tempel output)
- `tree -L 6 internal/transport/http`
- `tree -L 6 internal/modules/[TARGET_MODULE]`
- `cat internal/transport/http/router/v1/router.go`
- `cat internal/transport/http/presenter/error.go`
- `rg -n "^package " internal/transport/http/router`
- [EXTRA FILES] (kalau task nyentuh area spesifik)

### DECISIONS MINIMAL (jawab yang relevan)
- Target client: web browser / API-to-API / keduanya?
- Token delivery: refresh cookie HttpOnly atau body?
- Expiry access & refresh?
- Step-up/trust score perlu?
- Audit events minimal?

### OUTPUT YANG DIHARAPKAN
1) Requirement ringkas (bullet).
2) Blueprint ringkas (flow + file list).
3) Implementasi (minimal perubahan, tidak menjalar).
4) DoD command + expected output ringkas.
