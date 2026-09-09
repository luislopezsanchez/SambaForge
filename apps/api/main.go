package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/luislopezsanchez/SambaForge/audit"
	"github.com/luislopezsanchez/SambaForge/auth"
	"github.com/luislopezsanchez/SambaForge/backup"
	"github.com/luislopezsanchez/SambaForge/directory"
	"github.com/luislopezsanchez/SambaForge/dns"
	"github.com/luislopezsanchez/SambaForge/gpo"
	"github.com/luislopezsanchez/SambaForge/multidc"
	"github.com/luislopezsanchez/SambaForge/preflight"
	"github.com/luislopezsanchez/SambaForge/provision"
	"github.com/luislopezsanchez/SambaForge/twofa"
)

var version = "1.0.0"

func main() {
	auth.Init()
	audit.Init()

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))

	// Rate limit login specifically (anti brute-force)
	loginLimiter := middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{Rate: 5, Burst: 5},
	)

	// Public routes (no auth)
	e.GET("/api/health", healthHandler)
	e.POST("/api/auth/login", loginHandler, middleware.RateLimiter(loginLimiter))
	e.GET("/api/server/preflight", preflightHandler)
	e.POST("/api/server/preflight/fix/:id", preflightFixHandler)
	e.POST("/api/domain/provision", provisionHandler)
	e.GET("/api/domain/provision/stream", provisionStreamHandler)
	e.GET("/api/domain/health", domainHealthHandler)

	// Protected routes (require JWT)
	api := e.Group("/api", auth.Middleware)
	api.GET("/dashboard", dashboardHandler)
	api.GET("/users", listUsersHandler)
	api.POST("/users", createUserHandler)
	api.DELETE("/users/:username", deleteUserHandler)
	api.POST("/users/:username/password", setUserPasswordHandler)
	api.POST("/users/:username/disable", disableUserHandler)
	api.POST("/users/:username/enable", enableUserHandler)
	api.GET("/groups", listGroupsHandler)
	api.POST("/groups", createGroupHandler)
	api.DELETE("/groups/:name", deleteGroupHandler)
	api.GET("/groups/:name/members", listGroupMembersHandler)
	api.POST("/groups/:name/members", addGroupMemberHandler)
	api.DELETE("/groups/:name/members/:member", removeGroupMemberHandler)
	api.GET("/computers", listComputersHandler)
	api.GET("/ous", listOUsHandler)
	api.POST("/ous", createOUHandler)
	api.DELETE("/ous/:name", deleteOUHandler)
	api.GET("/password-policy", getPasswordPolicyHandler)
	api.PUT("/password-policy", setPasswordPolicyHandler)
	api.GET("/dns/zones", listDnsZonesHandler)
	api.GET("/dns/zones/:zone/records", listDnsRecordsHandler)
	api.POST("/dns/records", addDnsRecordHandler)
	api.DELETE("/dns/records", deleteDnsRecordHandler)
	api.GET("/dns/forwarders", getDnsForwardersHandler)
	api.GET("/gpos", listGPOsHandler)
	api.POST("/gpos", createGPOHandler)
	api.DELETE("/gpos/:id", deleteGPOHandler)
	api.GET("/gpos/templates", listGPOTemplatesHandler)
	api.GET("/audit", listAuditHandler)
	api.POST("/backup", backupHandler)
	api.GET("/backups", listBackupsHandler)
	api.DELETE("/backups", deleteBackupHandler)
	api.GET("/2fa/status/:username", get2FAStatusHandler)
	api.POST("/2fa/setup", setup2FAHandler)
	api.POST("/2fa/verify", verify2FAHandler)
	api.DELETE("/2fa/:username", disable2FAHandler)
	api.GET("/fsmo", fsmoShowHandler)
	api.POST("/fsmo/transfer", fsmoTransferHandler)
	api.GET("/trusts", listTrustsHandler)
	api.POST("/auth/change-password", changeOwnPasswordHandler)

	// Serve frontend
	webDir := os.Getenv("SAMBAFORGE_WEB_DIR")
	if webDir == "" {
		webDir = "/opt/sambaforge/apps/web/dist"
	}
	if _, err := os.Stat(webDir); err == nil {
		e.Use(middleware.Static(webDir))
		e.GET("/*", func(c echo.Context) error {
			path := filepath.Join(webDir, c.Param("*"))
			if _, err := os.Stat(path); os.IsNotExist(err) {
				return c.File(filepath.Join(webDir, "index.html"))
			}
			return c.File(path)
		})
	}

	port := os.Getenv("SAMBAFORGE_PORT")
	if port == "" {
		port = "8444"
	}
	e.Logger.Fatal(e.Start(":" + port))
}

