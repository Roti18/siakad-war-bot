package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Roti18/siakad-war-bot/internal/config"
	"github.com/Roti18/siakad-war-bot/internal/security"
	"github.com/Roti18/siakad-war-bot/internal/ui"
)

func main() {
	// 1. Run setup wizard if .env is not fully configured, and load env variables
	if err := config.SetupPrompt(".env"); err != nil {
		fmt.Printf("Configuration Error: %v\n", err)
		return
	}

	ui.Header("KRS WAR BOT CLIENT - LICENSE DEMO")

	cfg := config.NewConfigManager("configs/config.json")
	if err := cfg.Load(); err != nil {
		_ = cfg.Reset()
	}

	hash, err := security.GetDeviceFingerprint()
	if err != nil {
		ui.LogError("Gagal membaca device fingerprint: " + err.Error())
		ui.Footer()
		return
	}
	ui.BulletPoint("Device Fingerprint", hash)

	// 2. Initialize Secret Store & Auth Client
	store := security.NewSecretStore("credentials.dat")
	auth := security.NewAuthClient(store)
	ctx := context.Background()

	apiServerURL := os.Getenv("API_SERVER_URL")
	if apiServerURL == "" {
		apiServerURL = "http://localhost:8080"
	}
	ui.BulletPoint("API Server URL", apiServerURL)

	// 3. Load or Prompt License Key
	licenseKey, err := auth.GetSavedLicenseKey(ctx)

	if err != nil || strings.TrimSpace(licenseKey) == "" {
		ui.LogWarning("Lisensi tidak ditemukan di penyimpanan lokal.")
		licenseKey = ui.PromptRequired("Masukkan Kunci Lisensi (License Key)")
	} else {
		ui.LogInfo("Lisensi ditemukan di lokal: " + licenseKey)
		if !ui.PromptConfirm("Gunakan lisensi lokal ini?", true) {
			licenseKey = ui.PromptRequired("Masukkan Kunci Lisensi Baru")
		}
	}

	// 4. Verify License with Server
	ui.LogInfo("Memverifikasi lisensi ke server...")
	resp, err := auth.VerifyLicense(ctx, apiServerURL, licenseKey)
	if err != nil {
		ui.LogError("Verifikasi Lisensi Gagal!")
		ui.BulletPoint("ErrorDetail", err.Error())
		// Bersihkan sesi jika gagal
		_ = auth.ClearSession(ctx)
		ui.Footer()
		return
	}

	// 5. Display Verification Success Details
	ui.LogSuccess("VERIFIKASI LISENSI BERHASIL!")
	ui.BulletPoint("License Key", resp.LicenseKey)
	ui.BulletPoint("Status", resp.Status)
	
	if resp.IsLifetime {
		ui.BulletPoint("Tipe Lisensi", "Lifetime (Seumur Hidup)")
	} else if resp.IsTrial {
		ui.BulletPoint("Tipe Lisensi", "Uji Coba (Trial)")
		ui.BulletPoint("Kedaluwarsa", resp.ExpiresAt.Format("2006-01-02 15:04:05"))
	} else {
		ui.BulletPoint("Tipe Lisensi", "Standard")
		ui.BulletPoint("Kedaluwarsa", resp.ExpiresAt.Format("2006-01-02 15:04:05"))
	}

	ui.Footer()
}
