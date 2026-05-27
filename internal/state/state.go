package state

import (
	"sync"
	"time"
)

// AIMessage represents a single message in the AI conversation history.
type AIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

var (
	// Directories
	DataDir     = "/home/parothegreat/recon-engine/data"
	LogsDir     = "/home/parothegreat/recon-engine/logs"
	ScopeFile   = "/home/parothegreat/recon-engine/data/scope.txt"
	TargetsFile = "/home/parothegreat/recon-engine/data/targets.txt"
	HomeDir     = "/home/parothegreat"

	// Bot Config & State
	ChatID       int64
	DelayMinutes = 10
	Paused       = false
	LastSent     time.Time

	// Recon State
	ScanRunning bool
	ScanMu      sync.Mutex

	// AI State
	AIMu        sync.Mutex
	AIHistory   []AIMessage
	ScanSummary string

	// Autonomy State
	AutonomousMode bool
	MissionLog     []string

	// Queues
	AIScanQueue chan string
)

func Init() {
	AIScanQueue = make(chan string, 10)
}
