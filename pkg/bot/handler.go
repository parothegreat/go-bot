package bot

import (
	"context"
	"fmt"
	"os/exec"
	"recon-bot/internal/state"
	"recon-bot/pkg/ai"
	"recon-bot/pkg/recon"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	m := update.Message
	text := strings.TrimSpace(m.Text)
	lower := strings.ToLower(text)

	// Logging (consider moving to a proper logger later)
	fmt.Printf("[MSG] @%s: %s\n", m.From.UserName, text)

	// Dependency injection via closures for recon package
	sendFunc := func(txt string) { Send(bot, m.Chat.ID, txt) }
	sendFileFunc := func(path, caption string) { SendFile(bot, m.Chat.ID, path, caption) }
	sendVulnFunc := func(path string) { SendVulnReport(bot, m.Chat.ID, path) }

	go func(m *tgbotapi.Message, text, lower string) {
		isPrivate := m.Chat.Type == "private"
		isCommand := strings.HasPrefix(text, "/")
		isAITrigger := strings.HasPrefix(lower, "/ai") || (!isCommand && isPrivate)

		switch {
		case isAITrigger:
			query := text
			if strings.HasPrefix(lower, "/ai") {
				query = strings.TrimSpace(text[3:])
			}
			if query == "" {
				sendFunc("🤖 *Nodebuntu Recon Agent*\n\n" +
					"Aku punya akses langsung ke server. Aku bisa:\n" +
					"• Lihat file & direktori di server\n" +
					"• Jalankan command bash\n" +
					"• Baca hasil scan terkini\n" +
					"• Analisis vulnerability\n")
				return
			}
			
			response, actions := ai.RunAI(query)
			if response != "" {
				sendFunc(response)
			}
			for _, action := range actions {
				switch {
				case action == "SCAN_NOW":
					sendFunc("🤖 _AI memutuskan scan perlu dijalankan..._")
					go recon.RunScan(sendFunc, sendFileFunc, sendVulnFunc)
				case strings.HasPrefix(action, "ADDSCOPE:"):
					domain := strings.TrimPrefix(action, "ADDSCOPE:")
					sendFunc(fmt.Sprintf("🤖 _AI merekomendasikan tambah scope: `%s`_", domain))
					handleAddScope(bot, m.Chat.ID, domain)
				}
			}

		case lower == "/status":
			sendFunc(recon.GetStatusReport())

		case lower == "/scan":
			go recon.RunScan(sendFunc, sendFileFunc, sendVulnFunc)

		case lower == "/stop":
			// Implementing stop by killing recon processes
			sendFunc("🛑 *Menghentikan semua proses recon...*")
			exec.Command("bash", "-c", "pkill -f 'subfinder|nuclei|httpx|katana|naabu|monitor'").Run()
			state.ScanMu.Lock()
			state.ScanRunning = false
			state.ScanMu.Unlock()
			sendFunc("✅ *Proses dihentikan.*")

		case strings.HasPrefix(lower, "/addscope"):
			args := strings.TrimSpace(text[9:])
			handleAddScope(bot, m.Chat.ID, args)

		case strings.HasPrefix(lower, "/rmscope"):
			args := strings.TrimSpace(text[8:])
			handleRemoveScope(bot, m.Chat.ID, args)

		case strings.HasPrefix(lower, "/setscope"):
			args := strings.TrimSpace(text[9:])
			handleSetScope(bot, m.Chat.ID, args)

		case lower == "/sysinfo":
			sendFunc(recon.GetStatusReport()) // Reuse status report or expand later

		case lower == "/clearai":
			ai.ResetAIHistory()
			sendFunc("🧹 Memory AI direset")

		case lower == "/clean":
			sendFunc("🧹 Membersihkan sistem...")
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			// FIXED: No hardcoded password. User must update sudoers.
			cmd := exec.CommandContext(ctx, "sudo", "apt", "autoremove", "-y")
			out, err := cmd.CombinedOutput()
			result := string(out)
			if err != nil {
				result += "\nERROR: " + err.Error()
				result += "\n\n_Pastikan kamu sudah menambahkan user ke sudoers (NOPASSWD)._"
			}
			sendFunc("✅ *Selesai!*\n```\n" + result + "\n```")

		case lower == "/pause":
			state.Paused = true
			sendFunc("⏸ Notifikasi otomatis di-*pause*")

		case lower == "/resume":
			state.Paused = false
			state.LastSent = time.Time{}
			sendFunc("▶️ Notifikasi otomatis *aktif* kembali")

		case lower == "/autoon":
			state.AIMu.Lock()
			state.AutonomousMode = true
			state.AIMu.Unlock()
			sendFunc("🤖 *Autonomous Mode: ON*\nAI sekarang akan bekerja sendiri di background dan memberikan update berkala.")

		case lower == "/autooff":
			state.AIMu.Lock()
			state.AutonomousMode = false
			state.AIMu.Unlock()
			sendFunc("🤖 *Autonomous Mode: OFF*\nAI berhenti bekerja secara otonom.")

		case lower == "/missions":
			state.AIMu.Lock()
			logs := state.MissionLog
			state.AIMu.Unlock()
			if len(logs) == 0 {
				sendFunc("📝 *Mission Log kosong.* AI belum menjalankan aksi otonom.")
				return
			}
			msg := "📝 *Recent AI Missions (Autonomous):*\n\n"
			// Show last 5 missions
			start := len(logs) - 5
			if start < 0 { start = 0 }
			for _, entry := range logs[start:] {
				msg += fmt.Sprintf("• %s\n\n", entry)
			}
			sendFunc(msg)

		case lower == "/thought":
			state.AIMu.Lock()
			history := state.AIHistory
			state.AIMu.Unlock()
			if len(history) == 0 {
				sendFunc("💭 *AI belum memikirkan apa pun.*")
				return
			}
			// Show the last AI response to see its "current" thought
			var lastThought string
			for i := len(history) - 1; i >= 0; i-- {
				if history[i].Role == "assistant" {
					lastThought = history[i].Content
					break
				}
			}
			if lastThought == "" {
				sendFunc("💭 *Tidak ada memori pemikiran AI terakhir.*")
			} else {
				sendFunc("💭 *Last AI Thought Process:*\n\n" + lastThought)
			}

		case lower == "/start" || lower == "/help":
			sendFunc("🤖 *Nodebuntu Recon Agent*\n\n" +
				"*📡 Server:*\n" +
				"• `/status` — Status lengkap server\n" +
				"• `/clean` — Bersihkan sistem\n\n" +
				"*🎯 Recon:*\n" +
				"• `/scan` — Jalankan full recon cycle\n" +
				"• `/stop` — Hentikan scan paksa\n" +
				"• `/addscope domain.com` — Tambah target\n\n" +
				"*🧠 AI Agent:*\n" +
				"• `/ai [tanya]` — Tanya AI secara manual\n" +
				"• `/autoon` — AKTIFKAN Mode Otonom (AI kerja sendiri)\n" +
				"• `/autooff` — Matikan Mode Otonom\n" +
				"• `/clearai` — Reset memory percakapan AI")
		}
	}(m, text, lower)
}

