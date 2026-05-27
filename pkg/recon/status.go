package recon

import (
	"fmt"
	"net"
	"path/filepath"
	"recon-bot/internal/state"
	"recon-bot/pkg/tools"
	"sort"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	netutil "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

func GetStatusReport() string {
	v, _ := mem.VirtualMemory()
	c, _ := cpu.Percent(0, false)
	l, _ := load.Avg()
	h, _ := host.Info()
	d, _ := disk.Usage("/")
	n, _ := netutil.IOCounters(false)

	state.ScanMu.Lock()
	isRunning := state.ScanRunning
	state.ScanMu.Unlock()

	scanStatus := "💤 idle"
	if isRunning {
		scanStatus = "🔄 sedang berjalan"
	}

	// Net Stats
	var netSent, netRecv uint64
	if len(n) > 0 {
		netSent = n[0].BytesSent
		netRecv = n[0].BytesRecv
	}

	// Services Status
	services := []string{"cloudflared", "nginx", "apache2", "ssh", "tor", "zerotier-one"}
	svcLabels := map[string]string{
		"cloudflared":  "Cloudflared",
		"nginx":        "Nginx",
		"apache2":      "Apache",
		"ssh":          "SSH",
		"tor":          "Tor",
		"zerotier-one": "ZeroTier",
	}
	var svcResults []string
	for _, s := range services {
		status := "❌"
		if isServiceRunning(s) {
			status = "✅"
		}
		svcResults = append(svcResults, fmt.Sprintf("%s:%s", svcLabels[s], status))
	}

	// Ports Status
	ports := []int{22, 80, 443, 8080, 9050, 20241}
	var portResults []string
	for _, p := range ports {
		status := "❌"
		if isPortOpen(p) {
			status = "✅"
		}
		portResults = append(portResults, fmt.Sprintf("%d:%s", p, status))
	}

	// Top Processes
	topProcs := getTopProcesses(3)

	// Recon Stats
	subs := tools.CountLines(filepath.Join(state.DataDir, "known_subs.txt"))
	alive := tools.CountLines(filepath.Join(state.DataDir, "known_alive.txt"))
	scope := LoadScope()
	scopeStr := strings.Join(scope, ", ")
	if scopeStr == "" {
		scopeStr = "(kosong)"
	}

	return fmt.Sprintf("📡 *NODEBUNTU ADVANCED MONITOR*\n"+
		"🕒 %s\n"+
		"⏱ Uptime: %s\n"+
		"🐧 OS: %s %s\n\n"+
		"💻 CPU: %s %.1f%%\n"+
		"📈 Load: %.2f %.2f %.2f\n\n"+
		"🧠 RAM: %s %.1f%%\n"+
		"📊 %s / %s\n\n"+
		"💽 DISK: %s %.1f%%\n"+
		"📊 %s / %s\n\n"+
		"🌐 NET: ⬆ %s ⬇ %s\n\n"+
		"🛠 SERVICES:\n%s\n\n"+
		"🧱 PORTS:\n%s\n\n"+
		"🔝 TOP PROCESSES:\n%s\n\n"+
		"🎯 SCOPE: %s\n"+
		"🛡 RECON: %s Subs | %s Alive\n"+
		"⚙️ SCAN: %s",
		time.Now().Format("2006-01-02 15:04:05"),
		formatUptime(h.Uptime),
		h.OS, h.KernelVersion,
		GetBar(c[0]), c[0],
		l.Load1, l.Load5, l.Load15,
		GetBar(v.UsedPercent), v.UsedPercent, tools.GetSize(v.Used), tools.GetSize(v.Total),
		GetBar(d.UsedPercent), d.UsedPercent, tools.GetSize(d.Used), tools.GetSize(d.Total),
		tools.GetSize(netSent), tools.GetSize(netRecv),
		strings.Join(svcResults, "  "),
		strings.Join(portResults, "  "),
		strings.Join(topProcs, "\n"),
		scopeStr, subs, alive, scanStatus)
}

func GetBar(percent float64) string {
	filled := int(percent / 10)
	if filled > 10 {
		filled = 10
	}
	if filled < 0 {
		filled = 0
	}
	return "[" + strings.Repeat("■", filled) + strings.Repeat("□", 10-filled) + "]"
}

func formatUptime(u uint64) string {
	h := u / 3600
	m := (u % 3600) / 60
	s := u % 60
	return fmt.Sprintf("%dh %dm %ds", h, m, s)
}

func isServiceRunning(name string) bool {
	procs, _ := process.Processes()
	for _, p := range procs {
		n, _ := p.Name()
		if strings.Contains(strings.ToLower(n), name) {
			return true
		}
	}
	return false
}

func isPortOpen(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 500*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func getTopProcesses(count int) []string {
	procs, _ := process.Processes()
	type procStat struct {
		name string
		cpu  float64
	}
	var stats []procStat
	for _, p := range procs {
		n, _ := p.Name()
		c, _ := p.CPUPercent()
		if n != "" && n != "go" {
			stats = append(stats, procStat{n, c})
		}
	}
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].cpu > stats[j].cpu
	})

	var result []string
	for i := 0; i < count && i < len(stats); i++ {
		result = append(result, fmt.Sprintf("• %s (%.1f%%)", stats[i].name, stats[i].cpu))
	}
	return result
}
