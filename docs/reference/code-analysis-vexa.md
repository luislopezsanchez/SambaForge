# Análisis de Código Fuente: Vexa

> **Proyecto**: SambaForge — Fase 0: Investigación  
> **Repositorio analizado**: https://github.com/griffinwebnet/Vexa  
> **Versión**: 0.3.129  
> **Fecha**: 2026-09-09  
> **Stack**: Go (Gin) + React 19 (Vite + TypeScript + Tailwind)  
> **Propósito**: Plataforma web para gestionar Samba AD DC (provisioning, users, groups, computers, DNS, overlay networking con Headscale/Tailscale)

---

## 1. Estructura de Directorios Completa

```
Vexa/
├── api/                          # Backend Go
│   ├── main.go                   # Entry point, routing, server bootstrap
│   ├── go.mod / go.sum           # Dependencias Go
│   ├── config/
│   │   └── config.go             # Config desde env vars (SERVER_HOSTNAME, SERVER_NAMES)
│   ├── handlers/                 # HTTP handlers (capa presentación)
│   │   ├── auth.go               # Login, BootstrapStatus, BootstrapAdmin
│   │   ├── users.go              # CRUD users, ChangePassword, UpdateProfile
│   │   ├── user_actions.go       # ResetPassword, DisableUser, EnableUser (samba-tool directo)
│   │   ├── groups.go             # CRUD groups, add/remove members
│   │   ├── domain.go             # DomainStatus, GetDomainInfo, ProvisionDomainWithOutput (SSE)
│   │   ├── computers.go          # List/Get/Delete computers
│   │   ├── deployment.go         # Generación y serving de scripts .bat de deployment
│   │   ├── dns.go                # DNS status + forwarders (zones/records = stub)
│   │   ├── audit.go              # Audit logs, system logs, log stats, log level
│   │   ├── policies.go           # Domain policies (GPO-like)
│   │   ├── overlay.go            # Headscale overlay networking
│   │   └── updates.go            # Self-update mechanism
│   ├── services/                 # Lógica de negocio (capa servicios)
│   │   ├── auth_service.go       # Autenticación SAMBA + PAM, JWT generation
│   │   ├── domain_service.go     # Domain status, provisioning wizard logic
│   │   ├── user_service.go       # CRUD users vía samba-tool + ldbmodify
│   │   ├── group_service.go      # CRUD groups vía samba-tool
│   │   ├── computer_service.go   # List computers + Tailscale overlay integration
│   │   ├── dns_service.go        # DNS forwarders
│   │   ├── headscale_service.go  # Headscale/Tailscale overlay management
│   │   ├── overlay_service.go    # Overlay networking setup
│   │   ├── system_service.go     # System-level operations
│   │   ├── update_service.go     # Self-update logic
│   │   ├── ddns_service.go       # Dynamic DNS
│   │   └── vexa_admin_service.go # Bootstrap admin (bcrypt, /var/lib/vexa/admin.json)
│   ├── exec/                     # Wrappers de ejecución de comandos
│   │   ├── samba.go              # SambaTool: ejecuta samba-tool vía os/exec
│   │   ├── system.go             # System: systemctl, file ops, command exec
│   │   └── headscale.go          # HeadscaleTool: ejecuta headscale CLI
│   ├── middleware/
│   │   ├── auth.go               # JWT validation (Bearer token)
│   │   ├── cors.go               # CORS (Allow-Origin: *)
│   │   └── provisioning.go       # ProvisioningGate: bloquea APIs si no provisionado
│   ├── models/                   # Data models / DTOs
│   │   ├── auth.go               # LoginRequest, LoginResponse, AuthResult
│   │   ├── user.go               # User, CreateUserRequest, UpdateUserRequest
│   │   ├── group.go              # Group, CreateGroupRequest, etc.
│   │   ├── computer.go           # Computer model
│   │   ├── domain.go             # ProvisionDomainRequest, DomainStatusResponse, DomainInfo
│   │   └── system.go             # System models
│   ├── utils/
│   │   ├── command.go            # CommandSanitizer: políticas de ejecución segura
│   │   ├── auth.go               # AuthenticatePAM, AuthenticateSAMBA (smbclient)
│   │   ├── jwt.go                # JWT secret management (env or random)
│   │   ├── audit.go              # AuditContext, Log* functions, AuditMiddleware
│   │   ├── logger.go             # Logger con rotación (debug/info/warn/error/audit)
│   │   └── users.go              # User utilities
│   ├── scripts/
│   │   └── deployment/
│   │       ├── domain-join-only.bat
│   │       ├── domain-join-with-tailscale.bat
│   │       └── tailnet-add.bat
│   └── version/
│       └── version.go           # Version constant
├── web/                          # Frontend React
│   ├── package.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── tailwind.config.js
│   ├── postcss.config.js
│   ├── index.html
│   ├── src/
│   │   ├── main.tsx              # React entry point
│   │   ├── App.tsx               # Router + route guards (ProtectedRoute, AdminRoute, etc.)
│   │   ├── index.css             # Tailwind imports
│   │   ├── lib/
│   │   │   ├── api.ts            # Axios instance con interceptores JWT
│   │   │   └── utils.ts           # Utility functions (cn, etc.)
│   │   ├── stores/
│   │   │   └── authStore.ts      # Zustand store con persist (localStorage)
│   │   ├── pages/
│   │   │   ├── LoginPage.tsx
│   │   │   ├── SetupWizard.tsx   # Wizard de provisioning con SSE streaming
│   │   │   ├── Dashboard.tsx
│   │   │   ├── Users.tsx
│   │   │   ├── Groups.tsx
│   │   │   ├── Computers.tsx
│   │   │   ├── MachineDetails.tsx
│   │   │   ├── DNS.tsx
│   │   │   ├── DomainManagement.tsx
│   │   │   ├── DomainOUs.tsx
│   │   │   ├── DomainPolicies.tsx
│   │   │   ├── OrganizationalUnits.tsx
│   │   │   ├── OverlayNetworking.tsx
│   │   │   ├── Security.tsx
│   │   │   ├── SelfService.tsx
│   │   │   └── Settings.tsx
│   │   ├── components/
│   │   │   ├── ComputerDeploymentModal.tsx  # Modal de deployment scripts
│   │   │   ├── Logo.tsx
│   │   │   ├── ThemeProvider.tsx  # Dark/light theme
│   │   │   ├── ui/
│   │   │   │   ├── Button.tsx
│   │   │   │   ├── Card.tsx
│   │   │   │   ├── Dialog.tsx
│   │   │   │   └── Input.tsx
│   │   │   └── modals/
│   │   │       ├── AddGroupModal.tsx
│   │   │       ├── AddOUModal.tsx
│   │   │       ├── AddUserModal.tsx
│   │   │       ├── EditOUModal.tsx
│   │   │       ├── EditUserModal.tsx
│   │   │       ├── ManageGroupModal.tsx
│   │   │       └── ManageUserModal.tsx
│   │   ├── layouts/
│   │   │   └── DashboardLayout.tsx  # Sidebar + main content layout
│   │   ├── types/
│   │   │   └── updates.ts
│   │   └── utils/
│   │       └── domainUtils.ts
│   └── public/
│       ├── favicon.svg
│       └── logo dewsign black/white.svg
├── bootstrap.sh                  # Installer script (apt, Go build, npm build, nginx, systemd)
├── update.sh                      # Self-update script
├── README.md
├── DEVELOPMENT.md
└── vexa.png
```

