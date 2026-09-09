# Síntesis de análisis de código fuente — Lecciones para SambaForge

> Documento de la Fase 0.2.9. Consolida los hallazgos de los 4 análisis de código fuente
> de los proyectos existentes y define los patrones concretos que SambaForge adoptará.

---

## Hallazgos clave por proyecto

### 1. Vexa (Go + Gin + React)

**Lo que hace bien:**
- **Provisioning wizard con SSE streaming**: el backend ejecuta `samba-tool domain provision` y envía el output en tiempo real al frontend via Server-Sent Events. El admin ve el progreso del provisioning línea por línea.
- **Password admin generado con crypto/rand**: genera una contraseña aleatoria segura durante el provisioning en vez de pedirle al usuario.
- **Doble autenticación (Samba + PAM)**: primero intenta `smbclient -L //localhost -U user%pass` para validar credenciales del dominio, si falla intenta PAM (para login inicial antes de que el dominio exista).
- **Computer deployment scripts**: genera scripts `.bat` con template injection server-side (`{{DOMAIN_NAME}}`, `{{AUTH_KEY}}`), NUNCA inyecta passwords en los scripts — el password se pide interactivo.
- **bootstrap.sh**: instala Node 20, Go, Samba, Nginx, systemd service en un solo script.

**Lo que hace mal:**
- **Usa Gin, no go-ldap/gokrb5**: TODA la interacción con AD es via CLI (`samba-tool` + `smbclient` con `os/exec`). No usa LDAP directo ni Kerberos. Esto significa que no puede hacer búsquedas eficientes del directorio.
- **Command sanitizer en modo PERMISSIVE**: existe un sanitizer pero solo loguea, no bloquea. Inyección de comandos posible.
- **CORS Allow-Origin:* con credentials**: configuración CORS insegura.
- **DNS zones/records son stubs**: no implementado, solo placeholders.
- **Sin TLS nativo, sin rate limiting**.
- **Varios TODOs sin implementar**: Join DC, migrate son placeholders.

### 2. go-samba4 (Go + Echo + go-ldap + gokrb5 + GORM)

**Lo que hace bien:**
- **LDAP wrapper con auto-reconnect**: cliente go-ldap v3 con reconexión automática, doble-bind auth (service account busca DN → user rebind para validar).
- **Mapeo AD→structs Go**: structs `User`, `Group`, `OU` con tags JSON, mapeo directo de atributos LDAP.
- **Password UTF-16LE**: codifica passwords en UTF-16LE antes de enviar a LDAP (requerido por AD para el atributo `unicodePwd`).
- **Sesiones DB-backed**: sesiones con `crypto/rand` (64 bytes), guardadas en SQLite via GORM.
- **CSRF doble-cookie**: implementación correcta de CSRF protection.
- **Rate limiter in-memory**: 5 intentos por 5 minutos en login.
- **2FA TOTP con pquerna/otp**: `GenerateSecret`, `Validate`, `RecoveryCodes` — pero no integrado en el login (código muerto).
- **i18n con gotext + .po embebidos**: archivos .po compilados en el binario con go-embed, `Translate()` con fallback en→msgID.
- **Cobra CLI**: `serve`, `migrate`, `version` con ldflags para versión de build.
- **Docker multi-stage Alpine + UPX**: imagen final mínima.

**Lo que hace mal:**
- **Kerberos NO implementado**: gokrb5 NO está en go.mod. `SPNEGOMiddleware` es placeholder sin lógica. El README miente.
- **2FA NO integrado**: el código existe pero `require_totp` se ignora. Código muerto.
- **Sin paginación LDAP**: trae todos los resultados de una vez, delega paginación a DataTables client-side. No escala para directorios grandes.
- **RBAC binario**: solo admin / no-admin. Sin roles granulares.
- **Sin connection pool LDAP**: una sola conexión compartida.
- **Sin tests**.

### 3. Samba Conductor (Meteor 3.4 + React 19 + MongoDB)

