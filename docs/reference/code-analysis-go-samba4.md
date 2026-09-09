# Análisis de Código: go-samba4

> **Repositorio:** https://github.com/jniltinho/go-samba4
> **Versión analizada:** v1.1.2 (commit HEAD al 2026-09-09)
> **Fecha de análisis:** 2026-09-09
> **Propósito:** Investigación Fase 0 — SambaForge. Entender cómo go-samba4 implementa LDAP, Kerberos, 2FA TOTP e i18n a nivel de código.

---

## Tabla de Contenidos

1. [Estructura de Directorios](#1-estructura-de-directorios)
2. [Backend Go: Arquitectura](#2-backend-go-arquitectura)
3. [Integración LDAP](#3-integración-ldap)
4. [Integración Kerberos](#4-integración-kerberos)
5. [Autenticación](#5-autenticación)
6. [i18n: Internacionalización](#6-i18n-internacionalización)
7. [CLI Tooling (Cobra)](#7-cli-tooling-cobra)
8. [Docker](#8-docker)
9. [Snippets Reutilizables](#9-snippets-reutilizables)
10. [Limitaciones](#10-limitaciones)

---

## 1. Estructura de Directorios

```
go-samba4/
├── main.go                          # Entry point — llama cmd.Execute() con embed.FS
├── embed.go                         # go:embed para templates, static y locales
├── go.mod / go.sum                   # Módulo Go 1.26, dependencias principales
├── config.toml.example               # Config de ejemplo (TOML via Viper)
├── Makefile                          # Build con ldflags + UPX + Tailwind CSS
├── Dockerfile                        # Multi-stage: golang:1.26-alpine → alpine
├── docker-compose.yml                # Servicio único con volumen y env vars
├── package.json                      # Tailwind CSS v4.2 devDependencies
│
├── cmd/                              # CLI Cobra commands
│   ├── root.go                       # rootCmd + initConfig (Viper)
│   ├── serve.go                      # `serve` — inicia Echo server
│   ├── migrate.go                    # `migrate` — GORM AutoMigrate
│   ├── user.go                       # `user` — stub (placeholder)
│   └── version.go                    # `version` — imprime build info
│
├── internal/
│   ├── config/
│   │   └── config.go                 # Config struct + Viper loader (TOML + env SAMBA4_*)
│   │
│   ├── auth/
│   │   ├── ldap.go                    # AuthenticateUser — LDAP bind + search + rebind
│   │   ├── kerberos.go                # SPNEGOMiddleware — ESQUELETO (no implementado)
│   │   ├── session.go                 # SessionManager — GORM-backed sessions
│   │   └── totp.go                    # TOTP: pquerna/otp, secret + validate + recovery
│   │
│   ├── ldap/
│   │   ├── client.go                  # Client wrapper: connect, bind, reconnect, Search
│   │   ├── users.go                   # CRUD users: GetAllUsers, GetUserBySAM, CreateUser, UpdateUser, DeleteUser, SetPassword
│   │   ├── groups.go                  # GetAllGroups
│   │   ├── ous.go                     # GetAllOUs
│   │   └── schema.go                  # Structs User/Group/OU + constantes atributos AD
│   │
│   ├── handlers/                      # HTTP handlers (Echo)
│   │   ├── handlers.go                # AppContext: Config, DB, LDAPClient, SessionMgr
│   │   ├── auth.go                    # Login GET/POST, Logout, isAdminUser
│   │   ├── users.go                   # CRUD users + writeAudit
│   │   ├── groups.go                  # Groups list/form
│   │   ├── ous.go                     # OU tree
│   │   ├── dashboard.go               # Stats: count users/groups/OUs + recent audit
│   │   ├── search.go                  # Búsqueda LDAP wildcard
│   │   ├── settings.go                # Settings GET/POST (stub POST)
│   │   ├── audit.go                   # Audit log list
│   │   └── lang.go                    # SetLanguage + LangFromRequest
│   │
│   ├── middleware/
│   │   ├── auth.go                    # RequireAuth + OptionalAuth
│   │   ├── csrf.go                    # CSRF: cookie + form/header token
│   │   ├── rbac.go                    # RequireAdmin + RequireAdminOrSelf
│   │   └── ratelimit.go               # In-memory rate limiter (5 req / 5 min)
│   │
│   ├── models/                         # GORM models
│   │   ├── session.go                 # Session: Token, Username, IsAdmin, ExpiresAt
│   │   ├── audit.go                   # AuditLog: AdminUser, Action, ObjectDN, Details
│   │   └── setting.go                 # Setting: Key, Value
│   │
│   ├── routes/
│   │   └── routes.go                  # RegisterRoutes — todos los endpoints
│   │
│   ├── server/
│   │   ├── serve.go                   # Serve() — wiring: DB, LDAP, i18n, Echo, TLS
│   │   ├── render.go                  # TemplateRegistry — html/template + layout
│   │   └── template_funcs.go          # FuncMap: T, TData, version, dict, etc.
│   │
│   ├── buildinfo/
│   │   └── buildinfo.go               # Version + BuildDate (ldflags injection)
│   │
│   └── i18n/
│       └── i18n.go                    # gotext PO loader + Translate() + normalizeLang
│
├── locales/                           # Archivos .po (GNU gettext)
│   ├── en/default.po                  # ~100 msgid/msgstr
│   ├── es/default.po                  # Español
│   └── pt_BR/default.po               # Portugués Brasil
│
├── web/
│   ├── templates/                     # HTML templates (go:embed)
│   │   ├── layout/
│   │   │   ├── base.html              # Layout principal (sidebar + header condicional)
│   │   │   └── sidebar.html           # Navegación + language switcher
│   │   ├── auth/login.html            # Form login con CSRF
│   │   ├── dashboard.html             # Cards stats + DataTables audit
│   │   ├── users/
│   │   │   ├── list.html              # DataTables + delete modal
│   │   │   ├── form.html             # Create/edit form
│   │   │   └── detail.html            # User detail
│   │   ├── groups/
│   │   │   ├── list.html
│   │   │   └── form.html
│   │   ├── ous/tree.html
│   │   └── audit/list.html
│   │
│   └── static/                        # Assets embebidos (go:embed)
│       ├── css/
│       │   ├── input.css              # Tailwind v4 source + @theme tokens
│       │   ├── app.css                # Tailwind compiled (output)
│       │   └── datatables.css
│       ├── js/
│       │   ├── app.js                 # jQuery init + CSRF + modals + UPN autofill
│       │   ├── jquery.min.js          # jQuery 4.0
│       │   ├── datatables.min.js      # DataTables 2.3.7
│       │   ├── lucide.min.js          # Lucide icons
│       │   └── i18n/
│       │       ├── es-ES.json          # DataTables lang ES
│       │       └── pt-BR.json          # DataTables lang PT
│       └── fonts/
│           ├── jetbrains-mono-regular.woff2
│           └── jetbrains-mono-bold.woff2
│
├── ssl/                               # Certificados SSL de desarrollo
│   ├── server.crt
│   └── server.key
│
├── .github/workflows/release.yml      # CI: tag → build → release GitHub
├── .agent/skills/                     # Skills del autor (no parte del app)
├── DOCUMENTS/                         # PRD + setup docs
│   ├── PRD-Samba4-AD-WebAdmin.md
│   ├── PROMPT-Antigravity-Samba4Admin.md
│   └── setup/setup-samba4.md
└── README.md
```

### Métricas rápidas

| Métrica | Valor |
|---------|-------|
| Archivos Go | ~25 |
| Líneas Go (estimado) | ~1,200 |
| Dependencias directas | 7 (go-ldap, echo/v5, pquerna/otp, cobra, viper, gorm + drivers) |
| Locales (.po) | 3 (en, es, pt_BR) — ~100 claves cada uno |
| Plantillas HTML | ~12 |
| Modelos GORM | 3 (Session, AuditLog, Setting) |
| Endpoints HTTP | ~17 |
| Comandos CLI | 4 (serve, migrate, user, version) |

---

## 2. Backend Go: Arquitectura

### 2.1 Patrón general

```
main.go → cmd.Execute(embed.FS) → Cobra rootCmd → initConfig(Viper)
                                        ↓
                               serveCmd → server.Serve(cfg, embed.FS)
                                        ↓
                               ┌────────────────────────┐
                               │  GORM (SQLite/MySQL)   │
                               │  LDAP Client (go-ldap) │
                               │  SessionManager        │
                               │  i18n.Init(embed)      │
                               └────────────────────────┘
                                        ↓
                               Echo v5 + TemplateRegistry
                                        ↓
                               routes.RegisterRoutes()
                                        ↓
                               handlers.AppContext methods
```

### 2.2 Entry Point (`main.go` + `embed.go`)

El binario es **auto-contenido**: templates, static assets y locales se embeben con `go:embed`.

```go
//go:embed all:web/templates
var TemplatesFS embed.FS

//go:embed all:web/static
var StaticFS embed.FS

//go:embed all:locales
var LocalesFS embed.FS
```

`main.go` es minimalista: pasa los 3 `embed.FS` a `cmd.Execute()`.

### 2.3 Configuración (`internal/config/config.go`)

Usa **Viper** con soporte TOML + environment variables:

```go
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    LDAP     LDAPConfig     `mapstructure:"ldap"`
    Database DatabaseConfig `mapstructure:"database"`
    Session  SessionConfig  `mapstructure:"session"`
    Security SecurityConfig `mapstructure:"security"`
    RBAC     RBACConfig     `mapstructure:"rbac"`
}
```

- **Búsqueda de config**: `./config.toml` → `/etc/go-samba4/config.toml` → `--config` flag
- **Env vars**: prefijo `SAMBA4_` (ej: `SAMBA4_LDAP_PASS` para bind password)
- **Defaults**: SQLite + `data.db`, port 8080, host `0.0.0.0`

### 2.4 Server Bootstrap (`internal/server/serve.go`)

`server.Serve()` es el orquestador:

1. **DB**: GORM Open (SQLite o MySQL según `driver`)
2. **AutoMigrate**: Session, AuditLog, Setting — automático en cada startup
3. **LDAP**: `ldap.NewClient()` — si falla, sigue (lazy reconnect)
4. **SessionManager**: `auth.NewSessionManager(db, cfg)`
5. **i18n**: `i18n.Init(localesFS)` — carga .po embebidos
6. **Echo**: RequestLogger + Recover middleware
7. **TemplateRegistry**: parsea `base.html` + `sidebar.html` + cada página
8. **Static files**: `DevMode` → filesystem real; producción → `embed.FS`
9. **Routes**: `routes.RegisterRoutes(e, appCtx, sm)`
10. **TLS**: si `tls_cert` + `tls_key` configurados → HTTPS; sino HTTP plano

### 2.5 Template Registry (`internal/server/render.go`)

Cada plantilla se parsea con `base.html` + `sidebar.html` + la página específica. La función `Render()` inyecta automáticamente:

```go
viewData["CSRFToken"] = c.Get("csrf")
viewData["Username"]  = c.Get("username")
viewData["Lang"]      = handlers.LangFromRequest(c)
viewData["CurrentPath"] = c.Request().URL.Path
```

### 2.6 Template Funcs (`internal/server/template_funcs.go`)

```go
"T":           func(lang, msgID string) string           // i18n simple
"TData":       func(lang, msgID string, data map[string]any) string  // i18n con sustitución
"version":     func() string                             // build version
"unescapeHTML": func(s string) template.HTML             // HTML raw
"safeCSS":     func(s string) template.CSS               // CSS raw
"hasPrefix":   strings.HasPrefix                         // active nav
"dict":        func(kvs ...any) (map[string]any, error)  // construir map en templates
```

### 2.7 Handlers y AppContext

```go
type AppContext struct {
    Config     *config.Config
    DB         *gorm.DB
    LDAPClient *ldap.Client
    SessionMgr *auth.SessionManager
}
```

Todos los handlers son métodos de `*AppContext`, estilo Echo v5.

### 2.8 Modelos GORM

Solo 3 modelos — todos en DB local (no AD):

```go
// Session — sesiones de usuario
type Session struct {
    ID        uint      `gorm:"primarykey"`
    Token     string    `gorm:"type:varchar(128);uniqueIndex"`
    Username  string    `gorm:"type:varchar(255);index"`
    IsAdmin   bool      `gorm:"default:false"`
    IPAddress string    `gorm:"type:varchar(64)"`
    UserAgent string    `gorm:"type:varchar(512)"`
    ExpiresAt time.Time `gorm:"index"`
    CreatedAt time.Time
}

// AuditLog — registro de acciones administrativas
type AuditLog struct {
    ID        uint      `gorm:"primarykey"`
    CreatedAt time.Time `gorm:"index"`
    AdminUser string    `gorm:"type:varchar(255);index"`
    IPAddress string    `gorm:"type:varchar(64)"`
    Action    string    `gorm:"type:varchar(100)"`
    ObjectDN  string    `gorm:"type:text"`
    Details   string    `gorm:"type:text"`
}

// Setting — configuración persistida (key-value)
type Setting struct {
    ID        uint   `gorm:"primarykey"`
    Key       string `gorm:"type:varchar(100);uniqueIndex"`
    Value     string `gorm:"type:text"`
    UpdatedAt time.Time
}
```

### 2.9 Rutas API (`internal/routes/routes.go`)

```go
// Públicas
GET  /                    → redirect /dashboard
GET  /lang/:code          → SetLanguage (cookie + redirect)

// Auth group (con CSRF)
GET  /auth/login          → AuthLoginGET
POST /auth/login          → AuthLoginPOST (con RateLimit)

// Protegidas (RequireAuth + CSRF)
GET  /dashboard           → DashboardGET
GET  /users               → UsersListGET
GET  /users/new           → UsersFormGET (RequireAdmin)
POST /users/new           → UsersCreatePOST (RequireAdmin)
GET  /users/:id           → UsersDetailGET
GET  /users/:id/edit      → UsersFormGET (RequireAdminOrSelf)
POST /users/:id/edit      → UsersUpdatePOST (RequireAdminOrSelf)
POST /users/:id/delete    → UsersDeletePOST (RequireAdmin)
GET  /groups              → GroupsListGET
GET  /groups/new          → GroupsFormGET
GET  /ous                 → OUsTreeGET
GET  /search              → SearchGET
GET  /audit               → AuditListGET
GET  /settings            → SettingsGET
POST /settings            → SettingsPOST
```

### 2.10 Middleware

| Middleware | Archivo | Función |
|-----------|---------|---------|
| RequireAuth | `auth.go` | Verifica cookie de sesión → redirect login si no válida |
| OptionalAuth | `auth.go` | Adjunta sesión si existe, no bloquea |
| CSRF | `csrf.go` | Token en cookie + validación en POST/PUT/DELETE |
| RequireAdmin | `rbac.go` | 403 si `!session.IsAdmin` |
| RequireAdminOrSelf | `rbac.go` | Admin o `:id == session.Username` |
| RateLimit | `ratelimit.go` | In-memory: 5 intentos / 5 min por IP |
| RequestLogger | Echo built-in | Logging de requests |
| Recover | Echo built-in | Panic recovery |

---

## 3. Integración LDAP

### 3.1 Librería y versión

- **`github.com/go-ldap/ldap/v3` v3.4.12** — la librería estándar de facto para LDAP en Go.
- Se importa con alias `goldap` para evitar colisión con el paquete local `ldap`.

### 3.2 Cliente LDAP (`internal/ldap/client.go`)

```go
type Client struct {
    conn   *goldap.Conn
    config *config.LDAPConfig
}
```

**Conexión** — soporta 3 modos:

| Modo | Condición | URL |
|------|-----------|-----|
| LDAP plano | `!UseTLS` | `ldap://host:port` |
| LDAPS (TLS directo) | `UseTLS && port==636` | `ldaps://host:636` |
| StartTLS | `UseTLS && port!=636` | `ldap://host:port` → `conn.StartTLS()` |

**Bind**: conexión con service account (`BindUser` + `BindPass` de config).

**Reconnect automático**: `Search()` detecta `ErrorNetwork` y reconecta transparentemente:

```go
func (c *Client) Search(searchRequest *goldap.SearchRequest) (*goldap.SearchResult, error) {
    res, err := c.conn.Search(searchRequest)
    if err != nil {
        if goldap.IsErrorWithCode(err, goldap.ErrorNetwork) {
            if reconnErr := c.Reconnect(); reconnErr == nil {
                return c.conn.Search(searchRequest)
            }
        }
        return res, err
    }
    return res, nil
}
```

**Domain helper**: deriva el DNS domain del BaseDN:

```go
func (c *Client) Domain() string {
    // "DC=corp,DC=local" → "corp.local"
}
```

### 3.3 Mapeo AD → Structs Go (`internal/ldap/schema.go`)

```go
type User struct {
    DN                 string
    SAMAccountName     string
    UserPrincipalName  string
    DisplayName        string
    GivenName          string
    SN                 string
    Mail               string
    TelephoneNumber    string
    Title              string
    Department         string
    UserAccountControl int
    MemberOf           []string
}

type Group struct {
    DN             string
    SAMAccountName string
    Description    string
    GroupType      int
    Member         []string
}

type OU struct {
    DN          string
    Name        string
    Description string
}
```

Constantes de atributos centralizadas:

```go
const (
    AttrSAMAccountName     = "sAMAccountName"
    AttrUserPrincipalName  = "userPrincipalName"
    AttrDisplayName        = "displayName"
    AttrGivenName          = "givenName"
    AttrSN                 = "sn"
    AttrMail               = "mail"
    AttrTelephoneNumber    = "telephoneNumber"
    AttrTitle              = "title"
    AttrDepartment         = "department"
    AttrUserAccountControl = "userAccountControl"
    AttrMemberOf           = "memberOf"
    AttrDescription        = "description"
    AttrGroupType          = "groupType"
    AttrMember             = "member"
    AttrObjectClass        = "objectClass"
    AttrOU                 = "ou"
)
```

### 3.4 Queries LDAP

**GetAllUsers** — lista completa con filtro personalizable:

```go
searchRequest := goldap.NewSearchRequest(
    c.config.BaseDN,
    goldap.ScopeWholeSubtree, goldap.NeverDerefAliases, 0, 0, false,
    fmt.Sprintf("(&(objectCategory=person)%s)", filter),
    []string{
        AttrSAMAccountName, AttrUserPrincipalName, AttrDisplayName,
        AttrGivenName, AttrSN, AttrMail, AttrTelephoneNumber, AttrTitle,
        AttrDepartment, AttrUserAccountControl, AttrMemberOf,
    },
    nil,
)
```

**GetUserBySAM** — búsqueda individual con escape de filtro:

```go
safeSAM := goldap.EscapeFilter(sam)
filter := fmt.Sprintf("(&(objectCategory=person)(objectClass=user)(sAMAccountName=%s))", safeSAM)
```

**Search handler** — wildcard multi-atributo:

```go
escapedQuery := goldap.EscapeFilter(fmt.Sprintf("*%s*", query))
filter := fmt.Sprintf("(|(sAMAccountName=%s)(displayName=%s)(mail=%s))",
    escapedQuery, escapedQuery, escapedQuery)
```

### 3.5 Paginación

**No hay paginación LDAP del lado del servidor.** Las búsquedas usan `SizeLimit=0, TimeLimit=0` (sin límite), lo que significa que traen **todos los resultados** de una sola vez. La paginación visual se maneja **client-side** con DataTables (jQuery).

> ⚠️ **Esto es una limitación significativa** — en dominios grandes (miles de usuarios), la consulta sin paginación puede agotar memoria o timeout. AD tiene un límite default de 1000 resultados por búsqueda no paginada.

### 3.6 Creación de Usuario — flujo de 3 pasos

```go
func (c *Client) CreateUser(u User, password, ouDN string) error {
    // 1. Crear cuenta deshabilitada (UAC=514)
    addRequest := goldap.NewAddRequest(dn, nil)
    addRequest.Attribute("objectClass", []string{"top", "person", "organizationalPerson", "user"})
    addRequest.Attribute(AttrUserAccountControl, []string{"514"}) // disabled
    c.conn.Add(addRequest)

    // 2. Set password (requiere LDAPS/StartTLS)
    c.SetPassword(dn, password)

    // 3. Habilitar cuenta (UAC=512 o valor personalizado)
    modEnable := goldap.NewModifyRequest(dn, nil)
    modEnable.Replace(AttrUserAccountControl, []string{fmt.Sprintf("%d", finalUAC)})
    c.conn.Modify(modEnable)
}
```

Si falla el SetPassword, elimina la cuenta creada (rollback manual).

### 3.7 SetPassword — codificación UTF-16LE

```go
func encodePassword(password string) ([]byte, error) {
    quoted := `"` + password + `"`
    runes := utf16.Encode([]rune(quoted))
    buf := make([]byte, len(runes)*2)
    for i, r := range runes {
        binary.LittleEndian.PutUint16(buf[i*2:], r)
    }
    return buf, nil
}
```

Esto reemplaza el atributo `unicodePwd` con la contraseña encerrada en comillas y codificada en UTF-16 Little Endian, como requiere Active Directory.

### 3.8 UserAccountControl (UAC) — bits

El código usa flags UAC directamente:

| Valor | Significado | Uso en código |
|-------|-------------|---------------|
| 512 | NORMAL_ACCOUNT | Base default |
| 514 | ACCOUNTDISABLE | Creación inicial |
| 2 | ACCOUNTDISABLE bit | `newUAC \|= 2` |
| 65536 | DONT_EXPIRE_PASSWORD | `newUAC \|= 65536` |
| 66048 | 512 + 65536 | Enabled + pwd never expires |
| 66050 | 514 + 65536 | Disabled + pwd never expires |

Los templates verifican estado con `{{if or (eq .UserAccountControl 512) (eq .UserAccountControl 66048)}}` para mostrar icono verde (activo) o rojo (deshabilitado).

---

## 4. Integración Kerberos

### 4.1 Estado: NO IMPLEMENTADO (esqueleto)

El archivo `internal/auth/kerberos.go` es un **placeholder**:

```go
// kerberos.go handles SPNEGO middleware for Windows SSO.
// Due to complexity and size constraints, a skeleton is instantiated here.
// In a full implementation, it uses github.com/jcmturner/gokrb5/v8
// to decode NegTokenInit and validate tickets against a keytab.

func SPNEGOMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader != "" {
            slog.Debug("SPNEGO header received, validation not fully implemented yet")
        }
        next.ServeHTTP(w, r)
    })
}
```

### 4.2 Observaciones

- **gokrb5 NO está en go.mod** — a pesar de mencionarse en el README, la librería `github.com/jcmturner/gokrb5/v8` no es una dependencia del proyecto.
- El middleware **ni siquiera está registrado** en las rutas de Echo — `SPNEGOMiddleware` existe pero nunca se monta.
- No hay keytab loading, no hay kinit, no hay validación de tickets.
- **Para SambaForge**: este es un punto donde go-samba4 no sirve como referencia. Habrá que implementar Kerberos desde cero.

### 4.3 Lo que el README promete vs. lo que existe

| Feature README | Estado real |
|---------------|-------------|
| "Kerberos (SSO)" | ❌ No implementado |
| "gokrb5" | ❌ No en go.mod |
| "SPNEGO" | ❌ Skeleton sin lógica |
| "kinit" | ❌ No existe |
| "Ticket validation" | ❌ No existe |

---

## 5. Autenticación

### 5.1 Flujo de Login LDAP (`internal/auth/ldap.go`)

El patrón es **bind → search → rebind** (doble bind):

```go
func AuthenticateUser(cfg *config.Config, username, password string) (*ldap.User, error) {
    // 1. Conectar al LDAP
    conn, err := goldap.DialURL(url, ...)

    // 2. Bind con service account
    conn.Bind(cfg.LDAP.BindUser, cfg.LDAP.BindPass)

    // 3. Buscar el DN real del usuario
    searchRequest := goldap.NewSearchRequest(
        cfg.LDAP.BaseDN,
        goldap.ScopeWholeSubtree, goldap.NeverDerefAliases, 0, 0, false,
        fmt.Sprintf("(&(objectClass=user)(sAMAccountName=%s))", goldap.EscapeFilter(username)),
        []string{"dn", ldap.AttrSAMAccountName, ldap.AttrDisplayName, ldap.AttrMemberOf},
        nil,
    )
    sr, err := conn.Search(searchRequest)

    // 4. Re-bind con el DN del usuario + password del form
    userDN := sr.Entries[0].DN
    err = conn.Bind(userDN, password)  // si falla → "invalid credentials"

    // 5. Retornar User struct
    return &ldap.User{DN: ..., SAMAccountName: ..., DisplayName: ..., MemberOf: ...}, nil
}
```

**Seguridad**: No hace bind directo con `username@domain` + password; primero resuelve el DN real vía service account. Esto es más robusto pero requiere una cuenta de servicio con permisos de lectura.

### 5.2 Determinación de Admin (RBAC)

```go
func isAdminUser(memberOf []string, adminGroup string) bool {
    if adminGroup == "" {
        return true  // si no hay grupo configurado, todos son admins
    }
    for _, g := range memberOf {
        if strings.EqualFold(strings.TrimSpace(g), strings.TrimSpace(adminGroup)) {
            return true
        }
    }
    return false
}
```

**Limitación**: solo verifica `admin_group`. Los grupos `operator_group`, `helpdesk_group`, `readonly_group` están en config pero **no se usan en el código** — el RBAC es binario (admin o no-admin).

### 5.3 Sesiones (`internal/auth/session.go`)

**DB-backed** (GORM), no JWT ni cookies firmadas:

```go
func (sm *SessionManager) CreateSession(username, ip, userAgent string, isAdmin bool) (*models.Session, error) {
    tokenBytes := asSecret(64)  // crypto/rand 64 bytes
    token := base64.URLEncoding.EncodeToString(tokenBytes)
    // ... persistir en DB con ExpiresAt
}
```

- Token: 64 bytes aleatorios base64url → ~86 chars
- Cookie: `samba4_admin_session`, HttpOnly, SameSite (Strict o Lax según config)
- Expiración: configurable (`session.timeout_minutes`, default 30)
- Validación: `WHERE token = ? AND expires_at > now()`

### 5.4 2FA TOTP (`internal/auth/totp.go`)

**Librería**: `github.com/pquerna/otp` v1.5.0

```go
// Generar secret + URL otpauth://
func GenerateSecret(accountName string) (string, string, error) {
    key, err := totp.Generate(totp.GenerateOpts{
        Issuer:      "Samba4 Admin",
        AccountName: accountName,
    })
    return key.Secret(), key.URL(), nil
}

// Validar código TOTP
func ValidatePasscode(passcode string, secret string) bool {
    return totp.Validate(passcode, secret)
}

// Códigos de recuperación (8 códigos base32)
func GenerateRecoveryCodes() []string {
    codes := make([]string, 8)
    for i := 0; i < 8; i++ {
        b := make([]byte, 5)  // 40 bits → 8 chars base32
        rand.Read(b)
        codes[i] = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
    }
    return codes
}
```

**Algoritmo**: TOTP RFC 6238 (SHA-1, 6 dígitos, 30s ventana) — defaults de `pquerna/otp`.

### 5.5 Estado del 2FA en el flujo

**TOTP existe pero NO está integrado en el login.** La config tiene `require_totp` pero:

- `AuthLoginPOST` no verifica TOTP después del login LDAP.
- No hay endpoint de setup/enroll TOTP.
- Los secrets y recovery codes no se persisten en ningún modelo GORM (no hay campo TOTP en `Session` ni en `Setting`).
- `require_totp` en config no se lee en ningún handler.

> ⚠️ **El 2FA es código muerto** — las funciones existen pero no se invocan desde ningún flujo de autenticación.

### 5.6 CSRF (`internal/middleware/csrf.go`)

Doble-submit cookie pattern:

1. GET → genera token aleatorio 32 bytes base64url → cookie `samba4_csrf` + `c.Set("csrf", token)`
2. POST/PUT/DELETE → compara cookie vs form `_csrf` o header `X-CSRF-Token`
3. Si no coincen → 403 Forbidden

El token se inyecta en templates via `{{.CSRFToken}}` y en meta tag para AJAX:

```html
<meta name="csrf-token" content="{{.CSRFToken}}">
```

```js
// app.js
$.ajaxSetup({ headers: { 'X-CSRF-Token': csrfToken } });
```

### 5.7 Rate Limiting (`internal/middleware/ratelimit.go`)

In-memory, **no distribuido** (no funciona con múltiples réplicas):

```go
var loginRatelimiter = &rateLimiter{
    visitors: make(map[string]*visitor),
    limit:    5,         // 5 intentos
    window:   5 * time.Minute,
}
```

- Goroutine background limpia visitantes expirados cada minuto
- Solo se aplica a `POST /auth/login`
- Retorna 429 Too Many Requests

---

## 6. i18n: Internacionalización

### 6.1 Arquitectura

```
locales/
├── en/default.po      ← GNU gettext .po (source de verdad)
├── es/default.po
└── pt_BR/default.po
```

**Librería**: `github.com/leonelquinteros/gotext` v1.7.2

### 6.2 Carga (`internal/i18n/i18n.go`)

```go
func Init(fsys embed.FS) {
    locales = make(map[string]*gotext.Po)
    entries, _ := fs.ReadDir(fsys, "locales")
    for _, entry := range entries {
        if !entry.IsDir() { continue }
        lang := entry.Name()
        poPath := "locales/" + lang + "/default.po"
        data, _ := fsys.ReadFile(poPath)
        po := gotext.NewPo()
        po.Parse(data)
        locales[lang] = po
    }
}
```

Los `.po` se cargan desde el `embed.FS` al iniciar la app.

### 6.3 Traducción con sustitución de variables

```go
func Translate(lang, msgID string, data map[string]any) string {
    // 1. Normalizar lang (pt, pt-br, pt_BR → "pt_BR"; es* → "es"; default → "en")
    // 2. Buscar PO
    // 3. po.Get(msgID)
    // 4. Si no encuentra, fallback a "en"
    // 5. Si data != nil, aplicar Go text/template con {{.Var}}
    translated := po.Get(msgID)
    if data != nil {
        tmpl, _ := template.New("").Parse(translated)
        tmpl.Execute(&buf, data)
    }
    return buf.String()
}
```

### 6.4 Uso en templates

```html
<!-- Simple -->
{{ T $.Lang "Nav_Dashboard" }}

<!-- Con datos -->
{{ TData $.Lang "Msg_AutoFillUPN" (dict "User" .Username) }}
```

### 6.5 Detección de idioma (`internal/handlers/lang.go`)

```go
func LangFromRequest(c *echo.Context) string {
    // 1. Cookie "samba4_lang"
    if cookie, err := c.Cookie("samba4_lang"); err == nil && validLangs[cookie.Value] {
        return cookie.Value
    }
    // 2. Accept-Language header (primeros 2 chars)
    accept := c.Request().Header.Get("Accept-Language")
    if len(accept) >= 2 {
        prefix := accept[:2]
        switch prefix {
        case "pt": return "pt"
        case "es": return "es"
        }
    }
    // 3. Default: "en"
    return "en"
}
```

Cambio de idioma: `GET /lang/:code` → set cookie 1 año → redirect al referer.

### 6.6 DataTables i18n (frontend)

Separado del backend — archivos JSON en `web/static/js/i18n/`:

```js
// app.js
if (lang === 'pt' || lang === 'pt_BR') {
    window.dtLang = { url: '/static/js/i18n/pt-BR.json' };
} else if (lang === 'es') {
    window.dtLang = { url: '/static/js/i18n/es-ES.json' };
} else {
    window.dtLang = {};
}
```

### 6.7 Catálogo de claves (~100 entradas por idioma)

| Categoría | Prefijo | Ejemplos |
|-----------|---------|----------|
| Navegación | `Nav_` | Nav_Dashboard, Nav_Users, Nav_Groups |
| Páginas | `Page_` | Page_Login, Page_NewUser, Page_EditUser |
| Botones | `Btn_` | Btn_NewUser, Btn_Save, Btn_Delete |
| Columnas | `Col_` | Col_Username, Col_Email, Col_Status |
| Labels | `Label_` | Label_Username, Label_Password |
| Mensajes | `Msg_` | Msg_DeleteConfirm, Msg_Loading |
| Errores | `Err_` | Err_InvalidCredentials, Err_Forbidden |
| Dashboard cards | `Card_` | Card_Users, Card_DCStatus |
| Settings | `Settings_` | Settings_Host, Settings_BaseDN |

---

## 7. CLI Tooling (Cobra)

### 7.1 Estructura

```
cmd/
├── root.go      → rootCmd + Execute() + initConfig()
├── serve.go     → serveCmd
├── migrate.go   → migrateCmd
├── user.go      → userCmd (stub)
└── version.go   → versionCmd
```

### 7.2 rootCmd

```go
var rootCmd = &cobra.Command{
    Use:   "go-samba4",
    Short: "Samba 4 Active Directory Web Administration",
}

func Execute(templates, static, locales embed.FS) error {
    tplFS = templates
    statFS = static
    localesFS = locales
    return rootCmd.Execute()
}

// Flag persistente: --config
rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
    "config file (default is ./config.toml or /etc/go-samba4/config.toml)")
```

### 7.3 Comandos

#### `serve`

```bash
./go-samba4 serve [--port 8080] [--debug] [--config /path/to/config.toml]
```

Inicia el server Echo con todo el wiring (DB, LDAP, i18n, templates, routes).

#### `migrate`

```bash
./go-samba4 migrate [--config /path/to/config.toml]
```

Ejecuta `db.AutoMigrate(&models.Session{}, &models.AuditLog{}, &models.Setting{})`.

> Nota: `server.Serve()` también ejecuta AutoMigrate en startup, así que `migrate` es redundante para el flujo normal.

#### `user`

```bash
./go-samba4 user
```

**Stub**: solo imprime "User management CLI tool." No tiene subcomandos ni lógica.

#### `version`

```bash
./go-samba4 version
# Go-Samba4 version v1.1.2 (Build Date: 2026-09-09 12:00:00)
```

Usa `buildinfo.Version` y `buildinfo.BuildDate` inyectados via ldflags en el Makefile:

```makefile
VERSION  = v1.1.2
LDFLAGS  = -X '$(PREFIX).Version=$(VERSION)' -X '$(PREFIX).BuildDate=$(DATE)'
```

### 7.4 Config con Viper

`initConfig()` se ejecuta via `cobra.OnInitialize()`:

```go
func initConfig() {
    cfg, err := config.LoadConfig(cfgFile)
    if err != nil {
        fmt.Println("Error reading configuration:", err)
        os.Exit(1)
    }
    globalCfg = cfg
}
```

---

## 8. Docker

### 8.1 Dockerfile (multi-stage)

```dockerfile
# Build Stage
FROM golang:1.26.0-alpine AS builder
RUN apk add --no-cache upx
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o go-samba4 .
RUN upx --best --lzma go-samba4

# Runtime Stage
FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata sqlite-libs
COPY --from=builder /app/go-samba4 .
COPY config.toml /etc/go-samba4/config.toml
EXPOSE 8080
CMD ["./go-samba4", "serve", "--config", "/etc/go-samba4/config.toml"]
```

**Observaciones**:
- `CGO_ENABLED=1` — necesario para `go-sqlite3` (SQLite driver)
- UPX comprime el binario final
- Runtime Alpine con solo `ca-certificates`, `tzdata`, `sqlite-libs`
- **Falta `npm`/node** en el builder para compilar Tailwind CSS — el Dockerfile copia directamente el source asumiendo que `app.css` ya está compilado

### 8.2 docker-compose.yml

```yaml
version: '3.8'
services:
  go-samba4:
    build: .
    container_name: go-samba4
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - samba4_data:/var/lib/go-samba4
      - ./config.toml:/etc/go-samba4/config.toml:ro
    environment:
      - SAMBA4_LDAP_PASS=secretpassword
volumes:
  samba4_data:
```

**Observaciones**:
- Un solo servicio (no hay Samba4 AD DC en el compose)
- Volume para persistir SQLite DB
- Config montada read-only
- Password LDAP via environment variable
- **No hay healthcheck**

### 8.3 CI/CD (`.github/workflows/release.yml`)

Trigger: tag `v*` → build (make build) → tar.gz → GitHub Release (softprops/action-gh-release).

---

## 9. Snippets Reutilizables

### Snippet 1: Cliente LDAP con auto-reconnect

```go
package ldap

import (
    "crypto/tls"
    "fmt"
    "net"
    "time"
    goldap "github.com/go-ldap/ldap/v3"
)

type Client struct {
    conn   *goldap.Conn
    config LDAPConfig
}

func NewClient(cfg *LDAPConfig) (*Client, error) {
    url := fmt.Sprintf("ldap://%s:%d", cfg.Host, cfg.Port)
    if cfg.UseTLS && cfg.Port == 636 {
        url = fmt.Sprintf("ldaps://%s:%d", cfg.Host, cfg.Port)
    }

    dialOpts := []goldap.DialOpt{
        goldap.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}),
    }

    var conn *goldap.Conn
    var err error

    if cfg.UseTLS {
        tlsConfig := &tls.Config{
            InsecureSkipVerify: cfg.SkipTLSVerify,
            ServerName:         cfg.Host,
        }
        if cfg.Port == 636 {
            conn, err = goldap.DialURL(url, goldap.DialWithTLSConfig(tlsConfig))
        } else {
            conn, err = goldap.DialURL(url, dialOpts...)
            if err == nil {
                err = conn.StartTLS(tlsConfig)
            }
        }
    } else {
        conn, err = goldap.DialURL(url, dialOpts...)
    }

    if err != nil {
        return nil, fmt.Errorf("failed to connect to LDAP: %w", err)
    }

    if err = conn.Bind(cfg.BindUser, cfg.BindPass); err != nil {
        conn.Close()
        return nil, fmt.Errorf("failed to bind to LDAP: %w", err)
    }

    return &Client{conn: conn, config: *cfg}, nil
}

// Search con auto-reconnect
func (c *Client) Search(req *goldap.SearchRequest) (*goldap.SearchResult, error) {
    res, err := c.conn.Search(req)
    if err != nil {
        if goldap.IsErrorWithCode(err, goldap.ErrorNetwork) {
            if reconnErr := c.Reconnect(); reconnErr == nil {
                return c.conn.Search(req)
            }
        }
    }
    return res, err
}
```

### Snippet 2: Codificación de password AD (UTF-16LE)

```go
import (
    "encoding/binary"
    "unicode/utf16"
)

// encodePassword returns the AD-compatible unicodePwd value:
// UTF-16LE encoding of the password surrounded by double quotes.
func EncodePassword(password string) ([]byte, error) {
    quoted := `"` + password + `"`
    runes := utf16.Encode([]rune(quoted))
    buf := make([]byte, len(runes)*2)
    for i, r := range runes {
        binary.LittleEndian.PutUint16(buf[i*2:], r)
    }
    return buf, nil
}

// SetPassword sets a user's password via unicodePwd (requires LDAPS or StartTLS).
func (c *Client) SetPassword(dn, newPassword string) error {
    encoded, err := EncodePassword(newPassword)
    if err != nil {
        return fmt.Errorf("failed to encode password: %w", err)
    }
    modRequest := goldap.NewModifyRequest(dn, nil)
    modRequest.Replace("unicodePwd", []string{string(encoded)})
    return c.conn.Modify(modRequest)
}
```

### Snippet 3: Creación de usuario AD en 3 pasos

```go
// CreateUser creates a new AD user in the given OU DN.
// Steps: 1) Add disabled account, 2) Set password, 3) Enable account.
func (c *Client) CreateUser(u User, password, ouDN string) error {
    dn := fmt.Sprintf("CN=%s,%s", goldap.EscapeFilter(u.DisplayName), ouDN)

    // 1. Create disabled account (UAC=514)
    addRequest := goldap.NewAddRequest(dn, nil)
    addRequest.Attribute("objectClass", []string{"top", "person", "organizationalPerson", "user"})
    addRequest.Attribute("sAMAccountName", []string{u.SAMAccountName})
    addRequest.Attribute("userAccountControl", []string{"514"}) // disabled

    if u.DisplayName != "" {
        addRequest.Attribute("displayName", []string{u.DisplayName})
    }
    if u.Mail != "" {
        addRequest.Attribute("mail", []string{u.Mail})
    }
    // ... otros atributos opcionales

    if err := c.conn.Add(addRequest); err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }

    // 2. Set password (requires LDAPS)
    if password != "" {
        if err := c.SetPassword(dn, password); err != nil {
            _ = c.conn.Del(goldap.NewDelRequest(dn, nil)) // rollback
            return fmt.Errorf("failed to set password: %w", err)
        }

        // 3. Enable account (UAC=512)
        finalUAC := 512
        if u.UserAccountControl > 0 {
            finalUAC = u.UserAccountControl
        }
        modEnable := goldap.NewModifyRequest(dn, nil)
        modEnable.Replace("userAccountControl", []string{fmt.Sprintf("%d", finalUAC)})
        if err := c.conn.Modify(modEnable); err != nil {
            return fmt.Errorf("failed to enable account: %w", err)
        }
    }
    return nil
}
```

### Snippet 4: Autenticación LDAP doble-bind (service → user)

```go
func AuthenticateUser(cfg *Config, username, password string) (*User, error) {
    // 1. Connect + bind with service account
    conn, err := goldap.DialURL(url)
    defer conn.Close()
    conn.Bind(cfg.LDAP.BindUser, cfg.LDAP.BindPass)

    // 2. Search for user's real DN
    searchRequest := goldap.NewSearchRequest(
        cfg.LDAP.BaseDN,
        goldap.ScopeWholeSubtree, goldap.NeverDerefAliases, 0, 0, false,
        fmt.Sprintf("(&(objectClass=user)(sAMAccountName=%s))",
            goldap.EscapeFilter(username)),
        []string{"dn", "sAMAccountName", "displayName", "memberOf"},
        nil,
    )
    sr, err := conn.Search(searchRequest)
    if err != nil || len(sr.Entries) == 0 {
        return nil, fmt.Errorf("user not found")
    }
    userDN := sr.Entries[0].DN

    // 3. Re-bind with user's DN + provided password
    if err := conn.Bind(userDN, password); err != nil {
        return nil, fmt.Errorf("invalid credentials")
    }

    return &User{
        DN:             userDN,
        SAMAccountName: sr.Entries[0].GetAttributeValue("sAMAccountName"),
        DisplayName:    sr.Entries[0].GetAttributeValue("displayName"),
        MemberOf:       sr.Entries[0].GetAttributeValues("memberOf"),
    }, nil
}
```

### Snippet 5: TOTP 2FA con pquerna/otp

```go
import "github.com/pquerna/otp/totp"

// Generate TOTP secret + otpauth:// URL for QR code
func GenerateTOTPSecret(accountName string) (secret, otpauthURL string, err error) {
    key, err := totp.Generate(totp.GenerateOpts{
        Issuer:      "SambaForge",
        AccountName:  accountName,
        // Defaults: SHA1, 6 digits, 30s period (RFC 6238)
    })
    if err != nil {
        return "", "", err
    }
    return key.Secret(), key.URL(), nil
}

// Validate a TOTP passcode
func ValidateTOTP(passcode, secret string) bool {
    return totp.Validate(passcode, secret)
}

// Generate 8 recovery codes (base32, no padding)
func GenerateRecoveryCodes() []string {
    codes := make([]string, 8)
    for i := 0; i < 8; i++ {
        b := make([]byte, 5)
        rand.Read(b)
        codes[i] = base32.StdEncoding.WithPadding(base32.NoPadding).
            EncodeToString(b)
    }
    return codes
}
```

### Snippet 6: i18n con gotext + go:embed

```go
package i18n

import (
    "bytes"
    "embed"
    "io/fs"
    "text/template"
    "github.com/leonelquinteros/gotext"
)

var locales map[string]*gotext.Po

func Init(fsys embed.FS) {
    locales = make(map[string]*gotext.Po)
    entries, _ := fs.ReadDir(fsys, "locales")
    for _, entry := range entries {
        if !entry.IsDir() { continue }
        lang := entry.Name()
        data, _ := fsys.ReadFile("locales/" + lang + "/default.po")
        po := gotext.NewPo()
        po.Parse(data)
        locales[lang] = po
    }
}

func Translate(lang, msgID string, data map[string]any) string {
    normalized := normalizeLang(lang)
    po, ok := locales[normalized]
    if !ok {
        po, ok = locales["en"]
        if !ok { return msgID }
    }
    translated := po.Get(msgID)
    if translated == "" || translated == msgID {
        if normalized != "en" {
            if enPo, ok := locales["en"]; ok {
                translated = enPo.Get(msgID)
            }
        }
        if translated == "" { return msgID }
    }
    if data == nil { return translated }
    // {{.Var}} substitution via Go templates
    tmpl, err := template.New("").Parse(translated)
    if err != nil { return translated }
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil { return translated }
    return buf.String()
}
```

### Snippet 7: Session manager con GORM + crypto/rand

```go
type SessionManager struct {
    db  *gorm.DB
    cfg *config.SessionConfig
}

func (sm *SessionManager) CreateSession(username, ip, userAgent string, isAdmin bool) (*Session, error) {
    tokenBytes := make([]byte, 64)
    rand.Read(tokenBytes)
    token := base64.URLEncoding.EncodeToString(tokenBytes)

    expiresAt := time.Now().
        Add(time.Duration(sm.cfg.TimeoutMinutes) * time.Minute)

    session := &Session{
        Token:     token,
        Username:  username,
        IsAdmin:   isAdmin,
        IPAddress: ip,
        UserAgent: userAgent,
        ExpiresAt: expiresAt,
        CreatedAt: time.Now(),
    }
    if err := sm.db.Create(session).Error; err != nil {
        return nil, err
    }
    return session, nil
}

func (sm *SessionManager) GetSession(token string) (*Session, error) {
    var session Session
    if err := sm.db.Where("token = ? AND expires_at > ?",
        token, time.Now()).First(&session).Error; err != nil {
        return nil, err
    }
    return &session, nil
}
```

### Snippet 8: CSRF middleware (doble-submit cookie)

```go
func CSRF() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c *echo.Context) error {
            req := c.Request()
            if req.Method == http.MethodGet {
                // Generate token
                tokenBytes := make([]byte, 32)
                rand.Read(tokenBytes)
                token := base64.URLEncoding.EncodeToString(tokenBytes)
                c.Set("csrf", token)
                http.SetCookie(c.Response(), &http.Cookie{
                    Name:     "csrf_token",
                    Value:    token,
                    Path:     "/",
                    HttpOnly: true,
                    SameSite: http.SameSiteStrictMode,
                })
            } else if req.Method == http.MethodPost ||
                       req.Method == http.MethodPut ||
                       req.Method == http.MethodDelete {
                // Validate token
                cookie, err := c.Cookie("csrf_token")
                if err != nil {
                    return echo.NewHTTPError(403, "CSRF cookie missing")
                }
                provided := c.FormValue("_csrf")
                if provided == "" {
                    provided = req.Header.Get("X-CSRF-Token")
                }
                if provided == "" || provided != cookie.Value {
                    return echo.NewHTTPError(403, "invalid CSRF token")
                }
            }
            return next(c)
        }
    }
}
```

### Snippet 9: Template registry con layout para Echo

```go
type TemplateRegistry struct {
    Templates map[string]*template.Template
}

func NewTemplateRegistry(tplFS embed.FS) (*TemplateRegistry, error) {
    t := &TemplateRegistry{Templates: make(map[string]*template.Template)}
    layout := "web/templates/layout/base.html"
    sidebar := "web/templates/layout/sidebar.html"

    fs.WalkDir(tplFS, "web/templates", func(filePath string, d fs.DirEntry, err error) error {
        if d.IsDir() || !strings.HasSuffix(d.Name(), ".html") { return nil }
        if filePath == layout || filePath == sidebar { return nil }

        tmplKey := strings.TrimSuffix(
            strings.TrimPrefix(filePath, "web/templates/"), ".html")

        tmpl := template.New(path.Base(filePath)).Funcs(TemplateFuncMap())
        tmpl, err := tmpl.ParseFS(tplFS, layout, sidebar, filePath)
        if err != nil { return err }

        t.Templates[tmplKey] = tmpl
        return nil
    })
    return t, nil
}

func (t *TemplateRegistry) Render(c *echo.Context, w io.Writer, name string, data any) error {
    tmpl := t.Templates[name]
    viewData := map[string]interface{}{}
    if d, ok := data.(map[string]interface{}); ok { viewData = d }
    viewData["CSRFToken"] = c.Get("csrf")
    viewData["Username"] = c.Get("username")
    viewData["Lang"] = LangFromRequest(c)
    return tmpl.ExecuteTemplate(w, "base", viewData)
}
```

### Snippet 10: Derivar dominio DNS del BaseDN

```go
// Domain derives the DNS domain from BaseDN.
// "DC=corp,DC=local" → "corp.local"
func (c *Client) Domain() string {
    var parts []string
    for _, component := range strings.Split(c.config.BaseDN, ",") {
        if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(component)), "DC=") {
            parts = append(parts, component[strings.Index(component, "=")+1:])
        }
    }
    return strings.Join(parts, ".")
}
```

---

## 10. Limitaciones

### 10.1 Limitaciones Críticas

| # | Limitación | Impacto | Archivo |
|---|-----------|---------|---------|
| 1 | **Kerberos/SPNEGO no implementado** | No SSO Windows; el README miente | `auth/kerberos.go` |
| 2 | **2FA TOTP no integrado en login** | Código muerto; `require_totp` ignorado | `auth/totp.go`, `handlers/auth.go` |
| 3 | **Sin paginación LDAP server-side** | Dominios >1000 usuarios pueden fallar (límite AD) | `ldap/users.go` |
| 4 | **RBAC binario (admin/no-admin)** | operator/helpdesk/readonly definidos pero no implementados | `handlers/auth.go`, `middleware/rbac.go` |
| 5 | **Rate limiter in-memory** | No funciona con múltiples réplicas | `middleware/ratelimit.go` |
| 6 | **Sin connection pool LDAP** | Una sola conexión compartida para todo | `ldap/client.go` |

### 10.2 Limitaciones Moderadas

| # | Limitación | Detalle |
|---|-----------|---------|
| 7 | **Dockerfile no compila Tailwind** | Asume `app.css` pre-compilado; falta node/npm en builder |
| 8 | **Sin healthcheck en Docker** | No hay endpoint `/health` ni HEALTHCHECK en Dockerfile |
| 9 | **`user` CLI es stub** | Solo imprime texto, no gestiona usuarios |
| 10 | **SettingsPOST no implementado** | Solo redirect, no procesa el form |
| 11 | **Sin tests** | No hay archivos `_test.go` en el repo |
| 12 | **Sin manejo de grupos (CRUD)** | Solo list y form GET; no hay create/update/delete de grupos |
| 12 | **Sin manejo de OUs (CRUD)** | Solo vista tree, no create/move/delete |
| 13 | **Sin reset de password por admin** | El campo password en edit es opcional pero no hay flujo de reset |
| 14 | **Sin API REST** | Todo es SSR con forms HTML; no hay JSON API |
| 15 | **Sin logging estructurado configurable** | `slog` a stdout, sin rotación ni niveles configurables |
| 16 | **Secrets en config.toml** | `session.secret` en texto plano en el config file |
| 17 | **Sin HTTPS redirect** | Si TLS no está configurado, corre en HTTP plano |

### 10.3 Limitaciones Menores

| # | Limitación |
|---|-----------|
| 18 | `fmt.Printf` en `dashboard.go` en vez de `slog` (logging inconsistente) |
| 19 | `systemAccounts` hardcodeado (krbtgt, Guest, etc.) en `handlers/users.go` |
| 20 | No hay validación de complejidad de password al crear/actualizar usuarios |
| 21 | UPN autofill en frontend depende de jQuery; no hay fallback |
| 22 | `.po` files no tienen `.mo` compilados (gotext los parsea en runtime) |
| 23 | `skip_tls_verify: true` en config de ejemplo (inseguro por defecto) |
| 24 | `go 1.26.0` — versión inusualmente alta (puede ser beta/RC) |

### 10.4 Lecciones para SambaForge

1. **Kerberos**: Implementar desde cero con `gokrb5/v8` — go-samba4 no sirve como referencia.
2. **2FA**: El approach con `pquerna/otp` es correcto, pero hay que integrarlo en el flujo de login y persistir secrets.
3. **Paginación LDAP**: Usar `goldap.NewSearchRequest` con `PagingSize` o implementar paged results control (OID 1.2.840.113556.1.4.319).
4. **RBAC**: Implementar los 4 roles (admin, operator, helpdesk, readonly) con middleware que verifique grupo AD.
5. **Connection pool**: Múltiples conexiones LDAP o pool para concurrencia.
6. **i18n**: El approach con gotext + .po + go:embed es sólido y reutilizable.
7. **Templates**: El pattern de TemplateRegistry con layout + FuncMap es limpio.
8. **Docker**: Agregar node al builder para compilar Tailwind; agregar healthcheck.

---

## Apéndice A: Dependencias Go (`go.mod`)

| Dependencia | Versión | Propósito |
|-------------|---------|-----------|
| `github.com/go-ldap/ldap/v3` | v3.4.12 | Cliente LDAP |
| `github.com/labstack/echo/v5` | v5.0.4 | Framework HTTP |
| `github.com/pquerna/otp` | v1.5.0 | TOTP 2FA |
| `github.com/spf13/cobra` | v1.10.2 | CLI framework |
| `github.com/spf13/viper` | v1.21.0 | Config (TOML + env) |
| `gorm.io/driver/mysql` | v1.6.0 | MySQL driver para GORM |
| `gorm.io/driver/sqlite` | v1.6.0 | SQLite driver para GORM |
| `gorm.io/gorm` | v1.31.1 | ORM |
| `github.com/leonelquinteros/gotext` | v1.7.2 | i18n (GNU gettext) |

**NOTA**: `gokrb5` NO está en go.mod a pesar de mencionarse en el README.

## Apéndice B: Stack Frontend

| Tecnología | Versión | Uso |
|------------|---------|-----|
| Tailwind CSS | 4.2+ | Framework CSS (Neo-Brutalism) |
| jQuery | 4.0.0 | AJAX, DOM, modals |
| DataTables | 2.3.7 | Tablas con paginación/búsqueda |
| Lucide Icons | — | Iconografía SVG |
| JetBrains Mono | — | Fuente monoespaciada embebida |

## Apéndice C: Config de ejemplo (`config.toml.example`)

```toml
[server]
host     = "0.0.0.0"
port     = 8080
tls_cert = ""
tls_key  = ""
dev_mode = false

[ldap]
host            = "dc1.empresa.local"
port            = 636
use_tls         = true
skip_tls_verify = true
base_dn         = "DC=empresa,DC=local"
bind_user       = "CN=samba4admin,CN=Users,DC=empresa,DC=local"
# bind_pass via SAMBA4_LDAP_PASS env var

[database]
driver = "sqlite"
path   = "data.db"
# MySQL: driver = "mysql", dsn = "user:pass@tcp(127.0.0.1:3306)/samba4admin"

[session]
secret           = "YOUR_SECRET_KEY_HERE"
timeout_minutes  = 30
cookie_secure    = true
cookie_same_site = "lax"

[security]
max_login_attempts = 5
lockout_minutes    = 15
require_totp       = false

[rbac]
admin_group    = "Domain Admins"
operator_group = "SambaWebOperators"
helpdesk_group = "SambaWebHelpdesk"
readonly_group = "SambaWebReadOnly"
```

---

*Análisis generado el 2026-09-09 para el proyecto SambaForge.*