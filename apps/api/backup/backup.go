package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// BackupResult holds the result of a backup operation.
type BackupResult struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	File     string `json:"file"`
	Size     int64  `json:"size"`
	Error    string `json:"error"`
}

// BackupDomain performs an online backup of the Samba AD DC.
func BackupDomain(ctx context.Context, destDir string, logWriter io.Writer) (*BackupResult, error) {
	if destDir == "" {
		destDir = "/var/backups/sambaforge"
	}
	os.MkdirAll(destDir, 0755)

	timestamp := time.Now().Format("2006-01-02_150405")
	backupFile := filepath.Join(destDir, fmt.Sprintf("samba-ad-backup-%s.tar", timestamp))

	fmt.Fprintf(logWriter, "[backup] Starting online backup to %s...\n", backupFile)

	cmd := exec.CommandContext(ctx, "samba-tool", "domain", "backup", "online", "--targetdir="+destDir, "-P", "--color=never")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &BackupResult{
			Success: false,
			Message: "Backup failed",
			Error:   fmt.Sprintf("%s: %v", string(out), err),
		}, nil
	}

	// Find the backup file (samba-tool names it automatically)
	matches, _ := filepath.Glob(filepath.Join(destDir, "*.tar"))
	var latestFile string
	var latestSize int64
	for _, m := range matches {
		info, err := os.Stat(m)
		if err == nil && info.Size() > latestSize {
			latestFile = m
			latestSize = info.Size()
		}
	}

	if latestFile == "" {
		// Use the expected filename
		latestFile = backupFile
	}

	fmt.Fprintf(logWriter, "[backup] Backup completed: %s (%d bytes)\n", latestFile, latestSize)

	return &BackupResult{
		Success: true,
		Message: "Backup completed successfully",
		File:    latestFile,
		Size:    latestSize,
	}, nil
}

// ListBackups returns a list of existing backups.
func ListBackups(destDir string) ([]map[string]interface{}, error) {
	if destDir == "" {
		destDir = "/var/backups/sambaforge"
	}
	matches, _ := filepath.Glob(filepath.Join(destDir, "*.tar"))
	var backups []map[string]interface{}
	for _, m := range matches {
		info, err := os.Stat(m)
		if err != nil {
			continue
		}
		backups = append(backups, map[string]interface{}{
			"file":      filepath.Base(m),
			"path":      m,
			"size":      info.Size(),
			"sizeHuman": humanSize(info.Size()),
			"date":      info.ModTime().Format(time.RFC3339),
		})
	}
	return backups, nil
}

// RestoreDomain restores a backup file.
func RestoreDomain(ctx context.Context, backupFile, newDomain string, logWriter io.Writer) (*BackupResult, error) {
	fmt.Fprintf(logWriter, "[restore] Starting restore from %s...\n", backupFile)

	args := []string{"domain", "backup", "restore", "--backup-file=" + backupFile, "-P", "--color=never"}
	if newDomain != "" {
		args = append(args, "--new-domain-name="+newDomain)
	}

	cmd := exec.CommandContext(ctx, "samba-tool", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &BackupResult{
			Success: false,
			Message: "Restore failed",
			Error:   fmt.Sprintf("%s: %v", string(out), err),
		}, nil
	}

	fmt.Fprintf(logWriter, "[restore] Restore completed\n")
	return &BackupResult{
		Success: true,
		Message: "Restore completed successfully",
		File:    backupFile,
	}, nil
}

// DeleteBackup removes a backup file.
func DeleteBackup(filePath string) error {
	return os.Remove(filePath)
}

func humanSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}