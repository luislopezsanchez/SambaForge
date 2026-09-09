package main

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// API routes
	e.GET("/api/health", healthHandler)

	// Serve embedded frontend (static files)
	webDir := os.Getenv("SAMBAFORGE_WEB_DIR")
	if webDir == "" {
		webDir = "/opt/sambaforge/apps/web/dist"
	}

	// Check if web dist exists
	if _, err := os.Stat(webDir); err == nil {
		e.Use(middleware.Static(webDir))
		// SPA fallback: serve index.html for non-API, non-file routes
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
		"version": "0.0.1-dev",
	})
}