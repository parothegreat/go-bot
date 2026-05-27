package tools

import (
	"strings"
)

var allowedCommands = []string{
	"ls", "cat", "grep", "ps", "df", "free", "uptime",
	"wc", "head", "tail", "pwd", "whoami", "uname", "hostname",
	"netstat", "ss", "ip", "dig", "nslookup", "lsof",
	"subfinder", "httpx", "nuclei", "katana", "naabu",
}

// IsCommandAllowed checks if the command is in the allowlist and doesn't contain dangerous patterns.
func IsCommandAllowed(cmd string) bool {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return false
	}

	// Block destructive commands and shells
	blocked := []string{
		"rm -rf", "mkfs", "dd if=", "> /dev/", "chmod 777 /", "chown root",
		"shutdown", "reboot", "halt", "init 0", ":(){ :|:& };:", "curl | sh",
		"wget | sh", "bash -i", "nc -e", "/bin/sh -i", "python -c", "perl -e",
		"sudo", "apt", "dpkg",
	}
	lower := strings.ToLower(cmd)
	for _, b := range blocked {
		if strings.Contains(lower, b) {
			return false
		}
	}

	// Check if it starts with an allowed command
	found := false
	for _, allowed := range allowedCommands {
		if strings.HasPrefix(cmd, allowed+" ") || cmd == allowed {
			found = true
			break
		}
	}
	if !found {
		return false
	}

	// Block dangerous operators but allow simple pipes for filtering
	// We allow '| grep' as a special case if needed, but in general we want to be strict.
	blockedOps := []string{">", "<", "`", "$(", "&&", "||", ";", "&"}
	for _, op := range blockedOps {
		if strings.Contains(cmd, op) {
			// Allow '| grep' or '| head' or '| tail' or '| wc'
			if op == "|" {
				// Simple heuristic for safe pipes
				safePipe := false
				safeTools := []string{"grep", "head", "tail", "wc", "sort", "uniq"}
				for _, st := range safeTools {
					if strings.Contains(cmd, "| "+st) || strings.Contains(cmd, "|"+st) {
						safePipe = true
						break
					}
				}
				if !safePipe {
					return false
				}
			} else {
				return false
			}
		}
	}

	return true
}
