package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Entry represents an audit log entry.
type Entry struct {
	Timestamp string `json:"timestamp"`
	User      string `json:"user"`
	Action    string `json:"action"`
	Resource  string `json:"resource"`
	Detail    string `json:"detail"`
	IP        string `json:"ip"`
}

var (
	auditFile string
	once     sync.Once
)

func Init() {
	once.Do(func() {
		auditFile = "/var/log/sambaforge-audit.log"
		os.MkdirAll(filepath.Dir(auditFile), 0755)
		// Create if not exists
		f, err := os.OpenFile(auditFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err == nil {
			f.Close()
		}
	})
}

// Log writes an audit entry to the append-only log file.
func Log(user, action, resource, detail, ip string) {
	Init()
	entry := Entry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		User:      user,
		Action:    action,
		Resource:  resource,
		Detail:    detail,
		IP:        ip,
	}
	// Append-only: CSV format for easy parsing and export
	line := fmt.Sprintf("%s,%s,%s,%s,%s,%s\n",
		entry.Timestamp,
		escapeCSV(entry.User),
		escapeCSV(entry.Action),
		escapeCSV(entry.Resource),
		escapeCSV(entry.Detail),
		entry.IP,
	)
	f, err := os.OpenFile(auditFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err == nil {
		f.WriteString(line)
		f.Close()
	}
}

// List returns recent audit entries.
func List(limit int) ([]Entry, error) {
	Init()
	data, err := os.ReadFile(auditFile)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	// Parse from end (most recent first)
	var entries []Entry
	count := 0
	for i := len(lines) - 1; i >= 0 && count < limit; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		parts := splitCSV(line)
		if len(parts) >= 6 {
			entries = append(entries, Entry{
				Timestamp: parts[0],
				User:      parts[1],
				Action:    parts[2],
				Resource:  parts[3],
				Detail:    parts[4],
				IP:        parts[5],
			})
			count++
		}
	}
	return entries, nil
}

func escapeCSV(s string) string {
	if strings.Contains(s, ",") || strings.Contains(s, "\"") {
		return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
	}
	return s
}

func splitCSV(line string) []string {
	// Simple CSV parser - handles quoted fields
	var fields []string
	var current strings.Builder
	inQuotes := false
	for _, ch := range line {
		switch ch {
		case '"':
			inQuotes = !inQuotes
		case ',':
			if !inQuotes {
				fields = append(fields, current.String())
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}
	fields = append(fields, current.String())
	return fields
}