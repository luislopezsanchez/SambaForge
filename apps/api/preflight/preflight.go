package preflight

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

// CheckResult represents the result of a single preflight check.
type CheckResult struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Status      string `json:"status"`       // "pass", "warning", "fail"
	Value       string `json:"value"`        // current value detected
	Message     string `json:"message"`      // explanation
	Remediable  bool   `json:"remediable"`   // can SambaForge fix it?
	AutoFix     bool   `json:"autoFix"`      // can fix without user input?
	Critical    bool   `json:"critical"`     // blocks provisioning?
}

// PreflightResult is the full result of all checks.
type PreflightResult struct {
	Checks        []CheckResult `json:"checks"`
	Ready         bool          `json:"ready"`
	BlockingCount int           `json:"blockingCount"`
	OS            OSInfo        `json:"os"`
}

// OSInfo holds detected OS information.
type OSInfo struct {
	Distro    string `json:"distro"`
	Version   string `json:"version"`
	Supported bool   `json:"supported"`
}

// Run executes all preflight checks and returns the aggregated result.
func Run() *PreflightResult {
	osInfo := detectOS()
	result := &PreflightResult{OS: osInfo}

	checks := []CheckResult{
		checkOS(osInfo),
		checkSambaInstalled(),
		checkSambaVersion(),
		checkServerRole(),
		checkIPStatic(),
		checkEtcHosts(),
		checkHostname(),
		checkAvahi(),
		checkSystemdResolved(),
		checkDnsmasq(),
		checkSmbConf(),
		checkSambaDatabases(),
		checkPortsFree(),
		checkChrony(),
		checkDiskSpace(),
		checkRAM(),
	}

	result.Checks = checks
	for _, c := range checks {
		if c.Critical && c.Status != "pass" {
			result.BlockingCount++
		}
	}
	result.Ready = result.BlockingCount == 0

	return result
}

// --- OS detection ---

func detectOS() OSInfo {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return OSInfo{Supported: false}
	}

	info := OSInfo{}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "ID=") {
			info.Distro = strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
		}
		if strings.HasPrefix(line, "VERSION_ID=") {
			info.Version = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
		}
	}

	// Check support: Debian >= 13, Ubuntu >= 24.04
	switch info.Distro {
	case "debian":
		major, _ := strconv.Atoi(info.Version)
		info.Supported = major >= 13
	case "ubuntu":
		info.Supported = strings.HasPrefix(info.Version, "24.0") || strings.HasPrefix(info.Version, "25.") || strings.HasPrefix(info.Version, "26.")
	default:
		info.Supported = false
	}

	return info
}

// --- Individual checks ---

func checkOS(osi OSInfo) CheckResult {
	c := CheckResult{
		ID:       "P-01",
		Label:    "Sistema operativo",
		Critical: true,
	}
	if !osi.Supported {
		c.Status = "fail"
		c.Value = fmt.Sprintf("%s %s", osi.Distro, osi.Version)
		c.Message = fmt.Sprintf("OS no soportado. SambaForge requiere Debian 13+ o Ubuntu 24.04+. Detectado: %s %s", osi.Distro, osi.Version)
		c.Remediable = false
	} else {
		c.Status = "pass"
		c.Value = fmt.Sprintf("%s %s", osi.Distro, osi.Version)
		c.Message = "OS soportado"
	}
	return c
}

func checkSambaInstalled() CheckResult {
	c := CheckResult{
		ID:       "P-02",
		Label:    "Samba instalada",
		Critical: true,
	}
	path, err := exec.LookPath("samba-tool")
	if err != nil {
		c.Status = "fail"
		c.Value = "no instalada"
		c.Message = "samba-tool no encontrado. Instalar con: apt install samba samba-ad-dc"
		c.Remediable = true
		c.AutoFix = true
	} else {
		c.Status = "pass"
		c.Value = path
		c.Message = "samba-tool encontrado"
	}
	return c
}