---

## 2. Backend Go

### 2.1 Framework HTTP

**Gin** (`github.com/gin-gonic/gin v1.9.1`) — framework HTTP minimalista y de alto rendimiento.

```go
// main.go — estructura del router
router := gin.Default()
router.Use(middleware.CORS())           // CORS allow-all
router.Use(utils.AuditMiddleware())     // Audit logging global

// Rutas públicas
public := router.Group("/api/v1")
public.POST("/auth/login", authHandler.Login)

// Rutas protegidas (JWT + ProvisioningGate)
protected := router.Group("/api/v1")
protected.Use(middleware.AuthRequired())
protected.Use(middleware.ProvisioningGate())
```

### 2.2 Organización Handlers / Services / Middleware

**Patrón**: Handler → Service → Exec wrapper → os/exec

| Capa | Responsabilidad | Ejemplo |
|------|----------------|---------|
| `handlers/` | Parsear HTTP request, validar input, formatear response | `UserHandler.CreateUser()` |
| `services/` | Lógica de negocio, orquestar comandos | `UserService.CreateUser()` |
| `exec/` | Wrappers tipados para comandos del sistema | `SambaTool.UserCreate()` |
| `utils/` | Cross-cutting: auth, JWT, logging, command sanitizer | `SafeCommand()` |
| `middleware/` | Auth, CORS, provisioning gate | `AuthRequired()` |

Los handlers se instancian en `main.go` con constructor pattern:
```go
authHandler := handlers.NewAuthHandler()
userHandler := handlers.NewUserHandler()
```

Cada handler tiene una struct con dependencias inyectadas:
```go
type UserHandler struct {
    userService *services.UserService
    authService *services.AuthService
}
```

### 2.3 Paquetes Go Importados

| Paquete | Versión | Uso |
|---------|---------|-----|
| `github.com/gin-gonic/gin` | v1.9.1 | Framework HTTP |
| `github.com/golang-jwt/jwt/v5` | v5.2.0 | JWT generation/validation |
| `golang.org/x/crypto` | v0.17.0 | bcrypt (admin bootstrap) |
| `os/exec` | stdlib | Ejecución de samba-tool, systemctl, etc. |

**NOTABLE**: No usa `go-ldap`, `gokrb5`, ni ninguna librería LDAP/Kerberos nativa de Go. Toda la interacción con AD se hace vía **samba-tool CLI** y **smbclient**.

### 2.4 Ejecución de samba-tool (os/exec)

El paquete `exec/samba.go` es el wrapper central. Ejecuta `samba-tool` vía `os/exec`:

```go
// exec/samba.go
func (s *SambaTool) Run(args ...string) (string, error) {
    cmd, cmdErr := utils.SafeCommand("samba-tool", args...)
    if cmdErr != nil {
        return "", cmdErr
    }
    output, err := cmd.CombinedOutput()
    return string(output), err
}

// Ejemplos de comandos envueltos:
func (s *SambaTool) UserCreate(username, password string, options UserCreateOptions) (string, error) {
    args := []string{"user", "create", username, password}
    if options.FullName != "" { args = append(args, "--given-name="+options.FullName) }
    if options.Email != ""    { args = append(args, "--mail-address="+options.Email) }
    if options.OUPath != ""   { args = append(args, "--userou="+options.OUPath) }
    return s.Run(args...)
}
```

Comandos samba-tool usados:
- `samba-tool user create/list/show/delete/enable/disable/setpassword`
- `samba-tool group add/list/listmembers/delete/addmembers/removemembers/modify`
- `samba-tool domain provision --realm=X --domain=Y --adminpass=Z --server-role=dc --dns-backend=SAMBA_INTERNAL --use-rfc2307`
- `samba-tool domain info`
- `samba-tool computer list/delete`
- `samba-tool dns server set forwarder`

