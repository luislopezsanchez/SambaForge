# ADR-010: Preflight checks con auto-remediación interactiva

**Estado:** ✅ Aceptado  
**Fecha:** 2026-09-09

## Contexto

Durante la Fase 1, al verificar la VM de prueba (172.30.36.91), se detectaron manualmente varios problemas que impedirían un provisioning correcto de Samba AD DC:

1. **IP en DHCP** — un DC requiere IP estática. La VM tenía IP dinámica asignada por Proxmox.
2. **/etc/hosts incorrecto** — `127.0.1.1` apuntaba al FQDN en vez de la IP real de la LAN.
3. **systemd-resolved** — podía interferir con el DNS de Samba (estaba inactivo, pero había que verificarlo).
4. **avahi-daemon** — conflicta con el DNS interno de Samba (estaba inactivo, pero había que verificarlo).
5. **Puertos** — 53, 88, 389, 445, 464 debían estar libres.
6. **Samba previo** — no debía haber smb.conf ni bases de datos de instalaciones anteriores.

Estas verificaciones se hicieron manualmente via SSH. Pero cuando un usuario final despliega SambaForge en una VM recién instalada, no tiene a nadie que haga estas verificaciones por él. **El propio SambaForge debe hacerlas, informar al usuario, y corregir lo que se pueda corregir automáticamente.**

## Decisión

SambaForge tendrá un **módulo de preflight checks** que se ejecuta:

1. **Al instalar SambaForge** (script de instalación) — verifica que el servidor cumple los requisitos mínimos.
2. **Antes del provisioning del dominio** (wizard web) — verifica que el servidor está listo para ser un DC.
3. **Como endpoint API** — `GET /api/server/preflight` — para que el dashboard muestre el estado del servidor en tiempo real.

### Principios del preflight

1. **Verificar, no asumir.** Cada check prueba el estado real del servidor, no supone nada.
2. **Auto-remediar cuando sea seguro.** Si se puede corregir sin riesgo, se corrige automáticamente.
3. **Preguntar cuando se necesita input del usuario.** Si la corrección requiere una decisión (ej: ¿qué IP estática quieres?), se pregunta.
4. **Explicar el "por qué".** Cada check explica por qué es necesario y qué pasa si no se cumple.
5. **No continuar si hay checks críticos sin resolver.** El provisioning no puede proceder si la IP es DHCP o si los puertos están ocupados.

### Lista de checks del preflight

