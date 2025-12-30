Buatkan Aplikasi Web Application Firewall (WAF) berbasis web menggunakan golang dan nginx atau yang lebih compatible dengan golang dengan fitur sebagai berikut
- Reserve Proxy ✅
- Rate Limiting ✅
- Blocking by IP, Region, URL dan lainnya ✅
- Management Certificate SSL ✅
- Http Flood ✅
- Anti Bot ✅
- Management Group IP baik single IP maupun 1 blok IP ✅
- bisa Custom vhost, vhost location ✅
- Dashboard Monitoring Traffic, Monitoring Serangan ✅
- UI menggunakan reactJS + Tailwindcss ✅
- Application Branding (Custom Name & Logo) ✅

## Fitur Tambahan yang Sudah Diimplementasikan

### Advanced VHost Configuration
- WebSocket Support
- HTTP Version Selection (HTTP/1.1, HTTP/2)
- TLS Version Selection (TLS 1.2, TLS 1.3, atau kombinasi)
- Max Upload Size Configuration
- Proxy Timeouts (Read & Connect)
- Custom Headers (Key-Value pairs)
- Custom Location Blocks dengan backend URL validation

### Application Branding
- Custom Application Name - nama aplikasi dapat diubah sesuai kebutuhan
- Custom Application Logo - upload logo dengan format PNG, JPG, atau SVG (max 2MB)
- Settings disimpan di database dan diterapkan di seluruh aplikasi
- Logo dan nama muncul di sidebar dan browser title

## Database Migration

Tabel baru: `app_settings`
```sql
CREATE TABLE app_settings (
    id INT PRIMARY KEY DEFAULT 1,
    app_name VARCHAR(255) NOT NULL DEFAULT 'Docode WAF',
    app_logo TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT single_row CHECK (id = 1)
);
```

## API Endpoints

### Settings
- GET `/api/v1/settings/app` - Get application settings
- POST `/api/v1/settings/app` - Save application settings

## Cara Menggunakan

1. Login ke aplikasi
2. Buka menu "Settings"
3. Di bagian "Application Settings":
   - Masukkan nama aplikasi yang diinginkan
   - Upload logo (opsional) dengan format PNG/JPG/SVG
   - Klik "Save Application Settings"
4. Halaman akan reload otomatis dan menampilkan branding baru di sidebar