func checkSambaVersion() CheckResult {
	c := CheckResult{
		ID:       "P-02b",
		Label:    "Versión de Samba",
		Critical: true,
	}
	out, err := exec.Command("samba-tool", "--version").Output()
	if err != nil {
		c.Status = "fail"
		c.Value = "desconocida"
		c.Message = "No se pudo determinar la versión de Samba"
		return c
	}

	version := strings.TrimSpace(string(out))
	// Extract major.minor from "4.22.10-Debian-..."
	re := regexp.MustCompile(`(\d+)\.(\d+)`)
	m := re.FindStringSubmatch(version)
	if len(m) >= 3 {
		major, _ := strconv.Atoi(m[1])
		minor, _ := strconv.Atoi(m[2])
		if major > 4 || (major == 4 && minor >= 22) {
			c.Status = "pass"
		} else {
			c.Status = "warning"
			c.Message = fmt.Sprintf("Versión %d.%d es menor a 4.22 recomendada", major, minor)
		}
	} else {
		c.Status = "warning"
		c.Message = "No se pudo parsear la versión"
	}
	c.Value = version
	return c
}

func checkServerRole() CheckResult {
	c := CheckResult{
		ID:       "P-02c",
		Label:    "Rol del servidor Samba",
		Critical: false,
	}
	out, err := exec.Command("testparm", "--parameter-name=server role", "-s").Output()
	if err != nil {
		c.Status = "pass"
		c.Value = "no config"
		c.Message = "No hay smb.conf — listo para provisioning"
		return c
	}

	role := strings.TrimSpace(string(out))
	if strings.Contains(role, "active directory domain controller") {
		c.Status = "warning"
		c.Value = "AD DC ya provisionado"
		c.Message = "El servidor ya es un AD DC. ¿Re-provisionar?"
	} else {
		c.Status = "pass"
		c.Value = role
		c.Message = fmt.Sprintf("Rol actual: %s — se cambiará a AD DC durante provisioning", role)
	}
	return c
}

func checkIPStatic() CheckResult {
	c := CheckResult{
		ID:       "P-03",
		Label:    "IP estática",
		Critical: true,
	}

	// Get the primary interface IP
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		c.Status = "fail"
		c.Value = "sin red"
		c.Message = "No se pudo determinar la IP del servidor"
		return c
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip := localAddr.IP.String()

	// Check if DHCP by looking at networkctl or lease files
	isDHCP := false

	// Method 1: check for dhclient process
	if _, err := exec.LookPath("dhclient"); err == nil {
		if out, err := exec.Command("pgrep", "dhclient").Output(); err == nil && len(out) > 0 {
			isDHCP = true
		}
	}

	// Method 2: check networkctl
	out, err := exec.Command("networkctl", "list", "--no-legend").Output()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "routable") && (strings.Contains(line, "eth0") || strings.Contains(line, "ens") || strings.Contains(line, "enp")) {
				// Check if it mentions DHCP
				if strings.Contains(strings.ToLower(line), "dhcp") {
					isDHCP = true
				}
			}
		}
	}

	// Method 3: check /run/systemd/netif/leases/ existence
	matches, _ := filepathGlob("/run/systemd/netif/leases/*")
	if len(matches) > 0 {
		isDHCP = true
	}

	// Method 4: check /var/lib/dhcp/ for lease files
	matches2, _ := filepathGlob("/var/lib/dhcp/*.leases")
	if len(matches2) > 0 {
		isDHCP = true
	}

	// Method 5: check /etc/network/interfaces for "inet dhcp" (Debian)
	if data, err := os.ReadFile("/etc/network/interfaces"); err == nil {
		if strings.Contains(string(data), "inet dhcp") {
			isDHCP = true
		} else if strings.Contains(string(data), "inet static") {
			isDHCP = false // static config overrides other heuristics
		}
	}

	// Method 6: check netplan for DHCP (Ubuntu)
	if matches, _ := filepathGlob("/etc/netplan/*.yaml"); len(matches) > 0 {
		for _, f := range matches {
			if data, err := os.ReadFile(f); err == nil {
				if strings.Contains(string(data), "dhcp4: true") {
					isDHCP = true
				} else if strings.Contains(string(data), "dhcp4: false") {
					isDHCP = false
				}
			}
		}
	}

	if isDHCP {
		c.Status = "warning"
		c.Value = fmt.Sprintf("%s (DHCP)", ip)
		c.Message = fmt.Sprintf("La IP %s parece ser DHCP. Un DC requiere IP estática. SambaForge puede configurarla.", ip)
		c.Remediable = true
		c.AutoFix = false // needs user input for IP
	} else {
		c.Status = "pass"
		c.Value = ip
		c.Message = "IP estática detectada"
	}
	return c
}