Para LDAP modify directo, usa `ldbmodify`:
```go
cmd, _ := utils.SafeCommand("ldbmodify", "-H", "/var/lib/samba/private/sam.ldb")
cmd.Stdin = strings.NewReader(ldif)
cmd.CombinedOutput()
```

### 2.5 Gestión de Autenticación

**Estrategia**: Doble autenticación — SAMBA domain primero, PAM fallback.

```go
// services/auth_service.go
func (s *AuthService) authenticateUser(username, password string) (bool, bool, bool) {
    // 1. Intentar SAMBA (smbclient //localhost/ipc$)
    if utils.AuthenticateSAMBA(username, password) {
        isAdmin := utils.CheckDomainAdminStatus(username) // via samba-tool group listmembers
        return true, isAdmin, true  // domain user
    }
    // 2. Fallback a PAM (pamtester)
    if utils.AuthenticatePAM(username, password) {
        isAdmin := utils.CheckLocalAdminStatus(username) // id -nG
        if isAdmin { return true, true, false } // local admin
        return false, false, false // local non-admin rejected
    }
    return false, false, false
}
```

**SAMBA auth** usa `smbclient` con 4 estrategias de fallback:
1. `smbclient //localhost/ipc$ -U username%password -c exit`
2. `smbclient //localhost/netlogon -U username%password -c exit`
3. `smbclient //localhost/ipc$ -U DOMAIN\username%password -c exit`
4. `smbclient //localhost/ipc$ -U username@realm%password -c exit`

**PAM auth** usa `pamtester login <username> authenticate` (stdin: password).

**Admin detection**:
- Domain: `samba-tool group listmembers "Domain Admins"` y `"Administrators"`
- Local: `id -nG <user>` → busca `sudo`, `wheel`, `admin` groups

**No usa Kerberos directamente** (no kinit/klist en el flujo de auth). kinit/kdestroy están en el sanitizer policy pero no se invocan en el código de auth.

### 2.6 Rutas API Expuestas

**Públicas** (`/api/v1`):
| Método | Ruta | Handler |
|--------|------|---------|
| POST | `/auth/login` | `authHandler.Login` |
| GET | `/auth/bootstrap-status` | `authHandler.BootstrapStatus` |
| POST | `/auth/bootstrap-admin` | `authHandler.BootstrapAdmin` |
| GET | `/version` | inline |
| GET | `/updates/check` | `handlers.CheckForUpdates` |
| POST | `/updates/upgrade` | `handlers.PerformUpgrade` |
| GET | `/domain/status` | `domainHandler.DomainStatus` |
| GET | `/health` | inline |

**Protegidas** (JWT + ProvisioningGate):
| Método | Ruta | Handler |
|--------|------|---------|
| POST | `/domain/provision-with-output` | `domainHandler.ProvisionDomainWithOutput` (SSE) |
| GET | `/domain/info` | `domainHandler.GetDomainInfo` |
| PUT | `/domain/configure` | `domainHandler.ConfigureDomain` |
| GET/POST/PUT/DELETE | `/users`, `/users/:id` | CRUD users |
| POST | `/users/:id/reset-password` | `ResetUserPassword` |
| POST | `/users/:id/disable\|enable` | Enable/disable user |
| POST | `/users/:id/toggle-must-change-password` | Toggle flag |
| POST | `/users/change-password` | Self-service password change |
| POST | `/users/update-profile` | Self-service profile |
| GET/POST/PUT/DELETE | `/groups`, `/groups/:id` | CRUD groups |
| POST/DELETE | `/groups/:id/members` | Add/remove members |
| GET | `/computers` | `computerHandler.ListComputers` |
| GET | `/computers/:id` | `computerHandler.GetComputer` |
| GET | `/machines/:id` | `computerHandler.GetMachineDetails` |
| DELETE | `/computers/:id` | `computerHandler.DeleteComputer` |
| GET | `/dns/status` | `dnsHandler.DNSStatus` |
| PUT | `/dns/forwarders` | `dnsHandler.UpdateDNSForwarders` |
| GET/POST/DELETE | `/dns/zones`, `/dns/records` | **STUB** (501 Not Implemented) |
| GET/PUT | `/domain/policies` | Domain policies |
| GET/POST/DELETE | `/domain/ous`, `/domain/ous/:path` | Organizational Units |
| GET | `/deployment/scripts` | List deployment scripts |
| POST | `/deployment/generate` | Generate deployment command |
| GET | `/deployment/scripts/:script` | Serve .bat file |
| GET | `/audit/logs`, `/audit/stats` | Audit logs |
| GET | `/audit/logs/:type` | System logs (debug/info/warn/error) |
| POST | `/audit/log-level` | Change log level |
| GET | `/system/overlay-status` | Headscale status |
| POST | `/system/setup-overlay` | Setup Headscale |
| POST | `/system/test-fqdn` | Test FQDN connectivity |

---

## 3. Frontend React

### 3.1 Librerías

| Librería | Versión | Uso |
|----------|---------|-----|
| React | 19.1.1 | UI framework |
| React DOM | 19.1.1 | DOM rendering |
| React Router DOM | 7.9.3 | Client-side routing |
| Zustand | 5.0.8 | State management (con persist middleware) |
| TanStack React Query | 5.17.9 | Server state / data fetching |
| Axios | 1.6.5 | HTTP client |
| Tailwind CSS | 3.4.18 (dev) / 4.1.14 | Styling |
| lucide-react | 0.544.0 | Iconos |
| class-variance-authority | 0.7.0 | Component variants |
| clsx + tailwind-merge | — | className utilities |
| TypeScript | 5.3.3 | Type safety |
| Vite | 7.1.7 | Build tool / dev server |
| ESLint | 9.36.0 | Linting |

