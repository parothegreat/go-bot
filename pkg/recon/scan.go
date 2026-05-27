package recon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"recon-bot/internal/state"
	"recon-bot/pkg/tools"
	"strings"
	"time"
)

// RunScan executes the full recon cycle.
func RunScan(send func(string), sendFile func(string, string), sendVuln func(string)) {
	state.ScanMu.Lock()
	if state.ScanRunning {
		state.ScanMu.Unlock()
		send("⚠️ Scan sedang berjalan. Tunggu selesai.")
		return
	}
	state.ScanRunning = true
	state.ScanMu.Unlock()

	defer func() {
		state.ScanMu.Lock()
		state.ScanRunning = false
		state.ScanMu.Unlock()
	}()

	scope := LoadScope()
	if len(scope) == 0 {
		send("⚠️ *Scope kosong!*\n\nTambahkan target dulu:\n`/addscope target.com`")
		return
	}

	send(fmt.Sprintf(
		"🚀 *Memulai Recon Cycle...*\n\n🎯 Scope: `%s`\n_Bisa memakan beberapa menit_",
		strings.Join(scope, ", "),
	))

	// Executing the monitor script
	cmd := exec.Command("/bin/bash", "/home/parothegreat/recon-engine/scripts/monitor.sh")
	out, err := cmd.CombinedOutput()

	outStr := string(out)
	if len(outStr) > 800 {
		outStr = outStr[:800] + "\n...(terpotong)"
	}

	if err != nil {
		send(fmt.Sprintf("⚠️ *Scan selesai dengan error*\n```\n%s\n```", err.Error()))
	} else if outStr != "" {
		send(fmt.Sprintf("✅ *Scan selesai!*\n```\n%s\n```", outStr))
	} else {
		send("✅ *Scan selesai!*")
	}

	// Send alive hosts file
	aliveFile := filepath.Join(state.DataDir, "known_alive.txt")
	if _, err := os.Stat(aliveFile); err == nil {
		sendFile(aliveFile, fmt.Sprintf("📋 Alive Hosts (%s hosts)", tools.CountLines(aliveFile)))
	}

	// Check for vulnerabilities
	for _, vf := range []string{
		filepath.Join(state.DataDir, "nuclei_results.json"),
		filepath.Join(state.DataDir, "vuln.txt"),
		filepath.Join(state.LogsDir, "nuclei_results.json"),
	} {
		if info, err := os.Stat(vf); err == nil && info.Size() > 0 {
			sendVuln(vf)
			break
		}
	}
}

// StartWatcher starts a background process to watch for new vulnerabilities.
func StartWatcher(sendVuln func(string)) {
	lastMod := make(map[string]time.Time)
	files := []string{
		filepath.Join(state.DataDir, "nuclei_results.json"),
		filepath.Join(state.LogsDir, "nuclei_results.json"),
	}

	for {
		time.Sleep(1 * time.Minute)
		for _, f := range files {
			info, err := os.Stat(f)
			if err != nil {
				continue
			}
			prev, exists := lastMod[f]
			if !exists {
				lastMod[f] = info.ModTime()
				continue
			}
			if info.ModTime().After(prev) && info.Size() > 0 {
				lastMod[f] = info.ModTime()
				sendVuln(f)
			}
		}
	}
}