func checkEtcHosts() CheckResult {
	c := CheckResult{
		ID:       "P-04",
		Label:    "/etc/hosts",
		Critical: true,
	}

	data, err := os.ReadFile("/etc/hosts")
	if err != nil {
		c.Status = "fail"
		c.Value = "no se pudo leer"
		c.Message = "No se pudo leer /etc/hosts"
		return c
	}

	hostname, _ := os.Hostname()
	lines := strings.Split(string(data), "\n")

	// Get primary IP
	conn, _ := net.Dial("udp", "8.8.8.8:80")
	if conn == nil {
		c.Status = "fail"
		c.Message = "Sin red"
		return c
	}
	localIP := conn.LocalAddr().(*net.UDPAddr).IP.String()
	conn.Close()

	// Check if FQDN resolves to the real IP (not 127.x)
	fqdn := hostname
	// Try to get FQDN
	if out, err := exec.Command("hostname", "-f").Output(); err == nil {
		fqdn = strings.TrimSpace(string(out))
	}

	correct := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "127.") || strings.HasPrefix(line, "::") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			ipInFile := fields[0]
			if ipInFile == localIP {
				for _, name := range fields[1:] {
					if name == fqdn || name == hostname {
						correct = true
					}
				}
			}
		}
	}

	if correct {
		c.Status = "pass"
		c.Value = fmt.Sprintf("%s → %s", localIP, fqdn)
		c.Message = "/etc/hosts correcto"
	} else {
		c.Status = "fail"
		c.Value = fmt.Sprintf("FQDN no apunta a %s", localIP)
		c.Message = fmt.Sprintf("/etc/hosts: %s debería resolver a %s (no a 127.x). SambaForge puede corregirlo.", fqdn, localIP)
		c.Remediable = true
		c.AutoFix = true
	}
	return c
}

func checkHostname() CheckResult {
	c := CheckResult{
		ID:       "P-06",
		Label:    "Hostname",
		Critical: true,
	}
	hostname, _ := os.Hostname()

	if len(hostname) > 15 {
		c.Status = "fail"
		c.Value = fmt.Sprintf("%s (%d chars)", hostname, len(hostname))
		c.Message = fmt.Sprintf("Hostname '%s' tiene %d caracteres. NetBIOS limita a 15.", hostname, len(hostname))
		c.Remediable = true
		c.AutoFix = false
	} else if strings.Contains(strings.ToUpper(hostname), "PDC") || strings.Contains(strings.ToUpper(hostname), "BDC") {
		c.Status = "warning"
		c.Value = hostname
		c.Message = "Hostname contiene 'PDC' o 'BDC' — términos NT4 deprecados en AD"
	} else {
		c.Status = "pass"
		c.Value = hostname
		c.Message = fmt.Sprintf("Hostname '%s' válido (%d chars)", hostname, len(hostname))
	}
	return c
}

func checkAvahi() CheckResult {
	return checkServiceNotRunning("P-07", "avahi-daemon", "Avahi conflictúa con el DNS interno de Samba")
}

func checkSystemdResolved() CheckResult {
	return checkServiceNotRunning("P-08", "systemd-resolved", "systemd-resolved ocupa el puerto 53 que Samba necesita para DNS")
}

