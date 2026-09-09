# Análisis de proyectos existentes — Lecciones para SambaForge

> Documento de referencia generado el 2026-09-09. Debe completarse con lectura de código fuente en Fase 0.

## Metodología de análisis

Para cada proyecto se evalúa:
- **Qué hace bien** → tomar como patrón
- **Qué hace mal o le falta** → corregir en SambaForge
- **Arquitectura** → qué stack, cómo organiza el código
- **Patrones reutilizables** → wrappers, LDAP, auth, UI
- **Estado real** → activo, abandonado, WIP, producción

---

## 1. cockpit-samba-ad-dc (GSoC 2020 — Samba Team)

**Repo:** No es un repo standalone. Es un plugin de Cockpit desarrollado en GSoC 2020.
**Mentor:** Alexander Bokovoy (Samba Team, Red Hat).
**Estado:** Prototipo, estancado desde 2020. Pero con cobertura más completa de `samba-tool`.
**Stack:** JavaScript + Cockpit framework (PatternFly CSS).

### Cobertura de samba-tool (LA MÁS COMPLETA)
- ✅ Provisioning AD DC (con detección de server role)
- ✅ Computer Management (create/delete/list/show/move)
- ✅ Contact Management
- ✅ Delegation Management
- ✅ DNS Management (zones, records, roothints, serverinfo)
- ✅ Domain Management (provisioning, demote, join, classicupgrade, trusts, backup)
- ✅ Forest Management (DSHeuristics, FSMO)
- ✅ FSMO Management (seize, show, transfer)
- ✅ GPO Management (create/delete/backup/restore/link/inheritance/list)
- ✅ Group Management (create/delete/list/members/move/show)
- ✅ OU Management (create/delete/list/move/rename)
- ✅ Sites Management (create/delete/list/subnets)
- ✅ SPN Management
- ✅ User Management
- ✅ DS ACLs manipulation (get/modify access lists on directory objects)
- ✅ NT ACLs manipulation (get/set ACLs on files, sysvol ACL reset)
- ✅ Server Time

### Qué tomar
- **El mapeo completo de subcomandos de samba-tool** es nuestra checklist de features.
- El patrón de "verificar smb.conf → detectar server role → si no es DC, lanzar wizard de provisioning".
- La integración de DS ACLs y NT ACLs (ningún otro proyecto web lo hace).

### Qué corregir
- Independizar de Cockpit (Cockpit limita el UI y requiere instalación separada).
- Actualizar a samba-tool 4.24+ (el plugin usa APIs de 4.11).
- UI moderna (PatternFly es funcional pero feo y limitado).

---

## 2. Samba Conductor (edimarlnx)

**Repo:** https://github.com/edimarlnx/samba-conductor
**Stars:** 5 | **Forks:** 1 | **Creado:** 2026-03-25
**Estado:** 🟡 WIP — "not yet ready for production use"
**Stack:** React + Tailwind + Docker. (Backend no claramente especificado, parece Meteor por los topics.)

### Features
- ✅ Users, Groups, OUs, Computers
- ✅ DNS Management
- ✅ GPO Management
- ✅ gMSA (Service Accounts)
- ✅ OAuth2 Server (Authorization Code flow)
- ✅ Self-service portal (password change, profile edit)
- ✅ Disaster Recovery (encrypted AD backups to S3)
- ✅ DC Replication (automatic replica setup via env vars)
- ✅ Mobile-first responsive design
- ✅ Theme switching (Wine, Classic, Light)
- ✅ Zero stored credentials (per-session encrypted credentials)
- ✅ i18n

### Qué tomar
- **Zero stored credentials** — patrón de seguridad crítico. Credenciales por sesión, nunca en disco.
- **OAuth2 server** — para integraciones de terceros (Grafana, Portainer).
- **Disaster recovery con S3** — backup almacenamiento cloud.
- **Self-service portal** — usuarios cambian su propio password.
- **Themes** — dark/light/wine para UX.

### Qué corregir
- Estabilizar (están WIP).
- Backend poco claro (Meteor no es ideal para un DC — trae runtime pesado).
- Docker-only, sin opción de binario nativo.

---

## 3. Vexa (griffinwebnet)

**Repo:** https://github.com/griffinwebnet/Vexa
**Stars:** 2 | **Commits:** 156 | **Creado:** 2025-10-01
**Estado:** 🟢 Activo
**Stack:** Go (backend REST) + React/Vite/TypeScript/Tailwind (frontend) + Headscale (mesh networking)

