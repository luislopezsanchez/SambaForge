package auth

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

var jwtSecret []byte

func Init() {
	secret := os.Getenv("SAMBAFORGE_JWT_SECRET")
	if secret == "" {
		// Generate a random secret on startup
		b := make([]byte, 32)
		rand.Read(b)
		secret = hex.EncodeToString(b)
	}
	jwtSecret = []byte(secret)
}

// Login authenticates a user against the local Samba AD LDAP.
func Login(username, password, realm string) (*TokenResponse, error) {
	if realm == "" {
		// Detect realm from smb.conf
		realm = detectRealm()
	}
	if realm == "" {
		return nil, fmt.Errorf("no se pudo detectar el realm del dominio")
	}

	// Connect to local LDAP with STARTTLS (Samba AD requires encrypted bind)
	l, err := ldap.DialURL("ldap://127.0.0.1:389")
	if err != nil {
		return nil, fmt.Errorf("conectar LDAP: %w", err)
	}
	defer l.Close()

	// Upgrade to TLS (STARTTLS)
	err = l.StartTLS(&tls.Config{InsecureSkipVerify: true}) // self-signed cert on localhost
	if err != nil {
		return nil, fmt.Errorf("STARTTLS: %w", err)
	}

	// Bind with user credentials (UPN format: user@REALM)
	upn := fmt.Sprintf("%s@%s", username, realm)
	err = l.Bind(upn, password)
	if err != nil {
		return nil, fmt.Errorf("credenciales inválidas: %w", err)
	}

	// Search for the user to get details
	searchReq := ldap.NewSearchRequest(
		fmt.Sprintf("DC=%s", strings.Join(strings.Split(strings.ToLower(realm), "."), ",DC=")),
		ldap.ScopeWholeSubtree, ldap.DerefAlways, 0, 0, false,
		fmt.Sprintf("(sAMAccountName=%s)", ldap.EscapeFilter(username)),
		[]string{"cn", "sAMAccountName", "memberOf", "userAccountControl", "mail"},
		nil,
	)
	result, err := l.Search(searchReq)
	if err != nil || len(result.Entries) == 0 {
		return nil, fmt.Errorf("usuario no encontrado en el directorio")
	}

	entry := result.Entries[0]
	cn := entry.GetAttributeValue("cn")
	mail := entry.GetAttributeValue("mail")

	// Check if user is Domain Admin (memberOf contains Domain Admins)
	isAdmin := false
	for _, attr := range entry.Attributes {
		if attr.Name == "memberOf" {
			for _, val := range attr.Values {
				if strings.Contains(val, "Domain Admins") {
					isAdmin = true
				}
			}
		}
	}

	// Generate JWT
	sessionID := generateSessionID()
	claims := jwt.MapClaims{
		"sub":      username,
		"realm":    realm,
		"cn":       cn,
		"mail":     mail,
		"admin":    isAdmin,
		"exp":      jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
		"iat":      jwt.NewNumericDate(time.Now()),
		"jti":      sessionID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("generar token: %w", err)
	}

	// If admin, also create a Kerberos ticket cache for internal DNS operations
	if isAdmin {
		createKerberosTicket(username, password, realm)
	}

	return &TokenResponse{
		Token:    tokenString,
		Username: username,
		CN:       cn,
		Mail:     mail,
		IsAdmin:  isAdmin,
		Realm:    realm,
	}, nil
}

type TokenResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	CN       string `json:"cn"`
	Mail     string `json:"mail"`
	IsAdmin  bool   `json:"isAdmin"`
	Realm    string `json:"realm"`
}

// Middleware validates JWT and injects user info into context.
func Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Skip auth for login and health
		path := c.Path()
		if path == "/api/auth/login" || path == "/api/health" || path == "/api/server/preflight" || path == "/api/domain/provision" || path == "/api/domain/health" {
			return next(c)
		}

		auth := c.Request().Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "no token"})
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		claims := token.Claims.(jwt.MapClaims)
		c.Set("username", claims["sub"])
		c.Set("realm", claims["realm"])
		c.Set("cn", claims["cn"])
		c.Set("isAdmin", claims["admin"])

		return next(c)
	}
}

func detectRealm() string {
	// Try to read from smb.conf
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

// DetectRealmPublic returns the realm for use by other packages.
func DetectRealmPublic() string {
	return detectRealm()
}

func generateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// createKerberosTicket runs kinit to create a ticket cache for internal samba-tool operations.
func createKerberosTicket(username, password, realm string) {
	// Write password to a temp file and pipe to kinit
	// This avoids passing password on the command line
	cmd := exec.Command("kinit", fmt.Sprintf("%s@%s", username, realm))
	cmd.Stdin = strings.NewReader(password + "\n")
	cmd.Run() // ignore error - best effort
}