func checkDnsmasq() CheckResult {
	return checkServiceNotRunning("P-09", "dnsmasq", "dnsmasq ocupa el puerto 53 que Samba necesita para DNS")
}

func checkServiceNotRunning(id, name, reason string) CheckResult {
	c := CheckResult{
		ID:       id,
		Label:    name,
		Critical: true,
	}
	out, err := exec.Command("systemctl", "is-active", name).Output()
	status := strings.TrimSpace(string(out))

	if err != nil && status != "inactive" && status != "failed" {
		// systemctl returns non-zero for inactive, which is what we want
	}

	if status == "active" {
		c.Status = "fail"
		c.Value = "activo"
		c.Message = fmt.Sprintf("%s está activo. %s. SambaForge puede detenerlo.", name, reason)
		c.Remediable = true
		c.AutoFix = true
	} else {
		c.Status = "pass"
		c.Value = "inactivo"
		c.Message = fmt.Sprintf("%s no está corriendo", name)
	}
	return c
}

func checkSmbConf() CheckResult {
	c := CheckResult{
		ID:       "P-10",
		Label:    "smb.conf previo",
		Critical: true,
	}
	if _, err := os.Stat("/etc/samba/smb.conf"); err == nil {
		c.Status = "warning"
		c.Value = "existe"
		c.Message = "Hay un smb.conf previo. Se hará backup a smb.conf.initial antes del provisioning."
		c.Remediable = true
		c.AutoFix = true
	} else {
		c.Status = "pass"
		c.Value = "no existe"
		c.Message = "No hay smb.conf previo — listo para provisioning"
	}
	return c
}

func checkSambaDatabases() CheckResult {
	c := CheckResult{
		ID:       "P-11",
		Label:    "Bases de datos previas",
		Critical: true,
	}
	matches, _ := filepathGlob("/var/lib/samba/private/*.ldb")
	if len(matches) > 0 {
		c.Status = "warning"
		c.Value = fmt.Sprintf("%d archivos .ldb", len(matches))
		c.Message = "Hay bases de datos de una instalación anterior. Se eliminarán antes del provisioning."
		c.Remediable = true
		c.AutoFix = true
	} else {
		c.Status = "pass"
		c.Value = "limpio"
		c.Message = "No hay bases de datos previas"
	}
	return c
}

func checkPortsFree() CheckResult {
	c := CheckResult{
		ID:       "P-12",
		Label:    "Puertos AD libres",
		Critical: true,
	}

	ports := []int{53, 88, 389, 445, 464}
	var occupied []string

	for _, port := range ports {
		addr := fmt.Sprintf(":%d", port)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			occupied = append(occupied, strconv.Itoa(port))
		} else {
			ln.Close()
		}
	}

	// Also check UDP port 53
	udpAddr, err := net.ListenPacket("udp", ":53")
	if err != nil {
		// Already checked in TCP, don't double-report
	} else {
		udpAddr.Close()
	}

	if len(occupied) > 0 {
		c.Status = "fail"
		c.Value = fmt.Sprintf("ocupados: %s", strings.Join(occupied, ", "))
		c.Message = fmt.Sprintf("Puertos ocupados: %s. Liberarlos antes del provisioning.", strings.Join(occupied, ", "))
		c.Remediable = false
	} else {
		c.Status = "pass"
		c.Value = "53, 88, 389, 445, 464 libres"
		c.Message = "Todos los puertos de AD están libres"
	}
	return c
}

