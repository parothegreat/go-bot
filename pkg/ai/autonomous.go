package ai

import (
	"fmt"
	"recon-bot/internal/state"
	"time"
)

// StartAutonomousLoop runs the AI in the background to make decisions and act.
func StartAutonomousLoop(sendUpdate func(string)) {
	for {
		// Only run if AutonomousMode is enabled
		state.AIMu.Lock()
		active := state.AutonomousMode
		state.AIMu.Unlock()

		if !active {
			time.Sleep(5 * time.Minute)
			continue
		}

		fmt.Println("[AUTONOMOUS] AI sedang berpikir untuk langkah selanjutnya...")
		
		// 1. Ask the AI what it wants to do based on the current state
		prompt := "SEKARANG KAMU DALAM MODE AUTONOMOUS. Lihat kondisi server saat ini. Apa langkah recon paling strategis yang harus dilakukan sekarang? Kamu boleh menjalankan tool, membaca file, atau memulai scan. Berikan alasanmu dan lakukan tindakan tersebut via tool. Jika tidak ada yang mendesak, berikan ringkasan status saat ini."
		
		response, actions := RunAI(prompt)

		// 2. Notify the user about what the AI decided to do
		if response != "" {
			updateMsg := fmt.Sprintf("🤖 *Autonomous Agent Update:*\n\n%s", response)
			sendUpdate(updateMsg)
		}

		// 3. Handle specific actions (like SCAN_NOW)
		for _, action := range actions {
			if action == "SCAN_NOW" {
				state.AIScanQueue <- fmt.Sprintf("%d", state.ChatID)
			}
		}

		// 4. Record the mission in history/log
		state.AIMu.Lock()
		state.MissionLog = append(state.MissionLog, fmt.Sprintf("[%s] %s", time.Now().Format("15:04"), response))
		if len(state.MissionLog) > 50 {
			state.MissionLog = state.MissionLog[1:]
		}
		state.AIMu.Unlock()

		// 5. Sleep between autonomous cycles to save tokens and resources
		// Recommended: 30-60 minutes for a balanced agent
		time.Sleep(60 * time.Minute)
	}
}
