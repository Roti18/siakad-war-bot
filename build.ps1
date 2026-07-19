# PowerShell Build Script for KRS War Bot Client
Clear-Host
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host "      KRS WAR BOT - BUILD SYSTEM             " -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan

Write-Host "[*] Cleaning build cache..." -ForegroundColor Yellow
go clean -cache

Write-Host "[*] Compiling client executable (krs-war-bot.exe)..." -ForegroundColor Yellow
go build -ldflags="-s -w" -o krs-war-bot.exe cmd/bot/main.go

if ($LASTEXITCODE -eq 0) {
    $fileSize = (Get-Item krs-war-bot.exe).Length / 1MB
    Write-Host ""
    Write-Host "[+] BUILD SUCCESSFUL!" -ForegroundColor Green
    Write-Host "[+] Output: krs-war-bot.exe" -ForegroundColor Green
    Write-Host "[+] File Size: $([math]::Round($fileSize, 2)) MB" -ForegroundColor Green
    Write-Host "=============================================" -ForegroundColor Cyan
} else {
    Write-Host "[-] BUILD FAILED! Please check the compiler errors above." -ForegroundColor Red
}
