# Samba — Referencia técnica para SambaForge

> Documento de referencia generado el 2026-09-09. Debe revisarse antes de cada fase del proyecto.

## Versiones de Samba (septiembre 2026)

| Serie | Última versión | Estado | Fecha release |
|---|---|---|---|
| **4.24.x** | **4.24.6** | ✅ Stable (latest) | 2026-08-13 |
| 4.23.x | 4.23.12 | 🟡 Maintenance mode | 2026-09-01 |
| 4.22.x | 4.22.11 | 🟡 Security fixes only | 2026-07-28 |
| 4.25.x | 4.25.0rc1 | 🔴 Release candidate | 2026-08-11 |

**Recomendación:** SambaForge debe desarrollar contra **4.24.x** (stable latest) y soportar **4.23.x** (maintenance). No soportar 4.22 ni anteriores (security-only o EOL).

## Cambios clave por versión que afectan a SambaForge

### Samba 4.24 (marzo 2026) — Nuestra versión target

Nuevos features relevantes:
- **Audit de autenticación**: Logging estructurado de cambios en atributos críticos (servicePrincipalName, dNSHostName, msDS-KeyCredentialLink). SambaForge puede integrar este log en el módulo de auditoría.
- **Password management remoto (Entra ID SSPR, Keycloak)**: Permite resets de password desde sistemas externos respetando políticas del dominio. SambaForge puede usar esto para el self-service portal.
- **Kerberos PKINIT SID extension**: Soporte de certificados con SID extension. SambaForge debe soportar gestión de certificados.
- **KDC incluye PAC por defecto**: Cambio de comportamiento. `kdc always generate pac = no` para desactivar.
- **`kdc require canonicalization`**: Nuevo smb.conf para prevenir "dollar ticket attack".
- **AES default encryption**: AES-128/AES-256 ahora default para dominios functional level 2008+. Hay que detectar y respetar esto en el provisioning.
- **Nuevos subcomandos `samba-tool`**: `keytrust`, `generate-csr` (para computer y user).
- **smb.conf changes**: `strong certificate binding enforcement` (new), `kdc require canonicalization` (new), `kdc name match implicit dollar without canonicalization` (new).

### Samba 4.23 (septiembre 2025)

- DNS scavenging improvements (BUG 16223: scavenging ocurre incluso si fAging=FALSE — bug corregido en 4.23.12).
- DNS internal: ahora maneja switch UDP→TCP cuando paquete > 4k (BUG 15988).
- Nuevos parámetros smb.conf: `smbd profiling share`, `client smb transports`, `server smb transports`, `winbind varlang service`.
- BUG 15905: samba 4.21 falla al unirse a AD cuando múltiples DCs responden — corregido. Importante para Fase 7 (multi-DC).

### Samba 4.22 (abril 2025)

- **Microsoft Netlogon hardening**: Cambios en RPC Netlogon que afectan member servers con `ad` idmapping backend. SambaForge debe documentar esto.
- AD DC schema upgrade y provision más rápidos (LDB index cache más grande).
- Nuevos parámetros: `smb3 directory leases`, `client netlogon ping protocol` (new default: cldap), `himmelblaud hello enabled`, `himmelblaud hsm pin path`, `himmelblaud sfa fallback`.
- Experimentales: `client use krb5 netlogon`, `reject aes netlogon servers`, `server reject aes schannel`, `server support krb5 netlogon`.
- REMOVIDO: `cldap port` (CLDAP siempre 389 UDP), `fruit:posix_rename`, `nmbd proxy logon`.

## `samba-tool` — Mapa completo de subcomandos

Fuente: man page oficial samba-tool.8 (versión 4.23.0). Documentación completa en:
https://www.samba.org/samba/samba/docs/current/man-html/samba-tool.8.html

### Clasificación por módulo de SambaForge