### 3.2 State Management

**Zustand** con `persist` middleware (localStorage):

```typescript
// stores/authStore.ts
export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      username: null,
      isAdmin: false,
      isDomainUser: false,
      isAuthenticated: false,
      login: (token, username, isAdmin, isDomainUser) =>
        set({ token, username, isAdmin, isDomainUser, isAuthenticated: true }),
      logout: () => {
        set({ token: null, username: null, isAdmin: false, isDomainUser: false, isAuthenticated: false })
        window.location.href = '/login'
      },
    }),
    { name: 'vexa-auth' }  // localStorage key
  )
)
```

**TanStack Query** para server state (cache, invalidation, refetching).

### 3.3 Routing

React Router DOM v7 con route guards basados en roles:

```tsx
// App.tsx
<Router>
  <Routes>
    <Route path="/login" element={<LoginPage />} />
    <Route path="/wizard" element={<SetupRoute><SetupWizard /></SetupRoute>} />
    <Route path="/" element={<ProtectedRoute><DashboardLayout /></ProtectedRoute>}>
      <Route index element={<Dashboard />} />
      <Route path="self-service" element={<DomainUserRoute><SelfService /></DomainUserRoute>} />
      <Route path="domain" element={<AdminRoute><DomainManagement /></AdminRoute>} />
      <Route path="users" element={<AdminRoute><Users /></AdminRoute>} />
      <Route path="computers" element={<AdminRoute><Computers /></AdminRoute>} />
      {/* ... */}
    </Route>
  </Routes>
</Router>
```

4 route guards:
- `ProtectedRoute` — solo autenticado
- `AdminRoute` — autenticado + admin
- `DomainUserRoute` — autenticado + domain user
- `SetupRoute` — autenticado + local admin (no domain user)

### 3.4 API Client

```typescript
// lib/api.ts
const api = axios.create({
  baseURL: '/api/v1',
  headers: { 'Content-Type': 'application/json' },
})

// Interceptor: añade Bearer token
api.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// Interceptor: logout automático en 401
api.interceptors.response.use((response) => response, (error) => {
  if (error.response?.status === 401 && window.location.pathname !== '/login') {
    useAuthStore.getState().logout()
    window.location.href = '/login'
  }
  return Promise.reject(error)
})
```

### 3.5 UI Components

Design system minimal basado en **shadcn/ui pattern**:
- `ui/Button.tsx` — variantes (default, outline, ghost, destructive) + sizes
- `ui/Card.tsx` — Card, CardHeader, CardTitle, CardDescription, CardContent
- `ui/Dialog.tsx` — Modal dialog
- `ui/Input.tsx` — Input estilizado
- `ThemeProvider.tsx` — Dark/light theme con localStorage persistence
- `Logo.tsx` — SVG logo component

### 3.6 Organización Pages/Components/Stores/Lib

```
pages/        → 15 páginas (una por feature: Users, Groups, Computers, DNS, etc.)
components/   → Componentes reutilizables + modales de form (Add/Edit/Manage)
  ui/         → Primitivas de UI (Button, Card, Dialog, Input)
  modals/     → Modales de CRUD (AddUserModal, EditUserModal, etc.)
stores/       → Zustand stores (authStore)
lib/          → Utilities (api.ts, utils.ts)
layouts/      → Layout components (DashboardLayout con sidebar)
utils/        → domainUtils.ts
types/        → TypeScript types (updates.ts)
```

---

## 4. Provisioning Wizard

### 4.1 Flujo del Wizard

1. **Login como local admin** (PAM) → redirect a `/wizard` si dominio no provisionado
2. **Selección de modo**: New Domain | Join as DC (stub) | Migrate Domain (stub)
3. **Formulario de provisioning**:
   - Input: **Realm** (ej: `mydomain.local`)
   - Domain name auto-generado del realm (`realm.split('.')[0].toUpperCase()`)
   - Selector de **DNS Provider**: Cloudflare / Google / Quad9 / Custom
   - DNS forwarders mapeados del provider seleccionado
4. **POST** `/api/v1/domain/provision-with-output` con `{ domain, realm, dns_forwarder }`
5. **Streaming SSE** del output del provisioning en tiempo real
6. **Redirect a login** al completar

### 4.2 Comandos samba-tool en Provisioning

```go
// services/domain_service.go — ProvisionDomainWithOutput()
// 1. Verificar samba-tool: samba-tool --version
// 2. Limpiar config: rm /etc/samba/smb.conf
// 3. Detener servicios: systemctl stop smbd nmbd winbind samba-ad-dc
// 4. Limpiar DBs: rm /var/lib/samba/private/sam.ldb, secrets.ldb, /var/cache/samba/gencache.tdb
// 5. Configurar systemd-resolved: DNSStubListener=no + systemctl restart systemd-resolved
// 6. Generar password admin aleatorio (crypto/rand, 16 chars)
// 7. EJECUTAR PROVISION:
//    samba-tool domain provision \
//      --realm=REALM --domain=DOMAIN --adminpass=GENERATED \
//      --server-role=dc --dns-backend=SAMBA_INTERNAL --use-rfc2307 \
//      --option="dns forwarder = FORWARDER"
// 8. Crear grupos default: IT Staff, Finance, Sales, HR (samba-tool group add)
// 9. Iniciar servicio: systemctl enable --now samba-ad-dc
```

