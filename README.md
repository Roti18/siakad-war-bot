# Go Client - Bot War KRS SIAKAD

Aplikasi client berbasis **Go (Golang)** yang dirancang untuk melakukan otomatisasi pengisian Kartu Rencana Studi (KRS) pada portal SIAKAD menggunakan engine browser **Rod** (Chrome DevTools Protocol).

Aplikasi ini dipisahkan secara independen dari Backend Server untuk menjaga arsitektur modular, modularitas, keamanan kredensial, dan performa tinggi.

---

## Struktur Folder Client

```text
client/
├── cmd/
│   └── bot/                  # Entrypoint utama bot (main.go)
├── configs/
│   └── config.json           # File konfigurasi utama bot (browser, target matkul, dll)
└── internal/
    ├── browser/              # Driver otomasi browser (Rod) & Screenshot Service
    ├── config/               # Manager konfigurasi JSON
    ├── domain/               # Deklarasi interface dan model entitas
    └── security/             # Device Fingerprint & OS-level Secret Store
```

---

## Cara Setup & Konfigurasi

### 1. File Konfigurasi (`client/configs/config.json`)
Sesuaikan parameter browser, jadwal standby, dan daftar mata kuliah target di dalam `client/configs/config.json`.

Contoh isi konfigurasi:
```json
{
  "browser": {
    "headless": false,
    "width": 1920,
    "height": 1080,
    "block_images": true,
    "block_css": true,
    "block_fonts": true,
    "block_media": true
  },
  "schedule": {
    "time": "09:00:00",
    "retry_delay_seconds": 0.5,
    "max_retry": 100
  },
  "screenshot": {
    "enable": true,
    "save_directory": "warResult"
  },
  "courses": [
    {
      "nama": "Pengembangan Sistem Berbasis Framework",
      "kelas": "IF 4A"
    }
  ]
}
```

### 2. Kredensial Pengguna (`.env`)
Untuk keamanan, kredensial sensitif disimpan pada file `.env` di root folder project (searah dengan `go.work`):
```env
BASE_URL=https://siakad.xxxxx.ac.id/
NIM=1234567890
PASSWORD=password_siakad_kamu
```

---

## Sistem Screenshot (warResult)

Aplikasi client ini mengimplementasikan **`AsyncScreenshotService`** secara asinkron menggunakan goroutine worker pool:
1. **Asinkron & Non-Blocking**: Proses penangkapan layar dan penulisan file gambar tidak mengganggu proses "war" utama, sehingga bot tetap berjalan secepat mungkin tanpa delay I/O disk.
2. **Penyimpanan**: Hasil screenshot akan disimpan secara otomatis pada folder `warResult/` dengan subfolder:
   - `warResult/success/` - Saat mata kuliah target berhasil dicentang dan disubmit.
   - `warResult/failed/` - Saat terjadi error sistem, timeout, atau kegagalan transaksi war.

---

## Cara Menjalankan

Jalankan perintah berikut di root folder repository:

```bash
# Menjalankan bot client
go run cmd/bot/main.go
```

> **Catatan Status Migrasi**: Saat ini client masih dalam tahap kerangka arsitektur (skeleton). Struktur dasar seperti Driver Browser (Rod), Enkripsi/Keamanan (Fingerprint), Konfigurasi (config.json), dan Screenshot Service sudah siap, sedangkan modul War Engine utama sedang dalam proses pemindahan dari Python (final.py) ke Go.