| Módulo SambaForge | Subcomando samba-tool | Notas |
|---|---|---|
| **Provisioning** | `domain provision` | `--realm`, `--domain`, `--server-role=dc`, `--dns-backend`, `--adminpass`, `--use-rfc2307` |
| **Provisioning** | `domain classicupgrade` | Migrar NT4 → AD (no-goal inicial) |
| **Provisioning** | `domain join` | Unir DC adicional (Fase 7) |
| **Provisioning** | `domain demote` | Eliminar DC (Fase 7) |
| **Usuarios** | `user create` | Soporta `--given-name`, `--surname`, `--initials`, `--profile-path`, `--script-path`, `--home-drive`, `--home-directory` |
| **Usuarios** | `user delete` | |
| **Usuarios** | `user list` | Output JSON con `--json` |
| **Usuarios** | `user disable` / `user enable` | |
| **Usuarios** | `user setpassword` | Reset password |
| **Usuarios** | `user password` | Cambiar propio password (requiere auth) |
| **Usuarios** | `user setexpiry` | Expiración de cuenta |
| **Usuarios** | `user rename` | Cambiar CN, given-name, surname |
| **Usuarios** | `user edit` | Editor arbitrario (vi por defecto) |
| **Usuarios** | `user addunixattrs` | Añadir RFC2307 (UID, GID, shell, home) |
| **Usuarios** | `user sensitive` | UF_NOT_DELEGATED on/off/show |
| **Grupos** | `group add` / `group create` | |
| **Grupos** | `group delete` | |
| **Grupos** | `group list` | |
| **Grupos** | `group listmembers` | |
| **Grupos** | `group addmembers` | |
| **Grupos** | `group removemembers` | |
| **Grupos** | `group rename` | |
| **Computadoras** | `computer add` / `computer create` | |
| **Computadoras** | `computer delete` | |
| **Computadoras** | `computer list` | |
| **Computadoras** | `computer move` | |
| **OUs** | `ou add` / `ou create` | |
| **OUs** | `ou delete` | |
| **OUs** | `ou list` | |
| **OUs** | `ou move` | |
| **OUs** | `ou rename` | |
| **Contactos** | `contact add` / `contact create` | |
| **Contactos** | `contact delete` | |
| **Contactos** | `contact list` | |
| **Contactos** | `contact move` | |
| **Contactos** | `contact rename` | |
| **DNS** | `dns add` | A, AAAA, PTR, CNAME, NS, MX, SRV, TXT |
| **DNS** | `dns delete` | |
| **DNS** | `dns query` | |
| **DNS** | `dns roothints` | |
| **DNS** | `dns serverinfo` | |
| **DNS** | `dns update` | |
| **DNS** | `dns zonecreate` | |
| **DNS** | `dns zonedelete` | |
| **DNS** | `dns zoneinfo` | |
| **DNS** | `dns zonelist` | |
| **GPO** | `gpo create` (implícito en `gpo add`) | |
| **GPO** | `gpo del` | |
| **GPO** | `gpo listall` | |
| **GPO** | `gpo getinheritance` / `gpo setinheritance` | |
| **GPO** | `gpo getlink` / `gpo setlink` | |
| **GPO** | `gpo list` | GPOs para una cuenta |
| **GPO** | `gpo listcontainers` | Contenedores linkeados a un GPO |
| **GPO** | `gpo load` | Cargar policy en GPO |
| **GPO** | `gpo manage smb_conf list/set` | Gestionar smb.conf via GPO |
| **GPO** | `gpo manage sudoers add` | Sudoers via GPO |
| **GPO** | `gpo manage firewall` | Firewall via GPO |
| **GPO** | `gpo cse register/unregister` | Client Side Extensions |
| **Políticas de password** | `domain passwordsettings set` | `--complexity`, `--store-plaintext`, `--history-length`, `--min-pwd-length`, `--min-pwd-age`, `--max-pwd-age`, `--account-lockout-duration`, `--account-lockout-threshold`, `--reset-account-lockout` |
| **Políticas de password** | `domain passwordsettings pso create/delete/list/set` | Fine-grained PSO |
| **Políticas de password** | `domain passwordsettings pso apply` | Aplicar PSO a usuario/grupo |
| **Delegación** | `delegation add-service` / `delegation del-service` | |
| **Delegación** | `delegation for-any-protocol` / `delegation for-any-service` | |
| **Delegación** | `delegation show` | |
| **FSMO** | `domain fsmo show` / `domain fsmo transfer` / `domain fsmo seize` | Fase 7 |
| **Trusts** | `domain trust create` / `domain trust delete` | Fase 7 |
| **Trusts** | `domain trust list` / `domain trust show` / `domain trust validate` | |
| **Replicación** | `domain backup online` / `domain backup offline` | Fase 6 |
| **Replicación** | `domain backup restore` | |
| **Replicación** | `visualize ntdsconn` / `visualize reps` / `visualize uptodateness` | Monitoreo replicación |
| **Auth policies** | `domain auth policy ...` | Authentication policies (4.24+) |
| **Auth silos** | `domain auth silo ...` | Authentication silos (4.24+) |
| **Claims** | `domain claim claim-type create/modify` | Claims (4.24+) |
| **KDS** | `domain kds root-key create/list/view/delete` | gMSA |
| **DB check** | `dbcheck` / `dbcheck --cross-ncs` | Integridad de DB |
| **Sites** | `domain sites ...` | Sites y subnets (replication topology) |
| **SPN** | `spn add` / `spn delete` / `spn list` | Service Principal Names |
| **Time** | `time` | Hora del servidor |

