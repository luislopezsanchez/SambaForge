package directory

import (
	"fmt"
	"os/exec"
	"strings"
)

// OU represents an Organizational Unit.
type OU struct {
	Name string `json:"name"`
	DN   string `json:"dn"`
}

// ListOUs returns all OUs in the directory.
func ListOUs() ([]OU, error) {
	cmd := exec.Command("samba-tool", "ou", "list", "--color=never")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("samba-tool ou list: %w", err)
	}
	var ous []OU
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "samba-tool") {
			continue
		}
		ous = append(ous, OU{Name: line})
	}
	return ous, nil
}

// CreateOU creates a new OU.
func CreateOU(name string) error {
	cmd := exec.Command("samba-tool", "ou", "add", name, "--color=never")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("samba-tool ou add: %s: %w", string(out), err)
	}
	return nil
}

// DeleteOU deletes an OU.
func DeleteOU(name string) error {
	cmd := exec.Command("samba-tool", "ou", "delete", name, "--color=never")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("samba-tool ou delete: %s: %w", string(out), err)
	}
	return nil
}

// PasswordPolicy holds the domain password policy.
type PasswordPolicy struct {
	Complexity         string `json:"complexity"`
	MinLength           string `json:"minLength"`
	HistoryLength      string `json:"historyLength"`
	MinAge             string `json:"minAge"`
	MaxAge             string `json:"maxAge"`
	LockoutDuration    string `json:"lockoutDuration"`
	LockoutThreshold   string `json:"lockoutThreshold"`
	LockoutWindow      string `json:"lockoutWindow"`
}

// GetPasswordPolicy returns the current domain password policy.
func GetPasswordPolicy() (*PasswordPolicy, error) {
	cmd := exec.Command("samba-tool", "domain", "passwordsettings", "show", "--color=never")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("samba-tool domain passwordsettings show: %w", err)
	}

	policy := &PasswordPolicy{}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "complexity") {
			policy.Complexity = extractValue(line)
		} else if strings.Contains(line, "history length") {
			policy.HistoryLength = extractValue(line)
		} else if strings.Contains(line, "minimum password length") {
			policy.MinLength = extractValue(line)
		} else if strings.Contains(line, "minimum password age") {
			policy.MinAge = extractValue(line)
		} else if strings.Contains(line, "maximum password age") {
			policy.MaxAge = extractValue(line)
		} else if strings.Contains(line, "lockout duration") {
			policy.LockoutDuration = extractValue(line)
		} else if strings.Contains(line, "lockout threshold") {
			policy.LockoutThreshold = extractValue(line)
		} else if strings.Contains(line, "reset account lockout") {
			policy.LockoutWindow = extractValue(line)
		}
	}
	return policy, nil
}

// SetPasswordPolicy updates the domain password policy.
type SetPasswordPolicyRequest struct {
	Complexity       *string `json:"complexity"`
	MinLength         *int    `json:"minLength"`
	HistoryLength     *int    `json:"historyLength"`
	MinAge            *int    `json:"minAge"`
	MaxAge            *int    `json:"maxAge"`
	LockoutDuration   *int    `json:"lockoutDuration"`
	LockoutThreshold  *int    `json:"lockoutThreshold"`
}

func SetPasswordPolicy(req SetPasswordPolicyRequest) error {
	args := []string{"domain", "passwordsettings", "set", "--color=never"}
	if req.Complexity != nil {
		args = append(args, "--complexity="+*req.Complexity)
	}
	if req.MinLength != nil {
		args = append(args, fmt.Sprintf("--min-pwd-length=%d", *req.MinLength))
	}
	if req.HistoryLength != nil {
		args = append(args, fmt.Sprintf("--history-length=%d", *req.HistoryLength))
	}
	if req.MinAge != nil {
		args = append(args, fmt.Sprintf("--min-pwd-age=%d", *req.MinAge))
	}
	if req.MaxAge != nil {
		args = append(args, fmt.Sprintf("--max-pwd-age=%d", *req.MaxAge))
	}
	if req.LockoutDuration != nil {
		args = append(args, fmt.Sprintf("--account-lockout-duration=%d", *req.LockoutDuration))
	}
	if req.LockoutThreshold != nil {
		args = append(args, fmt.Sprintf("--account-lockout-threshold=%d", *req.LockoutThreshold))
	}

	cmd := exec.Command("samba-tool", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("samba-tool domain passwordsettings set: %s: %w", string(out), err)
	}
	return nil
}

func extractValue(line string) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}