// --- Health ---

func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"service": "SambaForge",
		"version": version,
	})
}

// --- Auth ---

func loginHandler(c echo.Context) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Realm    string `json:"realm"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	resp, err := auth.Login(req.Username, req.Password, req.Realm)
	if err != nil {
		audit.Log(req.Username, "login_failed", "auth", err.Error(), c.RealIP())
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	audit.Log(req.Username, "login", "auth", "success", c.RealIP())
	return c.JSON(http.StatusOK, resp)
}

// --- Dashboard ---

func dashboardHandler(c echo.Context) error {
	counts, err := directory.GetCounts()
	if err != nil {
		counts = &directory.Counts{}
	}

	// Get domain info
	domainLevel := ""
	if out, err := exec.Command("samba-tool", "domain", "level", "show").Output(); err == nil {
		domainLevel = string(out)
	}

	// Get Samba version
	sambaVersion := ""
	if out, err := exec.Command("samba-tool", "--version").Output(); err == nil {
		sambaVersion = strings.TrimSpace(string(out))
	}

	// Check samba service
	sambaActive := false
	if out, _ := exec.Command("systemctl", "is-active", "samba-ad-dc").Output(); strings.TrimSpace(string(out)) == "active" {
		sambaActive = true
	}

	// Get DNS forwarders
	forwarders, _ := dns.GetForwarders()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"counts":        counts,
		"sambaVersion":  sambaVersion,
		"sambaActive":   sambaActive,
		"domainLevel":   domainLevel,
		"forwarders":    forwarders,
		"realm":         auth.DetectRealmPublic(),
	})
}

// --- Users ---

func listUsersHandler(c echo.Context) error {
	users, err := directory.ListUsers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, users)
}

func createUserHandler(c echo.Context) error {
	var req directory.CreateUserRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := directory.CreateUser(req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	audit.Log(c.Get("username").(string), "create", "user", req.Username, c.RealIP())
	return c.JSON(http.StatusCreated, map[string]string{"status": "created", "username": req.Username})
}

func deleteUserHandler(c echo.Context) error {
	username := c.Param("username")
	if err := directory.DeleteUser(username); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	audit.Log(c.Get("username").(string), "delete", "user", username, c.RealIP())
	return c.JSON(http.StatusOK, map[string]string{"status": "deleted", "username": username})
}

func setUserPasswordHandler(c echo.Context) error {
	username := c.Param("username")
	var req struct{ Password string `json:"password"` }
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := directory.SetUserPassword(username, req.Password); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "password_changed", "username": username})
}

func disableUserHandler(c echo.Context) error {
	username := c.Param("username")
	if err := directory.DisableUser(username); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "disabled", "username": username})
}

func enableUserHandler(c echo.Context) error {
	username := c.Param("username")
	if err := directory.EnableUser(username); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "enabled", "username": username})
}

// --- Groups ---

func listGroupsHandler(c echo.Context) error {
	groups, err := directory.ListGroups()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, groups)
}

func createGroupHandler(c echo.Context) error {
	var req struct{ Name string `json:"name"` }
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := directory.CreateGroup(req.Name); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]string{"status": "created", "name": req.Name})
}

func deleteGroupHandler(c echo.Context) error {
	name := c.Param("name")
	if err := directory.DeleteGroup(name); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "deleted", "name": name})
}

func listGroupMembersHandler(c echo.Context) error {
	name := c.Param("name")
	members, err := directory.ListGroupMembers(name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, members)
}

func addGroupMemberHandler(c echo.Context) error {
	group := c.Param("name")
	var req struct{ Member string `json:"member"` }
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := directory.AddGroupMember(group, req.Member); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "member_added"})
}

func removeGroupMemberHandler(c echo.Context) error {
	group := c.Param("name")
	member := c.Param("member")
	if err := directory.RemoveGroupMember(group, member); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "member_removed"})
}

// --- Computers ---

func listComputersHandler(c echo.Context) error {
	computers, err := directory.ListComputers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, computers)
}

// --- OUs ---

func listOUsHandler(c echo.Context) error {
	ous, err := directory.ListOUs()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, ous)
}

func createOUHandler(c echo.Context) error {
	var req struct{ Name string `json:"name"` }
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := directory.CreateOU(req.Name); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]string{"status": "created", "name": req.Name})
}

func deleteOUHandler(c echo.Context) error {
	name := c.Param("name")
	if err := directory.DeleteOU(name); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "deleted", "name": name})
}

// --- Password Policy ---

func getPasswordPolicyHandler(c echo.Context) error {
	policy, err := directory.GetPasswordPolicy()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, policy)
}

func setPasswordPolicyHandler(c echo.Context) error {
	var req directory.SetPasswordPolicyRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := directory.SetPasswordPolicy(req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "updated"})
}

// --- DNS ---

func listDnsZonesHandler(c echo.Context) error {
	zones, err := dns.ListZones()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, zones)
}

func listDnsRecordsHandler(c echo.Context) error {
	zone := c.Param("zone")
	name := c.QueryParam("name")
	if name == "" {
		name = "@"
	}
	records, err := dns.QueryRecords(zone, name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, records)
}

func addDnsRecordHandler(c echo.Context) error {
	var req dns.AddRecordRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := dns.AddRecord(req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]string{"status": "created"})
}

func deleteDnsRecordHandler(c echo.Context) error {
	var req struct {
		Zone string `json:"zone"`
		Name string `json:"name"`
		Type string `json:"type"`
		Data string `json:"data"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := dns.DeleteRecord(req.Zone, req.Name, req.Type, req.Data); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
}

