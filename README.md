# 🤖 Nodebuntu Recon Agent

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Status](https://img.shields.io/badge/Status-Autonomous-orange.svg)]()

**Nodebuntu Recon Agent** adalah operator cybersecurity otonom berbasis AI yang berjalan langsung di server Ubuntu. Bot ini bukan sekadar chatbot biasa; ia adalah agen yang memiliki "siklus pemikiran" sendiri untuk melakukan pengintaian (reconnaissance), analisis kerentanan, dan pemantauan server secara proaktif.

---

## ✨ Fitur Utama

- **🧠 Agentic AI Loop:** Menggunakan LLM (Groq/OpenRouter) untuk berpikir, memilih tool, dan bertindak secara otonom.
- **🛡️ Autonomous Mode:** AI dapat bekerja sendiri di background (`/autoon`) untuk mencari celah keamanan tanpa perintah manual.
- **📡 Advanced Monitoring:** Laporan status server real-time dengan visualisasi penggunaan CPU, RAM, Disk, serta status layanan (Nginx, SSH, Tor, dll).
- **🔍 Integrated Recon Toolkit:** Orkestrasi otomatis untuk `subfinder`, `nuclei`, `httpx`, `naabu`, dan `katana`.
- **🔒 Security First:** 
    - **No Hardcoded Passwords:** Menggunakan sistem `sudoers` untuk operasi root.
    - **Injection Shielding:** Melindungi AI dari manipulasi input melalui isolasi output tool.
    - **Command Allowlist:** Membatasi perintah yang dapat dijalankan AI demi keamanan sistem.

---

## 🏗️ Arsitektur Proyek

Proyek ini dibangun dengan struktur modular yang bersih:

- `internal/state/`: Pusat kendali status global dan memori AI.
- `pkg/ai/`: Logika mesin AI, pemrosesan perintah, dan loop otonom.
- `pkg/tools/`: Layer eksekusi perintah sistem dan validasi keamanan.
- `pkg/recon/`: Logika inti reconnaissance dan manajemen scope.
- `pkg/bot/`: Interface Telegram dan penanganan perintah pengguna.

---

## 🚀 Instalasi & Persiapan

### 1. Prasyarat
- Go 1.22 atau lebih baru.
- Token Bot Telegram.
- API Key Groq atau OpenRouter.

### 2. Kloning & Konfigurasi
```bash
git clone https://github.com/parothegreat/go-bot.git
cd go-bot
```

Buat file `.env` di direktori utama:
```env
TELEGRAM_TOKEN=your_bot_token
TELEGRAM_CHAT_ID=your_chat_id
AI_API_KEY=your_ai_api_key
# Opsional:
# AI_BASE_URL=https://openrouter.ai/api/v1
# AI_MODEL=llama-3.3-70b-versatile
```

### 3. Keamanan Sudo (Penting!)
Agar fitur `/clean` berfungsi aman, tambahkan user Anda ke sudoers:
```bash
sudo visudo
# Tambahkan baris berikut di akhir file:
username ALL=(ALL) NOPASSWD: /usr/bin/apt autoremove, /usr/bin/apt clean
```

### 4. Build & Run
```bash
go build -o recon-bot main.go
./recon-bot
```

---

## 🎮 Cara Penggunaan

### Perintah Telegram
- `/status` - Melihat kondisi server dan progres recon.
- `/scan` - Menjalankan siklus pengintaian penuh secara manual.
- `/autoon` - Mengaktifkan **Mode Otonom** (AI akan bekerja sendiri).
- `/missions` - Melihat apa yang telah dikerjakan AI secara otonom.
- `/thought` - Mengintip proses berpikir/analisis terakhir AI.
- `/ai <tanya>` - Berinteraksi langsung dengan AI Agent.

---

## 🤝 Kontribusi
Kontribusi sangat terbuka! Silakan buka *Issue* atau kirimkan *Pull Request*.

---

## 📄 Lisensi
Proyek ini dilisensikan di bawah [MIT License](LICENSE).

---
*Dibuat dengan ❤️ untuk komunitas Cybersecurity.*
