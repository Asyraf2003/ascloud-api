#!/bin/bash

# 1. Hitung jumlah commit saat ini + 1 untuk nomor urut
count=$(($(git rev-list --count HEAD 2>/dev/null || echo 0) + 1))

# 2. Tambahkan semua perubahan
git add .

# 3. Eksekusi commit dengan format "Update [nomor]"
# Misal: Update 1, Update 2, dst.
if git commit -m "Update $count"; then
    echo "Berhasil commit: Update $count"
    
    # 4. Push ke remote origin branch main
    git push origin main
else
    echo "Gagal: Tidak ada perubahan untuk di-commit atau terjadi error."
fi