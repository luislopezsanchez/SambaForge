# SambaForge — Bitácora del proyecto

> **Regla de oro:** Al inicio de cada sesión, leer este archivo primero. Al final de cada sesión, actualizarlo.

---

## Metadatos del proyecto

| Campo | Valor |
|---|---|
| **Nombre** | SambaForge |
| **Repo GitHub** | https://github.com/luislopezsanchez/SambaForge |
| **Ruta local** | `C:\Users\llope\OneDrive\GOOGLE DRIVE\21-SambaForge` |
| **Licencia** | MIT |
| **Stack** | Go (backend) + React/TS (frontend) |
| **Arquitectura** | Local-only (instalado en el mismo servidor DC) |
| **Versión Samba target** | 4.24.x (stable), 4.23.x (maintenance) |
| **Fecha inicio** | 2026-09-09 |

---

## Fase actual

**Fase 0 — Investigación y análisis técnico**

---

## Tabla de fases

| Fase | Estado | Inicio | Fin |
|---|---|---|---|
| 0 — Investigación y análisis | 🟡 En progreso | 2026-09-09 | — |
| 1 — Fundaciones del proyecto | ⏳ Pendiente | — | — |
| 2 — Spike provisioning | ⏳ Pendiente | — | — |
| 3 — MVP gestión básica | ⏳ Pendiente | — | — |
| 4 — Gestión avanzada | ⏳ Pendiente | — | — |
| 5 — DNS + GPO | ⏳ Pendiente | — | — |
| 6 — Seguridad y auditoría | ⏳ Pendiente | — | — |
| 7 — Backup y DR | ⏳ Pendiente | — | — |
| 8 — Multi-DC | ⏳ Pendiente | — | — |
| 9 — Diferenciadores | ⏳ Pendiente | — | — |

---

## Sesiones de trabajo

### Sesión 1 — 2026-09-09: Inicio del proyecto

**Objetivo:** Investigar viabilidad, analizar proyectos existentes, definir plan de desarrollo, crear repo.

**Acciones realizadas:**
1. Investigación web de proyectos existentes que gestionan Samba4 AD DC desde web UI
2. Encontrados 10 proyectos: cockpit-samba-ad-dc, Samba Conductor, Vexa, go-samba4, samba4-manager, samba-admin-webapp, EasyDC, Samba-AD-DC-Controller, Zentyal, NethServer, UCS
3. Análisis comparativo de features — matriz 30+ features × 8 proyectos
4. Estudio de versiones de Samba: 4.24.6 (stable), 4.23.12 (maintenance), 4.25.0rc1
5. Mapeo completo de ~60 subcomandos de `samba-tool` clasificados por módulo
6. Estudio de changelogs 4.22/4.23/4.24 — identificados breaking changes y features nuevos
7. Documentación de CVEs 2026 relevantes y versiones mínimas seguras
8. Estudio del flujo oficial de provisioning (SambaWiki)
9. Definición de stack tecnológico: Go + React (con justificación)
10. Creación del plan de desarrollo (10 fases con gates de aprobación)
11. Creación de repo GitHub: https://github.com/luislopezsanchez/SambaForge
12. Configuración de .gitignore, LICENSE, README
13. Push inicial al repo

**Archivos creados:**
- `docs/project-plan.md` — Plan completo de 10 fases
- `docs/reference/samba-reference.md` — Referencia técnica de Samba
- `docs/reference/existing-projects-analysis.md` — Análisis de 10 proyectos
- `README.md`, `LICENSE`, `.gitignore`

---

### Sesión 2 — 2026-09-09: Fase 0 profunda — investigación técnica

**Objetivo:** Ejecutar la Fase 0 completa: análisis de código fuente de los 4 repos principales, evaluación de librerías Go, wireframes UI, y cierre de todos los ADRs.

**Acciones realizadas:**
1. **Tarea 0.3 — Evaluación de librerías Go** ✅ Completada
   - go-ldap/ldap/v3: cliente LDAP, soporta GSSAPI bind, paginación, DirSync. Snippets documentados.
   - gokrb5/v8: Kerberos client, SPNEGO para SSO, ChangePasswd para self-service. Snippets documentados.
   - os/exec: 5 patrones seguros para envolver samba-tool (env var passwords, allowlist, timeout, JSON output, path resolution). Reglas de seguridad documentadas.
   - Echo vs Gin: **Echo seleccionado** (Auto-TLS, error handling central, middleware built-in, validado por go-samba4)
   - modernc.org/sqlite: **Seleccionado** (pure Go, sin CGO, cross-compilation trivial)
   - pquerna/otp: TOTP 2FA, compatible Google Authenticator. Snippet documentado.
   - golang-jwt/v5: JWT sesiones, HS256. Snippet documentado.
