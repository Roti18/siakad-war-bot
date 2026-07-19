package browser

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/Roti18/siakad-war-bot/internal/domain"
	"github.com/Roti18/siakad-war-bot/internal/ui"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// getActivePage returns the frame page context if it exists, otherwise the main page
func getActivePage(page *rod.Page, name string) *rod.Page {
	el, err := page.Element(fmt.Sprintf("frame[name='%s'], iframe[name='%s']", name, name))
	if err == nil && el != nil {
		framePage, err := el.Frame()
		if err == nil && framePage != nil {
			return framePage
		}
	}
	return page
}

// loginLogic automates the logging in process and waits for the dashboard to appear
func loginLogic(ctx context.Context, page *rod.Page, baseURL, nim, password string, timeoutSec int) bool {
	// A. Cek apakah saat ini kita SUDAH berada di dashboard (misal jika browser sudah terbuka dan session aktif)
	hasKrs, _ := page.Timeout(500 * time.Millisecond).ElementX("//a[contains(., 'Kartu Rencana') or contains(., 'Logout') or contains(., 'Keluar')]")
	hasFrame, _ := page.Timeout(500 * time.Millisecond).ElementsX("//frame[@name='menu'] | //iframe[@name='menu']")
	if (hasKrs != nil) || (len(hasFrame) > 0) {
		ui.LogSuccess("Sesi aktif terdeteksi! Langsung menuju Dashboard.")
		return true
	}

	// B. Daftarkan handler dialog untuk mendeteksi alert (misal password salah) agar tidak nge-hang
	go page.EachEvent(func(e *proto.PageJavascriptDialogOpening) {
		ui.LogError("Alert SIAKAD: " + e.Message)
		// Setujui dialog untuk menutup popup alert agar halaman tidak terkunci
		_ = proto.PageHandleJavaScriptDialog{
			Accept:     true,
			PromptText: "",
		}.Call(page)
	})

	// C. Jika belum di dashboard, baru navigasi ke index.php
	url := baseURL + "index.php"
	ui.LogInfo("Menghubungi website: " + url + " ...")
	if err := page.Context(ctx).Navigate(url); err != nil {
		ui.LogError("Gagal memuat halaman login: " + err.Error())
		return false
	}

	// D. Cek apakah ada form login (#username) di halaman yang baru dimuat
	usernameEl, err := page.Timeout(5 * time.Second).Element("#username")
	if err != nil {
		// Jika form login tidak ada, mungkin langsung diarahkan ke dashboard
		links, err2 := page.Timeout(2 * time.Second).ElementsX("//a[contains(., 'Kartu Rencana') or contains(., 'Logout') or contains(., 'Keluar')]")
		if err2 == nil && len(links) > 0 {
			ui.LogSuccess("Sesi aktif terdeteksi (Auto-Redirect)! Menuju Dashboard.")
			return true
		}
		ui.LogError("Halaman login dimuat, tetapi form input 'username' tidak ditemukan!")
		return false
	}

	// E. Jika form login ditemukan, lakukan input credentials
	ui.LogInfo("Halaman login berhasil dimuat. Menyiapkan autentikasi...")
	usernameEl.MustInput(nim)

	passwordEl, err := page.Timeout(5 * time.Second).Element("#password")
	if err != nil {
		ui.LogError("Input field 'password' tidak ditemukan!")
		return false
	}
	passwordEl.MustInput(password)

	// F. Klik Tombol Login
	loginBtn, err := page.Timeout(5 * time.Second).ElementX("//input[@type='button' and @value='Login']")
	if err != nil {
		// Fallback ke input submit standar
		loginBtn, err = page.Timeout(5 * time.Second).Element("input[type='submit']")
	}
	if err == nil && loginBtn != nil {
		loginBtn.MustClick()
	} else {
		ui.LogError("Tombol login tidak ditemukan di halaman!")
		return false
	}

	ui.LogInfo("Mengirim data masuk... Menunggu halaman dashboard...")

	// G. Deteksi Dashboard Berhasil
	start := time.Now()
	for time.Since(start) < time.Duration(timeoutSec)*time.Second {
		info, _ := page.Info()
		if info != nil && strings.Contains(strings.ToLower(info.URL), "utama.php") {
			ui.LogSuccess("Login Berhasil! (Dashboard utama.php)")
			return true
		}

		// A. Cek keberadaan frame menu (Non-blocking dengan 500ms timeout)
		el, err := page.Timeout(500 * time.Millisecond).ElementsX("//frame[@name='menu'] | //iframe[@name='menu']")
		if err == nil && len(el) > 0 {
			ui.LogSuccess("Login Berhasil! (Dashboard Frame)")
			return true
		}

		// B. Cek link KRS langsung (Portal Baru) (Non-blocking dengan 500ms timeout)
		links, err := page.Timeout(500 * time.Millisecond).ElementsX("//a[contains(., 'Kartu Rencana') or contains(., 'Logout') or contains(., 'Keluar')]")
		if err == nil && len(links) > 0 {
			ui.LogSuccess("Login Berhasil! (Portal Tanpa Frame)")
			return true
		}

		// C. Cek Fallback: Form login (#username) sudah hilang, dan URL bukan halaman login awal
		_, err = page.Timeout(200 * time.Millisecond).Element("#username")
		if err != nil {
			if info != nil && !strings.Contains(strings.ToLower(info.URL), "index.php") && !strings.Contains(strings.ToLower(info.URL), "login") {
				ui.LogSuccess("Login Berhasil! (Form login hilang & Halaman berpindah)")
				return true
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	ui.LogError("Gagal mendeteksi halaman dashboard (Timeout). Periksa kembali NIM/Password Anda.")
	return false
}

// waitUntilWar halts execution with keep-alive page reloads until target time is reached
func waitUntilWar(ctx context.Context, page *rod.Page, targetTimeStr string, refreshInterval int, baseURL, nim, password string, timeoutSec int) {
	if targetTimeStr == "" {
		return
	}
	
	ui.LogInfo("STANDBY: Menunggu jadwal KRS aktif pada jam " + targetTimeStr + "...")
	lastRefresh := time.Now()

	for {
		nowStr := time.Now().Format("15:04:05")
		if nowStr >= targetTimeStr {
			ui.LogSuccess("WAKTU WAR TIBA (" + nowStr + ")! MEMULAI EKSEKUSI...")
			break
		}

		// Keep-Alive Refresh agar sesi login tidak expired
		if time.Since(lastRefresh) > time.Duration(refreshInterval)*time.Second {
			ui.LogInfo("Menyegarkan sesi browser (Keep-Alive)...")
			_ = page.Context(ctx).Reload()
			lastRefresh = time.Now()
			time.Sleep(2 * time.Second)

			// Deteksi jika ter-logout otomatis saat refresh
			usernameField, err := page.Element("#username")
			if err == nil && usernameField != nil {
				ui.LogWarning("Sesi Anda terputus! Mencoba login ulang otomatis...")
				loginLogic(ctx, page, baseURL, nim, password, timeoutSec)
			}
		}

		nextRefresh := int(time.Duration(refreshInterval)*time.Second-time.Since(lastRefresh)) / int(time.Second)
		fmt.Printf("\r[🕒] Jam: %s | Target: %s | Refresh Sesi: %ds ", nowStr, targetTimeStr, nextRefresh)
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Println()
}

// StartWarEngine runs the main automated KRS ticking loop
func StartWarEngine(ctx context.Context, 
	driver domain.BrowserDriver, 
	courses []domain.TargetCourse, 
	schTime string,
	schRefreshIntervalSec int,
	schRetryDelaySec float64,
	schMaxRetry int,
	nim, password, baseURL string,
	ssSaveDir string,
	ssService domain.ScreenshotService) {
	
	rodDriver, ok := driver.(*RodDriver)
	if !ok || rodDriver.page == nil {
		ui.LogError("Driver browser tidak valid atau belum diinisialisasi!")
		return
	}

	page := rodDriver.page
	sabarWait := 600 // default wait timeout (10 Menit)

	// 1. Jalankan Logika Login
	if !loginLogic(ctx, page, baseURL, nim, password, sabarWait) {
		return
	}

	// 2. Jalankan Timer Tunggu War KRS Aktif
	waitUntilWar(ctx, page, schTime, schRefreshIntervalSec, baseURL, nim, password, sabarWait)

	// 3. Navigasi Menuju Halaman KRS
	xpathTambah := "//*[(self::a or self::input) and (contains(@value, 'Tambah') or contains(., 'Tambah'))]"
	
	// Pengecekan awal: Apakah kita sudah berada di halaman Tambah KRS atau halaman pemilihan kelas?
	mainPageCheck := getActivePage(page, "main")
	hasTambah, _ := mainPageCheck.Timeout(1 * time.Second).ElementX(xpathTambah)
	hasTable, _ := mainPageCheck.Timeout(1 * time.Second).Element(".table-common")

	if hasTambah != nil || hasTable != nil {
		ui.LogSuccess("Sudah berada di halaman war KRS! Melewati navigasi menu.")
	} else {
		for {
			// Cek logout otomatis (Non-blocking check)
			usernameField, err := page.Timeout(500 * time.Millisecond).Element("#username")
			if err == nil && usernameField != nil {
				ui.LogWarning("Sesi putus saat navigasi! Mengulang login...")
				loginLogic(ctx, page, baseURL, nim, password, sabarWait)
			}

			menuPage := getActivePage(page, "menu")
			krsLink, err := menuPage.Timeout(10 * time.Second).ElementX("//a[contains(., 'Kartu Rencana Studi')]")
			if err == nil && krsLink != nil {
				ui.LogInfo("Mengakses menu Kartu Rencana Studi...")
				krsLink.MustClick()
				break
			}

			ui.LogError("Menu KRS tidak ditemukan! Merefresh Dashboard...")
			_ = page.Context(ctx).Navigate(baseURL + "index.php")
			time.Sleep(2 * time.Second)
		}
	}

	// 4. Klik Tombol "Tambah" KRS
	tambahAttempt := 1
	for {
		mainPage := getActivePage(page, "main")

		// Pengecekan awal: Jika sudah berada di halaman pemilihan kelas (tabel KRS sudah muncul), langsung bypass!
		_, errTable := mainPage.Timeout(1 * time.Second).Element(".table-common")
		if errTable == nil {
			ui.LogSuccess("Halaman Pemilihan Kelas terdeteksi aktif! Melewati tombol 'Tambah'.")
			break
		}

		tambahBtn, err := mainPage.Timeout(5 * time.Second).ElementX(xpathTambah)
		if err == nil && tambahBtn != nil {
			ui.LogSuccess("Tombol 'Tambah' ditemukan! Membuka Halaman Pemilihan Kelas...")
			tambahBtn.MustClick()
			break
		}

		ui.LogInfo(fmt.Sprintf("[%s] Tombol 'Tambah' belum aktif (Percobaan %d). Menyegarkan...", time.Now().Format("15:04:05"), tambahAttempt))
		
		// Deteksi keberadaan frame (Non-blocking)
		el, err := page.Timeout(500 * time.Millisecond).Element("frame[name='menu'], iframe[name='menu']")
		if err == nil && el != nil {
			// Klik ulang menu KRS pada sidebar jika ber-frame
			menuPage := getActivePage(page, "menu")
			krsLink, err := menuPage.Timeout(2 * time.Second).ElementX("//a[contains(., 'Kartu Rencana Studi')]")
			if err == nil && krsLink != nil {
				krsLink.MustClick()
			}
		} else {
			// Refresh biasa jika portal tanpa frame
			_ = page.Context(ctx).Reload()
		}

		tambahAttempt++
		time.Sleep(time.Duration(schRetryDelaySec * float64(time.Second)))
	}

	// Ambil URL form war saat ini
	info, _ := page.Info()
	warURL := info.URL
	attempt := 1

	// 5. Loop War Utama (Centang & Submit Kelas)
	for attempt <= schMaxRetry {
		ui.LogInfo(fmt.Sprintf("--- PERCOBAAN KRS WAR KE-%d ---", attempt))
		mainPage := getActivePage(page, "main")

		// Tunggu hingga tabel mata kuliah termuat (Non-blocking)
		_, err := mainPage.Timeout(15 * time.Second).Element(".table-common")
		if err != nil {
			ui.LogWarning("Tabel mata kuliah belum muncul. Refreshing...")
			_ = page.Context(ctx).Navigate(warURL)
			time.Sleep(time.Duration(schRetryDelaySec * float64(time.Second)))
			attempt++
			continue
		}

		// Otomatis klik ekspansi semester jika menggunakan accordion (Non-blocking)
		expands, err := mainPage.Timeout(5 * time.Second).ElementsX("//a[contains(text(), 'Paket Semester')]")
		if err == nil {
			for _, exp := range expands {
				_, _ = exp.Eval("this.click()")
			}
		}
		time.Sleep(1500 * time.Millisecond)

		// Memindai baris-baris kelas pada tabel (Non-blocking)
		rows, err := mainPage.Timeout(500 * time.Millisecond).Elements("tr")
		if err != nil {
			ui.LogError("Gagal memindai baris tabel: " + err.Error())
			attempt++
			continue
		}

		tickedCount := 0
		for _, row := range rows {
			cols, err := row.Timeout(200 * time.Millisecond).Elements("td")
			if err == nil && len(cols) >= 4 {
				kw := strings.TrimSpace(cols[2].MustText()) // Kolom kelas
				mw := strings.TrimSpace(cols[3].MustText()) // Kolom nama mata kuliah
				
				for _, target := range courses {
					if strings.Contains(strings.ToLower(mw), strings.ToLower(target.Nama)) && kw == target.Kelas {
						cb, err := cols[1].Timeout(200 * time.Millisecond).Element("input[name='kodeMkul[]']")
						if err == nil && cb != nil {
							// Cek apakah sudah tercentang
							checkedVal, err := cb.Property("checked")
							if err == nil && !checkedVal.Bool() {
								cb.MustClick()
								ui.LogSuccess(fmt.Sprintf("DICENTANG: %s (%s)", mw, kw))
								tickedCount++
								time.Sleep(300 * time.Millisecond) // jeda setelah mencentang agar sinkron
							}
						}
					}
				}
			}
		}

		// Jika ada mata kuliah target yang tercentang, lakukan submit
		if tickedCount > 0 {
			ui.LogInfo(fmt.Sprintf("Menyerahkan %d mata kuliah target ke SIAKAD...", tickedCount))
			time.Sleep(500 * time.Millisecond) // jeda sebelum klik submit agar halaman stabil
			submitBtn, err := mainPage.Timeout(200 * time.Millisecond).ElementX("//input[@name='btnProses' or @name='btnAdd']")
			if err == nil && submitBtn != nil {
				submitBtn.MustClick()
			} else {
				ui.LogError("Tombol submit (btnProses/btnAdd) tidak ditemukan!")
				attempt++
				continue
			}

			// Tunggu respon server
			ui.LogInfo(fmt.Sprintf("Menunggu konfirmasi server... Sabar %ds", sabarWait))
			time.Sleep(2 * time.Second)

			// Ambil screenshot hasil submit
			ssData, ssErr := driver.TakeScreenshot(ctx)
			if ssErr == nil && ssService != nil {
				filename := fmt.Sprintf("success/war_success_attempt_%d.png", attempt)
				ssService.QueueScreenshot(ctx, filename, ssData)
			}

			ui.LogSuccess("Mata kuliah berhasil disubmit! Silakan periksa dashboard SIAKAD Anda.")
			break
		} else {
			ui.LogWarning("Mata kuliah target tidak ditemukan pada halaman ini. Menyegarkan...")
			_ = page.Context(ctx).Navigate(warURL)
			time.Sleep(time.Duration(schRetryDelaySec * float64(time.Second)))
			attempt++
		}
	}

	ui.LogSuccess("Proses war selesai. Browser dibiarkan standby untuk verifikasi manual.")
	ui.LogInfo("Membuka folder hasil tangkapan layar (" + ssSaveDir + ")...")
	openFolder(ssSaveDir)
	ui.Prompt("Tekan Enter jika sudah selesai untuk menutup browser", "")
}

// openFolder opens the given directory path in the OS default file explorer
func openFolder(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default: // linux
		cmd = exec.Command("xdg-open", path)
	}
	_ = cmd.Start()
}
