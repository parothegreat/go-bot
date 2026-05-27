package tools

import (
	"strings"
)

var allowedPaths = []string{
	"/home/parothegreat",
	"/tmp",
	"/var/log",
	"/etc/hosts",
	"/proc/version",
	"/proc/cpuinfo",
	"/proc/meminfo",
}

// IsPathAllowed checks if the given path is safe for AI access.
func IsPathAllowed(path string) bool {
	// Prevent traversal
	if strings.Contains(path, "..") {
		return false
	}
	// Block sensitive files
	blocked := []string{
		"/etc/shadow", "/etc/passwd", "id_rsa", ".ssh/", 
		"private_key", ".gnupg", "authorized_keys", 
		".netrc", "credentials", ".env",
	}
	for _, b := range blocked {
		if strings.Contains(path, b) {
			return false
		}
	}
	for _, allowed := range allowedPaths {
		if strings.HasPrefix(path, allowed) || path == allowed {
			return true
		}
	}
	return false
}
