package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Roti18/siakad-war-bot/internal/domain"
	"github.com/Roti18/siakad-war-bot/internal/ui"
)

// RunMainMenu shows a professional interactive settings panel for non-IT users
func (c *Config) RunMainMenu() {
	for {
		ui.Header("KRS WAR BOT CLIENT - MENU UTAMA")
		
		ui.LogInfo("Pilih opsi menu di bawah ini:")
		ui.BulletPoint("[1]", "Mulai KRS War (Start)")
		ui.BulletPoint("[2]", "Konfigurasi Browser (Headless, Block Assets)")
		ui.BulletPoint("[3]", "Atur Target Mata Kuliah (Courses)")
		ui.BulletPoint("[4]", "Konfigurasi Jadwal & Percobaan (Schedule)")
		ui.BulletPoint("[5]", "Konfigurasi Akun & Server (.env)")
		ui.BulletPoint("[6]", "Lihat Pengaturan Saat Ini")
		ui.BulletPoint("[7]", "Keluar (Exit)")
		
		ui.Footer()
		
		choice := ui.Prompt("Pilih menu (1-7)", "1")
		
		switch choice {
		case "1":
			ui.Header("MULAI KRS WAR")
			if len(c.Courses) == 0 {
				ui.LogError("Target mata kuliah masih kosong!")
				ui.LogWarning("Silakan atur mata kuliah target Anda terlebih dahulu di menu [3].")
				ui.Footer()
				break
			}
			ui.LogSuccess("Mata kuliah target terkonfigurasi. Memulai engine KRS War...")
			ui.LogInfo("Mencari kelas untuk " + strconv.Itoa(len(c.Courses)) + " mata kuliah...")
			// Di sini engine utama (browser automation Rod) akan berjalan
			ui.LogInfo("Bot standby... Menunggu jadwal KRS aktif.")
			ui.Footer()
			return // Keluar ke main untuk menjalankan loop utama bot jika sudah diimplementasikan
		case "2":
			c.ConfigureBrowser()
		case "3":
			c.ConfigureCourses()
		case "4":
			c.ConfigureSchedule()
		case "5":
			c.ConfigureAccount(".env")
		case "6":
			c.ShowCurrentSettings()
		case "7":
			ui.Header("KELUAR APLIKASI")
			ui.LogInfo("Terima kasih telah menggunakan KRS War Bot!")
			ui.Footer()
			return
		default:
			ui.Header("INPUT SALAH")
			ui.LogError("Pilihan tidak valid. Silakan pilih nomor 1 sampai 7.")
			ui.Footer()
		}
	}
}

func (c *Config) ConfigureBrowser() {
	ui.Header("KONFIGURASI BROWSER")

	c.mu.Lock()
	// 1. Headless
	c.Browser.Headless = ui.PromptConfirm("Jalankan Browser tanpa tampilan (Headless)?", c.Browser.Headless)

	// 2. Width
	widthStr := ui.Prompt("Lebar Browser (Width)", strconv.Itoa(c.Browser.Width))
	if w, err := strconv.Atoi(widthStr); err == nil {
		c.Browser.Width = w
	}

	// 3. Height
	heightStr := ui.Prompt("Tinggi Browser (Height)", strconv.Itoa(c.Browser.Height))
	if h, err := strconv.Atoi(heightStr); err == nil {
		c.Browser.Height = h
	}

	// 4. Block Images
	c.Browser.BlockImages = ui.PromptConfirm("Blokir Gambar (Menghemat Kuota & RAM)?", c.Browser.BlockImages)

	// 5. Block CSS
	c.Browser.BlockCSS = ui.PromptConfirm("Blokir CSS Styles (Sangat direkomendasikan agar cepat)?", c.Browser.BlockCSS)

	// 6. Block Fonts
	c.Browser.BlockFonts = ui.PromptConfirm("Blokir Fonts?", c.Browser.BlockFonts)

	// 7. Block Media
	c.Browser.BlockMedia = ui.PromptConfirm("Blokir Video/Audio?", c.Browser.BlockMedia)
	c.mu.Unlock()

	if err := c.Save(); err != nil {
		ui.LogError("Gagal menyimpan konfigurasi browser: " + err.Error())
	} else {
		ui.LogSuccess("Konfigurasi browser berhasil disimpan ke config.json!")
	}
	ui.Footer()
}