**Lo que hace bien:**
- **Zero stored credentials real**: AES-256-GCM en `Map` en memoria con TTL 30min. Clave random por boot o env var. Separación read/write (reads fallback a sync account, writes requieren sesión activa). Esto es el patrón más robusto visto.
- **OAuth2 server**: Authorization Code flow con página server-rendered (no React), UserInfo endpoint custom, fork de `leaonline:oauth2-server` con fixes RFC 6749.
- **Disaster Recovery**: sync AD→MongoDB cada 15min/6h, DR Key via PBKDF2 (100k iter, sha512), backup `mongodump`+`samba-tool` a S3.
- **DC replica via env vars**: `SAMBA_JOIN_AS_DC=true` + `SAMBA_PRIMARY_DC`, idempotente con `.provisioned`/`.joined` markers.
- **3 temas via CSS custom properties**: Tailwind 4, toggle con localStorage.
- **execFile (no shell)**: usa `execFile` para samba-tool, no `exec` con shell. Previene inyección.
- **3 modos de ejecución samba-tool**: local, docker exec, remoto via `-H LDAP URL`.
- **Self-service portal**: cambio de password, edición de perfil.
- **Prometheus metrics**: prom-client integrado.
- **E2E tests con Playwright**.

**Lo que hace mal:**
- **MongoDB obligatorio**: dependence pesada para un panel de administración.
- **Meteor DDP (WebSocket)**: no es REST estándar. Difícil integrar con terceros.
- **TLS `rejectUnauthorized: false`** por defecto — inseguro.
- **Sync account sobre-permisiva**: la cuenta de sync es Domain Admins (demasiados privilegios).
- **Sin MFA, sin rate limiting**.
- **Backup no encripta archivos en /tmp** antes de upload a S3.
- **Single-instance sin HA**.

### 4. cockpit-samba-ad-dc (React + PatternFly + Cockpit)

**Lo que hace bien:**
- **Mayor cobertura de samba-tool**: ~99 operaciones mapeadas en 17 grupos de subcomandos. Esta es nuestra checklist de cobertura.
- **cockpit.spawn (array-based)**: 10 llamadas usan `cockpit.spawn()` con arrays (seguro, sin shell). Pero 91 usan `cockpit.script()` con template strings (inseguro).
- **superuser: true via PolicyKit**: escalada de privilegios limpia sin sudoers manual.
- **Gate function**: `checkServerRole()` ejecuta `testparm --parameter-name=serverrole` y verifica si es AD DC. Si no, lanza wizard. SambaForge debe replicar este patrón.
- **Un módulo por directorio**: estructura limpia, un archivo por operación de samba-tool.

**Lo que hace mal:**
- **Parseo de texto ingenuo**: `data.split('\\n')` y render de líneas como `<li>`. Sin JSON, sin regex, sin parseo estructurado. Bokovoy (mentor del Samba Team) dijo: "Parsing human-oriented output is a waste of resources for robots."
- **Credenciales por pantalla**: cada operación de DNS pide password. FSMO pide user+pass. No hay auth centralizada, no hay ticket cache. UX desastroso.
- **Template string injection**: `` `${command} ${userInput}` `` pasa input sin sanitizar al shell. Riesgo de inyección.
- **Sin error recovery**: el provisioning wizard es un script multi-línea sin verificación de errores intermedios.
- **Se estancó después de GSoC**: último commit ago-2020. Nadie lo mantuvo. samba-tool cambia y el plugin se queda incompatible.

---

## Patrones que SambaForge ADOPTARÁ

