# DESCRIPTION

Kenapa template AI ini efektif:
- Memaksa snapshot → AI tidak ngarang path/isi file.
- Memaksa keputusan produk → AI tidak “milih sendiri” hal sensitif (token, expiry, client type).
- Memaksa blueprint sebelum eksekusi → minim efek berantai.
- Memaksa DoD → hasil bisa diaudit (gofmt/test/vet/sanity).

Hal yang wajib kamu isi tiap task:
- `[MODULE_PATH]` (sesuai `go.mod`)
- `[TARGET_MODULE]`
- `[EXTRA_SNAPSHOT_FILES]` kalau nyentuh area spesifik (platform/queue/db/middleware)