2. **Tarea 0.4 — Guía de estilo y wireframes** ✅ Completada
   - Filosofía de diseño: "Cockpit for Samba AD", dark mode first, densidad informativa
   - Paleta de colores completa (dark + light mode), accent indigo #6366f1
   - Tipografía: Inter + JetBrains Mono
   - Layout: Sidebar 240px + Topbar 56px + Status bar 32px
   - 6 wireframes: Login, Dashboard, Users list, Provisioning Wizard, DNS, GPO
   - i18n: react-i18next con namespaces por módulo, es/en/pt
   - Dependencias frontend definidas: Zustand, TanStack Query, react-hook-form+zod, React Router
3. **Tarea 0.5 — Cierre de ADRs** ✅ Completada (9 ADRs cerrados)
   - ADR-001: Stack Go+React
   - ADR-002: Arquitectura local-only
   - ADR-003: Auth model (LDAP bind + JWT + 2FA TOTP)
   - ADR-004: Seguridad (TLS, zero stored creds, RBAC, audit log)
   - ADR-005: Echo v4
   - ADR-006: modernc.org/sqlite
   - ADR-007: react-i18next
   - ADR-008: Caddy
   - ADR-009: Ejecución de samba-tool (sudoers restringido, PASSWD env var, allowlist)
4. **Tarea 0.2.9 — Análisis de código fuente** ✅ Completada (4 subagentes finalizaron)
   - Subagente 0: Vexa — Go+Gin+React, SSE streaming provisioning, doble auth Samba+PAM, bootstrap.sh. Limitaciones: no usa go-ldap/gokrb5, sanitizer permissive, CORS inseguro.
   - Subagente 1: go-samba4 — Go+Echo+go-ldap+GORM, LDAP wrapper auto-reconnect, password UTF-16LE, 2FA TOTP (no integrado), i18n gotext .po. Limitaciones: Kerberos no implementado (placeholder), 2FA código muerto, sin paginación LDAP, RBAC binario.
   - Subagente 2: Samba Conductor — Meteor 3.4+React+MongoDB, zero stored credentials (AES-256-GCM memoria TTL 30min), OAuth2 server, DR con PBKDF2+S3, 3 temas, execFile no shell, 3 modos ejecución samba-tool. Limitaciones: MongoDB obligatorio, Meteor DDP no REST, TLS rejectUnauthorized:false.
   - Subagente 3: cockpit-samba-ad-dc — 99 operaciones de samba-tool mapeadas en 17 grupos, cockpit.script/spawn con superuser:true, gate function testparm. Limitaciones: parseo texto ingenuo split('\n'), credenciales por pantalla, template string injection, estancado desde 2020.
   - Síntesis creada: 25 patrones a adoptar + 16 anti-patrones a evitar + checklist de 99 operaciones.

**Decisiones cerradas (esta sesión):**
- ADR-005: Echo v4 (antes P-001) ✅
- ADR-006: modernc.org/sqlite (antes P-002) ✅
- ADR-007: react-i18next (antes P-005) ✅
- ADR-008: Caddy (antes P-004) ✅
- ADR-009: Ejecución samba-tool con sudoers restringido (antes P-006) ✅
- UI-001 a UI-010: Decisiones de diseño visual ✅
- shadcn/ui confirmado como librería de componentes (antes P-003) ✅

**Decisiones pendientes:**
- (Ninguna — todos los ADRs originales están cerrados)

