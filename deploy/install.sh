#!/bin/bash
#
# SambaForge — Script de instalación interactiva
# 
# Este script prepara un servidor Debian 13+ o Ubuntu 24.04+ desde cero
# para que SambaForge pueda gestionar un controlador de dominio Samba AD DC.
#
# Uso:
#   curl -fsSL https://raw.githubusercontent.com/luislopezsanchez/SambaForge/main/deploy/install.sh | bash
#   o:
#   bash install.sh
#
set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

log()    { echo -e "${GREEN}[SambaForge]${NC} $1"; }
warn()   { echo -e "${YELLOW}[WARN]${NC} $1"; }
error()  { echo -e "${RED}[ERROR]${NC} $1"; }
info()   { echo -e "${BLUE}[INFO]${NC} $1"; }
prompt() { echo -e "${BOLD}$1${NC}"; }

# --- OS Detection ---
detect_os() {
    if [ ! -f /etc/os-release ]; then
        error "No se pudo detectar el sistema operativo. SambaForge requiere Debian 13+ o Ubuntu 24.04+."
        exit 1
    fi

    . /etc/os-release
    DISTRO="$ID"
    VERSION="$VERSION_ID"

    case "$DISTRO" in
        debian)
            MAJOR=$(echo "$VERSION" | cut -d. -f1)
            if [ "$MAJOR" -lt 13 ]; then
                error "Debian $VERSION no es soportado. Se requiere Debian 13+ (trixie)."
                exit 1
            fi
            log "OS detectado: Debian $VERSION ✓"
            ;;
        ubuntu)
            MAJOR=$(echo "$VERSION" | cut -d. -f1)
            if [ "$MAJOR" -lt 24 ]; then
                error "Ubuntu $VERSION no es soportado. Se requiere Ubuntu 24.04+."
                exit 1
            fi
            log "OS detectado: Ubuntu $VERSION ✓"
            ;;
        *)
            error "Distribución '$DISTRO' no soportada. SambaForge requiere Debian 13+ o Ubuntu 24.04+."
            exit 1
            ;;
    esac
}

# --- Check if running as root ---
check_root() {
    if [ "$(id -u)" -ne 0 ]; then
        error "Este script debe ejecutarse como root."
        exit 1
    fi
}

# --- Check if LXC (warn only) ---
check_virt() {
    if command -v systemd-detect-virt &>/dev/null; then
        VIRT=$(systemd-detect-virt)
        if [ "$VIRT" = "lxc" ]; then
            warn "Se detectó un contenedor LXC. Samba AD DC puede no funcionar correctamente en LXC."
            warn "Se recomienda usar una VM KVM completa en Proxmox."
            prompt "¿Continuar de todas formas? (s/N): "
            read -r CONTINUE
            if [ "$CONTINUE" != "s" ] && [ "$CONTINUE" != "S" ]; then
                log "Instalación cancelada."
                exit 0
            fi
        else
            log "Virtualización detectada: $VIRT ✓"
        fi
    fi
}

# --- Check network ---
check_network() {
    # Get primary IP
    PRIMARY_IP=$(ip -4 route get 1.1.1.1 2>/dev/null | awk '{print $7}' | head -1)
    if [ -z "$PRIMARY_IP" ]; then
        error "No se pudo detectar la IP del servidor. Verifica la configuración de red."
        exit 1
    fi
    GATEWAY=$(ip -4 route get 1.1.1.1 2>/dev/null | awk '{print $3}' | head -1)
    INTERFACE=$(ip -4 route get 1.1.1.1 2>/dev/null | awk '{print $5}' | head -1)
    log "IP detectada: $PRIMARY_IP (interfaz $INTERFACE, gateway $GATEWAY)"
}

