# Yesplis Auto Order Ticket Bot ⚡

Bot otomatisasi super cepat (War Ticket Bot) yang dirancang khusus untuk platform Yesplis. Menggunakan **Direct API** untuk kecepatan maksimal tanpa browser.

## Keunggulan

| Aspek | API Bot (ini) | Browser Bot |
|-------|---------------|-------------|
| **Kecepatan** | **50-100ms** | 500-600ms |
| **Dependencies** | Go standard lib | Chrome/Chromium |
| **Resource** | Ringan | Berat |
| **Stealth** | Tidak perlu | Perlu anti-detect |
| **Reliability** | Tinggi | Bisa crash |

## Fitur

1. **Direct API Calls** - Tembak API langsung, 10x lebih cepat dari browser
2. **War Mode** - Polling 100ms untuk menangkap tiket saat dijual
3. **Auto Checkout** - Langsung beli saat tiket tersedia
4. **Multiple Payment Methods** - QRIS, OVO, VA, Credit Card, dll
5. **Context Cancellation** - Ctrl+C untuk berhenti segera
6. **Global Timeout** - Auto stop setelah waktu yang ditentukan

## Instalasi

```bash
# Clone repository
git clone https://github.com/yourusername/yesplis_auto_order_ticket.git
cd yesplis_auto_order_ticket

# Build
go build -o war_ticket main.go
```

## Persiapan Token

1. Buka browser Chrome → Login ke Yesplis
2. F12 → Application → Cookies → `access_token`
3. Copy token JWT-nya

## Cara Penggunaan

```bash
./war_ticket \
  -token="JWT_TOKEN_KAMU" \
  -event="harbour-fest-2026" \
  -ticket="PLATINUM" \
  -payment="QRIS" \
  -qty=2 \
  -timeout=5m
```

### Parameter

| Parameter | Wajib | Default | Deskripsi |
|-----------|-------|---------|-----------|
| `-token` | Ya | - | JWT token dari cookies |
| `-event` | Ya | - | Event slug (dari URL) |
| `-ticket` | Ya | - | Nama kategori tiket |
| `-payment` | Tidak | QRIS | Metode pembayaran |
| `-qty` | Tidak | 1 | Jumlah tiket |
| `-timeout` | Tidak | 2m | Global timeout |

### Metode Pembayaran

| Kode | Metode | Fee |
|------|--------|-----|
| QRIS | QRIS | 2.5% + Rp 4.000 |
| OVO | OVO | 2.5% + Rp 4.000 |
| DANA | DANA | 2.5% + Rp 4.000 |
| BCA | BCA VA | 1.75% + Rp 8.000 |
| BRI | BRI VA | 1.75% + Rp 8.000 |
| MANDIRI | Mandiri VA | 1.75% + Rp 8.000 |
| CREDIT_CARD | Credit Card | 3.75% + Rp 6.000 |

## Alur Kerja

```
1. Get event details
2. Get payment methods
3. War mode: Polling 100ms
4. Get checkout detail (cek tiket tersedia)
5. Buy ticket (langsung tembak!)
6. Selesai → Payment URL
```

## Contoh Output

```
Getting event details for: harbour-fest-2026
Event: HARBOUR FEST 2026
Status: publish
Tax: 0%
Getting payment methods...
Payment method: QRIS (Fee: 2.5% + Rp 4000)
Waiting for tickets (War Mode)...
Target: PLATINUM x1
Attempting to buy ticket...
Found ticket: EARLY BIRD - PLATINUM (Price: Rp 110000)
SUCCESS! Order ID: abc123
Total: Rp 114750
Payment URL: https://payment.yesplis.com/...
```

## Tips War Ticket

1. **Jalankan 5-10 menit sebelum** tiket dijual
2. **Gunakan VPS** di Jakarta untuk latensi 1-5ms
3. **Token harus fresh** - login ulang 30 menit sebelum
4. **Jangan buka tab lain** yang pakai token yang sama

## API Endpoints

| Endpoint | Method | Fungsi |
|----------|--------|--------|
| `/api/v3/public/events/detail/{slug}` | GET | Event details |
| `/api/v3/transaction/checkout/detail` | POST | Cart details |
| `/api/v3/transaction/count-payment` | POST | Calculate payment |
| `/api/v3/public/payment-methods/{slug}` | GET | Payment methods |
| `/api/v3/transaction/buy-ticket` | POST | Beli tiket |

## Disclaimer

Bot ini dibuat untuk tujuan edukasi dan otomatisasi pribadi. Penulis tidak bertanggung jawab atas segala pemblokiran akun, IP di-banned, atau kerugian finansial yang diakibatkan oleh penggunaan sistem otomatisasi ini. Gunakan dengan bijak.