| # | Patrón | Origen | Cómo lo implementamos |
|---|---|---|---|
| P-01 | **SSE streaming de samba-tool** | Vexa | Backend Go ejecuta samba-tool y envía output via SSE al frontend. El admin ve progreso del provisioning en tiempo real. |
| P-02 | **Password admin generado con crypto/rand** | Vexa | Durante el provisioning, SambaForge genera una contraseña aleatoria segura. El admin puede cambiarla después. |
| P-03 | **Doble auth: Samba + fallback** | Vexa | Login principal via LDAP bind (samba AD). Fallback via Kerberos ticket cache si existe. Para el primer setup (antes de que el dominio exista), auth via PAM del sistema. |
| P-04 | **LDAP wrapper con auto-reconnect** | go-samba4 | Cliente go-ldap v3 con reconexión automática. Pool de conexiones. |
| P-05 | **Mapeo AD→structs Go** | go-samba4 | Structs `User`, `Group`, `OU`, `Computer`, `Contact` con tags JSON. Mapeo bidireccional LDAP↔Go. |
| P-06 | **Password UTF-16LE para unicodePwd** | go-samba4 | Codificar passwords en UTF-16LE antes de enviar a LDAP. Esto es un requisito de AD, no opcional. |
| P-07 | **Sesiones DB-backed con crypto/rand** | go-samba4 | Sesiones JWT pero con JTI guardado en SQLite para revocación. crypto/rand para generar IDs. |
| P-08 | **CSRF doble-cookie** | go-samba4 | Token CSRF en cookie + header. Verificación en cada mutación. |
| P-09 | **Rate limiting en login** | go-samba4 | 5 intentos por 5 minutos por IP. Bloqueo temporal. |
| P-10 | **2FA TOTP integrado en login** | go-samba4 (mejorado) | pquerna/otp, pero A DIFERENCIA de go-samba4, SÍ integrado en el flow de login. Step 1: user+pass, Step 2: TOTP code. |
| P-11 | **i18n con go-embed** | go-samba4 (adaptado) | En el frontend usamos react-i18next (no gotext). Pero el patrón de embeber traducciones en el binario es válido para el backend Go. |
| P-12 | **Cobra CLI** | go-samba4 | `sambaforge serve`, `sambaforge migrate`, `sambaforge version`, `sambaforge provision` (CLI provisioning sin UI). |
| P-13 | **Zero stored credentials (AES-256-GCM en memoria)** | Samba Conductor | Credenciales del dominio encriptadas en memoria con AES-256-GCM, TTL 30min. Clave random por boot. Separación read/write. |
| P-14 | **OAuth2 server (Authorization Code flow)** | Samba Conductor | Para integraciones de terceros (Grafana, Portainer, n8n). |
| P-15 | **Disaster Recovery con PBKDF2 + S3** | Samba Conductor (mejorado) | Backup `samba-tool domain backup` a S3. DR Key via PBKDF2. A DIFERENCIA de Samba Conductor, encriptar archivos antes de upload. |
| P-16 | **DC replica idempotente con markers** | Samba Conductor | `.provisioned`/`.joined` markers para idempotencia. Re-ejecutar no rompe nada. |
| P-17 | **execFile (no shell) para samba-tool** | Samba Conductor | `exec.Command("samba-tool", args...)` — NUNCA `exec.Command("sh", "-c", "...")`. Args como slice, no string concatenado. |
| P-18 | **3 modos de ejecución samba-tool** | Samba Conductor | Local (default), Docker exec (si Samba corre en contenedor), Remoto via `-H LDAP URL`. |
| P-19 | **Gate function: detectar server role** | cockpit-samba | `testparm --parameter-name=serverrole` → si no es "active directory domain controller", lanzar provisioning wizard. |
| P-20 | **Checklist de 99 operaciones de samba-tool** | cockpit-samba | La lista completa de operaciones que cockpit envuelve es nuestra checklist de cobertura para SambaForge. |
| P-21 | **Un módulo por directorio** | cockpit-samba | Estructura de código: un paquete Go por módulo (users, groups, dns, gpo, etc.). |
| P-22 | **Computer deployment scripts** | Vexa | Generar scripts de join offline para Windows (PowerShell) sin inyectar passwords. |
| P-23 | **Bootstrap script** | Vexa | `bootstrap.sh` que instala Samba, SambaForge, systemd service, Caddy en un solo comando. |
| P-24 | **Prometheus metrics** | Samba Conductor | Endpoint `/metrics` con prom-client para monitoreo del propio SambaForge. |
| P-25 | **Self-service portal** | Samba Conductor | Usuarios cambian su password y editan atributos propios sin intervención del admin. |

