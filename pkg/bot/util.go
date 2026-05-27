package bot

import (
	"fmt"
	"recon-bot/pkg/recon"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Send(bot *tgbotapi.BotAPI, id int64, text string) {
	msg := tgbotapi.NewMessage(id, text)
	msg.ParseMode = "Markdown"
	if _, err := bot.Send(msg); err != nil {
		msg2 := tgbotapi.NewMessage(id, text)
		bot.Send(msg2)
	}
}

func SendFile(bot *tgbotapi.BotAPI, id int64, path, caption string) {
	doc := tgbotapi.NewDocument(id, tgbotapi.FilePath(path))
	doc.Caption = caption
	bot.Send(doc)
}

func SendVulnReport(bot *tgbotapi.BotAPI, id int64, vulnFile string) {
	entries, reportText := recon.BuildVulnReport(vulnFile)
	if len(entries) == 0 {
		return
	}
	counts := map[string]int{}
	for _, e := range entries {
		counts[e.Severity]++
	}

	summary := fmt.Sprintf(
		"🚨 *VULNERABILITY FOUND — %d Temuan*\n\n"+
			"%s Critical: `%d`  %s High: `%d`\n"+
			"%s Medium: `%d`  %s Low: `%d`\n\n"+
			"*Domain/Subdomain terdampak:*\n",
		len(entries),
		recon.SeverityEmoji("CRITICAL"), counts["CRITICAL"],
		recon.SeverityEmoji("HIGH"), counts["HIGH"],
		recon.SeverityEmoji("MEDIUM"), counts["MEDIUM"],
		recon.SeverityEmoji("LOW"), counts["LOW"],
	)

	Send(bot, id, summary)

	// Send detailed report as file if it's too long
	if len(reportText) > 3500 {
		Send(bot, id, "📄 *Detail report dikirim sebagai file (terlalu panjang)*")
		SendFile(bot, id, vulnFile, "Detailed Vulnerability Report")
	} else {
		Send(bot, id, "```\n"+reportText+"\n```")
	}
}
