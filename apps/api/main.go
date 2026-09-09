package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/luislopezsanchez/SambaForge/preflight"
	"github.com/luislopezsanchez/SambaForge/provision"
)

var version = "0.0.1-dev"

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// API routes
	e.GET("/api/health", healthHandler)
	e.GET("/api/server/preflight", preflightHandler)
	e.POST("/api/server/preflight/fix/:id", preflightFixHandler)
	e.POST("/api/domain/provision", provisionHandler)
	e.GET("/api/domain/provision/stream", provisionStreamHandler)
	e.GET("/api/domain/health", domainHealthHandler)

	// Serve embedded frontend (static files)
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

func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"service": "SambaForge",
		"version": version,
	})
}

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
		if err := preflight.StopService("avahi-daemon"); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "fixed", "id": id})

	case "P-08":
		if err := preflight.StopService("systemd-resolved"); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "fixed", "id": id})

	case "P-09":
		if err := preflight.StopService("dnsmasq"); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "fixed", "id": id})

	case "P-10":
		if err := preflight.BackupSmbConf(); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "fixed", "id": id})

	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "no auto-fix available for " + id})
	}
}

func provisionHandler(c echo.Context) error {
	var req provision.ProvisionRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	// Run preflight first
	pf := preflight.Run()
	if !pf.Ready {
		return c.JSON(http.StatusPreconditionFailed, map[string]interface{}{
			"error":   "preflight checks failed",
			"preflight": pf,
		})
	}

	// Provision with 5 minute timeout
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Minute)
	defer cancel()

	var logBuf bytes.Buffer
	result, err := provision.Provision(ctx, req, &logBuf)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Post-provision
	if result.Success {
		if err := provision.PostProvision(ctx, req.Realm, req.DNSForwarder, &logBuf); err != nil {
			result.Message += fmt.Sprintf(" (post-provision warning: %v)", err)
		}
	}

	return c.JSON(http.StatusOK, result)
}

func provisionStreamHandler(c echo.Context) error {
	// SSE streaming for real-time provisioning output
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")

	// For now, return a placeholder — full SSE implementation in next iteration
	return c.String(http.StatusOK, "data: SSE streaming not yet implemented\\n\\n")
}

func domainHealthHandler(c echo.Context) error {
	// Check if Samba AD DC is running
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	health := map[string]interface{}{
		"sambaforge": "ok",
	}

	// Check samba-tool domain level show
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