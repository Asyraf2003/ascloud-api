REPO HEADER
- Module: [MODULE_PATH]
- Ikuti docs/AI_RULES.md (hard rules).
- Kontrak stabil: router.Register, v1.Register, presenter.HTTPErrorHandler.
- Error response: JSON envelope via HTTPErrorHandler (no secrets). JSONB hanya storage internal.
- Debug routes gated DEBUG_ROUTES=1.
- 1 folder = 1 package. File <=100 baris.

BUDGET MODE
- Anggap message budget terbatas.
- Minta semua snapshot sekaligus. Jangan ngarang isi file.

TASK
- [TULIS TASK]

SNAPSHOT WAJIB (saya akan paste output)
- tree -L 6 internal/http
- tree -L 6 internal/modules/[TARGET]
- cat internal/http/router/v1/router.go
- cat internal/http/presenter/error.go
- rg -n "^package " internal/http/router
- [EXTRA FILES]
