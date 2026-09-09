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
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "samba-tool") || strings.Contains(line, "pszZoneName") {
			continue
		}
		// Parse zone name from output
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			zones = append(zones, Zone{Name: parts[0]})
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
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "samba-tool") {
			continue
		}
		// Try to parse: Name Type Data
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			records = append(records, Record{
				Zone: zone,
				Name: parts[0],
				Type: parts[1],
				Data: strings.Join(parts[2:], " "),
			})
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
	cmd := exec.Command("samba-tool", "dns", "add", server, req.Zone, req.Name, req.Type, req.Data, "--color=never")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("samba-tool dns add: %s: %w", string(out), err)
	}
	return nil
}

// DeleteRecord deletes a DNS record.
func DeleteRecord(zone, name, rtype, data string) error {
	server := "127.0.0.1"
	cmd := exec.Command("samba-tool", "dns", "delete", server, zone, name, rtype, data, "--color=never")
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