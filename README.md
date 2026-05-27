# 🤖 Nodebuntu Recon Agent

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Status](https://img.shields.io/badge/Status-Autonomous-orange.svg)]()

**Nodebuntu Recon Agent** is an autonomous AI-driven cybersecurity operator running directly on an Ubuntu server. This bot is more than just a chatbot; it's an agent with its own "thought cycle" designed for proactive reconnaissance, vulnerability analysis, and server monitoring.

---

## ✨ Key Features

- **🧠 Agentic AI Loop:** Utilizes LLMs (Groq/OpenRouter) to reason, select tools, and act autonomously.
- **🛡️ Autonomous Mode:** The AI works independently in the background (`/autoon`) to find security flaws without manual commands.
- **📡 Advanced Monitoring:** Real-time server status reports with visualizations for CPU, RAM, Disk usage, and service health (Nginx, SSH, Tor, etc.).
- **🔍 Integrated Recon Toolkit:** Automated orchestration for `subfinder`, `nuclei`, `httpx`, `naabu`, and `katana`.
- **🔒 Security First:** 
    - **No Hardcoded Passwords:** Uses the `sudoers` system for root operations.
    - **Injection Shielding:** Protects the AI from malicious input through tool output isolation.
    - **Command Allowlist:** Restricts executable commands for system safety.

---

## 🏗️ Project Architecture

The project is built with a clean modular structure:

- `internal/state/`: Central hub for global state and AI memory.
- `pkg/ai/`: AI engine logic, command processing, and autonomous loops.
- `pkg/tools/`: System command execution layer and security validation.
- `pkg/recon/`: Core reconnaissance logic and scope management.
- `pkg/bot/`: Telegram interface and user command handling.

---

## 🚀 Installation & Setup

### 1. Prerequisites
- Go 1.22 or newer.
- Telegram Bot Token.
- Groq or OpenRouter API Key.

### 2. Clone & Configure
```bash
git clone https://github.com/parothegreat/go-bot.git
cd go-bot
```

Create a `.env` file in the root directory:
```env
TELEGRAM_TOKEN=your_bot_token
TELEGRAM_CHAT_ID=your_chat_id
AI_API_KEY=your_ai_api_key
# Optional:
# AI_BASE_URL=https://openrouter.ai/api/v1
# AI_MODEL=llama-3.3-70b-versatile
```

### 3. Sudo Security (Crucial!)
To enable the `/clean` feature safely, add your user to sudoers:
```bash
sudo visudo
# Add the following line at the end of the file:
username ALL=(ALL) NOPASSWD: /usr/bin/apt autoremove, /usr/bin/apt clean
```

### 4. Build & Run
```bash
go build -o recon-bot main.go
./recon-bot
```

---

## 🎮 Usage

### Telegram Commands
- `/status` - View server health and recon progress.
- `/scan` - Run a full reconnaissance cycle manually.
- `/autoon` - Enable **Autonomous Mode** (AI works independently).
- `/missions` - View autonomous AI activity logs.
- `/thought` - Inspect the AI's latest reasoning/analysis.
- `/ai <query>` - Interact directly with the AI Agent.

---

## 🤝 Contributing
Contributions are welcome! Please open an *Issue* or submit a *Pull Request*.

---

## 📄 License
This project is licensed under the [MIT License](LICENSE).

---
*Built with ❤️ for the Cybersecurity community.*