### 4.3 Validaciones

- **Realm**: requerido, no vacío (frontend `required`)
- **Domain**: auto-generado del realm
- **DNS Backend**: default `SAMBA_INTERNAL`
- **DNS Forwarder**: mapeado de provider preseleccionado
- **Admin password**: generado con `crypto/rand`, 16 chars, garantiza mayúscula/minúscula/dígito/special
- **Pre-flight**: verifica `samba-tool` disponible
- **Limpieza previa**: elimina smb.conf, sam.ldb, secrets.ldb, gencache.tdb
- **systemd-resolved**: deshabilita DNSStubListener para liberar puerto 53

### 4.4 Streaming SSE

El handler usa Server-Sent Events para transmitir el output de `samba-tool domain provision`:

```go
// handlers/domain.go
func (h *DomainHandler) ProvisionDomainWithOutput(c *gin.Context) {
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")

    outputChan := make(chan string, 100)
    go func() {
        defer close(outputChan)
        err := h.domainService.ProvisionDomainWithOutput(req, outputChan)
        if err != nil {
            outputChan <- "ERROR: " + err.Error()
        }
    }()

    for output := range outputChan {
        c.SSEvent("message", gin.H{"type": "output", "content": output, "timestamp": time.Now().Unix()})
        c.Writer.Flush()
    }
    c.SSEvent("message", gin.H{"type": "complete", "content": "Domain provisioning completed"})
}
```

El exec wrapper usa `StdoutPipe`/`StderrPipe` con goroutines para streaming línea por línea.

---

## 5. Computer Deployment Scripts

### 5.1 Scripts Disponibles

| Script | Descripción | Requiere Headscale |
|--------|-------------|-------------------|
| `domain-join-only.bat` | Join domain sin Tailscale | No |
| `domain-join-with-tailscale.bat` | Instalar Tailscale + join domain | Sí |
| `tailnet-add.bat` | Añadir a Tailnet existente | Sí |

### 5.2 Template Variables

Los scripts usan placeholders reemplazados server-side:
- `{{DOMAIN_NAME}}` — NetBIOS domain name
- `{{DOMAIN_REALM}}` — Kerberos realm (DNS domain)
- `{{LOGIN_SERVER}}` — Headscale login server URL
- `{{AUTH_KEY}}` — Pre-auth key de Headscale
- `{{ADMIN_USER}}` — `administrator@REALM` (no se inyecta password)

**Política de seguridad**: `{{ADMIN_PASSWORD}}` → `PROMPT_FOR_PASSWORD` (siempre se pide interactivamente).

### 5.3 domain-join-only.bat — Flujo

```
1. Auto-elevate con UAC (Start-Process -Verb RunAs)
2. Flush DNS: ipconfig /flushdns
3. Test resolution: nslookup %DOMAIN_REALM%
4. Prompt: admin username + password (set /p)
5. PowerShell: Add-Computer -DomainName '%DOMAIN_REALM%' -Credential $cred -Force
6. Reboot: shutdown /r /f /t 0
```

### 5.4 domain-join-with-tailscale.bat — Flujo

```
1. Auto-elevate UAC
2. Download Tailscale MSI (https://pkgs.tailscale.com/stable/)
3. Install: msiexec /i tailscale-setup.msi /quiet /norestart
4. Wait 15s for service start
5. Connect: tailscale up --authkey=KEY --login-server=SERVER --accept-routes --hostname=HOSTNAME --unattended
6. Wait 5s for network stabilization
7. Flush DNS + test resolution
8. Prompt: admin credentials
9. Add-Computer -DomainName
10. Cleanup installer
11. Reboot
```

### 5.5 Serving de Scripts

```go
// handlers/deployment.go — ServeDeploymentScript()
// 1. Validar script name contra allowlist
// 2. Leer archivo de scripts/deployment/
// 3. Reemplazar template variables con valores reales del dominio
// 4. Content-Disposition: attachment; filename="script.bat"
```

---

## 6. bootstrap.sh

Script de instalación automática que:

1. **Verifica root** (`$EUID -ne 0`)
2. **Instala paquetes del sistema**:
   - Node.js 20 (desde NodeSource)
   - `krb5-user`, `krb5-config`, `ldb-tools`, `attr`, `acl`, `build-essential`, `pkg-config`, `git`, `curl`, `wget`, `nginx`, `ddclient`, `jq`, `pamtester`, `vim-nox`, `golang-go`
   - Samba: `samba`, `samba-dsdb-modules`, `samba-common-bin`, `smbclient`, `winbind`
3. **Detecta LXC unprivileged** y advierte sobre bug de Samba 4.19.x
4. **Clona el repositorio** (release tag o main branch)
5. **Copia archivos** a `/var/www/vexa/`
6. **Build Go API**: `go build -o /usr/local/bin/vexa-api`
7. **Build React**: `npm install && npm run build`
8. **Configura Nginx**:
   - Sirve frontend desde `/var/www/vexa/web/dist`
   - Proxy `/api/` → `localhost:8080`
   - SPA fallback: `try_files $uri $uri/ /index.html`
9. **Crea systemd service**: `vexa-api.service` (User=root, Restart=always, ENV=production)
10. **Habilita servicios**: `systemctl enable vexa-api nginx`

---

## 7. Patrones de Seguridad

### 7.1 Credenciales

