# Shared Module - War KRS

Module ini berisi kode bersama (**Shared Package**) yang digunakan oleh aplikasi **Server** (backend) dan **Client** (bot). 

Modul ini **tidak diperkenankan memiliki logika bisnis (business logic)**, melainkan hanya struktur data dan fungsi helper umum.

---

## Struktur Folder Shared

```text
shared/
├── api/          # DTO Request/Response API komunikasi Client-Server
├── constants/    # Konstanta umum seperti status code, event, dan nama tipe
├── dto/          # Data Transfer Object umum
├── validation/   # Fungsi-fungsi validasi bersama (misal format lisensi/NIM)
└── version/      # Validasi versi aplikasi client untuk kompatibilitas
```

---

## Aturan Dependensi (Dependency Rules)

- **Server** dilarang keras meng-import package khusus **Client**.
- **Client** dilarang keras meng-import package khusus **Server** (termasuk tidak boleh mengakses database/SQLite secara langsung).
- Keduanya berinteraksi melalui HTTP REST API dan hanya diperkenankan menggunakan format data terstruktur (DTO) yang didefinisikan di dalam folder `shared/` ini.
