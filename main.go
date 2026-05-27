package main

import (
	"fmt"
	"os"
	"recon-bot/internal/state"
	"recon-bot/pkg/ai"
	"recon-bot/pkg/bot"
	"recon-bot/pkg/recon"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	godotenv.Load("/home/parothegreat/.env")

	// Initialize state
	state.Init()

	// Ensure directories exist
	os.MkdirAll(state.DataDir, 0755)
	os.MkdirAll(state.LogsDir, 0755)

	// Initialize Bot API
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	if telegramToken == "" {
		fmt.Println("❌ TELEGRAM_TOKEN not found in .env")
		os.Exit(1)
	}

	tgBot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		fmt.Println("❌ Gagal connect ke Telegram:", err)
		os.Exit(1)
	}

	idStr := os.Getenv("TELEGRAM_CHAT_ID")
	state.ChatID, _ = strconv.ParseInt(idStr, 10, 64)

	fmt.Printf("[INFO] Bot @%s aktif. ChatID: %d\n", tgBot.Self.UserName, state.ChatID)

	// Initial scan summary
	if scope := recon.LoadScope(); len(scope) > 0 {
		state.AIMu.Lock()
		state.ScanSummary = fmt.Sprintf("[SCOPE: %s]", strings.Join(scope, ", "))
		state.AIMu.Unlock()
	}

	// Set bot commands
	tgBot.Request(tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{Command: "status", Description: "Status server lengkap"},
		tgbotapi.BotCommand{Command: "scan", Description: "Jalankan recon scan"},
		tgbotapi.BotCommand{Command: "addscope", Description: "Tambah domain ke scope"},
		tgbotapi.BotCommand{Command: "rmscope", Description: "Hapus domain dari scope"},
		tgbotapi.BotCommand{Command: "setscope", Description: "Ganti semua scope sekaligus"},
		tgbotapi.BotCommand{Command: "ai", Description: "Tanya AI (opsional)"},
		tgbotapi.BotCommand{Command: "clearai", Description: "Reset memori AI"},
		tgbotapi.BotCommand{Command: "clean", Description: "Bersihkan sistem"},
		tgbotapi.BotCommand{Command: "pause", Description: "Pause notifikasi otomatis"},
		tgbotapi.BotCommand{Command: "resume", Description: "Resume notifikasi otomatis"},
		tgbotapi.BotCommand{Command: "help", Description: "Bantuan"},
	))

	// Start background workers
	startBackgroundWorkers(tgBot)

	// Start Autonomous AI Agent
	go ai.StartAutonomousLoop(func(txt string) {
		if state.ChatID != 0 {
			bot.Send(tgBot, state.ChatID, txt)
		}
	})

	// Main update loop
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := tgBot.GetUpdatesChan(u)

	for update := range updates {
		bot.HandleUpdate(tgBot, update)
	}
}

func startBackgroundWorkers(tgBot *tgbotapi.BotAPI) {
	// AI Scan Worker
	go func() {
		for targetChatID := range state.AIScanQueue {
			id, _ := strconv.ParseInt(targetChatID, 10, 64)
			bot.Send(tgBot, id, "🤖 *AI meminta scan dijalankan...*")
			recon.RunScan(
				func(txt string) { bot.Send(tgBot, id, txt) },
				func(path, cap string) { bot.SendFile(tgBot, id, path, cap) },
				func(path string) { bot.SendVulnReport(tgBot, id, path) },
			)
		}
	}()

	// Periodic Status Reports
	go func() {
		for {
			time.Sleep(30 * time.Second)
			if !state.Paused && state.ChatID != 0 && time.Since(state.LastSent) > time.Duration(state.DelayMinutes)*time.Minute {
				bot.Send(tgBot, state.ChatID, recon.GetStatusReport())
				state.LastSent = time.Now()
			}
		}
	}()

	// Vulnerability Watcher
	if state.ChatID != 0 {
		go recon.StartWatcher(func(path string) {
			bot.SendVulnReport(tgBot, state.ChatID, path)
		})
	}
}
