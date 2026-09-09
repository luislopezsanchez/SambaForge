# Evaluación de librerías Go para SambaForge

> Documento técnico de la Fase 0.3. Evaluación de cada librería con justificación de selección.

---

## 1. go-ldap/ldap/v3 — Cliente LDAP

**Repo:** https://github.com/go-ldap/ldap  
**Import:** `github.com/go-ldap/ldap/v3`  
**Stars:** 2.5k+ | **Madurez:** Alta, ampliamente usado en producción

### Features soportadas
- Conexión: non-TLS, TLS, STARTTLS, custom dialer
- Bind: Simple Bind, GSSAPI (Kerberos), SASL, Unauthenticated Bind, NTLM
- Search: normal, paging (`SearchWithPaging`), async (`SearchAsync`), Server Side Sorting
- CRUD: Add, Delete, Modify, Modify DN
- Password Modify (RFC 3062)
- LDAPv3 Filter Compile/Decompile (con `ldap.EscapeFilter`)
- Extended Operations, Controls
- DirSync (para sync con AD)
- Syncrepl (replication)

### Por qué lo elegimos
- Es **la** librería LDAP para Go. No hay alternativa seria.
- Soporta GSSAPI bind (Kerberos) — crítico para autenticación contra Samba AD.
- `SearchWithPaging` maneja grandes directorios sin OOM.
- `ldap.EscapeFilter()` previene LDAP injection.
- `DirSync` nos permite sincronizar cambios del directorio.

### Patrón de uso para SambaForge

```go
// Conectar al LDAP local del DC (Samba escucha en 127.0.0.1:389)
l, err := ldap.DialURL("ldap://127.0.0.1:389")
if err != nil {
    return fmt.Errorf("LDAP connect: %w", err)
}
defer l.Close()

// Bind con credenciales del usuario (por sesión, nunca persistidas)
err = l.Bind(fmt.Sprintf("%s@%s", username, realm), password)
if err != nil {
    return fmt.Errorf("LDAP bind: %w", err)
}

// Buscar usuarios con paginación
searchReq := ldap.NewSearchRequest(
    baseDN,                       // "DC=samdom,DC=example,DC=com"
    ldap.ScopeWholeSubtree,
    ldap.DerefAlways,
    0, 0, false,
    fmt.Sprintf("(&(objectClass=user)(sAMAccountName=%s))", 
        ldap.EscapeFilter(username)),
    []string{"cn", "sAMAccountName", "mail", "memberOf"},
    nil,
)
result, err := l.Search(searchReq)

// Paginación para directorios grandes
result, err := l.SearchWithPaging(searchReq, 100)
```

### Consideraciones
- Samba AD no soporta clear-text LDAP binds (requiere TLS o Kerberos/GSSAPI). SambaForge debe conectar via STARTTLS o usar GSSAPI.
- El bind se hace con `user@REALM` (UPN format), no con DN completo — Samba AD lo prefiere.
- Para operaciones que requieren privilegios de admin (provisioning, GPO, FSMO), usar Kerberos ticket cache del Administrator, no LDAP bind.

---

## 2. gokrb5/v8 — Cliente Kerberos

**Repo:** https://github.com/jcmturner/gokrb5  
**Import:** `github.com/jcmturner/gokrb5/v8`  
**Stars:** 1k+ | **Madurez:** Alta, pure Go, sin CGO

### Features soportados
- Cliente Kerberos 5 completo (AS-REQ, TGS-REQ, TGS-REP)
- Login con password o keytab
- SPNEGO/GSSAPI para HTTP (autenticación web con Kerberos)
- Cambio de password (RFC 3244, port 464)
- Keytab load/parse
- Credential cache (CCache) — leer tickets existentes
- Service ticket requests
- Auto-renewal de TGT
- Diagnostics

### Por qué lo elegimos
- Pure Go, sin CGO — compila a binario estático.
- SPNEGO nos permite implementar SSO web (el browser se autentica con Kerberos).
- `ChangePasswd` nos permite implementar self-service password change sin `samba-tool`.
- `NewFromCCache` nos permite leer tickets existentes del sistema (si el admin ya hizo `kinit`).
- Validado por go-samba4 que usa la misma librería.

