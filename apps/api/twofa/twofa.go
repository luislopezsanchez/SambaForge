package twofa

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/pquerna/otp/totp"
)

// Secret stores a user's TOTP secret.
type Secret struct {
	Username string `json:"username"`
	Secret   string `json:"secret"`
	Enabled  bool   `json:"enabled"`
}

// GenerateSecret generates a new TOTP secret for a user.
func GenerateSecret(username, realm string) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "SambaForge",
		AccountName: fmt.Sprintf("%s@%s", username, realm),
		Period:      30,
		Digits:      6,
	})
	if err != nil {
		return "", fmt.Errorf("generate TOTP: %w", err)
	}

	// Store secret in file (simple approach for now, SQLite later)
	secretsDir := "/etc/sambaforge/totp"
	os.MkdirAll(secretsDir, 0700)

	// Use hex-encoded filename to avoid issues
	filename := hex.EncodeToString([]byte(username)) + ".secret"
	path := strings.Join([]string{secretsDir, filename}, "/")

	secret := key.Secret()
	err = os.WriteFile(path, []byte(secret), 0600)
	if err != nil {
		return "", fmt.Errorf("store TOTP secret: %w", err)
	}

	return key.URL(), nil
}

// ValidateCode validates a TOTP code for a user.
func ValidateCode(username, code string) (bool, error) {
	secretsDir := "/etc/sambaforge/totp"
	filename := hex.EncodeToString([]byte(username)) + ".secret"
	path := strings.Join([]string{secretsDir, filename}, "/")

	data, err := os.ReadFile(path)
	if err != nil {
		// No secret file = 2FA not enabled for this user
		return true, nil // Allow login if 2FA not set up
	}

	secret := strings.TrimSpace(string(data))
	valid := totp.Validate(code, secret)
	return valid, nil
}

// Has2FA checks if a user has 2FA enabled.
func Has2FA(username string) bool {
	secretsDir := "/etc/sambaforge/totp"
	filename := hex.EncodeToString([]byte(username)) + ".secret"
	path := strings.Join([]string{secretsDir, filename}, "/")

	_, err := os.Stat(path)
	return err == nil
}

// Disable2FA removes the TOTP secret for a user.
func Disable2FA(username string) error {
	secretsDir := "/etc/sambaforge/totp"
	filename := hex.EncodeToString([]byte(username)) + ".secret"
	path := strings.Join([]string{secretsDir, filename}, "/")
	return os.Remove(path)
}

// GenerateBackupCodes generates 10 backup codes.
func GenerateBackupCodes() []string {
	codes := make([]string, 10)
	for i := range codes {
		b := make([]byte, 8)
		rand.Read(b)
		codes[i] = base32.StdEncoding.EncodeToString(b)[:10]
	}
	return codes
}