| Mecanismo | Implementación |
|-----------|---------------|
| Admin bootstrap | bcrypt hash en `/var/lib/vexa/admin.json` (permisos 0600) |
| JWT secret | `JWT_SECRET` env var, o random 32 bytes en runtime |
| JWT signing | HS256, expira en 24h |
| Admin password provisioning | `crypto/rand` 16 chars con complejidad garantizada |
| Password reset | Genera password aleatoria (adjetivo + sustantivo + símbolo + número) |
| Deployment scripts | NUNCA inyecta admin password — siempre prompt interactivo |

### 7.2 TLS

- **No hay TLS nativo** en Go API (escucha en `:8080` HTTP plano)
- TLS delegado a **Nginx** (configuración en bootstrap.sh — pero el config mostrado solo tiene `listen 80`)
- Sin certificates management, sin Let's Encrypt integration
- smbclient usa conexión local sin TLS

### 7.3 Sesiones

- **Stateless**: JWT en `Authorization: Bearer <token>`
- Frontend persiste token en **localStorage** (vexa-auth key)
- No hay refresh token mechanism
- No hay session revocation/blacklist
- Logout = limpiar localStorage + redirect

### 7.4 Command Sanitizer

Patrón notable: `utils/command.go` define un `CommandSanitizer` con políticas por comando:

```go
policies: map[string]CommandPolicy{
    "systemctl": { Allowed: true, StaticArgs: [...], PositionalArgs: {1: isSafeServiceName}, MaxArgs: 2 },
    "samba-tool": { Allowed: true, StaticArgs: [], PositionalArgs: {}, MaxArgs: 20 }, // PERMISSIVE
    // ...
}
```

**PERO**: el sanitizer está en modo PERMISSIVE — `SanitizeCommand()` solo loguea y retorna nil:
```go
func (cs *CommandSanitizer) SanitizeCommand(name string, args ...string) error {
    Info("[SafeExec] %s %v", name, args)
    return nil  // PERMISSIVE MODE - Just log and allow everything
}
```

### 7.5 CORS

```go
c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
```

⚠️ `Allow-Origin: *` con `Allow-Credentials: true` es una configuración insegura (los navegadores modernos lo rechazan, pero el server lo envía).

### 7.6 Provisioning Gate

Middleware que bloquea acceso a APIs protegidas si el dominio no está provisionado:
```go
func ProvisioningGate() gin.HandlerFunc {
    // Permite: auth, bootstrap, domain/status, domain/provision, version, health
    // Bloquea: todo lo demás si !Provisioned && !isAdmin
}
```

### 7.7 Audit Logging

Sistema de audit logging completo:
- Categorías: authentication, user_management, group_management, domain_management, computer_management, system_management, data_access, security
- Almacenado en `/var/log/vexa/audit.log` como JSON lines
- Rotación: 10MB, 6 archivos, compresión tar.gz
- Middleware global `AuditMiddleware()` loguea todas las requests/responses

---

## 8. Snippets Reutilizables para SambaForge

### Snippet 1: SafeCommand wrapper con sanitizer

```go
// Patrón: wrapper de os/exec con sanitización por política
package utils

type ArgValidator func(string) bool

type CommandPolicy struct {
    Allowed           bool
    StaticArgs        []string
    PositionalArgs    map[int]ArgValidator
    RequiresElevation bool
    MaxArgs           int
}

type CommandSanitizer struct {
    policies map[string]CommandPolicy
}

func SafeCommand(name string, args ...string) (*exec.Cmd, error) {
    // Validar contra política, retornar exec.Cmd
    return globalSanitizer.SafeExec(name, args...)
}

func SafeCommandContext(ctx context.Context, name string, args ...string) (*exec.Cmd, error) {
    return globalSanitizer.SafeExecContext(ctx, name, args...)
}
```

### Snippet 2: Streaming SSE de provisioning con goroutines

```go
// Patrón: streaming de output de comando CLI vía Server-Sent Events
func (h *DomainHandler) ProvisionDomainWithOutput(c *gin.Context) {
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")

    outputChan := make(chan string, 100)
    go func() {
        defer close(outputChan)
        err := h.domainService.ProvisionDomainWithOutput(req, outputChan)
        if err != nil {
            outputChan <- "ERROR: " + err.Error()
        }
    }()

    for output := range outputChan {
        c.SSEvent("message", gin.H{
            "type":      "output",
            "content":   output,
            "timestamp": time.Now().Unix(),
        })
        c.Writer.Flush()
    }
    c.SSEvent("message", gin.H{"type": "complete", "content": "Done"})
}
```

### Snippet 3: Doble autenticación SAMBA + PAM con fallback

```go
// Patrón: auth domain-first, PAM-fallback con admin check
func (s *AuthService) authenticateUser(username, password string) (authenticated, isAdmin, isDomainUser bool) {
    // 1. SAMBA domain auth (smbclient)
    if utils.AuthenticateSAMBA(username, password) {
        isAdmin := utils.CheckDomainAdminStatus(username) // samba-tool group listmembers
        return true, isAdmin, true
    }
    // 2. PAM fallback (pamtester)
    if utils.AuthenticatePAM(username, password) {
        isAdmin := utils.CheckLocalAdminStatus(username) // id -nG
        if isAdmin { return true, true, false }
        return false, false, false // reject non-admin local users
    }
    return false, false, false
}
```

### Snippet 4: SambaTool wrapper con métodos tipados