func getDnsForwardersHandler(c echo.Context) error {
	forwarders, err := dns.GetForwarders()
	if err != nil {
		return c.JSON(http.StatusOK, []string{})
	}
	return c.JSON(http.StatusOK, forwarders)
}

// --- GPO ---

func listGPOsHandler(c echo.Context) error {
	gpos, err := gpo.ListGPOs()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, gpos)
}

func createGPOHandler(c echo.Context) error {
	var req struct{ Name string `json:"name"` }
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := gpo.CreateGPO(req.Name); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	audit.Log(c.Get("username").(string), "create", "gpo", req.Name, c.RealIP())
	return c.JSON(http.StatusCreated, map[string]string{"status": "created", "name": req.Name})
}

func deleteGPOHandler(c echo.Context) error {
	id := c.Param("id")
	if err := gpo.DeleteGPO(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	audit.Log(c.Get("username").(string), "delete", "gpo", id, c.RealIP())
	return c.JSON(http.StatusOK, map[string]string{"status": "deleted", "id": id})
}

func listGPOTemplatesHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, gpo.ListTemplates())
}

// --- Audit ---

func listAuditHandler(c echo.Context) error {
	entries, err := audit.List(100)
	if err != nil {
		return c.JSON(http.StatusOK, []audit.Entry{})
	}
	return c.JSON(http.StatusOK, entries)
}

// --- Backup ---

func backupHandler(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Minute)
	defer cancel()
	var logBuf bytesBuffer
	result, err := backup.BackupDomain(ctx, "", &logBuf)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	audit.Log(c.Get("username").(string), "backup", "domain", result.File, c.RealIP())
	return c.JSON(http.StatusOK, result)
}

func listBackupsHandler(c echo.Context) error {
	backups, err := backup.ListBackups("")
	if err != nil {
		return c.JSON(http.StatusOK, []map[string]interface{}{})
	}
	return c.JSON(http.StatusOK, backups)
}

func deleteBackupHandler(c echo.Context) error {
	var req struct{ Path string `json:"path"` }
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := backup.DeleteBackup(req.Path); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	audit.Log(c.Get("username").(string), "delete", "backup", req.Path, c.RealIP())
	return c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Change own password (self-service) ---

func changeOwnPasswordHandler(c echo.Context) error {
	username := c.Get("username").(string)
	var req struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if req.NewPassword == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "newPassword es obligatorio"})
	}

	// Verify old password via LDAP bind
	realm := auth.DetectRealmPublic()
	if err := auth.VerifyCredentials(username, req.OldPassword, realm); err != nil {
		audit.Log(username, "password_change_failed", "self-service", err.Error(), c.RealIP())
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "contraseña actual incorrecta"})
	}

	// Change password via samba-tool
	if err := directory.SetUserPassword(username, req.NewPassword); err != nil {
		audit.Log(username, "password_change_failed", "self-service", err.Error(), c.RealIP())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	audit.Log(username, "password_change", "self-service", "success", c.RealIP())
	return c.JSON(http.StatusOK, map[string]string{"status": "password_changed"})
}

// --- 2FA TOTP ---

