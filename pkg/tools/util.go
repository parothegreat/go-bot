package tools

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

func GetSize(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func CountLines(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return "0"
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		count++
	}
	return strconv.Itoa(count)
}
