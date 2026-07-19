# PowerShell Build Script for KRS War Bot Client
# Compiles a standalone portable executable for distribution to clients.

Clear-Host
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host "       KRS WAR BOT - COMPILATION WIZARD            " -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host "[*] Membersihkan cache compiler..." -ForegroundColor Yellow

# Clean cache
go clean -cache

Write-Host "[*] Mengompilasi program client ke portable binary (.exe)..." -ForegroundColor Yellow

# Build release binary optimized for size (-s -w strips debug symbols and DWARF)
go build -ldflags="-s -w" -o krs-war-bot.exe cmd/bot/main.go

if ($LASTEXITCODE -eq 0) {
    $fileSize = (Get-Item krs-war-bot.exe).Length / 1MB
    Write-Host ""
    Write-Host "==================================================" -ForegroundColor Green
    Write-Host "🎉 COMPILATION SUCCESSFUL!" -ForegroundColor Green
    Write-Host "==================================================" -ForegroundColor Green
    Write-Host "📂 Output: krs-war-bot.exe" -ForegroundColor Green
    Write-Host "📊 Ukuran File: $([math]::Round($fileSize, 2)) MB" -ForegroundColor Green
    Write-Host "==================================================" -ForegroundColor Green
    Write-Host "[*] Distribusikan file 'krs-war-bot.exe' langsung ke pembeli Anda." -ForegroundColor Gray
    Write-Host "[*] Pembeli TIDAK PERLU menginstal Golang, Python, Selenium, dll." -ForegroundColor Gray
    Write-Host "[*] Saat dijalankan pertama kali oleh pembeli, file ini akan:" -ForegroundColor Gray
    Write-Host "    1. Menampilkan setup konfigurasi NIM/Password (.env)" -ForegroundColor Gray
    Write-Host "    2. Meminta/memverifikasi kunci lisensi" -ForegroundColor Gray
    Write-Host "    3. Mengunduh dependensi browser Chromium otomatis ke AppData" -ForegroundColor Gray
} else {
    Write-Host "[-] COMPILATION FAILED! Periksa error di atas." -ForegroundColor Red
}
