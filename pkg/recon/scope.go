package recon

import (
	"bufio"
	"fmt"
	"os"
	"recon-bot/internal/state"
	"strings"
	"time"
)

// LoadScope loads domains from the scope file.
func LoadScope() []string {
	f, err := os.Open(state.ScopeFile)
	if err != nil {
		return nil
	}
	defer f.Close()
	var domains []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			domains = append(domains, line)
		}
	}
	return domains
}

// SaveScope saves domains to the scope file and targets file.
func SaveScope(domains []string) error {
	var sb strings.Builder
	sb.WriteString("# Scope file — satu domain per baris\n")
	sb.WriteString(fmt.Sprintf("# Updated: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	for _, d := range domains {
		sb.WriteString(d + "\n")
	}
	if err := os.WriteFile(state.ScopeFile, []byte(sb.String()), 0644); err != nil {
		return err
	}
	return os.WriteFile(state.TargetsFile, []byte(strings.Join(domains, "\n")+"\n"), 0644)
}

// GetScopeText returns a formatted string of the current scope.
func GetScopeText() string {
	domains := LoadScope()
	if len(domains) == 0 {
		return "_Belum ada scope. Tambahkan dengan_ `/addscope domain.com`"
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 *Scope saat ini (%d domain):*\n\n", len(domains)))
	for i, d := range domains {
		sb.WriteString(fmt.Sprintf("`%d.` `%s`\n", i+1, d))
	}
	return sb.String()
}

// RebuildScopeInSummary updates the scope line in the scan summary.
func RebuildScopeInSummary(domains []string) string {
	state.AIMu.Lock()
	prev := state.ScanSummary
	state.AIMu.Unlock()

	scopeLine := fmt.Sprintf("[SCOPE: %s]", strings.Join(domains, ", "))
	if prev == "" {
		return scopeLine
	}
	var filtered []string
	for _, l := range strings.Split(prev, "\n") {
		if !strings.HasPrefix(l, "[SCOPE:") {
			filtered = append(filtered, l)
		}
	}
	return scopeLine + "\n" + strings.Join(filtered, "\n")
}