### Features
- ✅ One-click domain provisioning wizard
- ✅ PAM authentication (login con credenciales del sistema o del dominio)
- ✅ Users, Groups, OUs management
- ✅ Computer management con deployment scripts (PowerShell offline)
- ✅ DNS management con split DNS para mesh
- ✅ Light & Dark mode
- ✅ Mesh networking (Headscale/Tailscale)
- ✅ Bootstrap script + update script

### Qué tomar
- **Stack Go + React** — valida nuestra elección tecnológica.
- **One-click provisioning wizard** — UX del asistente de instalación.
- **PAM authentication** — alternativa a LDAP bind para el login inicial (cuando el dominio aún no existe).
- **Computer deployment scripts** — generar scripts de join offline.
- **Bootstrap script** — `bootstrap.sh` para instalación one-line.

### Qué corregir
- **Quitar Headscale/Tailscale** — añade complejidad innecesaria para la mayoría de deployments.
- **Mejorar cobertura de samba-tool** — Vexa cubre poco vs cockpit-samba-ad-dc.
- **Sin GPO, sin FSMO, sin trusts, sin backup** — todo eso falta.

---

## 4. go-samba4 (jniltinho)

**Repo:** https://github.com/jniltinho/go-samba4
**Stars:** 7 | **Commits:** 19 | **Creado:** 2026-03-07
**Estado:** 🟡 En desarrollo (v1.1.2)
**Stack:** Go 1.26+ + Echo + GORM + Tailwind + jQuery + DataTables + Lucide Icons + go-ldap + gokrb5

### Features
- ✅ Users/Groups CRUD (LDAP)
- ✅ OU tree navigation
- ✅ LDAP bind authentication
- ✅ Kerberos SSO support
- ✅ 2FA TOTP
- ✅ i18n (pt_BR, en, es) — incluyendo DataTables localization
- ✅ CLI tooling (Cobra/Viper)
- ✅ Docker support
- ✅ Config TOML

### Qué tomar
- **Validación del stack Go+LDAP+Kerberos** — mismo stack que proponemos.
- **go-ldap y gokrb5** — las librerías correctas, ya validadas en producción por go-samba4.
- **2FA TOTP** — implementación Go de TOTP.
- **i18n es/en/pt** — mismo enfoque multi-idioma.
- **Cobra CLI** — para comando `sambaforge serve`, `sambaforge migrate`, etc.
- **DataTables para tablas** — útil para listados de usuarios/grupos.

### Qué corregir
- **Frontend SSR con html/template + jQuery** — limitante. SambaForge usará React SPA.
- **Solo LDAP CRUD** — sin provisioning, sin GPO, sin DNS, sin backup, sin multi-DC.
- **Sin RBAC** — un solo nivel de admin.

---

## 5. samba4-manager (stgraber)

**Repo:** https://github.com/stgraber/samba4-manager
**Stars:** 46 | **Forks:** 13 | **Creado:** 2015-09-20
**Estado:** 🔴 Sin mantenimiento (último commit ~2015)
**Stack:** Python Flask + python-ldap + WTForms

### Features
- ✅ Operaciones de directorio comunes (users, groups)
- ✅ Autenticación con credenciales del propio usuario (LDAP bind)
- ✅ Docker support

### Qué tomar
- **LDAP bind con credenciales del usuario** — patrón de auth: el usuario se autentica con sus propias credenciales del dominio, no con un service account.
- **Simplicidad Flask** — el proyecto más simple que funciona. Lección: no sobreingeniar.

### Qué corregir
- **Abandonado** — sin soporte para Samba 4.20+.
- **Sin provisioning** — asume el dominio ya creado.
- **Sin GPO, DNS, backup, multi-DC**.
- **UI antigua** — Flask WTForms no es una UI moderna.

---

## 6. samba-admin-webapp (boonkerz)

**Repo:** https://github.com/boonkerz/samba-admin-webapp
**Estado:** Disponible
**Descripción:** Setup wizard que provisiona AD DC en bare Debian/Ubuntu, luego da web UI que replica RSAT (ADUC, GPO).

### Qué tomar
- **Setup wizard end-to-end** — de bare metal a dominio funcionando.
- **Replica de ADUC** — el objetivo de parecerse a las herramientas de Windows.

---

## 7. AdamVenn/Samba-AD-DC-Controller

**Repo:** https://github.com/AdamVenn/Samba-Active-Directory-Domain-Controller-Controller
**Stars:** 2 | **Estado:** Simple
**Stack:** Python wxPython + Paramiko (SSH)