# --- Check if IP is static or DHCP ---
check_ip_static() {
    IS_DHCP=false

    # Check /etc/network/interfaces (Debian)
    if [ -f /etc/network/interfaces ]; then
        if grep -q "inet dhcp" /etc/network/interfaces; then
            IS_DHCP=true
        elif grep -q "inet static" /etc/network/interfaces; then
            IS_DHCP=false
            log "IP estática configurada en /etc/network/interfaces ✓"
        fi
    fi

    # Check netplan (Ubuntu)
    if [ -d /etc/netplan ]; then
        for f in /etc/netplan/*.yaml; do
            if grep -q "dhcp4: true" "$f" 2>/dev/null; then
                IS_DHCP=true
            elif grep -q "dhcp4: false" "$f" 2>/dev/null; then
                IS_DHCP=false
            fi
        done
    fi

    if [ "$IS_DHCP" = "true" ]; then
        warn "La IP actual ($PRIMARY_IP) parece ser DHCP. Un DC requiere IP estática."
        prompt "¿Configurar IP estática? (recomendado) (s/N): "
        read -r CONFIG_IP
        if [ "$CONFIG_IP" = "s" ] || [ "$CONFIG_IP" = "S" ]; then
            prompt "Introduce la IP estática [${PRIMARY_IP}]: "
            read -r NEW_IP
            NEW_IP=${NEW_IP:-$PRIMARY_IP}

            prompt "Introduce la máscara en formato CIDR [/24]: "
            read -r CIDR
            CIDR=${CIDR:-/24}

            prompt "Introduce el gateway [${GATEWAY}]: "
            read -r NEW_GW
            NEW_GW=${NEW_GW:-$GATEWAY}

            FULL_IP="${NEW_IP}${CIDR}"

            if [ "$DISTRO" = "ubuntu" ] && [ -d /etc/netplan ]; then
                # Netplan config
                cat > "/etc/netplan/01-sambaforge.yaml" << NETPLAN
network:
  version: 2
  ethernets:
    $INTERFACE:
      dhcp4: false
      addresses:
        - $FULL_IP
      routes:
        - to: default
          via: $NEW_GW
      nameservers:
        addresses:
          - 127.0.0.1
          - $NEW_GW
NETPLAN
                log "Configuración netplan escrita."
                warn "Aplicando configuración de red... La conexión SSH puede perderse temporalmente."
                netplan apply 2>/dev/null || true
            else
                # /etc/network/interfaces (Debian)
                cp /etc/network/interfaces /etc/network/interfaces.bak 2>/dev/null || true
                cat > /etc/network/interfaces << INTERFACES
auto lo
iface lo inet loopback

auto $INTERFACE
iface $INTERFACE inet static
    address $FULL_IP
    gateway $NEW_GW
    dns-nameservers 127.0.0.1
INTERFACES
                log "Configuración /etc/network/interfaces escrita."
                warn "Aplicando configuración de red... La conexión SSH puede perderse temporalmente."
                systemctl restart networking 2>/dev/null || ifdown "$INTERFACE" && ifup "$INTERFACE" 2>/dev/null || true
            fi

            PRIMARY_IP="$NEW_IP"
            log "IP estática configurada: $FULL_IP ✓"
        fi
    fi
}

# --- Fix /etc/hosts ---
fix_hosts() {
    HOSTNAME=$(hostname)
    FQDN=$(hostname -f 2>/dev/null || echo "$HOSTNAME")

    # Check if FQDN resolves to real IP (not 127.x)
    if grep -qE "^${PRIMARY_IP}\s+${FQDN}" /etc/hosts; then
        log "/etc/hosts ya está correcto ✓"
        return
    fi

    warn "/etc/hosts no apunta el FQDN a la IP real ($PRIMARY_IP). Corrigiendo..."
    cp /etc/hosts /etc/hosts.bak 2>/dev/null || true

    # Write new /etc/hosts preserving localhost and IPv6
    cat > /etc/hosts << HOSTS
127.0.0.1	localhost
::1		localhost ip6-localhost ip6-loopback
ff02::1		ip6-allnodes
ff02::2		ip6-allrouters
${PRIMARY_IP}	${FQDN} ${HOSTNAME}
HOSTS

    log "/etc/hosts corregido: ${PRIMARY_IP} → ${FQDN} ✓"
}

# --- Check hostname ---
check_hostname() {
    HOSTNAME=$(hostname)
    LEN=${#HOSTNAME}

    if [ "$LEN" -gt 15 ]; then
        warn "Hostname '$HOSTNAME' tiene $LEN caracteres. NetBIOS limita a 15."
        prompt "Introduce un nuevo hostname (máx 15 chars): "
        read -r NEW_HOSTNAME
        if [ -n "$NEW_HOSTNAME" ] && [ ${#NEW_HOSTNAME} -le 15 ]; then
            hostnamectl set-hostname "$NEW_HOSTNAME" 2>/dev/null || hostname "$NEW_HOSTNAME"
            log "Hostname cambiado a: $NEW_HOSTNAME ✓"
        fi
    else
        log "Hostname: $HOSTNAME ($LEN chars) ✓"
    fi

    # Warn about PDC/BDC names
    UPPER=$(echo "$HOSTNAME" | tr '[:lower:]' '[:upper:]')
    if echo "$UPPER" | grep -qE "PDC|BDC"; then
        warn "El hostname contiene 'PDC' o 'BDC' — términos NT4 deprecados en AD."
    fi
}

# --- Stop conflicting services ---
stop_conflicts() {
    # avahi-daemon
    if systemctl is-active avahi-daemon &>/dev/null; then
        warn "avahi-daemon está activo. Conflictúa con el DNS de Samba. Deteniendo..."
        systemctl stop avahi-daemon
        systemctl disable avahi-daemon
        log "avahi-daemon detenido y deshabilitado ✓"
    else
        log "avahi-daemon: inactivo ✓"
    fi

    # systemd-resolved
    if systemctl is-active systemd-resolved &>/dev/null; then
        warn "systemd-resolved está activo. Ocupa el puerto 53. Deteniendo..."
        systemctl stop systemd-resolved
        systemctl disable systemd-resolved
        log "systemd-resolved detenido y deshabilitado ✓"
    else
        log "systemd-resolved: inactivo ✓"
    fi

    # dnsmasq
    if systemctl is-active dnsmasq &>/dev/null; then
        warn "dnsmasq está activo. Ocupa el puerto 53. Deteniendo..."
        systemctl stop dnsmasq
        systemctl disable dnsmasq
        log "dnsmasq detenido y deshabilitado ✓"
    else
        log "dnsmasq: inactivo ✓"
    fi
}

# --- Install Samba ---
install_samba() {
    if command -v samba-tool &>/dev/null; then
        VERSION=$(samba-tool --version 2>&1 | grep -oP '[\d]+\.[\d]+' | head -1)
        log "Samba ya instalada: $VERSION ✓"
        return
    fi

    info "Instalando Samba AD DC..."
    apt update -qq
    apt install -y samba samba-ad-dc smbclient winbind libnss-winbind libpam-winbind \
        chrony ldb-tools krb5-user python3-setproctitle 2>/dev/null

    if ! command -v samba-tool &>/dev/null; then
        error "La instalación de Samba falló."
        exit 1
    fi

    log "Samba instalada: $(samba-tool --version 2>&1 | tail -1) ✓"
}

# --- Stop standalone Samba ---
stop_standalone_samba() {
    # If smb.conf exists with standalone config, stop smbd/nmbd
    if [ -f /etc/samba/smb.conf ]; then
        ROLE=$(testparm --parameter-name='server role' -s 2>/dev/null || echo "")
        if [ -z "$ROLE" ] || [ "$ROLE" != "active directory domain controller" ]; then
            warn "Samba standalone detectada. Deteniendo smbd/nmbd..."
            systemctl stop smbd nmbd 2>/dev/null || true
            systemctl disable smbd nmbd 2>/dev/null || true

            prompt "¿Hacer backup de smb.conf existente y continuar? (s/N): "
            read -r BACKUP_CONF
            if [ "$BACKUP_CONF" = "s" ] || [ "$BACKUP_CONF" = "S" ]; then
                mv /etc/samba/smb.conf /etc/samba/smb.conf.initial
                log "smb.conf respaldado como smb.conf.initial ✓"
            fi
        fi
    fi
}

# --- Clean old databases ---
clean_dbs() {
    LDB_COUNT=$(ls /var/lib/samba/private/*.ldb 2>/dev/null | wc -l)
    if [ "$LDB_COUNT" -gt 0 ]; then
        warn "Se encontraron $LDB_COUNT bases de datos de una instalación anterior."
        prompt "¿Eliminar bases de datos previas? (necesario para provisioning limpio) (s/N): "
        read -r CLEAN
        if [ "$CLEAN" = "s" ] || [ "$CLEAN" = "S" ]; then
            rm -f /var/lib/samba/private/*.ldb /var/lib/samba/private/*.tdb 2>/dev/null
            rm -f /var/lib/samba/*.tdb /var/lib/samba/*.ldb 2>/dev/null
            log "Bases de datos eliminadas ✓"
        fi
    else
        log "No hay bases de datos previas ✓"
    fi
}

# --- Check ports ---
check_ports() {
    PORTS_BLOCKED=""
    for port in 53 88 389 445 464; do
        if ss -tlnp | grep -q ":${port} " 2>/dev/null; then
            PROCESS=$(ss -tlnp | grep ":${port} " | awk '{print $NF}' | head -1)
            PORTS_BLOCKED="${PORTS_BLOCKED} ${port}(${PROCESS})"
        fi
    done

    if [ -n "$PORTS_BLOCKED" ]; then
        warn "Puertos ocupados:${PORTS_BLOCKED}"
        error "Los puertos 53, 88, 389, 445, 464 deben estar libres para Samba AD DC."
        exit 1
    else
        log "Puertos 53, 88, 389, 445, 464 libres ✓"
    fi
}

# --- Install Go ---
install_go() {
    if command -v go &>/dev/null; then
        GO_VERSION=$(go version | awk '{print $3}')
        log "Go ya instalado: $GO_VERSION ✓"
        return
    fi

    info "Instalando Go toolchain..."
    apt install -y golang-go 2>/dev/null
    log "Go instalado: $(go version) ✓"
}

# --- Install Node.js (for building frontend) ---
install_node() {
    if command -v node &>/dev/null; then
        log "Node.js ya instalado: $(node --version) ✓"
        return
    fi

    info "Instalando Node.js..."
    apt install -y nodejs npm 2>/dev/null
    log "Node.js instalado: $(node --version) ✓"
}

# --- Download and compile SambaForge ---
install_sambaforge() {
    SAMBAFORGE_DIR="/opt/sambaforge"
    BINARY="/usr/local/bin/sambaforge"

    info "Descargando SambaForge desde GitHub..."

    # Clone the repo
    if [ -d "$SAMBAFORGE_DIR/.git" ]; then
        log "Directorio $SAMBAFORGE_DIR ya existe. Actualizando..."
        cd "$SAMBAFORGE_DIR"
        git pull 2>/dev/null || true
    else
        apt install -y git 2>/dev/null
        git clone https://github.com/luislopezsanchez/SambaForge.git "$SAMBAFORGE_DIR" 2>/dev/null
        cd "$SAMBAFORGE_DIR"
    fi

    # Build backend
    info "Compilando SambaForge backend (Go)..."
    cd "$SAMBAFORGE_DIR/apps/api"
    go mod tidy 2>/dev/null
    CGO_ENABLED=0 go build -ldflags '-s -w -X main.version=1.0.0' -o "$BINARY" .
    log "Backend compilado: $(ls -lh $BINARY | awk '{print $5}') ✓"

    # Build frontend
    info "Compilando SambaForge frontend (React)..."
    cd "$SAMBAFORGE_DIR/apps/web"
    npm install --silent 2>/dev/null
    npm run build 2>/dev/null
    log "Frontend compilado ✓"

    # Install systemd service
    cp "$SAMBAFORGE_DIR/deploy/sambaforge.service" /etc/systemd/system/sambaforge.service
    systemctl daemon-reload
    systemctl enable sambaforge
    systemctl restart sambaforge
    sleep 2

    if systemctl is-active sambaforge &>/dev/null; then
        log "SambaForge activo y habilitado ✓"
    else
        error "SambaForge no pudo iniciarse. Revisar: journalctl -u sambaforge"
        exit 1
    fi
}

# --- Summary ---
summary() {
    HOSTNAME=$(hostname)
    FQDN=$(hostname -f 2>/dev/null || echo "$HOSTNAME")

    echo ""
    echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}  SambaForge instalado correctamente${NC}"
    echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
    echo ""
    echo "  Host:        $FQDN ($PRIMARY_IP)"
    echo "  SambaForge:  http://$PRIMARY_IP:8444"
    echo "  Estado:      $(systemctl is-active sambaforge)"
    echo ""
    if ! command -v samba-tool &>/dev/null || [ ! -f /etc/samba/smb.conf ] || ! testparm --parameter-name='server role' -s 2>/dev/null | grep -q "active directory"; then
        echo -e "  ${YELLOW}El dominio aún no está provisionado.${NC}"
        echo "  Abre http://$PRIMARY_IP:8444 en tu navegador"
        echo "  y usa el asistente de instalación web."
    else
        echo "  Dominio:     $(testparm --parameter-name=realm -s 2>/dev/null)"
    fi
    echo ""
    echo "  Logs:         journalctl -u sambaforge -f"
    echo "  Restart:       systemctl restart sambaforge"
    echo ""
}

# --- Main ---
main() {
    echo ""
    echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}  SambaForge — Instalador interactivo${NC}"
    echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
    echo ""

    check_root
    detect_os
    check_virt
    check_network
    check_ip_static
    fix_hosts
    check_hostname
    stop_conflicts
    install_samba
    stop_standalone_samba
    clean_dbs
    check_ports
    install_go
    install_node
    install_sambaforge
    summary
}

main "$@"