// Handler wrappers for scope management to keep logic in recon package

func handleAddScope(bot *tgbotapi.BotAPI, id int64, args string) {
	raw := strings.Fields(strings.ToLower(args))
	if len(raw) == 0 {
		Send(bot, id, "❌ Format: `/addscope domain.com domain2.com`")
		return
	}
	existing := recon.LoadScope()
	existMap := map[string]bool{}
	for _, d := range existing {
		existMap[d] = true
	}
	var added, skipped []string
	for _, d := range raw {
		d = strings.TrimPrefix(d, "http://")
		d = strings.TrimPrefix(d, "https://")
		d = strings.TrimSuffix(d, "/")
		if existMap[d] {
			skipped = append(skipped, d)
		} else {
			existing = append(existing, d)
			added = append(added, d)
		}
	}
	if err := recon.SaveScope(existing); err != nil {
		Send(bot, id, "❌ Gagal menyimpan scope: "+err.Error())
		return
	}
	msg := ""
	if len(added) > 0 {
		msg += "✅ *Ditambahkan:*\n"
		for _, d := range added {
			msg += fmt.Sprintf("• `%s`\n", d)
		}
	}
	if len(skipped) > 0 {
		msg += "⚠️ *Sudah ada (skip):*\n"
		for _, d := range skipped {
			msg += fmt.Sprintf("• `%s`\n", d)
		}
	}
	msg += "\n" + recon.GetScopeText()
	state.AIMu.Lock()
	state.ScanSummary = recon.RebuildScopeInSummary(recon.LoadScope())
	state.AIMu.Unlock()
	Send(bot, id, msg)
}

func handleRemoveScope(bot *tgbotapi.BotAPI, id int64, args string) {
	raw := strings.Fields(strings.ToLower(args))
	if len(raw) == 0 {
		Send(bot, id, "❌ Format: `/rmscope domain.com`")
		return
	}
	removeMap := map[string]bool{}
	for _, d := range raw {
		removeMap[d] = true
	}
	existing := recon.LoadScope()
	var newScope, removed []string
	for _, d := range existing {
		if removeMap[d] {
			removed = append(removed, d)
		} else {
			newScope = append(newScope, d)
		}
	}
	if len(removed) == 0 {
		Send(bot, id, "⚠️ Domain tidak ditemukan di scope.")
		return
	}
	if err := recon.SaveScope(newScope); err != nil {
		Send(bot, id, "❌ Gagal menyimpan scope: "+err.Error())
		return
	}
	msg := "🗑 *Dihapus dari scope:*\n"
	for _, d := range removed {
		msg += fmt.Sprintf("• `%s`\n", d)
	}
	msg += "\n" + recon.GetScopeText()
	state.AIMu.Lock()
	state.ScanSummary = recon.RebuildScopeInSummary(recon.LoadScope())
	state.AIMu.Unlock()
	Send(bot, id, msg)
}

func handleSetScope(bot *tgbotapi.BotAPI, id int64, args string) {
	raw := strings.Fields(strings.ToLower(args))
	if len(raw) == 0 {
		Send(bot, id, "❌ Format: `/setscope domain1.com domain2.com`\n\n_Ini akan *mengganti* scope yang ada._")
		return
	}
	var domains []string
	for _, d := range raw {
		d = strings.TrimPrefix(d, "http://")
		d = strings.TrimPrefix(d, "https://")
		d = strings.TrimSuffix(d, "/")
		if d != "" {
			domains = append(domains, d)
		}
	}
	if err := recon.SaveScope(domains); err != nil {
		Send(bot, id, "❌ Gagal menyimpan scope: "+err.Error())
		return
	}
	state.AIMu.Lock()
	state.ScanSummary = recon.RebuildScopeInSummary(domains)
	state.AIMu.Unlock()
	Send(bot, id, "✅ *Scope berhasil diganti!*\n\n"+recon.GetScopeText())
}