### Qué tomar
- **Patrón SSH remoto** — gestiona el DC remotamente via SSH. (SambaForge es local, pero el patrón de envolver `samba-tool` via exec es el mismo.)

---

## 8. Zentyal (comercial)

**URL:** https://zentyal.com
**Estado:** 🟢 Producción, soporte pago
**Licencia:** Propietaria con componentes open-source

### Features integrales
- Samba4 AD DC + DNS + Kerberos + GPO
- File shares, mail, VPN, firewall
- Web UI completa
- Actualizaciones comerciales

### Qué tomar
- **Visión integral** — un panel que cubre todo (DNS, Kerberos, GPO, users, groups, computers).
- **UX de Windows Server** — familiar para admins que vienen de Windows.

### Qué corregir
- **Es comercial** — SambaForge será open-source puro.
- **Monolítico** — intenta hacer mail+VPN+firewall. SambaForge se focaliza en AD DC.

---

## 9. NethServer 8

**Repo:** https://github.com/NethServer/ns8-samba
**Estado:** 🟢 Producción activa
**Stack:** Podman containers + Python/Go modules + Web UI

### Features
- ✅ Multi-DC (múltiples dominios en un cluster)
- ✅ Provisioning modes: new-domain, join-domain, join-member
- ✅ LDAP proxy (Ldapproxy)
- ✅ File server integration
- ✅ Cluster management

### Qué tomar
- **Provisioning modes** — `new-domain` (crear DC), `join-domain` (unir DC), `join-member` (unir file server). SambaForge puede usar el mismo modelo conceptual.
- **LDAP proxy** — un proxy local para que apps externas consulten LDAP via TLS/GSSAPI sin acceso directo al DC.

### Qué corregir
- **Complejidad de cluster** — NethServer 8 es un sistema completo de cluster. SambaForge será single-server.
- **Sin gestión de GPOs desde la web** — remite a `samba-tool` por CLI.

---

## 10. Univention Corporate Server (UCS)

**URL:** https://www.univention.com
**Estado:** 🟢 Producción, soporte pago
**Stack:** Debian-based + OpenLDAP + Samba4 + S4 Connector + UMC (Univention Management Console)

### Features
- ✅ Samba AD DC via App Center
- ✅ S4 Connector (sincroniza OpenLDAP ↔ Samba LDAP)
- ✅ AD Connector (sincroniza con Windows AD existente)
- ✅ Management Console web completa
- ✅ App Center (instalación de apps con un click)

### Qué tomar
- **S4 Connector** — el patrón de sincronizar dos directorios (OpenLDAP y Samba LDAP). Si SambaForge alguna vez necesita integrar con un LDAP externo.
- **App Center** — idea de marketplace de apps instalables desde el panel.

### Qué corregir
- **Demasiada abstracción** — UCS tiene tres capas (UMC → S4 Connector → Samba LDAP). SambaForge hablará directo a Samba LDAP.
- **Es comercial** — la versión community tiene limitaciones.

---

## Matriz de features: comparación

| Feature | cockpit | S.Conductor | Vexa | go-samba4 | s4-manager | Zentyal | NethServer | SambaForge (objetivo) |
|---|---|---|---|---|---|---|---|---|
| Provisioning wizard | ✅ | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ |
| Users CRUD | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Groups CRUD | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| OUs | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ✅ |
| Computers | ✅ | ✅ | ✅ | ❌ | ❌ | ✅ | ❌ | ✅ |
| DNS management | ✅ | ✅ | ✅ | ❌ | ❌ | ✅ | ❌ | ✅ |
| GPO management | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ |
| GPO editor | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ (plantillas) |
| Password policy | ✅ | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ | ✅ |
| PSO (fine-grained) | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ |
| FSMO roles | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ |
| Trusts | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ |
| Multi-DC join | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ |
| Replication monitor | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ |
| Backup/Restore | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ |
| DS ACLs | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| NT ACLs/SYSVOL | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| SPN management | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Auth policies (4.24) | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Auth silos (4.24) | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Claims (4.24) | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Self-service portal | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| OAuth2 server | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| 2FA TOTP | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ |
| RBAC | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ |
| Audit log | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ |
| i18n es/en/pt | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ | ✅ | ✅ |
| Dark mode | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Zero stored creds | ❌ | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ✅ |
| Docker deployment | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ |
| Native binary | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ |
| Import CSV | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ |
| Kerberos SSO | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ | ✅ | ✅ |

**SambaForge cubrirá TODAS las features** que ningún proyecto individual cubre por completo.