| # | Check | Crítico | Auto-remediable | Cómo |
|---|---|---|---|---|
| P-01 | **OS soportado** (Debian 13+ o Ubuntu 24.04+) | ✅ | ❌ | Detectar via `/etc/os-release`. Si no es soportado, abortar con mensaje claro. |
| P-02 | **Samba instalada** y versión ≥ 4.22 | ✅ | ✅ | Si no instalada, ofrecer instalar via apt. Si versión vieja, ofrecer upgrade. |
| P-03 | **IP estática** (no DHCP) | ✅ | ✅ interactivo | Detectar via `ip route get 1.1.1.1` + `networkctl status`. Si DHCP, preguntar al usuario: "¿Qué IP estática deseas? (ej: 172.30.36.91/24)" y configurar via systemd-networkd o /etc/network/interfaces. |
| P-04 | **/etc/hosts correcto** | ✅ | ✅ | Verificar que el FQDN resuelve a la IP de la LAN (no 127.0.0.1 ni 127.0.1.1). Si incorrecto, reescribir automáticamente con la IP correcta. |
| P-05 | **/etc/resolv.conf** | ✅ | ✅ | Después del provisioning, debe apuntar a 127.0.0.1 (DNS de Samba). Antes, puede apuntar al gateway. Verificar que no haya entradas conflictivas. |
| P-06 | **Hostname ≤ 15 caracteres** | ✅ | ✅ interactivo | Verificar `hostname`. Si > 15 chars, preguntar nuevo hostname. Si contiene "PDC" o "BDC", advertir (términos NT4 deprecados). |
| P-07 | **avahi-daemon** | ✅ | ✅ | Si está activo, detenerlo y deshabilitarlo (`systemctl stop avahi-daemon && systemctl disable avahi-daemon`). Conflicta con DNS de Samba. |
| P-08 | **systemd-resolved** | ✅ | ✅ | Si está activo, detenerlo y deshabilitarlo. Su DNS stub en 127.0.0.53 conflictúa con el DNS de Samba que debe escuchar en 127.0.0.1:53. |
| P-09 | **dnsmasq** | ✅ | ✅ | Si está activo en puerto 53, detenerlo. Conflicta con el DNS interno de Samba. |
| P-10 | **smb.conf previo** | ✅ | ✅ interactivo | Si existe `/etc/samba/smb.conf`, preguntar: "Hay una configuración previa de Samba. ¿Backup + continuar, o cancelar?" Hacer backup a `smb.conf.initial` antes de proceder. |
| P-11 | **Bases de datos previas** (*.tdb, *.ldb) | ✅ | ✅ interactivo | Si existen DBs en /var/lib/samba/, preguntar: "Hay bases de datos de una instalación anterior. ¿Eliminar? (necesario para provisioning limpio)" |
| P-12 | **Puertos libres** (53, 88, 389, 445, 464) | ✅ | ❌ | Verificar via `ss -tlnp`. Si algún puerto está ocupado, reportar qué proceso lo usa. No se puede auto-remediar (el usuario debe decidir). |
| P-13 | **Kerberos krb5.conf** | ✅ | ✅ | Después del provisioning, copiar el krb5.conf generado por Samba a /etc/krb5.conf. Antes, verificar que no haya un realm configurado que conflicte. |
| P-14 | **resolvconf / NetworkManager** | ✅ | ✅ | Si resolvconf o NetworkManager están activos, pueden sobreescribir /etc/resolv.conf. Deshabilitar o configurar para que no toquen resolv.conf. |
| P-15 | **Chrony/NTP** | ✅ | ✅ | Un DC debe servir hora. Si chrony no está instalado, instalarlo. Si está instalado, verificar configuración. |
| P-16 | **Go toolchain** (solo para build desde fuente) | ❌ | ✅ | Si se compila en el servidor, verificar Go ≥ 1.23. Si no, ofrecer descargar e instalar desde go.dev. |
| P-17 | **Espacio en disco** | ❌ | ❌ | Verificar ≥ 10 GB libres en /. Si no, advertir. |
| P-18 | **RAM suficiente** | ❌ | ❌ | Verificar ≥ 2 GB. Si no, advertir. |

### Flujo del preflight interactivo