func checkChrony() CheckResult {
	c := CheckResult{
		ID:       "P-15",
		Label:    "Chrony/NTP",
		Critical: false, // warning only — AD works without NTP but it's recommended
	}
	out, _ := exec.Command("systemctl", "is-active", "chrony").Output()
	status := strings.TrimSpace(string(out))

	if status == "active" {
		c.Status = "pass"
		c.Value = "activo"
		c.Message = "Chrony está corriendo"
	} else if _, err := exec.LookPath("chronyd"); err == nil {
		c.Status = "warning"
		c.Value = fmt.Sprintf("instalado pero %s", status)
		c.Message = "Chrony está instalado pero no activo. Un DC debería servir hora."
		c.Remediable = true
		c.AutoFix = true
	} else {
		c.Status = "warning"
		c.Value = "no instalado"
		c.Message = "Chrony no está instalado. Recomendado para un DC."
		c.Remediable = true
		c.AutoFix = true
	}
	return c
}

func checkDiskSpace() CheckResult {
	c := CheckResult{
		ID:       "P-17",
		Label:    "Espacio en disco",
		Critical: false,
	}
	var stat syscall.Statfs_t
	syscall.Statfs("/", &stat)
	freeGB := float64(stat.Bavail*uint64(stat.Bsize)) / 1e9

	if freeGB < 10 {
		c.Status = "warning"
		c.Value = fmt.Sprintf("%.1f GB libres", freeGB)
		c.Message = "Menos de 10 GB libres. Puede no ser suficiente para un DC."
	} else {
		c.Status = "pass"
		c.Value = fmt.Sprintf("%.1f GB libres", freeGB)
		c.Message = "Espacio suficiente"
	}
	return c
}

func checkRAM() CheckResult {
	c := CheckResult{
		ID:       "P-18",
		Label:    "Memoria RAM",
		Critical: false,
	}
	out, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		c.Status = "warning"
		c.Value = "desconocida"
		return c
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, _ := strconv.ParseInt(fields[1], 10, 64)
				gb := float64(kb) / 1e6
				if gb < 2 {
					c.Status = "warning"
					c.Value = fmt.Sprintf("%.1f GB", gb)
					c.Message = "Menos de 2 GB de RAM. Puede no ser suficiente."
				} else {
					c.Status = "pass"
					c.Value = fmt.Sprintf("%.1f GB", gb)
					c.Message = "RAM suficiente"
				}
			}
			break
		}
	}
	return c
}

// --- Remediation functions ---

// FixEtcHosts corrects /etc/hosts so FQDN resolves to the real LAN IP.
func FixEtcHosts() error {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return fmt.Errorf("sin red: %w", err)
	}
	localIP := conn.LocalAddr().(*net.UDPAddr).IP.String()
	conn.Close()

	hostname, _ := os.Hostname()
	fqdn := hostname
	if out, err := exec.Command("hostname", "-f").Output(); err == nil {
		fqdn = strings.TrimSpace(string(out))
	}

	// Backup
	os.Rename("/etc/hosts", "/etc/hosts.bak")

	content := fmt.Sprintf("127.0.0.1\tlocalhost\n::1\t\tlocalhost ip6-localhost ip6-loopback\nff02::1\t\tip6-allnodes\nff02::2\t\tip6-allrouters\n%s\t%s %s\n", localIP, fqdn, hostname)
	return os.WriteFile("/etc/hosts", []byte(content), 0644)
}

// StopService stops and disables a systemd service.
func StopService(name string) error {
	exec.Command("systemctl", "stop", name).Run()
	exec.Command("systemctl", "disable", name).Run()
	return nil
}

// BackupSmbConf moves smb.conf to smb.conf.initial if it exists.
func BackupSmbConf() error {
	if _, err := os.Stat("/etc/samba/smb.conf"); err == nil {
		return os.Rename("/etc/samba/smb.conf", "/etc/samba/smb.conf.initial")
	}
	return nil
}

// --- Helpers ---

func filepathGlob(pattern string) ([]string, error) {
	// Use os.ReadDir for simple glob patterns
	dir := filepathDir(pattern)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var matches []string
	for _, entry := range entries {
		matches = append(matches, fmt.Sprintf("%s/%s", dir, entry.Name()))
	}
	return matches, nil
}

func filepathDir(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx == -1 {
		return "."
	}
	return path[:idx]
}