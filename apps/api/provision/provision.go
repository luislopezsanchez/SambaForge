package provision

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// ProvisionRequest holds the parameters for domain provisioning.
type ProvisionRequest struct {
	Realm       string `json:"realm"`       // e.g. SAMDOM.EXAMPLE.COM
	Domain      string `json:"domain"`      // NetBIOS name, e.g. SAMDOM
	DNSBackend  string `json:"dnsBackend"`  // SAMBA_INTERNAL (default), BIND9_DLZ
	AdminPass   string `json:"adminPass"`   // admin password
	UseRFC2307  bool   `json:"useRFC2307"`  // NIS extensions for Unix attrs
	DNSForwarder string `json:"dnsForwarder"` // e.g. 8.8.8.8
}

// ProvisionResult holds the result of the provisioning attempt.
type ProvisionResult struct {
	Success  bool     `json:"success"`
	Message  string   `json:"message"`
	Output   []string `json:"output"`   // line-by-line output
	Error    string   `json:"error"`    // stderr if failed
	ExitCode int      `json:"exitCode"`
}

// Provision executes samba-tool domain provision non-interactively.
// It streams output line-by-line to the logWriter (for SSE streaming to the web UI).
func Provision(ctx context.Context, req ProvisionRequest, logWriter io.Writer) (*ProvisionResult, error) {
	// Validate input
	if req.Realm == "" || req.Domain == "" || req.AdminPass == "" {
		return nil, fmt.Errorf("realm, domain y adminPass son obligatorios")
	}

	// Build samba-tool command args
	args := []string{
		"domain", "provision",
		"--server-role=dc",
		"--realm=" + req.Realm,
		"--domain=" + req.Domain,
		"--dns-backend=" + defaultStr(req.DNSBackend, "SAMBA_INTERNAL"),
		"--adminpass=" + req.AdminPass, // Note: passed via env var in production, here for spike
		"--color=never",
	}

	if req.UseRFC2307 {
		args = append(args, "--use-rfc2307")
	}

	// Resolve samba-tool absolute path
	sambaToolPath, err := exec.LookPath("samba-tool")
	if err != nil {
		return nil, fmt.Errorf("samba-tool no encontrado en PATH: %w", err)
	}

	// Create command with context (timeout)
	cmd := exec.CommandContext(ctx, sambaToolPath, args...)
	cmd.Env = append(os.Environ())

	// Capture stdout and stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start samba-tool: %w", err)
	}

	// Read stdout line by line and stream to logWriter
	result := &ProvisionResult{Output: []string{}}
	stdoutDone := make(chan bool)
	stderrDone := make(chan bool)

	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			result.Output = append(result.Output, line)
			fmt.Fprintf(logWriter, "%s\n", line)
		}
		stdoutDone <- true
	}()

	var stderrLines []string
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			stderrLines = append(stderrLines, line)
			fmt.Fprintf(logWriter, "[stderr] %s\n", line)
		}
		stderrDone <- true
	}()

	<-stdoutDone
	<-stderrDone

	err = cmd.Wait()
	result.ExitCode = cmd.ProcessState.ExitCode()

	if err != nil {
		result.Success = false
		result.Error = strings.Join(stderrLines, "\n")
		result.Message = fmt.Sprintf("Provisioning falló (exit code %d)", result.ExitCode)
		return result, nil
	}

	result.Success = true
	result.Message = "Dominio provisionado correctamente"
	return result, nil
}

// PostProvision performs the post-provisioning steps:
// 1. Copy krb5.conf to /etc/krb5.conf
// 2. Configure /etc/resolv.conf to use Samba DNS
// 3. Start samba service
func PostProvision(ctx context.Context, realm, dnsForwarder string, logWriter io.Writer) error {
	sambaPrivateDir := "/var/lib/samba/private"

	// Step 1: Copy krb5.conf
	fmt.Fprintf(logWriter, "\n[post-provision] Copiando krb5.conf...\n")
	srcKrb5 := sambaPrivateDir + "/krb5.conf"
	if _, err := os.Stat(srcKrb5); err == nil {
		src, err := os.Open(srcKrb5)
		if err != nil {
			return fmt.Errorf("abrir krb5.conf source: %w", err)
		}
		defer src.Close()

		dst, err := os.Create("/etc/krb5.conf")
		if err != nil {
			return fmt.Errorf("crear /etc/krb5.conf: %w", err)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return fmt.Errorf("copiar krb5.conf: %w", err)
		}
		fmt.Fprintf(logWriter, "[post-provision] krb5.conf copiado a /etc/krb5.conf\n")
	}

	// Step 2: Configure resolv.conf
	fmt.Fprintf(logWriter, "[post-provision] Configurando /etc/resolv.conf...\n")
	resolvContent := fmt.Sprintf("nameserver 127.0.0.1\nsearch %s\n", strings.ToLower(realm))
	if dnsForwarder != "" {
		// Samba internal DNS handles forwarding, but keep resolv.conf simple
		resolvContent = fmt.Sprintf("nameserver 127.0.0.1\nsearch %s\n", strings.ToLower(realm))
	}
	if err := os.WriteFile("/etc/resolv.conf", []byte(resolvContent), 0644); err != nil {
		return fmt.Errorf("escribir resolv.conf: %w", err)
	}
	// Protect resolv.conf from being overwritten by dhcpcd/NetworkManager
	exec.CommandContext(ctx, "chattr", "+i", "/etc/resolv.conf").Run()
	fmt.Fprintf(logWriter, "[post-provision] resolv.conf configurado y protegido (nameserver 127.0.0.1)\n")

	// Step 3: Kill stale winbindd from standalone Samba (if any)
	fmt.Fprintf(logWriter, "[post-provision] Limpiando procesos stale...\n")
	exec.CommandContext(ctx, "systemctl", "stop", "winbind").Run()
	exec.CommandContext(ctx, "systemctl", "disable", "winbind").Run()
	os.Remove("/run/samba/winbindd.pid")

	// Step 4: Start samba service
	fmt.Fprintf(logWriter, "[post-provision] Iniciando servicio Samba AD DC...\n")
	cmd := exec.CommandContext(ctx, "systemctl", "restart", "samba-ad-dc")
	if err := cmd.Run(); err != nil {
		// Try alternative service name
		cmd2 := exec.CommandContext(ctx, "systemctl", "restart", "samba")
		if err2 := cmd2.Run(); err2 != nil {
			return fmt.Errorf("iniciar samba: %w (también probé 'samba'): %v", err, err2)
		}
	}
	fmt.Fprintf(logWriter, "[post-provision] Servicio Samba AD DC iniciado\n")

	// Step 5: Verify
	fmt.Fprintf(logWriter, "[post-provision] Verificando dominio...\n")
	verifyCmd := exec.CommandContext(ctx, "samba-tool", "domain", "level", "show")
	verifyOut, err := verifyCmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(logWriter, "[post-provision] ADVERTENCIA: domain level show falló: %v\n", err)
	} else {
		fmt.Fprintf(logWriter, "[post-provision] Domain level:\n%s\n", string(verifyOut))
	}

	return nil
}

func defaultStr(s, def string) string {
	if s == "" {
		return def
	}
	return s
}