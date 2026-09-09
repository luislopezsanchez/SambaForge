package gpo

import (
	"fmt"
	"os/exec"
	"strings"
)

// GPO represents a Group Policy Object.
type GPO struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	DN      string `json:"dn"`
}

// ListGPOs returns all GPOs in the domain.
func ListGPOs() ([]GPO, error) {
	cmd := exec.Command("samba-tool", "gpo", "listall", "-P", "--color=never")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("samba-tool gpo listall: %w", err)
	}

	var gpos []GPO
	var current GPO
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "GPO") && strings.Contains(line, ":") {
			// "GPO          : {GUID}"
			id := strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			if current.ID != "" {
				gpos = append(gpos, current)
			}
			current = GPO{ID: id}
		} else if strings.HasPrefix(line, "display name") {
			current.Name = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		} else if strings.HasPrefix(line, "dn") {
			current.DN = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		}
	}
	if current.ID != "" {
		gpos = append(gpos, current)
	}
	return gpos, nil
}

// CreateGPO creates a new GPO.
func CreateGPO(name string) error {
	cmd := exec.Command("samba-tool", "gpo", "create", name, "-P", "--color=never")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("samba-tool gpo create: %s: %w", string(out), err)
	}
	return nil
}

// DeleteGPO deletes a GPO by ID.
func DeleteGPO(id string) error {
	cmd := exec.Command("samba-tool", "gpo", "del", id, "-P", "--color=never")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("samba-tool gpo del: %s: %w", string(out), err)
	}
	return nil
}

// GetInheritance returns GPO inheritance for a container.
func GetInheritance(containerDN string) (string, error) {
	cmd := exec.Command("samba-tool", "gpo", "getinheritance", containerDN, "-P", "--color=never")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("samba-tool gpo getinheritance: %w", err)
	}
	return string(out), nil
}

// SetInheritance sets GPO inheritance for a container.
func SetInheritance(containerDN string, enabled bool) error {
	flag := "1"
	if !enabled {
		flag = "0"
	}
	cmd := exec.Command("samba-tool", "gpo", "setinheritance", containerDN, flag, "-P", "--color=never")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("samba-tool gpo setinheritance: %s: %w", string(out), err)
	}
	return nil
}

// Template represents a preconfigured GPO template.
type Template struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// ListTemplates returns preconfigured GPO templates available.
func ListTemplates() []Template {
	return []Template{
		{ID: "password-policy", Name: "Política de Contraseña", Description: "Complejidad, longitud mínima, historial, expiración", Category: "Seguridad"},
		{ID: "drive-maps", Name: "Mapeo de Unidades", Description: "Mapear unidades de red automáticamente al iniciar sesión", Category: "Scripts"},
		{ID: "wallpaper", Name: "Fondo de Pantalla", Description: "Establecer wallpaper corporativo en todos los equipos", Category: "Apariencia"},
		{ID: "logon-script", Name: "Script de Inicio", Description: "Ejecutar script al iniciar sesión", Category: "Scripts"},
		{ID: "registry-settings", Name: "Configuración de Registro", Description: "Aplicar claves de registro a los equipos", Category: "Sistema"},
		{ID: "firewall-rules", Name: "Reglas de Firewall", Description: "Configurar reglas de firewall de Windows", Category: "Seguridad"},
	}
}