```go
// Patrón: wrapper tipado para samba-tool con options struct
type SambaTool struct{}

type UserCreateOptions struct {
    FullName    string
    Email       string
    OUPath      string
    Description string
}

func (s *SambaTool) UserCreate(username, password string, opts UserCreateOptions) (string, error) {
    args := []string{"user", "create", username, password}
    if opts.FullName != ""    { args = append(args, "--given-name="+opts.FullName) }
    if opts.Email != ""       { args = append(args, "--mail-address="+opts.Email) }
    if opts.OUPath != ""      { args = append(args, "--userou="+opts.OUPath) }
    if opts.Description != "" { args = append(args, "--description="+opts.Description) }
    return s.Run(args...)
}

func (s *SambaTool) Run(args ...string) (string, error) {
    cmd, err := utils.SafeCommand("samba-tool", args...)
    if err != nil { return "", err }
    output, err := cmd.CombinedOutput()
    return string(output), err
}
```

### Snippet 5: JWT generation con claims de rol

```go
// Patrón: JWT con claims is_admin + is_domain_user
func (s *AuthService) GenerateToken(username string, isAdmin, isDomainUser bool) (*models.LoginResponse, error) {
    expiresAt := time.Now().Add(24 * time.Hour)
    claims := jwt.MapClaims{
        "username":       username,
        "user_id":        username,
        "is_admin":       isAdmin,
        "is_domain_user": isDomainUser,
        "exp":            expiresAt.Unix(),
        "iat":            time.Now().Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(utils.GetJWTSecret()))
    // ...
}
```

### Snippet 6: JWT middleware con context injection

```go
// Patrón: middleware que valida JWT e inyecta claims en gin.Context
func AuthRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" { /* 401 */ }

        token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
            return []byte(utils.GetJWTSecret()), nil
        })
        if err != nil || !token.Valid { /* 401 */ }

        if claims, ok := token.Claims.(jwt.MapClaims); ok {
            c.Set("username", claims["username"])
            c.Set("is_admin", claims["is_admin"])
            c.Set("is_domain_user", claims["is_domain_user"])
            c.Set("claims", claims)
        }
        c.Next()
    }
}
```

### Snippet 7: Audit middleware con context extraction

```go
// Patrón: audit logging middleware que extrae user, IP, user-agent
func AuditMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := GetAuditContext(c)
        LogDataAccess(ctx, "api_request", c.Request.URL.Path, true, map[string]interface{}{
            "method": c.Request.Method,
            "path":   c.Request.URL.Path,
        })
        c.Next()
        status := c.Writer.Status()
        LogDataAccess(ctx, "api_response", c.Request.URL.Path, status < 400, map[string]interface{}{
            "status": status,
        })
    }
}
```

### Snippet 8: Deployment script serving con template injection segura

```go
// Patrón: servir scripts con template variables + allowlist de nombres
func (h *DeploymentHandler) ServeDeploymentScript(c *gin.Context) {
    scriptName := c.Param("script")
    allowedScripts := []string{"domain-join-only.bat", "domain-join-with-tailscale.bat", "tailnet-add.bat"}
    // Validar contra allowlist
    // Leer archivo
    // Reemplazar templates: {{DOMAIN_NAME}}, {{DOMAIN_REALM}}, {{LOGIN_SERVER}}, {{AUTH_KEY}}
    // NUNCA inyectar password: {{ADMIN_PASSWORD}} → "PROMPT_FOR_PASSWORD"
    c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", scriptName))
    c.String(http.StatusOK, processedContent)
}
```

### Snippet 9: Bootstrap admin con bcrypt

```go
// Patrón: bootstrap admin credentials con bcrypt y file storage seguro
func (s *VexaAdminService) SetPassword(password string) error {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    rec := vexaAdminRecord{
        Username:     "vexa",
        PasswordHash: string(hash),
        CreatedAt:    time.Now().Unix(),
    }
    b, _ := json.MarshalIndent(rec, "", "  ")
    os.MkdirAll(filepath.Dir(s.storagePath), 0700)
    return os.WriteFile(s.storagePath, b, 0600)
}

func (s *VexaAdminService) Verify(username, password string) bool {
    if username != "vexa" { return false }
    b, err := os.ReadFile(s.storagePath)
    var rec vexaAdminRecord
    json.Unmarshal(b, &rec)
    return bcrypt.CompareHashAndPassword([]byte(rec.PasswordHash), []byte(password)) == nil
}
```

### Snippet 10: Frontend SSE consumer con fetch streaming

```typescript
// Patrón: consumir SSE con fetch + ReadableStream (sin EventSource, para POST)
const response = await fetch('/api/v1/domain/provision-with-output', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
  body: JSON.stringify({ domain, realm, dns_forwarder: dnsServers })
})

const reader = response.body?.getReader()
const decoder = new TextDecoder()

while (true) {
  const { done, value } = await reader.read()
  if (done) break
  const chunk = decoder.decode(value)
  const lines = chunk.split('\n')
  for (const line of lines) {
    if (line.trim().startsWith('data: ')) {
      const data = JSON.parse(line.trim().slice(6))
      if (data.type === 'output') { /* mostrar progreso */ }
      if (data.type === 'complete') { /* éxito */ }
    }
  }
}
```

---

## 9. Limitaciones

### 9.1 Arquitectura

1. **No usa librerías LDAP/Kerberos nativas de Go** — toda interacción con AD es vía CLI (`samba-tool`, `smbclient`, `ldbmodify`). Esto implica:
   - Parsing frágil de output de texto (regex/string splitting)
   - Sin connection pooling LDAP
   - Sin manejo de errores LDAP estructurado
   - Performance inferior a LDAP directo

2. **Command sanitizer en modo PERMISSIVE** — las políticas existen pero `SanitizeCommand()` solo loguea y permite todo. Riesgo de command injection si input del usuario llega a args de shell.

3. **Sin TLS en API** — HTTP plano en `:8080`, delega a Nginx (que en el bootstrap.sh solo configura `listen 80`).

