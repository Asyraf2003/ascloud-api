Penjelasan “kenapa template ini kuat” dan titik yang harus disesuaikan
Kenapa ini kuat

Memaksa snapshot: GPT nggak bisa halu path/file yang udah berubah.

Ada boundary rules: ini yang mencegah “perbaikan berantai”.

Ada DoD: GPT harus kasih command dan expected output, jadi kamu bisa audit cepat.

Ada keputusan eksplisit: token delivery (cookie vs body) itu beda dunia security-nya.

Bagian yang harus kamu sesuaikan

[MODULE_PATH]
Wajib sesuai go.mod.

[TARGET_MODULE]
Pilih modul yang terkait: auth, account, hosting, dll.

[EXTRA SNAPSHOT FILES]
Kalau task nyentuh area spesifik, sebut file yang harus dibaca dulu. Contoh:

auth: internal/platform/google/*, internal/security/token/*, internal/modules/auth/*

billing: internal/modules/billing/* (kalau ada)

Pertanyaan keputusan
Kalau task beda, pertanyaan beda. Tapi selalu ada “decision list” supaya GPT nggak asal milih.