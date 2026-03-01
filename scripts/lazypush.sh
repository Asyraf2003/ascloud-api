count=$(($(git rev-list --count HEAD 2>/dev/null || echo 0) + 1))
git add .
if git commit -m "Update $count"; then
    echo "Berhasil commit: Update $count"
    git push -u origin main
else
    echo "Gagal: Tidak ada perubahan untuk di-commit atau terjadi error."
fi