### Output JSON

`samba-tool` soporta `--json` en varios subcomandos. Esto es CRÍTICO para SambaForge: debemos parsear JSON, no texto. Verificar qué subcomandos soportan `--json` en la Fase 0.

### Seguridad de credenciales en samba-tool

> "Be cautious about including passwords in scripts or passing user-supplied values onto the command line. For security it is better to let the Samba client tool ask for the password if needed, or obtain the password once with kinit."

> "While Samba will attempt to scrub the password from the process title (as seen in ps), this is after startup and so is subject to a race."

**Implicación para SambaForge:** NO pasar passwords como argumentos de línea de comandos. Usar:
- Variable de entorno `PASSWD` (samba-tool la lee)
- Kerberos ticket cache (`kinit` previo, `--use-kerberos=required`)
- stdin interactivo (piping)

## Provisioning — Pasos oficiales (SambaWiki)

Fuente: https://wiki.samba.org/index.php/Setting_up_Samba_as_an_Active_Directory_Domain_Controller

### Pre-requisitos
1. DNS domain elegido (no `.local`, no TLD principal de la organización)
2. Hostname < 15 caracteres (NetBIOS limit), no usar `PDC`/`BDC`
3. IP estática configurada
4. `/etc/hosts`: FQDN → LAN IP (no 127.0.0.1)
5. `/etc/resolv.conf`: nameserver = 127.0.0.1, search = dominio AD
6. Deshabilitar `resolvconf`/`systemd-resolved` que auto-modifican resolv.conf
7. Eliminar `smb.conf` previo: `mv /etc/samba/smb.conf /etc/samba/smb.conf.initial`
8. Eliminar DBs previos: `*.tdb`, `*.ldb` en LOCKDIR, STATEDIR, CACHEDIR, PRIVATE_DIR
9. Deshabilitar `avahi-daemon` (conflicto con DNS de Samba)

### Comando de provisioning (no-interactivo)
```bash
samba-tool domain provision \
  --server-role=dc \
  --use-rfc2307 \
  --dns-backend=SAMBA_INTERNAL \
  --realm=SAMDOM.EXAMPLE.COM \
  --domain=SAMDOM \
  --adminpass='Passw0rd'
```