## Patrones que SambaForge EVITARÁ

| # | Anti-patrón | Origen | Por qué evitarlo |
|---|---|---|---|
| A-01 | **Parseo de texto ingenuo (split('\\n'))** | cockpit-samba | samba-tool cambia formato entre versiones. Usar `--json` cuando exista, regex tolerante cuando no. |
| A-02 | **Template string con user input** | cockpit-samba | Inyección de shell. Usar `exec.Command` con args slice. |
| A-03 | **Credenciales por pantalla** | cockpit-samba | UX desastroso. Auth centralizada con sesión LDAP+Kerberos. |
| A-04 | **Sin error recovery en provisioning** | cockpit-samba | Cada paso del provisioning debe verificarse. Si falla, rollback limpio. |
| A-05 | **Kerberos como placeholder** | go-samba4 | Si decimos que soportamos Kerberos, implementarlo de verdad. No poner gokrb5 en el README si no está en go.mod. |
| A-06 | **2FA no integrado** | go-samba4 | Si decimos que tenemos 2FA, que funcione en el login. No código muerto. |
| A-07 | **Sin paginación LDAP** | go-samba4 | Directorios grandes (>1000 objetos) necesitan paginación server-side. `SearchWithPaging`. |
| A-08 | **RBAC binario** | go-samba4 | Roles granulares: Domain Admin, Helpdesk, Read-only, Self-service. |
| A-09 | **MongoDB obligatorio** | Samba Conductor | SQLite es suficiente para sesiones, audit log y config. Sin servicio extra. |
| A-10 | **Meteor DDP** | Samba Conductor | REST API estándar con Echo. Interoperable con cualquier cliente. |
| A-11 | **TLS rejectUnauthorized: false** | Samba Conductor | TLS estricto. Si el certificado no es válido, fallar. |
| A-12 | **Sync account como Domain Admin** | Samba Conductor | La cuenta de sync debe tener mínimos privilegios. Solo lectura para sync, escritura con credenciales de sesión. |
| A-13 | **Command sanitizer permissive** | Vexa | Sanitizer debe bloquear, no solo loguear. |
| A-14 | **CORS * con credentials** | Vexa | CORS restrictivo. Same-origin o origins explícitos. |
| A-15 | **Stubs sin implementar** | Vexa | Si una feature está en el README, debe funcionar. No placeholders. |
| A-16 | **Sin tests** | go-samba4, cockpit-samba | Tests desde Fase 1. Unit tests del wrapper de samba-tool, tests E2E del provisioning. |

---

## Insights críticos para la arquitectura de SambaForge

### 1. El problema fundamental de samba-tool (lección de Bokovoy)

Alexander Bokovoy (Samba Team, mentor del GSoC 2020) identificó el problema central:

> "Parsing human-oriented output is a waste of resources for robots. samba-tool needs machine-readable output (JSON output, DBus interface, or Python library bindings)."

**Implicación para SambaForge:**
- Usar `--json` en TODOS los subcomandos que lo soporten.
- Para los que no lo soportan, escribir parsers tolerantes que:
  - No asuman número fijo de líneas.
  - Usen regex con grupos nombrados.
  - Detecten errores por exit code + stderr, no por texto.
- Contribuir `--json` flags upstream para los subcomandos que no lo tienen (trabajo futuro).

### 2. La importancia del provisioning con feedback en tiempo real

Vexa implementa SSE streaming del provisioning. Samba Conductor usa execFile con callbacks. cockpit-samba usa un script blind sin feedback.