### Patrón de uso para SambaForge

```go
import (
    "github.com/jcmturner/gokrb5/v8/client"
    "github.com/jcmturner/gokrb5/v8/config"
    "github.com/jcmturner/gokrb5/v8/keytab"
)

// Crear cliente Kerberos con password
cfg, err := config.Load("/etc/krb5.conf")
cl := client.NewWithPassword(
    "Administrator",
    "SAMDOM.EXAMPLE.COM",
    password,
    cfg,
)

// Login (AS-REQ)
err = cl.Login()

// Cambiar password (self-service portal)
ok, err := cl.ChangePasswd(newPassword)

// SPNEGO para HTTP — autenticación web con Kerberos
import "github.com/jcmturner/gokrb5/v8/spnego"
h := http.HandlerFunc(appHandler)
http.Handle("/", spnego.SPNEGOKRB5Authenticate(h, &kt, service.Logger(l)))
```

### Consideraciones
- El `krb5.conf` se genera durante el provisioning. SambaForge debe detectar su path.
- Para SPNEGO web, necesitamos un keytab del servicio HTTP (samba-tool spn add HTTP/sambaforge.domain).
- Samba 4.24 cambia el comportamiento de PAC y canonicalización — verificar compatibilidad.

---

## 3. os/exec — Ejecución de samba-tool

**Stdlib:** `os/exec` (no requiere dependencia externa)

### Por qué es crítico
`samba-tool` es nuestro motor de operaciones. Todo lo que no hagamos via LDAP directo (provisioning, GPO, FSMO, trusts, backup) se hace ejecutando `samba-tool`.

### Patrones seguros para SambaForge

#### Patrón 1: Ejecución básica con captura de output

```go
func runSambaTool(args ...string) (stdout, stderr string, exitCode int, err error) {
    cmd := exec.CommandContext(ctx, "samba-tool", args...)
    
    var outBuf, errBuf bytes.Buffer
    cmd.Stdout = &outBuf
    cmd.Stderr = &errBuf
    
    err = cmd.Run()
    return outBuf.String(), errBuf.String(), cmd.ProcessState.ExitCode(), err
}
```

#### Patrón 2: Password via variable de entorno (NUNCA via CLI)

```go
// samba-tool lee PASSWD, PASSFD, PASSWD_FILE del entorno
// NUNCA pasar password como argumento (--adminpass=...) — es visible en ps
cmd := exec.CommandContext(ctx, "samba-tool", 
    "domain", "provision",
    "--server-role=dc",
    "--realm=SAMDOM.EXAMPLE.COM",
    "--domain=SAMDOM",
    "--dns-backend=SAMBA_INTERNAL",
    // No --adminpass aquí
)
cmd.Env = append(os.Environ(), "PASSWD="+adminPassword)
```

#### Patrón 3: Kerberos (sin password)

```go
// Si ya tenemos un ticket cache (kinit previo), samba-tool lo usa
cmd := exec.CommandContext(ctx, "samba-tool",
    "user", "create", username,
    "--use-kerberos=required",
)
// KRB5CCNAME env var apunta al ticket cache
cmd.Env = append(os.Environ(), "KRB5CCNAME="+ccachePath)
```

#### Patrón 4: Output JSON cuando está disponible

```go
cmd := exec.CommandContext(ctx, "samba-tool",
    "user", "list",
    "--json",
)
// Parsear JSON output en vez de texto
```

