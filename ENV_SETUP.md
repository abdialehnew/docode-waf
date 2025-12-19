# Panduan Konfigurasi dengan File .env

## Perubahan yang Dilakukan

Konfigurasi port dan kredensial sekarang menggunakan file `.env` untuk keamanan dan fleksibilitas yang lebih baik.

## File yang Ditambahkan/Diubah

### 1. File Baru
- ✅ `.env.example` - Template konfigurasi environment
- ✅ `.env` - File konfigurasi aktual (sudah ada untuk development)
- ✅ `ENV_GUIDE.md` - Panduan lengkap konfigurasi environment

### 2. File yang Diperbarui
- ✅ `internal/config/config.go` - Menambahkan dukungan environment variables
- ✅ `cmd/waf/main.go` - Menggunakan godotenv untuk load .env
- ✅ `go.mod` - Menambahkan dependency godotenv
- ✅ `docker-compose.yaml` - Menggunakan variabel dari .env
- ✅ `README.md` - Update dokumentasi konfigurasi
- ✅ `SETUP.md` - Update petunjuk setup

## Cara Menggunakan

### 1. Setup Awal

```bash
# Copy template .env
cp .env.example .env

# Edit file .env dengan credentials Anda
nano .env
```

### 2. Konfigurasi Penting

Berikut adalah konfigurasi yang HARUS diubah di production:

```env
# Database - Ganti password!
DATABASE_PASSWORD=your_secure_password_here

# Redis - Tambahkan password untuk keamanan
REDIS_PASSWORD=your_redis_password_here

# Port (opsional, bisa disesuaikan)
SERVER_PORT=8080
SERVER_ADMIN_PORT=9090
```

### 3. Menjalankan Aplikasi

#### Development
```bash
# Aplikasi akan otomatis membaca .env
go run cmd/waf/main.go
```

#### Production dengan Docker
```bash
# Docker Compose akan otomatis menggunakan .env
docker-compose up -d
```

## Prioritas Konfigurasi

Aplikasi membaca konfigurasi dengan urutan prioritas berikut:

1. **System Environment Variables** (prioritas tertinggi)
2. **File .env**
3. **config.local.yaml** (jika ada)
4. **config.yaml** (default)

Ini berarti environment variable akan override nilai dari file YAML.

## Keamanan

### ⚠️ PENTING

1. **JANGAN commit file .env ke Git!**
   - `.env` sudah ada di `.gitignore`
   - Hanya commit `.env.example`

2. **Gunakan password yang kuat**
   - Minimum 16 karakter
   - Kombinasi huruf, angka, dan simbol

3. **Berbeda untuk setiap environment**
   - Development: gunakan password sederhana
   - Production: HARUS gunakan password yang kuat dan unik

## Contoh Konfigurasi

### Development (.env)
```env
SERVER_HOST=localhost
DATABASE_HOST=localhost
DATABASE_PASSWORD=dev123
LOG_LEVEL=debug
```

### Production (.env)
```env
SERVER_HOST=0.0.0.0
DATABASE_HOST=db.production.com
DATABASE_PASSWORD=Str0ng!P@ssw0rd#2024
LOG_LEVEL=warn
SSL_AUTO_CERT=true
```

## Troubleshooting

### Error: "No .env file found"
Ini adalah warning saja, aplikasi akan menggunakan config.yaml atau environment variables sistem.

### Error: "Failed to connect to database"
1. Periksa `DATABASE_HOST` dan `DATABASE_PORT` di .env
2. Pastikan PostgreSQL berjalan
3. Verifikasi username dan password

### Error: "Port already in use"
Ubah `SERVER_PORT` atau `SERVER_ADMIN_PORT` di .env ke port yang available.

## Fitur Utama

### Environment Variables Tersedia

**Server:**
- `SERVER_HOST` - Host address
- `SERVER_PORT` - WAF proxy port
- `SERVER_ADMIN_PORT` - Admin API port

**Database:**
- `DATABASE_HOST` - PostgreSQL host
- `DATABASE_PORT` - PostgreSQL port
- `DATABASE_NAME` - Database name
- `DATABASE_USER` - Database username
- `DATABASE_PASSWORD` - Database password

**WAF Features:**
- `WAF_RATE_LIMIT_ENABLED` - Enable/disable rate limiting
- `WAF_RATE_LIMIT_RPS` - Requests per second
- `WAF_HTTP_FLOOD_ENABLED` - Enable/disable HTTP flood protection
- `WAF_ANTI_BOT_ENABLED` - Enable/disable anti-bot

Dan masih banyak lagi! Lihat [ENV_GUIDE.md](ENV_GUIDE.md) untuk daftar lengkap.

## Bantuan Lebih Lanjut

Untuk panduan lengkap konfigurasi environment variables, lihat:
- [ENV_GUIDE.md](ENV_GUIDE.md) - Panduan lengkap environment variables
- [SETUP.md](SETUP.md) - Panduan setup dan instalasi
- [README.md](README.md) - Dokumentasi umum aplikasi
