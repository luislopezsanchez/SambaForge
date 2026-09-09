package dns

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Zone represents a DNS zone.
type Zone struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Records int    `json:"records"`
}

// Record represents a DNS record.
type Record struct {
	Zone string `json:"zone"`
	Name string `json:"name"`
	Type string `json:"type"`
	Data string `json:"data"`
}

// ListZones returns all DNS zones.
func ListZones() ([]Zone, error) {
	server := "127.0.0.1"
	cmd := exec.Command("samba-tool", "dns", "zonelist", server, "-P", "--color=never")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("samba-tool dns zonelist: %w", err)
	}

	var zones []Zone
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "pszZoneName") {
			name := strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			zones = append(zones, Zone{Name: name, Type: "primary"})
		}
	}
	return zones, nil
}

// QueryRecords returns DNS records for a zone.
func QueryRecords(zone, name string) ([]Record, error) {
	server := "127.0.0.1"
	args := []string{"dns", "query", server, zone, name, "ALL", "-P", "--color=never"}
	cmd := exec.Command("samba-tool", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("samba-tool dns query: %w", err)
	}

	var records []Record
	var currentName string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "samba-tool") {
			continue
		}
		// Parse "Name=xyz, Records=N" header lines
		if strings.HasPrefix(line, "Name=") {
			parts := strings.SplitN(line, ",", 2)
			currentName = strings.TrimPrefix(parts[0], "Name=")
			continue
		}
		// Parse record lines: "A: 172.30.36.115 (flags=...)" or "CNAME: target." etc
		for _, rtype := range []string{"A:", "AAAA:", "CNAME:", "NS:", "MX:", "SRV:", "TXT:", "PTR:", "SOA:"} {
			if strings.HasPrefix(line, rtype) {
				data := strings.TrimSpace(strings.TrimPrefix(line, rtype))
				// Extract just the data part before (flags=...
				if idx := strings.Index(data, " (flags="); idx > 0 {
					data = strings.TrimSpace(data[:idx])
				}
				records = append(records, Record{
					Zone: zone,
					Name: currentName,
					Type: strings.TrimSuffix(rtype, ":"),
					Data: data,
				})
				break
			}
		}
	}
	return records, nil
}

// AddRecord adds a DNS record.
type AddRecordRequest struct {
	Zone string `json:"zone"`
	Name string `json:"name"`
	Type string `json:"type"` // A, AAAA, CNAME, MX, SRV, TXT, PTR, NS
	Data string `json:"data"`
}

func AddRecord(req AddRecordRequest) error {
	server := "127.0.0.1"
	// Use Kerberos auth - kinit must have been done previously
	// Set KRB5CCNAME from a cached ticket
	cmd := exec.Command("samba-tool", "dns", "add", server, req.Zone, req.Name, req.Type, req.Data, "-P", "--color=never")
	cmd.Env = append(os.Environ())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("samba-tool dns add: %s: %w", string(out), err)
	}
	return nil
}

// DeleteRecord deletes a DNS record.
func DeleteRecord(zone, name, rtype, data string) error {
	server := "127.0.0.1"
	cmd := exec.Command("samba-tool", "dns", "delete", server, zone, name, rtype, data, "-P", "--color=never")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("samba-tool dns delete: %s: %w", string(out), err)
	}
	return nil
}

// GetForwarders returns the DNS forwarders from smb.conf.
func GetForwarders() ([]string, error) {
	data, err := os.ReadFile("/etc/samba/smb.conf")
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "dns forwarder") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				forwarders := strings.Split(strings.TrimSpace(parts[1]), ",")
				for i := range forwarders {
					forwarders[i] = strings.TrimSpace(forwarders[i])
				}
				return forwarders, nil
			}
		}
	}
	return nil, nil
}