#### Patrón 5: Timeout y cancelación

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
cmd := exec.CommandContext(ctx, "samba-tool", "domain", "provision", ...)
// Si excede 30s, el proceso se mata automáticamente
```

### Reglas de seguridad para el wrapper de samba-tool

| Regla | Justificación |
|---|---|
| **NUNCA** pasar passwords como argumentos | Visibles en `ps`, race condition en scrub del process title |
| Usar `PASSWD` env var o `KRB5CCNAME` | samba-tool los lee nativamente |
| Allowlist de subcomandos | Mapear operación → subcomando fijo, nunca input del usuario como comando |
| Insertar `--` antes de positionals del usuario | Previene flag injection |
| Validar positionals (hostname, DN, username) | Schema estricto antes de ejecutar |
| Resolver path absoluto de samba-tool | `exec.LookPath("samba-tool")` — prevenir PATH planting |
| `exec.CommandContext` con timeout | Evitar procesos colgados |
| Capturar stderr siempre | samba-tool reporta errores en stderr |
| `--json` cuando esté disponible | Parseo confiable, no regex sobre texto |
| `--color=never` | Output limpio sin códigos ANSI |

---

## 4. Echo vs Gin — Framework HTTP

### Comparación

| Aspecto | Gin | Echo |
|---|---|---|
| GitHub stars | ~80k | ~30k |
| Primera release | 2014 | 2015 |
| HTTP library | net/http | net/http |
| Throughput | ~80k req/s | ~80k req/s |
| Memory per request | Low | Low |
| API style | `c *gin.Context` | `return error` (handler returns error) |
| Error handling | `c.Error()`, `c.AbortWithStatusJSON()` | Central error handler, handlers return error |
| Middleware ecosystem | Más grande (terceros) | Más built-in (logger, recover, CORS, JWT) |
| Auto TLS | No | **Sí** (Let's Encrypt integrado) |
| Template rendering | Buena | Excelente (multi-engine) |
| Binding/Validation | `c.ShouldBindJSON()` | `c.Bind()` (usa go-playground/validator) |
| Documentación | Sólida pero dispersa | Buena, más concisa |
| Acoplamiento | Alto (todo en gin.Context) | Medio (más cercano a net/http) |

### Decisión: **Echo**

**Justificación:**

1. **Auto TLS**: Echo integra Let's Encrypt nativamente (`e.StartAutoTLS()`). En un DC donde necesitamos TLS sin complicaciones, esto es invaluable. Gin requiere configuración manual.

2. **Error handling centralizado**: Los handlers retornan `error`, que un middleware central convierte a respuesta HTTP. Más limpio que el patrón `c.AbortWithStatusJSON()` de Gin.

3. **Middleware built-in**: Echo trae logger, recover, CORS, JWT middleware, rate limiter, CSRF — todo lo que SambaForge necesita sin depender de terceros.

4. **Menor acoplamiento**: Echo está más cerca de `net/http` estándar. Si algún día queremos mover un handler a net/http puro, es más fácil.

5. **go-samba4 usa Echo**: Ya validado en el mismo dominio (gestión de Samba AD DC). Podemos ver patrones concretos.

6. **Performance equivalente**: ~80k req/s en ambos. La diferencia no es significativa para un panel de administración (no es un API de alto throughput).

---

## 5. modernc.org/sqlite — Driver SQLite

### Comparación

| Aspecto | mattn/go-sqlite3 | modernc.org/sqlite |
|---|---|---|
| Implementación | Binding C (CGO) | Transpilación C→Go (pure Go) |
| CGO requerido | Sí | **No** |
| Cross-compilation | Difícil (necesita C toolchain) | **Trivial** (`CGO_ENABLED=0 go build`) |
| Performance | Ligeramente superior | Muy cercana (aceptable) |
| Binary size | Menor | Ligeramente mayor |
| Feature support | Completa (C extensions) | Casi completa |
| Docker scratch/Alpine | Requiere C toolchain | **Funciona directo** |
| Madurez | Muy alta (de facto standard) | Alta (battle-tested) |

### Decisión: **modernc.org/sqlite**

**Justificación:**

1. **Sin CGO**: SambaForge compila a un binario único estático. `CGO_ENABLED=0 go build` funciona sin C compiler. Esto es crítico para un DC donde puede no haber toolchain de compilación.

2. **Cross-compilation**: `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build` — compilar desde Windows a Linux sin configurar nada extra.

3. **Docker simplificado**: Imágenes `FROM scratch` o Alpine funcionan sin librerías C.

4. **Performance adecuada**: Un panel de administración no necesita el último 5% de performance. La diferencia es imperceptible para sesiones, audit log y configuración.

5. **Best practice 2026**: La comunidad Go recomienda modernc.org/sqlite como default para nuevos proyectos.

### Patrón de uso

```go
import (
    "database/sql"
    _ "modernc.org/sqlite"  // driver "sqlite"
)

