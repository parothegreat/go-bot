package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"recon-bot/internal/state"
	"strings"
	"time"
)

// ExecuteTool runs a tool requested by the AI.
func ExecuteTool(toolName, toolInput string) string {
	toolInput = strings.TrimSpace(toolInput)

	switch toolName {
	case "get_scan_status":
		state.ScanMu.Lock()
		isRunning := state.ScanRunning
		state.ScanMu.Unlock()

		statusStr := "idle"
		if isRunning {
			statusStr = "SEDANG BERJALAN"
		}

		// Check system processes
		procOut, _ := exec.Command("bash", "-c",
			"ps aux | grep -E '(subfinder|nuclei|httpx|katana|naabu|monitor)' | grep -v grep").CombinedOutput()
		procStr := strings.TrimSpace(string(procOut))
		if procStr == "" {
			procStr = "(tidak ada proses recon aktif)"
		}

		return fmt.Sprintf("Scan status (Go internal): %s\nProses recon di sistem:\n%s", statusStr, procStr)

	case "run_command":
		if !IsCommandAllowed(toolInput) {
			return fmt.Sprintf("[BLOCKED] Command tidak diizinkan atau mengandung operator berbahaya: %s", toolInput)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Still using bash -c for now because many recon tools rely on shell features (pipes),
		// but we've improved IsCommandAllowed.
		cmd := exec.CommandContext(ctx, "bash", "-c", toolInput)
		out, err := cmd.CombinedOutput()

		if ctx.Err() == context.DeadlineExceeded {
			return "[ERROR] Command timeout (15s)"
		}

		result := strings.TrimSpace(string(out))
		if len(result) > 2000 {
			result = result[:2000] + "\n...(terpotong)"
		}

		if err != nil && result == "" {
			return fmt.Sprintf("[ERROR] %s", err.Error())
		}
		if result == "" {
			return "[OK] Command selesai, tidak ada output"
		}
		return result

	case "read_file":
		if !IsPathAllowed(toolInput) {
			return fmt.Sprintf("[BLOCKED] Path tidak diizinkan: %s", toolInput)
		}
		content, err := os.ReadFile(toolInput)
		if err != nil {
			return fmt.Sprintf("[ERROR] Tidak bisa baca file: %s", err.Error())
		}
		result := strings.TrimSpace(string(content))
		if len(result) > 3000 {
			result = result[:3000] + "\n...(terpotong)"
		}
		return result

	case "list_dir":
		if !IsPathAllowed(toolInput) {
			return fmt.Sprintf("[BLOCKED] Path tidak diizinkan: %s", toolInput)
		}
		entries, err := os.ReadDir(toolInput)
		if err != nil {
			return fmt.Sprintf("[ERROR] Tidak bisa baca direktori: %s", err.Error())
		}
		var lines []string
		for _, e := range entries {
			info, _ := e.Info()
			size := ""
			if info != nil && !e.IsDir() {
				size = fmt.Sprintf(" (%s)", GetSize(uint64(info.Size())))
			}
			typeChar := "-"
			if e.IsDir() {
				typeChar = "d"
			}
			lines = append(lines, fmt.Sprintf("[%s] %s%s", typeChar, e.Name(), size))
		}
		return fmt.Sprintf("Isi %s (%d item):\n%s", toolInput, len(lines), strings.Join(lines, "\n"))

	case "write_file":
		parts := strings.SplitN(toolInput, "|||", 2)
		if len(parts) != 2 {
			return "[ERROR] Format: path|||content"
		}
		path := strings.TrimSpace(parts[0])
		content := parts[1]

		writableDir := []string{state.DataDir, state.LogsDir, "/tmp"}
		allowed := false
		for _, d := range writableDir {
			if strings.HasPrefix(path, d) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Sprintf("[BLOCKED] Hanya boleh tulis ke: %s", strings.Join(writableDir, ", "))
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Sprintf("[ERROR] Gagal tulis: %s", err.Error())
		}
		return fmt.Sprintf("[OK] File ditulis: %s (%d bytes)", path, len(content))

	case "get_recon_data":
		result := fmt.Sprintf("=== RECON DATA ===\n")
		files := map[string]string{
			"known_subs.txt":      "Subdomains",
			"known_alive.txt":     "Alive Hosts",
			"nuclei_results.json": "Vuln Results",
			"scope.txt":           "Scope",
		}
		for fname, label := range files {
			fpath := filepath.Join(state.DataDir, fname)
			info, err := os.Stat(fpath)
			if err != nil {
				result += fmt.Sprintf("\n[%s] tidak ada\n", label)
				continue
			}
			count := CountLines(fpath)
			result += fmt.Sprintf("\n[%s] %s lines, modified: %s\n",
				label, count, info.ModTime().Format("2006-01-02 15:04"))

			if f, err := os.Open(fpath); err == nil {
				scanner := bufio.NewScanner(f)
				i := 0
				for scanner.Scan() && i < 5 {
					result += "  " + scanner.Text() + "\n"
					i++
				}
				f.Close()
			}
		}
		return result
	}

	return fmt.Sprintf("[ERROR] Tool tidak dikenal: %s", toolName)
}
