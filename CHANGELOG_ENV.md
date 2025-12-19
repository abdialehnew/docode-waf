# Ringkasan Perubahan - Konfigurasi dengan .env

## ✅ Selesai - Konfigurasi Port dan Kredensial dengan File .env

### 📋 Perubahan yang Dilakukan

#### 1. File Environment Baru
- **`.env.example`** - Template konfigurasi untuk semua environment variables
- **`.env`** - File konfigurasi aktual untuk development (sudah ada)
- **`ENV_GUIDE.md`** - Panduan lengkap penggunaan environment variables
- **`ENV_SETUP.md`** - Quick start guide untuk setup .env

#### 2. Update Backend
- **`internal/config/config.go`**
  - ✅ Menambahkan method `overrideFromEnv()` untuk membaca environment variables
  - ✅ Menambahkan helper methods: `GetDSN()`, `GetRedisAddr()`, `GetServerAddr()`, `GetAdminAddr()`
  - ✅ Support prioritas konfigurasi: env vars > .env > config.yaml

- **`cmd/waf/main.go`**
  - ✅ Menambahkan import `github.com/joho/godotenv`
  - ✅ Auto-load file .env saat startup
  - ✅ Menggunakan helper methods untuk connection strings

- **`go.mod`**
  - ✅ Menambahkan dependency `github.com/joho/godotenv v1.5.1`

#### 3. Update Docker
- **`docker-compose.yaml`**
  - ✅ Menggunakan `env_file: - .env`
  - ✅ Semua service menggunakan variabel dari .env
  - ✅ Port mapping dinamis dari environment variables

#### 4. Update Dokumentasi
- **`README.md`** - Update bagian konfigurasi
- **`SETUP.md`** - Update petunjuk setup dengan .env

### 🔧 Environment Variables yang Tersedia

#### Server Configuration
```env
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_ADMIN_PORT=9090
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s
```

#### Database Configuration
```env
DATABASE_DRIVER=postgres
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=docode_waf
DATABASE_USER=waf_user
DATABASE_PASSWORD=waf_password
DATABASE_SSLMODE=disable
```

#### Redis Configuration
```env
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=10
```

#### WAF Features
```env
# Rate Limiting
WAF_RATE_LIMIT_ENABLED=true
WAF_RATE_LIMIT_RPS=100
WAF_RATE_LIMIT_BURST=200

# HTTP Flood Protection
WAF_HTTP_FLOOD_ENABLED=true
WAF_HTTP_FLOOD_MAX_RPM=1000
WAF_HTTP_FLOOD_BLOCK_DURATION=300

# Anti-Bot
WAF_ANTI_BOT_ENABLED=true
WAF_ANTI_BOT_CHALLENGE_MODE=js

# GeoIP (Optional)
WAF_GEOIP_ENABLED=false
WAF_GEOIP_DATABASE_PATH=./geoip/GeoLite2-Country.mmdb
```

#### SSL Configuration
```env
SSL_AUTO_CERT=false
SSL_CERT_DIR=./certs
```

#### Logging Configuration
```env
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout
LOG_FILE_PATH=./logs/waf.log
```

### 🚀 Cara Menggunakan

#### 1. Setup Pertama Kali
```bash
# Copy template .env
cp .env.example .env

# Edit sesuai kebutuhan
nano .env
```

#### 2. Jalankan Aplikasi

**Development:**
```bash
# Install dependencies
go mod download

# Run (akan otomatis load .env)
go run cmd/waf/main.go
```

**Production dengan Docker:**
```bash
# Edit .env dengan credentials production
nano .env

# Start semua services
docker-compose up -d
```

### 🔐 Keamanan

#### ⚠️ PENTING - Untuk Production:

1. **Ganti semua password default!**
   ```env
   DATABASE_PASSWORD=GunakaN_P@ssw0rd_Yang_Ku4t!
   REDIS_PASSWORD=J4ng4n_Gunakan_Y4ng_Mud4h!
   ```

2. **File .env sudah di .gitignore**
   - JANGAN commit .env ke Git
   - Hanya commit .env.example

3. **Gunakan SSL di production**
   ```env
   SSL_AUTO_CERT=true
   ```

4. **Set log level yang sesuai**
   ```env
   LOG_LEVEL=warn  # untuk production
   LOG_LEVEL=debug # untuk development
   ```

### 📊 Prioritas Konfigurasi

Aplikasi membaca konfigurasi dengan urutan:

1. **System Environment Variables** (tertinggi)
2. **File .env**
3. **config.local.yaml** (jika ada)
4. **config.yaml** (default)

Environment variable akan **override** nilai dari file YAML!

### ✨ Keuntungan Menggunakan .env

1. ✅ **Keamanan**: Kredensial tidak hardcoded di code
2. ✅ **Fleksibilitas**: Mudah ganti konfigurasi tanpa rebuild
3. ✅ **Multi-Environment**: Beda .env untuk dev/staging/production
4. ✅ **Docker-Friendly**: Terintegrasi sempurna dengan Docker Compose
5. ✅ **Best Practice**: Sesuai dengan 12-Factor App methodology

### 📚 Dokumentasi Tambahan

- **[ENV_GUIDE.md](ENV_GUIDE.md)** - Panduan lengkap semua environment variables
- **[ENV_SETUP.md](ENV_SETUP.md)** - Quick start guide
- **[SETUP.md](SETUP.md)** - Setup dan instalasi aplikasi
- **[README.md](README.md)** - Dokumentasi umum

### 🎯 Testing

```bash
# Verify .env loaded correctly
go run cmd/waf/main.go

# Check logs
# Starting WAF server on 0.0.0.0:8080
# Starting Admin API server on 0.0.0.0:9090
```

### 💡 Tips

1. **Development**: Gunakan password sederhana di .env
2. **Production**: HARUS gunakan password yang kuat
3. **Berbeda environment**: Buat .env.dev, .env.staging, .env.production
4. **Backup**: Simpan .env production di tempat aman (password manager)

---

## Status: ✅ SELESAI

Semua konfigurasi port dan kredensial sudah menggunakan file .env dengan aman!
