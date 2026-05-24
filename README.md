# Yesplis Auto Order Ticket Bot 🎫⚡

Bot otomatisasi super cepat (War Ticket Bot) yang dirancang khusus untuk platform Yesplis. Dibangun menggunakan Golang dan Chromedp, bot ini dilengkapi dengan arsitektur **Enterprise Stealth** untuk melewati sistem anti-bot modern seperti Cloudflare Turnstile dan DataDome.

## 🌟 Fitur Utama (2026 Stealth Architecture)

1. **Kecepatan Brutal (0ms Reaction):** Menggunakan *aggressive polling* (150ms) alih-alih `WaitVisible` standar. Langsung menembak event tanpa peduli layar *loading* atau *overlay*.
2. **True Physical Mouse Events:** Berbeda dengan bot amatir yang menggunakan JavaScript `element.click()` (`isTrusted: false`), bot ini menghitung koordinat BoundingBox elemen dan menembakkan perintah klik fisik via **Chrome DevTools Protocol (CDP)** (`isTrusted: true`).
3. **Webdriver Spoofing (Anti-Detect):** Menyuntikkan script secara instan sesaat sebelum halaman dimuat (`AddScriptToEvaluateOnNewDocument`) untuk menetralisir `navigator.webdriver` dan memalsukan *fingerprint* eksekusi Chrome.
4. **Automation Flags Cleansing:** Diluncurkan tanpa parameter yang mencurigakan (mematikan `AutomationControlled` dan menggunakan mode `--headless=new` yang lebih aman).
5. **Session Injection:** Mendukung injeksi `cookies.json` untuk *bypass* antrean halaman *login*.
6. **Case-Insensitive Locators:** Menggunakan ekspresi XPath tingkat lanjut yang tidak akan gagal meskipun UI mengubah teks dari "Beli Tiket" menjadi "BELI TIKET" atau "Buy Ticket".

---

## 🛠️ Persiapan & Instalasi

### 1. Kebutuhan Sistem
- **Golang** (Versi 1.20 atau lebih baru)
- **Google Chrome** terinstal di sistem operasi Anda.
- **Ekstensi Browser:** *EditThisCookie* atau sejenisnya (untuk mengambil cookie sesi).

### 2. Membangun Binary

Clone repositori ini dan lakukan kompilasi:

```bash
# Untuk Windows
go build -o bot_ticket.exe main.go

# Untuk Linux / Mac
go build -o bot_ticket main.go
```

### 3. Persiapan Cookies (Wajib!)

Bot ini menghemat waktu dengan melewati halaman *login*. Anda **harus** menyediakan *cookies* sesi login Anda.

1. Buka browser Chrome biasa Anda dan *login* ke akun Yesplis Anda.
2. Gunakan ekstensi *EditThisCookie* untuk mengekspor semua *cookies* pada halaman Yesplis.
3. Buat file baru bernama `cookies.json` di folder yang sama dengan file eksekusi bot Anda.
4. *Paste* data *cookies* tersebut ke dalam `cookies.json` dan simpan.

*(Lakukan langkah ini 15-30 menit sebelum war tiket dimulai agar sesi tetap fresh).*

---

## 🚀 Cara Penggunaan

Jalankan perintah berikut di terminal/command prompt Anda. Sangat disarankan untuk menjalankan bot dalam mode **Non-Headless** (`-headless=false`) agar Anda bisa menyelesaikan proses pembayaran.

```bash
./bot_ticket -url="https://yesplis.com/event/nama-konser-xyz" -ticket="CAT 1" -qty=2 -headless=false
```

### Parameter:
- `-url`: (Wajib) URL spesifik halaman *event* Yesplis. Jangan gunakan halaman beranda.
- `-ticket`: (Wajib) Nama kategori tiket yang ingin dibeli. Tulis sebagian kata kuncinya (misal: "CAT 1" atau "VIP"). Bersifat *case-insensitive*.
- `-qty`: (Opsional) Jumlah tiket. Default: `1`.
- `-headless`: (Opsional) Jika diset `true`, bot berjalan tanpa layar GUI. **Awas:** Jika Anda menggunakan *headless*, Anda tidak akan bisa melakukan klik pembayaran di tahap akhir! Default: `false`.

### Alur Kerja (Workflow):
1. Bot memuat sesi dari `cookies.json`.
2. Menyuntikkan script *stealth*.
3. Masuk ke halaman URL Event.
4. Menunggu tombol "Beli Tiket" muncul secara agresif (auto-reload setiap 300ms jika tombol belum ada).
5. Klik tombol "Beli Tiket" secara *physical click*.
6. Mencari kategori tiket yang sesuai dan menambahkan kuota sesuai parameter `-qty`.
7. Menekan tombol "Checkout/Lanjut/Bayar".
8. **Bot Berhenti:** Bot akan membiarkan Chrome terbuka selama 30 menit. **Tugas Anda adalah mengambil alih kontrol mouse dan menyelesaikan proses pembayaran (memilih VA/QRIS/E-Wallet).**

---

## ☁️ Menjalankan di VPS (Virtual Private Server)

Menjalankan bot di VPS sangat direkomendasikan untuk mendapatkan latensi/ping 1-5ms ke server tiket. Karena VPS biasanya tidak memiliki GUI, ikuti panduan `Xvfb` (Virtual Framebuffer) berikut (Contoh untuk Ubuntu/Debian):

### 1. Install Dependencies
```bash
sudo apt update
# Install Google Chrome
wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
sudo apt install ./google-chrome-stable_current_amd64.deb -y
# Install Virtual Screen dan Remote Desktop
sudo apt install xvfb x11vnc fluxbox -y
```

### 2. Setup Virtual Screen
Buat layar virtual (di *background*):
```bash
Xvfb :99 -screen 0 1920x1080x24 &
export DISPLAY=:99
```

### 3. Setup VNC Server
Agar Anda bisa mengendalikan browser jarak jauh untuk proses pembayaran:
```bash
x11vnc -display :99 -bg -nopw -listen localhost -xkb
```
*(Gunakan SSH Tunneling dari PC Anda ke Port 5900 VPS, lalu buka VNC Viewer di PC Anda).*

### 4. Jalankan Bot
```bash
# Pastikan Anda sudah mengupload bot_ticket dan cookies.json ke VPS
./bot_ticket -url="https://yesplis.com/event/nama-konser-xyz" -ticket="CAT 1" -qty=2 -headless=false
```

Dengan metode ini, VPS yang berkecepatan dewa akan melakukan perang rebutan (klik dan antre), dan saat sudah tembus ke halaman *checkout*, Anda bisa membayarnya dengan santai melalui VNC di laptop Anda.

---

## ⚠️ Peringatan (Disclaimer)
Bot ini dibuat untuk tujuan edukasi dan otomatisasi pribadi. Penulis tidak bertanggung jawab atas segala pemblokiran akun, IP di-*banned*, atau kerugian finansial yang diakibatkan oleh penggunaan sistem otomatisasi ini. Gunakan dengan bijak.
