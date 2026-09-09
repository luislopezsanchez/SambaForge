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

	"github.com/luislopezsanchez/SambaForge/auth"
	"github.com/luislopezsanchez/SambaForge/directory"
	"github.com/luislopezsanchez/SambaForge/dns"
	"github.com/luislopezsanchez/SambaForge/preflight"
	"github.com/luislopezsanchez/SambaForge/provision"
)

var version = "0.2.0-dev"

func main() {
	auth.Init()

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Public routes (no auth)
	e.GET("/api/health", healthHandler)
	e.POST("/api/auth/login", loginHandler)
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
	api.GET("/dns/zones", listDnsZonesHandler)
	api.GET("/dns/zones/:zone/records", listDnsRecordsHandler)
	api.POST("/dns/records", addDnsRecordHandler)
	api.DELETE("/dns/records", deleteDnsRecordHandler)
	api.GET("/dns/forwarders", getDnsForwardersHandler)

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
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
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
	return c.JSON(http.StatusCreated, map[string]string{"status": "created", "username": req.Username})
}

func deleteUserHandler(c echo.Context) error {
	username := c.Param("username")
	if err := directory.DeleteUser(username); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
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