db, err := sql.Open("sqlite", "file:sambaforge.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
db.SetMaxOpenConns(1)  // SQLite soporta un solo writer
```

---

## 6. pquerna/otp — TOTP 2FA

**Repo:** https://github.com/pquerna/otp  
**Import:** `github.com/pquerna/otp/totp`  
**Stars:** 2k+ | **Madurez:** Alta

### Features
- TOTP (RFC 6238) — compatible con Google Authenticator, Authy, etc.
- HOTP (RFC 4226)
- Generación de secretos
- Validación de códigos
- QR code generation (para setup)

### Por qué lo elegimos
- Es la librería TOTP estándar para Go.
- Validada por go-samba4 que la usa para 2FA.
- API simple: `totp.Generate()`, `totp.Validate()`.

### Patrón de uso

```go
import "github.com/pquerna/otp/totp"

// Generar secreto para un usuario nuevo
key, err := totp.Generate(totp.GenerateOpts{
    Issuer:      "SambaForge",
    AccountName: "admin@samdom.example.com",
    Period:      30,
    Digits:     otp.DigitsSix,
})

// key.URL() → string para QR code
// key.Secret() → secreto base32 para guardar en DB

// Validar código del usuario
valid := totp.Validate(userCode, key.Secret())
```

---

## 7. golang-jwt/jwt/v5 — Sesiones JWT

**Repo:** https://github.com/golang-jwt/jwt  
**Import:** `github.com/golang-jwt/jwt/v5`  
**Stars:** 8k+ | **Madurez:** Alta (sucesor de dgrijalva/jwt-go)

### Features
- HMAC (HS256/384/512) — simétrico
- RSA (RS256/384/512) — asimétrico
- ECDSA (ES256/384/512)
- RSA-PSS (PS256/384/512)
- EdDSA
- Custom claims
- Token parsing y validation

### Por qué lo elegimos
- Estándar de facto para JWT en Go.
- v5 es la versión mantenida activamente.
- Soporta todos los algoritmos que podamos necesitar.

### Patrón de uso

```go
import "github.com/golang-jwt/jwt/v5"

// Generar token de sesión
claims := jwt.MapClaims{
    "sub":      username,           // subject (username)
    "realm":    realm,               // dominio AD
    "role":     "domain-admin",      // rol RBAC
    "exp":      jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
    "iat":      jwt.NewNumericDate(time.Now()),
    "jti":      sessionID,           // ID único para revocación
}
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
tokenString, err := token.SignedString(jwtSecret)

// Validar token
token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
    if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
    }
    return jwtSecret, nil
})
```

### Decisión: HMAC (HS256) para sesiones internas
- SambaForge es un solo binario que firma y valida — no necesitamos clave pública separada.
- HMAC es más simple y rápido.
- El secreto se genera al instalar y se guarda en la config local.

---

## Resumen de dependencias Go

| Librería | Import | Propósito | CGO |
|---|---|---|---|
| Echo v4 | `github.com/labstack/echo/v4` | Framework HTTP, middleware, auto-TLS | No |
| go-ldap v3 | `github.com/go-ldap/ldap/v3` | Cliente LDAP (auth, CRUD directorio) | No |
| gokrb5 v8 | `github.com/jcmturner/gokrb5/v8` | Kerberos (login, SPNEGO, password change) | No |
| modernc.org/sqlite | `modernc.org/sqlite` | SQLite (sesiones, audit log, config) | **No** |
| golang-jwt v5 | `github.com/golang-jwt/jwt/v5` | JWT (sesiones web) | No |
| pquerna/otp | `github.com/pquerna/otp/totp` | TOTP (2FA) | No |
| os/exec | (stdlib) | Ejecución de samba-tool | No |

**Total: 6 dependencias externas + stdlib. Sin CGO. Binario estático.**

### Dependencias auxiliares (frontend, no Go)

| Librería | Propósito |
|---|---|
| React 19 + Vite | Frontend SPA |
| TypeScript | Type safety |
| Tailwind CSS | Styling |
| shadcn/ui | Componentes accesibles |
| Zustand | State management (ligero, simple) |
| React Router | Routing |
| react-i18next | i18n (es/en/pt) |
| TanStack Query | Data fetching + caching |
| axios | HTTP client |