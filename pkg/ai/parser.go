package ai

import (
	"strings"
)

// ParseToolCalls extracts tool calls from the AI response.
func ParseToolCalls(response string) (string, []struct{ Name, Input string }) {
	var tools []struct{ Name, Input string }
	var cleanLines []string

	lines := strings.Split(response, "\n")
	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		if strings.HasPrefix(line, "TOOL:") {
			toolName := strings.TrimSpace(strings.TrimPrefix(line, "TOOL:"))
			var inputLines []string
			i++
			for i < len(lines) && strings.TrimSpace(lines[i]) != "END_TOOL" {
				inputLine := lines[i]
				if strings.HasPrefix(strings.TrimSpace(inputLine), "INPUT:") {
					inputLine = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(inputLine), "INPUT:"))
				}
				inputLines = append(inputLines, inputLine)
				i++
			}
			toolInput := strings.TrimSpace(strings.Join(inputLines, "\n"))
			tools = append(tools, struct{ Name, Input string }{toolName, toolInput})
			i++
			continue
		}

		cleanLines = append(cleanLines, lines[i])
		i++
	}

	return strings.TrimSpace(strings.Join(cleanLines, "\n")), tools
}

// ParseAIActions extracts specific bot actions (like SCAN_NOW) from clean text.
func ParseAIActions(text string) (string, []string) {
	var actions []string
	var cleanLines []string

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "SCAN_NOW" {
			actions = append(actions, "SCAN_NOW")
		} else {
			cleanLines = append(cleanLines, line)
		}
	}
	return strings.TrimSpace(strings.Join(cleanLines, "\n")), actions
}