**SambaForge debe:**
- Ejecutar samba-tool con `cmd.Start()` + `StdoutPipe()`.
- Leer stdout línea por línea en una goroutine.
- Enviar cada línea via SSE (Echo soporta `c.Stream()`).
- Frontend muestra progreso en tiempo real con un log console.
- Si exit code != 0, mostrar stderr completo.

### 3. Credenciales: tres niveles

| Nivel | Cuándo | Cómo |
|---|---|---|
| **Pre-provisioning** | Antes de que el dominio exista | PAM auth del sistema (login Linux) |
| **Post-provisioning, sin 2FA** | Login normal del admin | LDAP bind con `user@REALM` |
| **Post-provisioning, con 2FA** | Si 2FA activado | LDAP bind + TOTP verification |
| **Operaciones que requieren Kerberos** | FSMO, trusts, backup | Ticket cache (kinit previo) o `--use-kerberos=required` con `KRB5CCNAME` |

### 4. El checklist de cobertura de samba-tool

Basado en el análisis de cockpit-samba-ad-dc, SambaForge debe cubrir estas 99 operaciones:

| Grupo | Operaciones | Fase SambaForge |
|---|---|---|
| computer (5) | create, delete, list, show, move | Fase 3 |
| contact (5) | create, delete, list, show, move | Fase 3 |
| delegation (5) | add-service, del-service, for-any-protocol, for-any-service, show | Fase 6 |
| dns (8) | add, delete, query, roothints, serverinfo, update, zonecreate, zonedelete | Fase 3+5 |
| domain (9) | info, join, demote, dcpromo, classicupgrade, backup online, backup offline, backup rename, backup restore | Fase 2+7+8 |
| domain trust (6) | create, delete, list, show, validate, namespaces | Fase 8 |
| forest (2) | show, dsheuristics | Fase 8 |
| fsmo (3) | show, seize, transfer | Fase 8 |
| gpo (14) | create, del, getinheritance, setinheritance, getlink, setlink, list, listall, listcontainers, backup, restore, fetch, addlink, dellink | Fase 5 |
| group (7) | add, delete, list, listmembers, addmembers, removemembers, show | Fase 3 |
| dsacl (2) | get, set | Fase 6 |
| ntacl (6) | get, set, getdosinfo, setdosinfo, check, reset | Fase 6 |
| ou (6) | add, delete, list, move, rename, show | Fase 4 |
| sites (5) | create, delete, list, subnet create, subnet delete | Fase 8 |
| spn (3) | add, delete, list | Fase 6 |
| user (10) | create, delete, list, disable, enable, setpassword, setexpiry, rename, edit, sensitive | Fase 3 |
| time (1) | show | Fase 3 |
| provision (1) | domain provision | Fase 2 |
| testparm (1) | --parameter-name=serverrole | Fase 2 |
| **Total** | **99 operaciones** | |

### 5. Lo que NINGÚN proyecto existente hace bien

| Feature | Ningún proyecto lo hace | SambaForge lo hará |
|---|---|---|
| Auth policies (Samba 4.24) | Ninguno | Fase 6 |
| Auth silos (Samba 4.24) | Ninguno | Fase 6 |
| Claims (Samba 4.24) | Ninguno | Fase 6 |
| DS ACLs manipulation | Solo cockpit (pero roto) | Fase 6 |
| NT ACLs/SYSVOL reset | Solo cockpit (pero roto) | Fase 6 |
| RBAC granular | Ninguno (todos binario) | Fase 6 |
| Audit log persistente e inmutable | Ninguno | Fase 6 |
| Paginación LDAP server-side | Ninguno | Fase 3 |
| Kerberos real (gokrb5) | Ninguno (go-samba4 lo finge) | Fase 6 |
| 2FA integrado en login | Ninguno (go-samba4 no lo integra) | Fase 6 |
| Importación masiva CSV | Ninguno | Fase 4 |
| API REST documentada (OpenAPI) | Ninguno | Fase 9 |
| i18n es/en/pt real | Solo go-samba4 (pero gotext, no React) | Fase 3 |