**Archivos creados (esta sesión):**
- `docs/reference/go-libraries-evaluation.md` — Evaluación completa de 7 librerías Go con snippets
- `docs/reference/ui-style-guide.md` — Guía de estilo, paleta, tipografía, 6 wireframes, i18n
- `docs/adr/adr-001-to-009.md` — 9 ADRs cerrados
- `docs/reference/code-analysis-vexa.md` — Análisis de código fuente de Vexa (42KB, 1048 líneas)
- `docs/reference/code-analysis-go-samba4.md` — Análisis de go-samba4 (53KB, 1572 líneas)
- `docs/reference/code-analysis-samba-conductor.md` — Análisis de Samba Conductor (50KB, 1346 líneas)
- `docs/reference/code-analysis-cockpit-samba.md` — Análisis de cockpit-samba-ad-dc (43KB, 975 líneas)
- `docs/reference/code-analysis-synthesis.md` — Síntesis: 25 patrones a adoptar + 16 anti-patrones a evitar + checklist 99 operaciones

### Sesión 3 — 2026-09-09: Fase 1 + ADR-010 Preflight

**Objetivo:** Scaffold del monorepo, compilación nativa en VM, y formalización del preflight interactivo.

**Acciones realizadas:**
1. Verificación de VM 172.30.36.91 (Debian 13 LXC, 4 vCPU, 4 GB RAM)
2. Instalación de herramientas base: git, make, curl, golang-go (Go 1.24.4)
3. Scaffold Go backend (Echo v4, /api/health, static file server)
4. Scaffold React frontend (Vite, TS, Tailwind, i18n es/en/pt, Login + Dashboard)
5. Compilación binario nativo: CGO_ENABLED=0, 6.2 MB estático
6. systemd service activo, 1.5 MB RAM
7. ADR-010: 18 preflight checks con auto-remediación interactiva

**Archivos creados:**
- `apps/api/go.mod`, `apps/api/main.go` — Backend Go + Echo
- `apps/web/` (14 archivos) — Frontend React
- `Makefile`, `deploy/sambaforge.service`, `.github/workflows/ci.yml`
- `docs/adr/adr-010-preflight.md`

---

### Sesión 4 — 2026-09-09: Fase 2+3+4 — Provisioning, Auth, CRUD completo

**Objetivo:** Provisionar dominio real, implementar auth LDAP, CRUD usuarios/grupos/DNS, OUs, password policy.

**Acciones realizadas:**
1. **Fase 2 — Spike provisioning:**
   - Primera VM (172.30.36.91) era LXC — samba-tool PANIC, samba-ad-dc no arranca. Doc: proxmox-lxc-limitations.md
   - Usuario creó VM KVM nueva (172.30.36.115, Debian 13, 4 vCPU, 3.8 GB RAM, 75 GB)
   - Instalación de Samba 4.22.10 + Go 1.24.4 en VM KVM
   - Corrección de /etc/hosts, IP estática, detención de smbd standalone
   - `POST /api/domain/provision` → dominio TELEDATA.LAB creado (Success: true)
   - Post-provision: krb5.conf, resolv.conf, matar winbindd stale, samba-ad-dc iniciado
   - Verificación completa: kinit OK, DNS SRV OK, DNS A OK, LDAP 389 OK, Kerberos 88 OK
   - 3 usuarios (Administrator, Guest, krbtgt), 38 grupos, 1 computadora (SAMBAFORGE$)

2. **Fase 3 — MVP gestión básica:**
   - auth.go: LDAP bind con STARTTLS, JWT HS256, middleware de autenticación
   - directory.go: CRUD usuarios (create/delete/password/disable/enable), grupos (create/delete/members), computers
   - dns.go: zones, records, forwarders (con -P para Kerberos)
   - 25+ endpoints API (auth, dashboard, users, groups, computers, dns)
   - Frontend actualizado: Login real con API, JWT en localStorage, axios interceptor
   - Dashboard con datos reales (usuarios, grupos, equipos, Samba version, realm)
   - Users page: tabla, búsqueda, crear usuario (modal), eliminar
   - Groups page: tabla con 38 grupos
   - Computers page: tabla con SAMBAFORGE$
   - ProtectedRoute para rutas autenticadas
   - **Verificado por el usuario:** login funciona, crear usuario funciona

3. **Fase 4 — Gestión avanzada (parcial):**
   - directory/ou.go: ListOUs, CreateOU, DeleteOU
   - directory/ou.go: GetPasswordPolicy, SetPasswordPolicy
   - 5 endpoints nuevos: GET/POST/DELETE /api/ous, GET/PUT /api/password-policy
   - Verificado: OU=Domain Controllers listada, password policy con complexity=on