func (c *Config) ConfigureCourses() {
	for {
		ui.Header("ATUR TARGET MATA KULIAH")
		
		c.mu.RLock()
		if len(c.Courses) == 0 {
			ui.LogWarning("Belum ada mata kuliah target yang terdaftar.")
		} else {
			ui.LogInfo("Daftar mata kuliah saat ini:")
			for i, course := range c.Courses {
				ui.BulletPoint("["+strconv.Itoa(i+1)+"]", course.Nama+" (Kelas: "+course.Kelas+")")
			}
		}
		c.mu.RUnlock()

		fmt.Println("")
		ui.LogInfo("Pilih tindakan:")
		ui.BulletPoint("[1]", "Tambah Mata Kuliah")
		ui.BulletPoint("[2]", "Hapus Mata Kuliah")
		ui.BulletPoint("[3]", "Kembali ke Menu Utama")
		ui.Footer()

		action := ui.Prompt("Pilih tindakan (1-3)", "1")

		switch action {
		case "1":
			ui.Header("TAMBAH MATA KULIAH BARU")
			nama := ui.PromptRequired("Nama Mata Kuliah")
			kelas := ui.PromptRequired("Kelas (contoh: IF 4A)")

			c.mu.Lock()
			c.Courses = append(c.Courses, domain.TargetCourse{
				Nama:  nama,
				Kelas: kelas,
			})
			c.mu.Unlock()

			if err := c.Save(); err != nil {
				ui.LogError("Gagal menyimpan mata kuliah: " + err.Error())
			} else {
				ui.LogSuccess("Mata kuliah '" + nama + "' (" + kelas + ") berhasil ditambahkan!")
			}
			ui.Footer()
		case "2":
			ui.Header("HAPUS MATA KULIAH")
			c.mu.RLock()
			totalCourses := len(c.Courses)
			c.mu.RUnlock()

			if totalCourses == 0 {
				ui.LogError("Tidak ada mata kuliah untuk dihapus.")
				ui.Footer()
				break
			}

			numStr := ui.PromptRequired("Masukkan nomor urut mata kuliah yang ingin dihapus (1-" + strconv.Itoa(totalCourses) + ")")
			idx, err := strconv.Atoi(numStr)
			if err != nil || idx < 1 || idx > totalCourses {
				ui.LogError("Nomor urut tidak valid!")
				ui.Footer()
				break
			}

			c.mu.Lock()
			removedName := c.Courses[idx-1].Nama
			c.Courses = append(c.Courses[:idx-1], c.Courses[idx:]...)
			c.mu.Unlock()

			if err := c.Save(); err != nil {
				ui.LogError("Gagal menghapus mata kuliah: " + err.Error())
			} else {
				ui.LogSuccess("Mata kuliah '" + removedName + "' berhasil dihapus!")
			}
			ui.Footer()
		case "3":
			return
		default:
			ui.LogError("Pilihan tidak valid.")
			ui.Footer()
		}
	}
}

func (c *Config) ConfigureSchedule() {
	ui.Header("KONFIGURASI JADWAL")

	c.mu.Lock()
	// 1. Time
	c.Schedule.Time = ui.Prompt("Waktu KRS Aktif (HH:MM:SS)", c.Schedule.Time)

	// 2. Date
	c.Schedule.Date = ui.Prompt("Tanggal KRS Aktif (YYYY-MM-DD, kosongkan jika hari ini)", c.Schedule.Date)

	// 3. Retry Delay
	delayStr := ui.Prompt("Delay Percobaan Ulang (dalam detik, misal 0.5)", strconv.FormatFloat(c.Schedule.RetryDelaySec, 'f', -1, 64))
	if d, err := strconv.ParseFloat(delayStr, 64); err == nil {
		c.Schedule.RetryDelaySec = d
	}

	// 4. Max Retry
	maxRetryStr := ui.Prompt("Maksimal Percobaan Ulang (Max Retry)", strconv.Itoa(c.Schedule.MaxRetry))
	if m, err := strconv.Atoi(maxRetryStr); err == nil {
		c.Schedule.MaxRetry = m
	}
	c.mu.Unlock()

	if err := c.Save(); err != nil {
		ui.LogError("Gagal menyimpan konfigurasi jadwal: " + err.Error())
	} else {
		ui.LogSuccess("Konfigurasi jadwal berhasil disimpan ke config.json!")
	}
	ui.Footer()
}

