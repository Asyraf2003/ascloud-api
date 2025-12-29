# AI_TUNING

Dokumen ini untuk mode hemat message/token (kalau chat limit atau mau cepat).

## Prinsip
- Snapshot tetap wajib.
- Hindari diskusi panjang.
- Fokus: requirement → blueprint → eksekusi.

## Budget Workflow
1) Minta semua snapshot sekaligus (sekali tembak).
2) Ambil keputusan minimal yang wajib.
3) Tulis blueprint ringkas (file list + flow).
4) Eksekusi.

## Snapshot minimal
- `tree -L 6 internal/transport/http`
- `tree -L 6 internal/modules/[TARGET_MODULE]`
- `cat internal/transport/http/router/v1/router.go`
- `cat internal/transport/http/presenter/error.go`
- `rg -n "^package " internal/transport/http/router`
- [EXTRA_SNAPSHOT_FILES]