**Dominio activo:**
- Realm: TELEDATA.LAB
- NetBIOS: TELEDATA
- DC: sambaforge.teledata.lab (172.30.36.115)
- Admin: administrator@TELEDATA.LAB / SambaForge@2026
- Samba: 4.22.10-Debian
- Functional level: Windows 2008 R2

**Archivos creados:**
- `apps/api/auth/auth.go` — LDAP bind + STARTTLS + JWT
- `apps/api/directory/directory.go` — CRUD usuarios/grupos/computers
- `apps/api/directory/ou.go` — OUs + password policy
- `apps/api/dns/dns.go` — DNS zones/records/forwarders
- `apps/web/src/lib/api.ts` — axios client con JWT interceptor
- `apps/web/src/stores/auth.ts` — Zustand auth store
- `apps/web/src/pages/Login.tsx` — Login real con API
- `apps/web/src/pages/Dashboard.tsx` — Dashboard con datos reales
- `apps/web/src/pages/Users.tsx` — CRUD usuarios
- `apps/web/src/pages/Groups.tsx` — Lista grupos
- `apps/web/src/pages/Computers.tsx` — Lista equipos
- `docs/reference/proxmox-lxc-limitations.md`

**Próximos pasos (Sesión 5):**
- Fase 4 restante: importación masiva CSV, atributos extendidos
- Fase 5: DNS avanzado (parser de zonelist corregir), GPOs con plantillas
- Fase 6: 2FA TOTP, RBAC, audit log, hardening
- Frontend: páginas de OUs, password policy, DNS

**Contexto para retomar:**
- SambaForge v0.3.0-dev corriendo en 172.30.36.115 (KVM, Debian 13)
- Dominio TELEDATA.LAB funcional con Samba AD DC 4.22.10
- Backend: 30+ endpoints API funcionando (auth, dashboard, users, groups, computers, ous, dns, password-policy, preflight, provision)
- Frontend: Login + Dashboard + Users + Groups + Computers funcionando
- 14 commits en GitHub, todo pusheado
- SSH VM: plink -hostkey "SHA256:pUuQplga4Gap3ZR4h5AHN9jTn3UU0QlMFCgDlQetubM" root@172.30.36.115
- Usuario verificó: login OK, crear usuario OK

**Riesgos identificados nuevos:**
- Samba AD no soporta clear-text LDAP binds — SambaForge debe usar STARTTLS o GSSAPI siempre
- `samba-tool` scrub de password en process title tiene race condition — usar PASSWD env var es obligatorio
- Subcomandos sin --json requerirán parseo de texto con regex — mapear cuáles en Fase 1

**Contexto para retomar:**
- **Fase 0 completa.** Todos los entregables están listos: 9 ADRs, 7 librerías Go evaluadas, 6 wireframes, guía de estilo, 4 análisis de código fuente + síntesis.
- 25 patrones a adoptar y 16 anti-patrones a evitar documentados en code-analysis-synthesis.md.
- Checklist de 99 operaciones de samba-tool a cubrir, mapeadas a fases de SambaForge.
- Total dependencias Go: 6 externas + stdlib, sin CGO.
- Total dependencias frontend: 8 (React, Vite, TS, Tailwind, shadcn/ui, Zustand, TanStack Query, react-i18next).
- Repos GitHub: https://github.com/luislopezsanchez/SambaForge — 5 commits, todo pusheado.
- Token de GitHub en credential helper (~/.git-credentials).
- Usuario GitHub: luislopezsanchez
- **Próxima sesión: revisar gate de Fase 0 con el usuario, luego iniciar Fase 1 (scaffold del proyecto).**

---

## Reglas para actualizar esta bitácora

1. **Al inicio de cada sesión:** Leer `docs/project-log.md` y `docs/project-plan.md` primero.
2. **Durante la sesión:** Actualizar el estado de tareas en el plan (⏳ → ✅).
3. **Al final de cada sesión:** Añadir una nueva entrada en "Sesiones de trabajo" con:
   - Objetivo de la sesión
   - Acciones realizadas
   - Decisiones cerradas/pendientes
   - Archivos creados/modificados
   - Próximos pasos
   - Riesgos nuevos
   - Contexto para retomar
4. **Al cerrar un ADR:** Crear documento en `docs/adr/` y marcar en el plan.
5. **Al completar un gate de fase:** Revisar el plan con el usuario antes de continuar a la siguiente fase.