### Post-provisioning
1. Configurar `/etc/resolv.conf`: `nameserver 127.0.0.1`, `search samdom.example.com`
2. Copiar krb5.conf: `cp /usr/local/samba/private/krb5.conf /etc/krb5.conf`
3. Iniciar servicio Samba: `samba` (o systemd unit)
4. Crear zona reversa: `samba-tool dns zonecreate <server> <zone>`
5. Verificar:
   - `samba-tool domain level show`
   - `kinit administrator@SAMDOM.EXAMPLE.COM`
   - `host -t SRV _ldap._tcp.samdom.example.com`
   - `smbclient -L //localhost -U%`

### Notas críticas
- **No provisionar un segundo DC en el mismo dominio.** Usar `samba-tool domain join`.
- **No renombrar dominio DNS ni realm Kerberos.** Samba no lo soporta.
- **BIND9_FLATFILE será removido.** Solo usar `SAMBA_INTERNAL` o `BIND9_DLZ`.
- **No usar `.local` como TLD.** Conflicto con Avahi/mDNS.
- **El DC debe ser dedicado a autenticación.** File/Print en member servers.
- **GPO support no se habilita automáticamente en modo interactivo.** Hay que editar smb.conf después.

## CVEs recientes relevantes (2026)

| CVE | Descripción | Versión afectada | Fix en |
|---|---|---|---|
| CVE-2026-58221 | AD LDAP domain takeover | < 4.24.5/4.23.10/4.22.11 | 4.24.5+ |
| CVE-2026-58222 | LDAP Compare filter injection | < 4.24.5/4.23.10/4.22.11 | 4.24.5+ |
| CVE-2026-6949 | TSIG DNS crash | < 4.24.5/4.23.10/4.22.11 | 4.24.5+ |
| CVE-2026-58218 | DNS TKEY cache exhaustion DoS | < 4.24.5/4.23.10/4.22.11 | 4.24.5+ |
| CVE-2026-3238 | WINS server DoS (NULL ptr) | < 4.24.3/4.23.8/4.22.10 | 4.24.3+ |
| CVE-2026-3012 | GPO CA cert over HTTP sin verif. | < 4.24.3/4.23.8/4.22.10 | 4.24.3+ |
| CVE-2026-4408 | RCE en DCE/RPC SAMR | < 4.24.3/4.23.8/4.22.10 | 4.24.3+ |

**Implicación:** SambaForge debe verificar la versión de Samba instalada y alertar si está por debajo de 4.24.5.

## Detección de versión y paths

SambaForge debe detectar automáticamente:

```bash
# Versión
samba --version | head -1
# o
samba-tool --version

# Paths (self-compiled vs distro)
smbd -b | grep "CONFIGFILE"     # /usr/local/samba/etc/samba/smb.conf o /etc/samba/smb.conf
smbd -b | egrep "LOCKDIR|STATEDIR|CACHEDIR|PRIVATE_DIR"

# Rol actual
testparm -s --section=global 2>/dev/null | grep "server role"
# o
samba-tool domain level show
```

## Links de referencia oficiales

- SambaWiki principal: https://wiki.samba.org/
- Setup AD DC: https://wiki.samba.org/index.php/Setting_up_Samba_as_an_Active_Directory_Domain_Controller
- samba-tool manpage: https://www.samba.org/samba/samba/docs/current/man-html/samba-tool.8.html
- Release history: https://www.samba.org/samba/history/
- Samba 4.24 features: https://samba.org/samba/history/samba-4.24.0.html
- Samba 4.23 features: https://wiki.samba.org/index.php/Samba_4.23_Features_added/changed
- Samba 4.22 features: https://wiki.samba.org/index.php/Samba_4.22_Features_added/changed
- Release planning: https://wiki.samba.org/index.php/Samba_Release_Planning
- Security CVEs: https://www.samba.org/samba/security/