func get2FAStatusHandler(c echo.Context) error {
	username := c.Param("username")
	enabled := twofa.Has2FA(username)
	return c.JSON(http.StatusOK, map[string]bool{"enabled": enabled})
}

func setup2FAHandler(c echo.Context) error {
	var req struct {
		Username string `json:"username"`
		Realm    string `json:"realm"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if req.Realm == "" {
		req.Realm = auth.DetectRealmPublic()
	}
	url, err := twofa.GenerateSecret(req.Username, req.Realm)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	audit.Log(c.Get("username").(string), "setup", "2fa", req.Username, c.RealIP())
	return c.JSON(http.StatusOK, map[string]string{"qrUrl": url})
}

func verify2FAHandler(c echo.Context) error {
	var req struct {
		Username string `json:"username"`
		Code     string `json:"code"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	valid, err := twofa.ValidateCode(req.Username, req.Code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if !valid {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "código TOTP inválido"})
	}
	return c.JSON(http.StatusOK, map[string]bool{"valid": true})
}

func disable2FAHandler(c echo.Context) error {
	username := c.Param("username")
	if err := twofa.Disable2FA(username); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	audit.Log(c.Get("username").(string), "disable", "2fa", username, c.RealIP())
	return c.JSON(http.StatusOK, map[string]string{"status": "disabled"})
}

// --- Multi-DC ---

func fsmoShowHandler(c echo.Context) error {
	roles, err := multidc.ShowFSMO()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, roles)
}

func fsmoTransferHandler(c echo.Context) error {
	var req struct {
		Role     string `json:"role"`
		TargetDC string `json:"targetDC"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := multidc.TransferRole(req.Role, req.TargetDC); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	audit.Log(c.Get("username").(string), "transfer", "fsmo", req.Role, c.RealIP())
	return c.JSON(http.StatusOK, map[string]string{"status": "transferred", "role": req.Role})
}

func listTrustsHandler(c echo.Context) error {
	trusts, err := multidc.ShowTrusts()
	if err != nil {
		return c.JSON(http.StatusOK, []string{})
	}
	return c.JSON(http.StatusOK, trusts)
}

// --- Preflight ---

func preflightHandler(c echo.Context) error {
	result := preflight.Run()
	return c.JSON(http.StatusOK, result)
}

func preflightFixHandler(c echo.Context) error {
	id := c.Param("id")
	switch id {
	case "P-04":
		if err := preflight.FixEtcHosts(); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "fixed", "id": id})
	case "P-07":
		preflight.StopService("avahi-daemon")
		return c.JSON(http.StatusOK, map[string]string{"status": "fixed", "id": id})
	case "P-08":
		preflight.StopService("systemd-resolved")
		return c.JSON(http.StatusOK, map[string]string{"status": "fixed", "id": id})
	case "P-09":
		preflight.StopService("dnsmasq")
		return c.JSON(http.StatusOK, map[string]string{"status": "fixed", "id": id})
	case "P-10":
		preflight.BackupSmbConf()
		return c.JSON(http.StatusOK, map[string]string{"status": "fixed", "id": id})
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "no auto-fix for " + id})
	}
}

// --- Provisioning ---

func provisionHandler(c echo.Context) error {
	var req provision.ProvisionRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	pf := preflight.Run()
	if !pf.Ready {
		return c.JSON(http.StatusPreconditionFailed, map[string]interface{}{"error": "preflight failed", "preflight": pf})
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Minute)
	defer cancel()
	result, err := provision.Provision(ctx, req, &bytesBuffer{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if result.Success {
		provision.PostProvision(ctx, req.Realm, req.DNSForwarder, &bytesBuffer{})
	}
	return c.JSON(http.StatusOK, result)
}

type bytesBuffer struct{ data []byte }

func (b *bytesBuffer) Write(p []byte) (int, error) { b.data = append(b.data, p...); return len(p), nil }

func provisionStreamHandler(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	return c.String(http.StatusOK, "data: SSE not yet implemented\n\n")
}

func domainHealthHandler(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	health := map[string]interface{}{"sambaforge": "ok"}
	cmd := exec.CommandContext(ctx, "samba-tool", "domain", "level", "show")
	out, err := cmd.CombinedOutput()
	if err != nil {
		health["domain"] = "not provisioned"
		health["samba"] = "stopped"
	} else {
		health["domain"] = "provisioned"
		health["samba"] = "running"
		health["level"] = string(out)
	}
	return c.JSON(http.StatusOK, health)
}