```
┌─────────────────────────────────────────────────────┐
│           SambaForge — Preflight Checks              │
├─────────────────────────────────────────────────────┤
│                                                     │
│  ✅ P-01  OS: Debian 13 (trixie) — soportado        │
│  ✅ P-02  Samba 4.22.10 instalada                    │
│  ⚠️ P-03  IP es DHCP (172.30.36.91/24)              │
│          ┌─────────────────────────────────┐        │
│          │  ¿Configurar IP estática?       │        │
│          │  IP: [172.30.36.91    ]         │        │
│          │  Máscara: [/24         ]        │
│          │  Gateway: [172.30.36.1]         │        │
│          │  [Cancelar] [Configurar]        │        │
│          └─────────────────────────────────┘        │
│  ⚠️ P-04  /etc/hosts: FQDN apunta a 127.0.1.1       │
│          → Corregir automáticamente                 │
│          [Corregir]                                  │
│  ✅ P-06  Hostname: SAMBA4 (6 chars) — OK            │
│  ✅ P-07  avahi-daemon: inactivo                     │
│  ✅ P-08  systemd-resolved: inactivo                 │
│  ✅ P-10  smb.conf: no existe                        │
│  ✅ P-12  Puertos 53,88,389,445,464: libres          │
│  ✅ P-15  chrony: instalado                          │
│                                                     │
│  2 checks requieren atención.                       │
│  [Resolver todos] [Continuar al provisioning]       │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### Implementación técnica

El preflight se implementa en dos niveles:

**1. Script de instalación (`install.sh`):**
- Se ejecuta antes de instalar SambaForge.
- Verifica P-01 (OS), P-16 (Go), P-17 (disco), P-18 (RAM).
- Instala Samba si no está (P-02).
- Configura IP estática si es necesario (P-03).
- Corrige /etc/hosts (P-04).
- Detiene servicios conflictivos (P-07, P-08, P-09).
- Instala chrony (P-15).
- Es interactivo: pregunta al usuario lo que necesita.

**2. API de preflight (`GET /api/server/preflight`):**
- Se ejecuta desde el wizard web antes del provisioning.
- Verifica P-03 a P-14 en tiempo real.
- Devuelve JSON con estado de cada check:
  ```json
  {
    "checks": [
      {"id": "P-03", "label": "IP estática", "status": "warning", "value": "DHCP", "remediable": true, "autoFix": false},
      {"id": "P-04", "label": "/etc/hosts", "status": "warning", "value": "127.0.1.1", "remediable": true, "autoFix": true},
      {"id": "P-12", "label": "Puertos libres", "status": "pass", "value": "all free", "remediable": false}
    ],
    "ready": false,
    "blockingCount": 2
  }
  ```
- El wizard web muestra los checks y permite resolverlos uno por uno.
- Los auto-remediables se corrigen con un click.
- Los que requieren input muestran un formulario.
- El botón "Continuar al provisioning" solo se habilita cuando `ready: true`.

### Detección de OS multi-distribución

```go
func detectOS() (distro string, version string, supported bool) {
    // Leer /etc/os-release
    // ID=debian → distro="debian", VERSION_ID="13"
    // ID=ubuntu → distro="ubuntu", VERSION_ID="24.04"
    // Supported: debian >= 13, ubuntu >= 24.04
}
```

### Detección de IP DHCP vs estática

```go
func checkIPStatic() (isStatic bool, currentIP string, err error) {
    // Método 1: networkctl status eth0 → muestra "DHCPv4: yes" o "Link File: /etc/systemd/network/*.network"
    // Método 2: ip route get 1.1.1.1 → obtiene IP de salida
    // Método 3: leer /etc/network/interfaces o /etc/systemd/network/*.network
    // Si la IP viene de DHCP (dhclient activo, networkctl dice DHCP), isStatic=false
}
```

### Corrección de /etc/hosts

```go
func fixEtcHosts(lanIP, hostname, fqdn string) error {
    // Backup /etc/hosts → /etc/hosts.bak
    // Reescribir con:
    //   127.0.0.1    localhost
    //   <lanIP>      <fqdn> <hostname>
    // Preservar otras líneas (IPv6, etc.)
}
```

### Configuración de IP estática

```go
func setStaticIP(ip, mask, gateway string) error {
    // Debian 13: usar systemd-networkd (.network file) o /etc/network/interfaces
    // Ubuntu 24.04+: usar netplan (.yaml file)
    // Detectar cuál método usa el sistema y escribir el config apropiado
    // Reiniciar red o aplicar via systemctl restart systemd-networkd / netplan apply
}
```

## Consecuencias

- El script de instalación (`install.sh`) es más complejo pero asegura que cualquier VM recién instalada funcione.
- El wizard web de provisioning tiene un paso de preflight antes del paso de configuración del dominio.
- El dashboard muestra el estado del servidor en tiempo real via `GET /api/server/preflight`.
- SambaForge soporta oficialmente Debian 13+ y Ubuntu 24.04+.
- Un usuario puede instalar una VM limpia, correr `install.sh`, y SambaForge detecta, corrige, y prepara todo automáticamente.