package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Roti18/siakad-war-bot/internal/config"
	"github.com/Roti18/siakad-war-bot/internal/security"
)

func main() {
	// 1. Run setup wizard if .env is not fully configured, and load env variables
	if err := config.SetupPrompt(".env"); err != nil {
		fmt.Printf("Configuration Error: %v\n", err)
		return
	}

	fmt.Println("==================================================")
	fmt.Println("       KRS WAR BOT CLIENT - LICENSE DEMO")
	fmt.Println("==================================================")

	cfg := config.NewConfigManager("configs/config.json")
	if err := cfg.Load(); err != nil {
		_ = cfg.Reset()
	}

	hash, err := security.GetDeviceFingerprint()
	if err != nil {
		fmt.Printf("Error fingerprint: %v\n", err)
		return
	}
	fmt.Printf("[✔] Device Fingerprint: %s\n", hash)

	// 2. Initialize Secret Store & Auth Client
	store := security.NewSecretStore("credentials.dat")
	auth := security.NewAuthClient(store)
	ctx := context.Background()

	apiServerURL := os.Getenv("API_SERVER_URL")
	if apiServerURL == "" {
		apiServerURL = "http://localhost:8080"
	}
	fmt.Printf("[i] Menghubungkan ke API Server: %s\n", apiServerURL)

	// 3. Load or Prompt License Key
	licenseKey, err := auth.GetSavedLicenseKey(ctx)
	reader := bufio.NewReader(os.Stdin)

	if err != nil || strings.TrimSpace(licenseKey) == "" {
		fmt.Println("\n[!] Lisensi tidak ditemukan di penyimpanan lokal.")
		fmt.Print("Masukkan Kunci Lisensi (License Key): ")
		input, _ := reader.ReadString('\n')
		licenseKey = strings.TrimSpace(input)

		if licenseKey == "" {
			fmt.Println("Error: Kunci lisensi tidak boleh kosong!")
			return
		}
	} else {
		fmt.Printf("[✔] Lisensi ditemukan di lokal: %s\n", licenseKey)
		fmt.Print("Gunakan lisensi lokal ini? (Y/n atau tekan Enter untuk ya): ")
		useLocal, _ := reader.ReadString('\n')
		useLocal = strings.TrimSpace(strings.ToLower(useLocal))
		if useLocal != "" && useLocal != "y" && useLocal != "yes" {
			fmt.Print("Masukkan Kunci Lisensi Baru: ")
			input, _ := reader.ReadString('\n')
			licenseKey = strings.TrimSpace(input)
			if licenseKey == "" {
				fmt.Println("Error: Kunci lisensi tidak boleh kosong!")
				return
			}
		}
	}

	// 4. Verify License with Server
	fmt.Println("[...] Memverifikasi lisensi ke server...")
	resp, err := auth.VerifyLicense(ctx, apiServerURL, licenseKey)
	if err != nil {
		fmt.Printf("\n[❌] Verifikasi Lisensi Gagal!\nError: %v\n", err)
		// Bersihkan sesi jika gagal
		_ = auth.ClearSession(ctx)
		return
	}

	// 5. Display Verification Success Details
	fmt.Println("\n[✔] VERIFIKASI LISENSI BERHASIL!")
	fmt.Printf("- Key: %s\n", resp.LicenseKey)
	fmt.Printf("- Status: %s\n", resp.Status)
	if resp.IsLifetime {
		fmt.Println("- Tipe: Lifetime (Seumur Hidup)")
	} else if resp.IsTrial {
		fmt.Println("- Tipe: Uji Coba (Trial)")
		fmt.Printf("- Kedaluwarsa: %s\n", resp.ExpiresAt.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Println("- Tipe: Standar")
		fmt.Printf("- Kedaluwarsa: %s\n", resp.ExpiresAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Println("==================================================")
}