4. **Sin refresh tokens** — JWT de 24h sin revocación. Logout solo limpia localStorage.

### 9.2 Funcionalidad

5. **DNS zones/records son STUBs** — retornan 501 Not Implemented
6. **"Join as DC" y "Migrate Domain"** en el wizard son placeholders (opacity-50, "Coming soon")
7. **DomainPolicies, OUs** — handlers existen pero funcionalidad limitada
8. **GetDomainInfo** retorna "UNKNOWN" — parsing no implementado
9. **ConfigureDomain** — retorna error "not implemented yet"
10. **ClearMustChangePassword** — retorna nil sin hacer nada (TODO)
11. **ToggleMustChangePassword** — solo setea el flag, no lo toggles realmente
12. **readSystemLogs** — tiene un bug: duplica el array y toma solo la mitad (`lines[len(lines)/2:]`)
13. **SetLogLevel** — no cambia el nivel real (TODO)

### 9.3 Seguridad

14. **CORS `Allow-Origin: *` + `Allow-Credentials: true`** — configuración inválida/insegura
15. **JWT secret**: si no hay `JWT_SECRET` env var, genera random en runtime — se pierde al reiniciar, invalidando todas las sesiones
16. **Contraseña de admin de provisioning**: generada pero no se muestra al usuario en el wizard (¿cómo la obtiene el admin?)
17. **Deployment scripts .bat**: pasan credenciales por variables de entorno visibles en el proceso
18. **Sin rate limiting** en endpoints de login
19. **Sin CSRF protection** (aunque JWT en header mitiga esto)

### 9.4 Frontend

20. **Muchos `console.log`** de debug en producción (App.tsx, LoginPage.tsx, SetupWizard.tsx)
21. **localStorage para JWT** — vulnerable a XSS
22. **Sin tests** (no hay archivos de test en el repo)
23. **React 19 + Vite 7** — versiones muy nuevas, posible inestabilidad

### 9.5 Operacional

24. **Solo soporta Debian/Ubuntu** (apt-get, systemd, systemctl)
25. **Requiere root** para bootstrap.sh y para el servicio (User=root en systemd)
26. **Bug conocido en LXC unprivileged** con Samba 4.19.x
27. **Sin Docker/container support** nativo
28. **Sin multi-DC** — solo gestiona un DC local

---

## 10. Patrones Destacables para SambaForge

### A Adoptar

| Patrón | Valor |
|--------|-------|
| Handler → Service → Exec wrapper | Separación limpia de capas |
| SambaTool tipado con options structs | Type safety sobre CLI strings |
| SSE streaming de provisioning | UX en tiempo real de comandos largos |
| Doble auth (domain + PAM fallback) | Resiliencia ante fallos de dominio |
| ProvisioningGate middleware | Bloqueo automático pre-provisioning |
| Audit logging con categorías | Trazabilidad completa |
| Bootstrap admin con bcrypt | Primer arranque seguro |
| Deployment scripts con template injection | Automatización de join sin exponer credenciales |
| SafeCommand wrapper | Capa de defensa para os/exec |
| Zustand + persist para auth | State management simple y efectivo |
| Axios interceptors para JWT | Token injection automático |
| Route guards por rol | RBAC en frontend |

### B Evitar/Mejorar

| Patrón | Problema | Mejora para SambaForge |
|--------|----------|----------------------|
| samba-tool CLI parsing | Frágil, sin tipos | Usar go-ldap-client o LDB directo |
| Command sanitizer PERMISSIVE | No valida nada | Implementar validación real |
| CORS `*` + credentials | Inseguro | Whitelist de origins |
| JWT en localStorage | XSS vulnerable | HttpOnly cookie + CSRF token |
| JWT random en runtime | Sesiones perdidas al reiniciar | Persistir secret en archivo |
| Sin TLS nativo | MITM en API interna | TLS termination con autocert |
| Sin rate limiting | Brute force | Rate limiter middleware |
| Sin refresh tokens | UX deficiente al expirar | Refresh token + rotation |
| Console.log en producción | Ruido, info leakage | Logger estructurado |
| readSystemLogs bug | Datos corruptos | Reescribir paginación |
| User=root en systemd | Privilegios excesivos | User=vexa con capabilities |

---

## 11. Lecciones Clave para SambaForge

1. **La capa exec wrapper es el patrón más reutilizable** — abstrae samba-tool detrás de tipos Go, pero el parsing de output sigue siendo frágil. SambaForge debería considerar go-ldap o LDB directo para operaciones que no requieran samba-tool específicamente.

2. **El provisioning wizard con SSE streaming es excelente UX** — el usuario ve progreso en tiempo real. SambaForge debe replicar este patrón.

3. **La doble autenticación (domain + PAM) es resiliente** — permite acceso de emergencia cuando el dominio falla. SambaForge debe mantener este patrón.

4. **El deployment script approach es pragmático** — generar .bat con templates server-side y no exponer credenciales. SambaForge debe soportar Linux (bash) además de Windows (bat/ps1).

5. **La falta de tests es un riesgo significativo** — SambaForge debe tener tests desde el día 1.

6. **El command sanitizer PERMISSIVE es un antipatrón** — tener la infraestructura pero no usarla es peor que no tenerla (da falsa sensación de seguridad).

7. **El audit logging está bien diseñado** — categorías, rotación, JSON lines, middleware automático. SambaForge debe adoptar este patrón.

8. **El bootstrap.sh es completo** — instala todo, construye, configura Nginx + systemd. SambaForge debería tener un installer similar pero con soporte Docker opcional.