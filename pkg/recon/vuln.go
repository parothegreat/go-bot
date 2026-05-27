package recon

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

type VulnEntry struct {
	TemplateID string
	Name       string
	Severity   string
	Host       string
	URL        string
	Time       string
}

func ExtractDomain(rawURL string) string {
	rawURL = strings.TrimPrefix(rawURL, "https://")
	rawURL = strings.TrimPrefix(rawURL, "http://")
	parts := strings.SplitN(rawURL, "/", 2)
	h := parts[0]
	if idx := strings.LastIndex(h, ":"); idx != -1 {
		h = h[:idx]
	}
	return h
}

func SeverityEmoji(sev string) string {
	switch strings.ToUpper(sev) {
	case "CRITICAL":
		return "🔴"
	case "HIGH":
		return "🟠"
	case "MEDIUM":
		return "🟡"
	case "LOW":
		return "🔵"
	default:
		return "⚪"
	}
}

func ParseVulnLine(line string) *VulnEntry {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}
	var nucleiJSON struct {
		TemplateID string `json:"template-id"`
		Info       struct {
			Name     string `json:"name"`
			Severity string `json:"severity"`
		} `json:"info"`
		Host      string `json:"host"`
		MatchedAt string `json:"matched-at"`
		Timestamp string `json:"timestamp"`
	}
	if err := json.Unmarshal([]byte(line), &nucleiJSON); err == nil && nucleiJSON.TemplateID != "" {
		return &VulnEntry{
			TemplateID: nucleiJSON.TemplateID,
			Name:       nucleiJSON.Info.Name,
			Severity:   strings.ToUpper(nucleiJSON.Info.Severity),
			Host:       ExtractDomain(nucleiJSON.MatchedAt),
			URL:        nucleiJSON.MatchedAt,
			Time:       nucleiJSON.Timestamp,
		}
	}

	// Try custom format: [SEVERITY] Name | Host | URL
	if strings.HasPrefix(line, "[") {
		entry := &VulnEntry{}
		parts := strings.Split(line, "|")
		for i, p := range parts {
			p = strings.TrimSpace(p)
			if i == 0 {
				if idx := strings.Index(p, "]"); idx != -1 {
					entry.Severity = strings.ToUpper(strings.Trim(p[1:idx], " "))
					val := strings.TrimSpace(p[idx+1:])
					if val != "" {
						entry.Name = val
					}
				}
			} else if strings.HasPrefix(p, "http") {
				entry.URL = p
				entry.Host = ExtractDomain(p)
			}
		}
		if entry.URL != "" {
			return entry
		}
	}
	return nil
}

func BuildVulnReport(vulnFile string) ([]VulnEntry, string) {
	f, err := os.Open(vulnFile)
	if err != nil {
		return nil, ""
	}
	defer f.Close()

	var entries []VulnEntry
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		if e := ParseVulnLine(scanner.Text()); e != nil {
			entries = append(entries, *e)
		}
	}
	if len(entries) == 0 {
		return nil, ""
	}

	domainMap := map[string][]VulnEntry{}
	for _, e := range entries {
		domainMap[e.Host] = append(domainMap[e.Host], e)
	}

	var domains []string
	for d := range domainMap {
		domains = append(domains, d)
	}
	sort.Strings(domains)

	var sb strings.Builder
	for _, domain := range domains {
		vulns := domainMap[domain]
		sb.WriteString(fmt.Sprintf("━━━ %s (%d findings) ━━━\n", domain, len(vulns)))
		for _, v := range vulns {
			sb.WriteString(fmt.Sprintf(
				"  [%s] %s\n  Template : %s\n  URL      : %s\n  Time     : %s\n\n",
				v.Severity, v.Name, v.TemplateID, v.URL, v.Time,
			))
		}
	}
	return entries, sb.String()
}