func (c *Config) ShowCurrentSettings() {
	ui.Header("DETAIL PENGATURAN SAAT INI")

	c.mu.RLock()
	ui.LogInfo("=== BROWSER ===")
	ui.BulletPoint("Headless Mode", strconv.FormatBool(c.Browser.Headless))
	ui.BulletPoint("Dimensi Layar", strconv.Itoa(c.Browser.Width)+"x"+strconv.Itoa(c.Browser.Height))
	ui.BulletPoint("Blokir Gambar", strconv.FormatBool(c.Browser.BlockImages))
	ui.BulletPoint("Blokir CSS", strconv.FormatBool(c.Browser.BlockCSS))
	
	ui.LogInfo("=== JADWAL & SISTEM ===")
	ui.BulletPoint("Waktu KRS", c.Schedule.Time)
	ui.BulletPoint("Tanggal KRS", c.Schedule.Date)
	ui.BulletPoint("Delay Retry", strconv.FormatFloat(c.Schedule.RetryDelaySec, 'f', -1, 64)+" detik")
	ui.BulletPoint("Max Retry", strconv.Itoa(c.Schedule.MaxRetry))
	
	ui.LogInfo("=== TARGET KURSUS ===")
	if len(c.Courses) == 0 {
		ui.BulletPoint("Courses", "Belum ada target mata kuliah")
	} else {
		for i, crs := range c.Courses {
			ui.BulletPoint("Target #"+strconv.Itoa(i+1), crs.Nama+" ("+crs.Kelas+")")
		}
	}
	c.mu.RUnlock()

	ui.Footer()
	ui.Prompt("Tekan Enter untuk kembali ke Menu Utama", "")
}

func (c *Config) ConfigureAccount(envFile string) {
	ui.Header("KONFIGURASI AKUN & SERVER")

	env, _ := LoadEnv(envFile)
	currentAPIServer := "http://localhost:8080"
	currentBaseURL := "https://siakad.example.ac.id/"
	currentNIM := ""
	currentPassword := ""

	if env != nil {
		if val, ok := env["API_SERVER_URL"]; ok && val != "" {
			currentAPIServer = val
		}
		if val, ok := env["BASE_URL"]; ok && val != "" {
			currentBaseURL = val
		}
		if val, ok := env["NIM"]; ok && val != "" {
			currentNIM = val
		}
		if val, ok := env["PASSWORD"]; ok && val != "" {
			currentPassword = val
		}
	}

	apiServerURL := ui.Prompt("Masukkan IP/URL Server API", currentAPIServer)
	baseURL := ui.Prompt("Masukkan BASE_URL SIAKAD", currentBaseURL)
	nim := ui.Prompt("Masukkan NIM Anda", currentNIM)
	password := ui.Prompt("Masukkan PASSWORD Anda", currentPassword)

	err := SaveEnv(envFile, apiServerURL, baseURL, nim, password)
	if err != nil {
		ui.LogError("Gagal menyimpan konfigurasi akun: " + err.Error())
	} else {
		// Update current OS env variables
		os.Setenv("API_SERVER_URL", apiServerURL)
		os.Setenv("BASE_URL", baseURL)
		os.Setenv("NIM", nim)
		os.Setenv("PASSWORD", password)
		ui.LogSuccess("Konfigurasi akun berhasil disimpan ke " + envFile + "!")
	}
	ui.Footer()
}
