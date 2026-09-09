package multidc

import (
	"fmt"
	"os/exec"
	"strings"
)

// FSMORoles holds the FSMO role assignments.
type FSMORoles struct {
	Schema       string `json:"schema"`
	Naming       string `json:"naming"`
	PDC          string `json:"pdc"`
	RID          string `json:"rid"`
	Infrastructure string `json:"infrastructure"`
}

// ShowFSMO returns the current FSMO role assignments.
func ShowFSMO() (*FSMORoles, error) {
	cmd := exec.Command("samba-tool", "fsmo", "show", "-P", "--color=never")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("samba-tool fsmo show: %w", err)
	}

	roles := &FSMORoles{}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "SchemaMaster") {
			roles.Schema = extractRole(line)
		} else if strings.Contains(line, "NamingMaster") {
			roles.Naming = extractRole(line)
		} else if strings.Contains(line, "PdcEmulator") {
			roles.PDC = extractRole(line)
		} else if strings.Contains(line, "RidManager") {
			roles.RID = extractRole(line)
		} else if strings.Contains(line, "InfrastructureMaster") {
			roles.Infrastructure = extractRole(line)
		}
	}
	return roles, nil
}

// TransferRole transfers a FSMO role to another DC.
func TransferRole(role, targetDC string) error {
	cmd := exec.Command("samba-tool", "fsmo", "transfer", "--role="+role, "-P", "--color=never")
	if targetDC != "" {
		cmd.Args = append(cmd.Args, "-H", targetDC)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("samba-tool fsmo transfer: %s: %w", string(out), err)
	}
	return nil
}

// JoinDC joins a second DC to the domain.
func JoinDC(realm, adminPass, dcIP string) error {
	args := []string{"domain", "join", realm, "DC", "-UAdministrator%" + adminPass, "--color=never"}
	if dcIP != "" {
		args = append(args, "-H", dcIP)
	}
	cmd := exec.Command("samba-tool", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("samba-tool domain join: %s: %w", string(out), err)
	}
	return nil
}

// DemoteDC demotes the current DC.
func DemoteDC() error {
	cmd := exec.Command("samba-tool", "domain", "demote", "-P", "--color=never")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("samba-tool domain demote: %s: %w", string(out), err)
	}
	return nil
}

// ShowTrusts lists domain trusts.
func ShowTrusts() ([]string, error) {
	cmd := exec.Command("samba-tool", "domain", "trust", "list", "-P", "--color=never")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("samba-tool domain trust list: %w", err)
	}

	var trusts []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "samba-tool") {
			continue
		}
		trusts = append(trusts, line)
	}
	return trusts, nil
}

func extractRole(line string) string {
	parts := strings.Split(line, ":")
	if len(parts) >= 2 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}