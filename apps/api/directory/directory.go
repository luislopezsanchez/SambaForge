package directory

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

// User represents an AD user.
type User struct {
	Username     string   `json:"username"`
	CN           string   `json:"cn"`
	DN           string   `json:"dn"`
	Email        string   `json:"email"`
	Enabled      bool     `json:"enabled"`
	Groups       []string `json:"groups"`
	OU           string   `json:"ou"`
	LastLogon    string   `json:"lastLogon"`
	Description  string   `json:"description"`
}

// Group represents an AD group.
type Group struct {
	Name        string   `json:"name"`
	DN          string   `json:"dn"`
	Description string   `json:"description"`
	Members     []string `json:"members"`
	Type        string   `json:"type"`
}

// Computer represents an AD computer account.
type Computer struct {
	Name    string `json:"name"`
	DN      string `json:"dn"`
	OU      string `json:"ou"`
	OS      string `json:"os"`
	Enabled bool   `json:"enabled"`
}

func detectRealm() string {
	data, err := os.ReadFile("/etc/samba/smb.conf")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "realm") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

func baseDN() string {
	realm := detectRealm()
	parts := strings.Split(strings.ToLower(realm), ".")
	return "DC=" + strings.Join(parts, ",DC=")
}

func ldapConnect() (*ldap.Conn, error) {
	l, err := ldap.DialURL("ldap://127.0.0.1:389")
	if err != nil {
		return nil, err
	}
	return l, nil
}

// ListUsers returns all users in the directory.
func ListUsers() ([]User, error) {
	l, err := ldapConnect()
	if err != nil {
		return nil, fmt.Errorf("LDAP connect: %w", err)
	}
	defer l.Close()

	// Anonymous bind (Samba AD allows anonymous search for listing)
	// Actually try simple bind with empty - use SASL or just search
	// For Samba AD, we can search with a simple bind as the machine
	// Let's use samba-tool instead for reliability
	return listUsersViaSambaTool()
}

func listUsersViaSambaTool() ([]User, error) {
	cmd := exec.Command("samba-tool", "user", "list", "--color=never")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("samba-tool user list: %w", err)
	}

	var users []User
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "samba-tool") {
			continue
		}
		users = append(users, User{
			Username: line,
			CN:       line,
		})
	}
	return users, nil
}

// ListGroups returns all groups in the directory.
func ListGroups() ([]Group, error) {
	cmd := exec.Command("samba-tool", "group", "list", "--color=never")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("samba-tool group list: %w", err)
	}

	var groups []Group
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "samba-tool") {
			continue
		}
		groups = append(groups, Group{
			Name: line,
		})
	}
	return groups, nil
}

// ListComputers returns all computer accounts.
func ListComputers() ([]Computer, error) {
	cmd := exec.Command("samba-tool", "computer", "list", "--color=never")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("samba-tool computer list: %w", err)
	}

	var computers []Computer
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "samba-tool") {
			continue
		}
		computers = append(computers, Computer{
			Name: line,
		})
	}
	return computers, nil
}

// CreateUser creates a new user via samba-tool.
type CreateUserRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `json:"email"`
	Description string `json:"description"`
}

func CreateUser(req CreateUserRequest) error {
	args := []string{"user", "create", req.Username, req.Password, "--color=never"}
	if req.FirstName != "" {
		args = append(args, "--given-name="+req.FirstName)
	}
	if req.LastName != "" {
		args = append(args, "--surname="+req.LastName)
	}

	cmd := exec.Command("samba-tool", args...)
	cmd.Env = append(os.Environ())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("samba-tool user create: %s: %w", string(out), err)
	}
	return nil
}

// DeleteUser deletes a user.
func DeleteUser(username string) error {
	cmd := exec.Command("samba-tool", "user", "delete", username, "--color=never")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("samba-tool user delete: %s: %w", string(out), err)
	}
	return nil
}

// SetUserPassword resets a user's password.
func SetUserPassword(username, password string) error {
	cmd := exec.Command("samba-tool", "user", "setpassword", username, "--newpassword="+password, "--color=never")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("samba-tool user setpassword: %s: %w", string(out), err)
	}
	return nil
}

// DisableUser disables a user account.
func DisableUser(username string) error {
	cmd := exec.Command("samba-tool", "user", "disable", username, "--color=never")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("samba-tool user disable: %s: %w", string(out), err)
	}
	return nil
}

// EnableUser enables a user account.
func EnableUser(username string) error {
	cmd := exec.Command("samba-tool", "user", "enable", username, "--color=never")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("samba-tool user enable: %s: %w", string(out), err)
	}
	return nil
}

// CreateGroup creates a new group.
func CreateGroup(name string) error {
	cmd := exec.Command("samba-tool", "group", "add", name, "--color=never")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("samba-tool group add: %s: %w", string(out), err)
	}
	return nil
}

// DeleteGroup deletes a group.
func DeleteGroup(name string) error {
	cmd := exec.Command("samba-tool", "group", "delete", name, "--color=never")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("samba-tool group delete: %s: %w", string(out), err)
	}
	return nil
}

// AddGroupMember adds a user to a group.
func AddGroupMember(group, member string) error {
	cmd := exec.Command("samba-tool", "group", "addmembers", group, member, "--color=never")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("samba-tool group addmembers: %s: %w", string(out), err)
	}
	return nil
}

// RemoveGroupMember removes a user from a group.
func RemoveGroupMember(group, member string) error {
	cmd := exec.Command("samba-tool", "group", "removemembers", group, member, "--color=never")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("samba-tool group removemembers: %s: %w", string(out), err)
	}
	return nil
}

// ListGroupMembers returns members of a group.
func ListGroupMembers(group string) ([]string, error) {
	cmd := exec.Command("samba-tool", "group", "listmembers", group, "--color=never")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("samba-tool group listmembers: %w", err)
	}

	var members []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "samba-tool") {
			continue
		}
		members = append(members, line)
	}
	return members, nil
}

// Count returns quick counts for the dashboard.
type Counts struct {
	Users     int `json:"users"`
	Groups    int `json:"groups"`
	Computers int `json:"computers"`
}

func GetCounts() (*Counts, error) {
	users, _ := ListUsers()
	groups, _ := ListGroups()
	computers, _ := ListComputers()

	return &Counts{
		Users:     len(users),
		Groups:    len(groups),
		Computers: len(computers),
	}, nil
}