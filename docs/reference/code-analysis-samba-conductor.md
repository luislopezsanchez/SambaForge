# Análisis de Código: Samba Conductor

> **Repositorio:** https://github.com/edimarlnx/samba-conductor
> **Fecha de análisis:** 2026-09-09
> **Propósito:** Investigación Fase 0 — SambaForge
> **Stack:** Meteor 3.4 + React 19 + Tailwind CSS 4 + Docker (Fedora minimal)

---

## 1. Estructura de Directorios

```
samba-conductor/
├── web/                          # Aplicación Meteor (Web UI)
│   ├── client/
│   │   ├── main.js               # Entry point React (createRoot)
│   │   ├── main.html             # HTML shell
│   │   └── main.css              # Tailwind CSS 4 + theme tokens
│   ├── server/
│   │   ├── main.js               # Server entry (importa todos los métodos)
│   │   ├── oauth2.js             # OAuth2 server (Authorization Code flow)
│   │   ├── rest.js               # REST API endpoints (/api, /api/metrics)
│   │   ├── health.js             # Health checks (memoria, heap, V8)
│   │   └── metrics.js            # Prometheus metrics (prom-client)
│   ├── app/
│   │   ├── auth/                 # Login, credentialStore, drKeyStore
│   │   ├── components/           # Button, DataTable, ThemeToggle, etc.
│   │   ├── computers/            # Gestión de equipos AD
│   │   ├── dashboard/            # Dashboard admin
│   │   ├── dns/                  # Gestión DNS
│   │   ├── domain/               # Info de dominio
│   │   ├── dr/                   # Disaster Recovery (sync, backup, restore)
│   │   ├── general/              # App, Router, RoutePaths, NotFound
│   │   ├── gpo/                  # Group Policy Objects
│   │   ├── groups/               # Gestión de grupos
│   │   ├── infra/                # Cron jobs, migrations
│   │   ├── layouts/              # AdminLayout, SelfServiceLayout, AdminGuard
│   │   ├── oauth/                # OAuth2 clients y realms (admin UI)
│   │   ├── ous/                  # Unidades Organizativas
│   │   ├── samba/                # Server-only: sambaAuth, sambaExec, sambaLdap...
│   │   ├── selfservice/          # Portal de autoservicio
│   │   ├── serviceaccounts/      # gMSA management
│   │   ├── settings/             # Settings (editable fields, sync account, S3)
│   │   ├── users/                # Gestión de usuarios
│   │   └── status/               # Server status
│   ├── .meteor/                  # Meteor config (packages, release, platforms)
│   ├── private/env/dev/settings.json  # Meteor settings (LDAP, realm, baseDn)
│   ├── Dockerfile                # Imagen web (Meteor + samba-tool)
│   ├── package.json
│   ├── rspack.config.js          # Build config (rspack + PostCSS)
│   └── babel.config.js           # React Compiler + e2e attr stripping
├── docker/
│   ├── docker-compose.yml        # Dev: DC primario + replica
│   ├── scripts/
│   │   └── samba-setup.sh        # Provisioning + join + TLS + Kerberos
│   ├── samba-ad-dc/              # Imagen standalone Fedora + Samba
│   └── all-in-one/               # Imagen todo-en-uno (Samba + MongoDB + Web)
├── docs/                         # Documentación de usuario
├── e2e/                          # Tests E2E (Playwright)
├── .github/workflows/            # CI: e2e-tests, docker-publish, deploy
├── CLAUDE.md                     # Guía de desarrollo
└── README.md
```

---

## 2. Backend: Framework y Arquitectura

### Framework: **Meteor 3.4** (no Express, no Node puro)

Meteor 3.4 es un framework fullstack que combina:
- **DDP** (Distributed Data Protocol) para comunicación cliente-servidor en tiempo real
- **MongoDB** como base de datos nativa (vía `meteor/mongo`)
- **Meteor Methods** para RPC (equivalente a API REST pero sobre WebSocket/long-poll)
- **WebApp handlers** para HTTP REST cuando se necesita (`meteor/webapp`)
- **Accounts** para autenticación integrada

### Meteor Packages clave (`.meteor/packages`):
```
meteor-base@1.5.2
mongo@2.2.0
accounts-password@3.0.6
react-meteor-data@4.0.1
quave:logged-user-react        # Hook useLoggedUser()
quave:alert-react-tailwind     # Sistema de alertas
quave:synced-cron              # Cron jobs
quave:migrations               # Migraciones de DB
quave:collections              # Wrapper de colecciones con schema
leaonline:oauth2-server        # OAuth2 server package
montiapm:agent                # Monitoring agent
rspack                        # Build tool moderno
```

### API Routes / Methods

**No hay rutas REST tradicionales (excepto OAuth2 y metrics).** Toda la comunicación es vía **Meteor Methods** (RPC sobre DDP):

| Método Meteor | Función |
|---|---|
| `auth.logout` | Limpia credenciales de sesión |
| `selfService.getProfile` | Perfil propio (READ, fallback a sync account) |
| `selfService.updateProfile` | Editar perfil propio (WRITE, requiere sesión) |
| `selfService.changePassword` | Cambiar contraseña propia |
| `users.*` (list/get/create/delete/move) | CRUD usuarios AD |
| `groups.*` (list/get/create/delete/addmembers) | CRUD grupos AD |
| `ous.*` | Unidades organizativas |
| `computers.*` | Equipos AD |
| `dns.*` | Zonas y records DNS |
| `gpo.*` | Group Policy Objects |
| `domain.*` | Info de dominio |
| `settings.get/set` | Configuración de la app |
| `settings.configureSyncAccount` | Crear cuenta sync en AD |
| `dr.*` (getStatus/triggerSync/configureS3/restore) | Disaster Recovery |
| `oauth.clients.*` | CRUD clientes OAuth2 |
| `oauth.realms.*` | CRUD realms OAuth2 |

**REST HTTP endpoints (via WebApp handlers):**
- `GET /api` — Health check simple
- `GET /api/metrics` — Prometheus metrics (con auth opcional via Basic Auth)
- `GET /oauth/authorize` — Página de autorización (server-rendered HTML)
- `POST /oauth/authorize` — Acepta/deniega autorización
- `POST /oauth/token` — Token endpoint (RFC 6749)
- `POST /oauth/meteor-login` — Login LDAP para OAuth2 flow
- `GET /oauth/check-token` — Valida token de sesión Meteor
- `GET /oauth/userinfo` — UserInfo endpoint (OIDC-style)

### Interacción con samba-tool

El módulo `app/samba/sambaExec.js` es el wrapper central:

```javascript
// Ejecuta samba-tool usando execFile (NO shell — previene inyección)
export async function runSambaTool({ args, credentials, timeout = 30000 }) {
  const { dockerContainer, sambaToolUrl } = getSambaConfig();

  // Agregar credenciales: -U username%password
  const authArgs = credentials
    ? [...args, '-U', `${credentials.username}%${credentials.password}`]
    : args;

  // Conexión remota opcional: -H ldap://dc1 --option=tls verify peer = no_check
  if (sambaToolUrl && !isCommandWithoutHostFlag({ args })) {
    authArgs.push('-H', sambaToolUrl, '--option=tls verify peer = no_check');
  }

  if (dockerContainer) {
    // Ejecutar dentro del contenedor Docker
    const dockerArgs = ['exec', dockerContainer, '/usr/bin/samba-tool', ...authArgs];
    ({ stdout, stderr } = await execFileAsync('/usr/bin/docker', dockerArgs, { timeout }));
  } else {
    // Ejecutar localmente (misma máquina)
    ({ stdout, stderr } = await execFileAsync('/usr/bin/samba-tool', authArgs, { timeout }));
  }

  return { stdout: stdout.trim(), stderr: stderr.trim() };
}
```

**Modos de operación:**
1. **Local** — samba-tool instalado en el mismo contenedor (all-in-one)
2. **Docker exec** — ejecuta samba-tool dentro del contenedor Samba DC
3. **Remoto** — vía `-H sambaToolUrl` (conexión LDAP remota)

**Sanitización de errores:** las credenciales se eliminan de los mensajes de error:
```javascript
const sanitizedMessage = message.replace(/-U\s+\S+/g, '-U ***');
```

### Interacción con LDAP

El módulo `app/samba/sambaLdap.js` usa `ldapjs` v3:

```javascript
import ldap from 'ldapjs';

// Crear cliente LDAP
export function createLdapClient() {
  const { ldapUrl, tlsRejectUnauthorized } = getSambaConfig();
  const options = { url: ldapUrl, connectTimeout: 10000, timeout: 30000 };
  if (ldapUrl.startsWith('ldaps://')) {
    options.tlsOptions = { rejectUnauthorized: tlsRejectUnauthorized };
  }
  return ldap.createClient(options);
}

// Bind con credenciales de sesión (UPN: username@realm)
export async function ldapBindWithCredentials({ client, credentials }) {
  const { realm } = getSambaConfig();
  const upn = `${credentials.username}@${realm}`;
  await ldapBind({ client, dn: upn, password: credentials.password });
}

// Search genérico — devuelve objetos JS
export function ldapSearch({ client, baseDn, filter, scope = 'sub', attributes = [] }) {
  return new Promise((resolve, reject) => {
    client.search(baseDn, { filter, scope, attributes }, (error, res) => {
      const entries = [];
      res.on('searchEntry', (entry) => {
        const obj = {};
        entry.pojo.attributes.forEach((attr) => {
          obj[attr.type] = attr.values.length === 1 ? attr.values[0] : attr.values;
        });
        obj.dn = entry.pojo.objectName;
        entries.push(obj);
      });
      res.on('end', () => resolve(entries));
    });
  });
}
```

**Patrón de uso:** cada operación LDAP crea un cliente, hace bind con las credenciales del usuario, ejecuta search/modify, y desconecta. No hay pool de conexiones — cada operación es atómica.

---

## 3. Frontend React

### Librerías principales (`package.json`):
| Librería | Versión | Propósito |
|---|---|---|
| `react` | 19.1.1 | UI framework |
| `react-dom` | 19.1.1 | DOM renderer |
| `react-router-dom` | 6.30.1 | Routing |
| `react-error-boundary` | 4.1.2 | Error boundaries |
| `tailwindcss` | 4.1.11 | Styling |
| `simpl-schema` | 3.4.6 | Validación de schemas |
| `@aws-sdk/client-s3` | 3.700.0 | S3 backups (server-side) |
| `ldapjs` | 3.0.7 | LDAP client (server-side) |
| `prom-client` | 15.1.3 | Prometheus metrics |

### State Management

**No usa Redux, Zustand, ni ningún state manager externo.** El state management se basa en:

1. **React hooks nativos** (`useState`, `useEffect`) para state local de componentes
2. **Meteor reactive data** vía `react-meteor-data` y `useLoggedUser` para datos reactivos del servidor
3. **Meteor Methods** (`Meteor.callAsync`) para toda comunicación servidor → patrones de fetch/mutation implícitos
4. **localStorage** para preferencias de UI (tema)

```javascript
// Patrón típico de componente
const [data, setData] = useState([]);
const [loading, setLoading] = useState(true);

useEffect(() => {
  Meteor.callAsync('users.list')
    .then(setData)
    .catch(err => openAlert(err.reason))
    .finally(() => setLoading(false));
}, []);
```

### Routing

`react-router-dom` v6 con `BrowserRouter`:

```javascript
// App.js
export function App() {
  return (
    <BrowserRouter>
      <Suspense fallback={<Loading />}>
        <AlertProvider>
          <Alert Component={MyAlert} />
          <Router />
        </AlertProvider>
      </Suspense>
    </BrowserRouter>
  );
}
```

**Layouts anidados:**
- `AnonymousLayout` — Login (sin auth)
- `LoggedLayout` → `ConditionalLayout` — Verifica sesión
- `SelfServiceLayout` — Portal de autoservicio (cualquier usuario)
- `AdminLayout` + `AdminGuard` — Panel admin (Domain Admins only)

**Estrategia de code-splitting:** imports estáticos (no lazy loading con `React.lazy`). El bundling lo maneja Rspack.

### UI Components

Componentes reutilizables en `app/components/`:
- `Button` — Botón con variantes (primary, secondary)
- `DataTable` — Tabla genérica con columnas configurables
- `ConfirmModal` — Modal de confirmación
- `StatCard` — Tarjeta de estadística para dashboard
- `ThemeToggle` — Switcher de temas
- `Loading` — Spinner
- `Logo` — Logo de la app
- `MyAlert` — Alert personalizado (via quave:alert-react-tailwind)
- `OUPicker` — Selector de OU

**Estilo:** Tailwind CSS 4 con tokens semánticos (`bg-surface`, `text-fg`, `border-border`). No usa componentes de terceros (no shadcn, no Material UI). Todo es CSS utility-first.

---

## 4. Zero Stored Credentials — Credenciales por Sesión

Esta es una de las features más notorias del proyecto. La implementación está en `app/auth/credentialStore.js`:

### Arquitectura

```
Login → LDAP bind → storeCredentials(userId, user, pass) → Map en memoria (AES-256-GCM)
  ↓
Operación AD → getCredentials(userId) → decrypt → usar → descartar
  ↓
Logout / TTL expira → clearCredentials(userId) → delete from Map
```

### Implementación clave:

```javascript
// credentialStore.js
import crypto from 'crypto';

const ALGORITHM = 'aes-256-gcm';
const IV_LENGTH = 16;
const TAG_LENGTH = 16;
const DEFAULT_TTL_MINUTES = 30;
const CLEANUP_INTERVAL_MS = 60 * 1000;

// Clave de encriptación: de env var (cluster-safe) o random por boot (más seguro)
const ENCRYPTION_KEY = process.env.CREDENTIAL_ENCRYPTION_KEY
  ? Buffer.from(process.env.CREDENTIAL_ENCRYPTION_KEY, 'hex')
  : crypto.randomBytes(32);

// Store en memoria: Map<userId, { encrypted, iv, tag, expiresAt }>
const store = new Map();

function encrypt({ text }) {
  const iv = crypto.randomBytes(IV_LENGTH);
  const cipher = crypto.createCipheriv(ALGORITHM, ENCRYPTION_KEY, iv);
  const encrypted = Buffer.concat([cipher.update(text, 'utf8'), cipher.final()]);
  const tag = cipher.getAuthTag();
  return { encrypted, iv, tag };
}

function decrypt({ encrypted, iv, tag }) {
  const decipher = crypto.createDecipheriv(ALGORITHM, ENCRYPTION_KEY, iv);
  decipher.setAuthTag(tag);
  return Buffer.concat([decipher.update(encrypted), decipher.final()]).toString('utf8');
}

export function storeCredentials({ userId, username, password }) {
  const payload = JSON.stringify({ username, password });
  const { encrypted, iv, tag } = encrypt({ text: payload });
  store.set(userId, { encrypted, iv, tag, expiresAt: Date.now() + getTtlMs() });
}

export function getCredentials({ userId }) {
  const entry = store.get(userId);
  if (!entry) throw new Meteor.Error('session-expired', 'No active session.');
  if (Date.now() > entry.expiresAt) {
    store.delete(userId);
    throw new Meteor.Error('session-expired', 'Session expired.');
  }
  return JSON.parse(decrypt({ encrypted: entry.encrypted, iv: entry.iv, tag: entry.tag }));
}
```

### Características de seguridad:
1. **AES-256-GCM** — cifrado autenticado (confidencialidad + integridad)
2. **IV aleatorio** por cada operación (16 bytes)
3. **Auth tag** GCM (16 bytes) — detecta tampering
4. **TTL de 30 minutos** (configurable via `Meteor.settings.samba.sessionTtlMinutes`)
5. **Cleanup automático** cada 60 segundos elimina entradas expiradas
6. **Clave random por boot** — si no se setea `CREDENTIAL_ENCRYPTION_KEY`, cada reinicio genera una nueva clave (las credenciales en memoria se pierden)
7. **Nada se escribe a disco** — el `Map` es volatile; al reiniciar el proceso, todo se pierde

### Separación Read/Write:

```javascript
// READ — fallback a sync account si la sesión expiró
export async function getReadCredentials({ userId }) {
  if (hasValidCredentials({ userId })) return getCredentials({ userId });
  const syncCreds = await getSyncCredentials(); // cuenta de servicio
  if (syncCreds) return syncCreds;
  throw new Meteor.Error('session-expired', '...');
}

// WRITE — siempre requiere sesión activa (sin fallback)
export function getWriteCredentials({ userId }) {
  return getCredentials({ userId });
}
```

Esto permite que operaciones de lectura (listar usuarios, ver perfil) funcionen incluso si la sesión del admin expiró, usando la "sync account" como fallback. Pero operaciones de escritura (crear usuario, cambiar password) requieren sesión activa del admin.

---

## 5. OAuth2 Server — Authorization Code Flow

### Implementación (`server/oauth2.js`)

Usa el package `leaonline:oauth2-server` (fork propio con fixes RFC 6749):

```javascript
import { OAuth2Server } from 'meteor/leaonline:oauth2-server';

const oauth2server = new OAuth2Server({
  serverOptions: {
    authorizationCodeLifetime: 300,   // 5 min
    accessTokenLifetime: 3600,         // 1 hora
    refreshTokenLifetime: 1209600,     // 14 días
    requireClientAuthentication: true, // client_secret en token exchange
    allowEmptyState: false,
  },
  model: {
    accessTokensCollectionName: 'oauth_access_tokens',
    refreshTokensCollectionName: 'oauth_refresh_tokens',
    clientsCollectionName: 'oauth_clients',
    authCodesCollectionName: 'oauth_auth_codes',
  },
  routes: {
    accessTokenUrl: '/oauth/token',
    authorizeUrl: '/oauth/authorize',
    errorUrl: '/oauth/error',
    fallbackUrl: '/oauth/*',
  },
});
```

### Flow completo:

1. **Cliente redirige a** `/oauth/authorize?client_id=X&redirect_uri=Y&response_type=code&state=Z&scope=openid profile`
2. **Server renderiza página HTML** de login/consentimiento (server-side rendered, no React):
   - Si el usuario ya tiene sesión Meteor (loginToken en localStorage), auto-autoriza
   - Si no, muestra formulario de login que hace POST a `/oauth/meteor-login`
3. **Login LDAP:** `POST /oauth/meteor-login` autentica contra Samba AD via `authenticateUser()`, genera token de sesión Meteor
4. **Autorización:** el browser hace POST automático a `/oauth/authorize` con el token y `allowed=true`
5. **Redirect:** el server redirige al cliente con `?code=AUTH_CODE&state=Z`
6. **Token exchange:** `POST /oauth/token` con `grant_type=authorization_code&code=X&client_id=Y&client_secret=Z`
7. **UserInfo:** `GET /oauth/userinfo` con `Bearer ACCESS_TOKEN` — devuelve claims OIDC-style

### UserInfo endpoint (custom implementation):

```javascript
// Valida Bearer token contra MongoDB
async function handleUserInfo(req, res) {
  const token = authHeader.startsWith('Bearer ') ? authHeader.slice(7) : null;
  const tokenDoc = await accessTokensRaw.findOne({ accessToken: token });

  if (tokenDoc.expiresAt < new Date()) { /* expired */ }

  const user = await Meteor.users.findOneAsync(tokenDoc.user?.id);
  const groups = (user.profile.memberOf || []).map((dn) => {
    const match = dn.match(/^CN=([^,]+)/);
    return match ? match[1] : dn;
  });

  res.end(JSON.stringify({
    sub: user._id,
    login: user.username,
    email: user.profile.email,
    name: user.profile.displayName,
    given_name: user.profile.givenName,
    family_name: user.profile.surname,
    groups,  // grupos AD extraídos del DN
  }));
}
```

### Gestión de clientes (admin UI):

```javascript
// oauthMethods.js
'oauth.clients.create': async function createClient({ clientName, redirectUris, realm, scopes, trusted }) {
  const clientId = crypto.randomUUID();
  const clientSecret = crypto.randomBytes(32).toString('hex'); // 64 chars hex

  // Registrar en el OAuth2 server
  await oauth2server.registerClient({
    title: clientName,
    redirectUris: redirectUris.join(','),
    grants: ['authorization_code', 'refresh_token'],
    clientId, secret: clientSecret,
  });

  // Guardar en colección propia para admin UI
  await OAuthClientsCollection.insertAsync({ clientId, clientName, redirectUris, realm, scopes, trusted, enabled: true });

  return { clientId, clientSecret }; // secret se muestra solo una vez
}
```

### Realms OAuth2:

Los realms permiten segmentar clientes OAuth2 por contexto (ej: "internal", "external"), con scopes permitidos y grupo AD de acceso:
```javascript
const OAuthRealmSchema = new SimpleSchema({
  name: { type: String },
  displayName: { type: String },
  allowedScopes: { type: Array },
  defaultScopes: { type: Array, optional: true },
  adGroupAccess: { type: String, optional: true }, // Grupo AD requerido
  enabled: { type: Boolean, defaultValue: true },
});
```

---

## 6. Self-Service Portal

### Rutas:
- `/` — Home del portal (perfil + quick actions)
- `/profile` — Editar perfil propio
- `/change-password` — Cambiar contraseña propia

### Implementación (`app/selfservice/`):

**Profile editing con campos configurables:**

```javascript
// selfServiceMethods.js
'selfService.updateProfile': async function updateOwnProfile({ fields }) {
  const credentials = getWriteCredentials({ userId: this.userId });

  // Solo campos habilitados por el admin en Settings
  const setting = await SettingsCollection.findOneAsync({ key: 'selfService.editableFields' });
  const editableFields = setting?.value || SETTINGS_DEFAULTS['selfService.editableFields'];

  const fieldToAdAttribute = {
    givenName: 'givenName',
    surname: 'sn',
    mail: 'mail',
    telephoneNumber: 'telephoneNumber',
    description: 'description',
    company: 'company',
    department: 'department',
    physicalDeliveryOffice: 'physicalDeliveryOfficeName',
  };

  // Filtrar solo campos habilitados
  Object.entries(fields).forEach(([key, value]) => {
    if (editableFields[key]?.enabled && fieldToAdAttribute[key]) {
      allowedAttributes[fieldToAdAttribute[key]] = value;
    }
  });

  // Usar sync account para LDAP modify (usuarios no pueden modificar su propio perfil)
  const syncCredentials = await getSyncCredentials();
  await updateUserAttributes({ username, attributes: allowedAttributes, credentials: syncCredentials });
}
```

**Cambio de contraseña con verificación:**

```javascript
'selfService.changePassword': async function changeOwnPassword({ currentPassword, newPassword }) {
  const meteorUser = await Meteor.users.findOneAsync(this.userId);
  const mustChange = meteorUser.profile?.mustChangePassword;

  // Verificar contraseña actual via LDAP bind (excepto si es cambio forzado)
  if (!mustChange) {
    await authenticateUser({ username: meteorUser.username, password: currentPassword });
  }

  // Usar credenciales actuales para samba-tool (o sin credenciales si es forzado)
  const credentials = mustChange ? undefined : { username: meteorUser.username, password: currentPassword };
  await resetPassword({ username: meteorUser.username, newPassword, credentials });

  // Actualizar flag y guardar nueva contraseña en sesión
  await Meteor.users.updateAsync(this.userId, { $set: { 'profile.mustChangePassword': false } });
  storeCredentials({ userId: this.userId, username: meteorUser.username, password: newPassword });
}
```

**Características:**
- Los usuarios ven su perfil AD (nombre, email, teléfono, grupos)
- Pueden editar campos que el admin haya habilitado en Settings
- Cambio de contraseña verifica la actual via LDAP bind
- Soporta password expirada (`pwdLastSet=0`) — redirige automáticamente al cambio
- Las modificaciones de perfil usan la sync account (no las credenciales del usuario, que no tiene permisos de escritura sobre su propio objeto)

---

## 7. Disaster Recovery — Backup Encriptado a S3

### Arquitectura DR:

```
┌─────────────┐     ┌──────────────┐     ┌───────────┐
│  AD Sync    │────▶│  MongoDB    │────▶│  S3 Backup│
│ (LDAP+tool) │     │ (snapshots)  │     │ (encriptado)│
└─────────────┘     └──────────────┘     └───────────┘
     ▲                    ▲                     │
     │                    │                     ▼
  Cron jobs          DR Key (AES)          S3-compatible
  (15min/6h/1h)      (in-memory)          (MinIO, AWS, etc.)
```

### DR Key Store (`app/auth/drKeyStore.js`):

El DR Key es una clave de encriptación separada de las credenciales de sesión:

```javascript
const ALGORITHM = 'aes-256-gcm';
const PBKDF2_ITERATIONS = 100000;
const KEY_LENGTH = 32;

// In-memory DR Key — null when locked
let drKey = null;

// Derivar clave de passphrase usando PBKDF2
function deriveKey({ key, salt }) {
  return crypto.pbkdf2Sync(key, salt, PBKDF2_ITERATIONS, KEY_LENGTH, 'sha512');
}

// Hash de verificación (no almacena la clave, solo un HMAC para validar)
function createVerificationHash({ key, salt }) {
  return crypto.createHmac('sha256', salt).update(key).digest('base64');
}

// Configurar DR Key (primera vez)
export async function configureDrKey({ key }) {
  if (key.length < 16) throw new Meteor.Error('dr.key.invalid', 'Min 16 chars');
  const salt = crypto.randomBytes(32);
  const verificationHash = createVerificationHash({ key, salt });
  const derivedKey = deriveKey({ key, salt });

  // Guardar salt + verificationHash en MongoDB (NO la clave)
  await SettingsCollection.upsertAsync({ key: 'dr.config' }, { $set: {
    value: { salt: salt.toString('base64'), verificationHash, configuredAt: new Date() }
  }});

  drKey = derivedKey; // En memoria solo
}

// Desbloquear después de restart — valida contra hash
export async function unlockDrKey({ key }) {
  const config = await SettingsCollection.findOneAsync({ key: 'dr.config' });
  const salt = Buffer.from(config.value.salt, 'base64');
  if (createVerificationHash({ key, salt }) !== config.value.verificationHash) {
    throw new Meteor.Error('dr.key.invalid', 'Invalid DR key');
  }
  drKey = deriveKey({ key, salt });
}
```

### Sync de datos AD a MongoDB (`app/dr/drSync.js`):

```javascript
// Sync metadata de usuarios via LDAP
export async function syncUsers({ credentials }) {
  const client = createLdapClient();
  await ldapBindWithCredentials({ client, credentials });
  const users = await ldapSearch({ client, baseDn, filter: '(&(objectClass=user)(objectCategory=person))', attributes: [...] });
  await AdSnapshotCollection.upsertAsync({ type: 'users' }, { $set: { snapshotAt: new Date(), data: users }});
  return { count: users.length };
}

// Sync password hashes — encriptados con DR Key
export async function syncUserHashes({ credentials }) {
  if (!isDrKeyUnlocked()) throw new Error('DR Key must be unlocked');

  const { stdout } = await runSambaTool({ args: ['user', 'list'], credentials });
  const usernames = parseListOutput({ output: stdout });

  const hashes = [];
  for (const username of usernames) {
    const { stdout: hashOutput } = await runSambaTool({
      args: ['user', 'getpassword', username, '--attributes=unicodePwd,supplementalCredentials'],
      credentials,
    });
    hashes.push({ username, hashData: hashOutput });
  }

  // Encriptar todos los hashes con DR Key antes de guardar
  const sensitiveData = drEncrypt({ text: JSON.stringify(hashes) });
  await AdSnapshotCollection.upsertAsync({ type: 'hashes' }, { $set: { sensitiveData, data: [{ count: hashes.length }] }});
}
```

### S3 Backup (`app/dr/drBackup.js`):

```javascript
import { S3Client, PutObjectCommand, ListObjectsV2Command, DeleteObjectCommand } from '@aws-sdk/client-s3';

// Ejecutar mongodump
async function runMongoDump() {
  const mongoUrl = process.env.MONGO_URL || 'mongodb://127.0.0.1:27017/samba-conductor';
  const outputPath = path.join(BACKUP_TMP_DIR, `mongodump-${timestamp}.archive`);
  await execFileAsync('mongodump', [`--uri=${mongoUrl}`, `--archive=${outputPath}`, '--gzip'], { timeout: 300000 });
  return outputPath;
}

// Ejecutar samba-tool domain backup
async function runSambaBackup() {
  if (dockerContainer) {
    // Backup dentro del contenedor DC
    await execFileAsync('docker', ['exec', dockerContainer, 'samba-tool', 'domain', 'backup', 'online', `--targetdir=/tmp/samba-backup`]);
    await execFileAsync('docker', ['cp', `${dockerContainer}:/tmp/samba-backup/.`, backupDir]);
  } else {
    await execFileAsync('samba-tool', ['domain', 'backup', 'online', `--targetdir=${backupDir}`]);
  }
  return path.join(backupDir, tarFile);
}

// Upload a S3
async function uploadToS3({ s3Client, bucket, key, filePath }) {
  const fileStream = fs.createReadStream(filePath);
  await s3Client.send(new PutObjectCommand({ Bucket: bucket, Key: key, Body: fileStream, ContentLength: stat.size }));
}

// S3 compatible (MinIO, Wasabi, etc.)
function createS3Client({ endpoint, region, accessKeyId, secretAccessKey }) {
  const config = { region, credentials: { accessKeyId, secretAccessKey } };
  if (endpoint) { config.endpoint = endpoint; config.forcePathStyle = true; }
  return new S3Client(config);
}
```

### Cron jobs (`app/infra/cron.js`):

```javascript
// Metadata sync: cada 15 minutos
SyncedCron.add({ name: 'DR: Sync AD metadata', schedule: parser => parser.text('every 15 minutes'),
  async job() { const creds = await getSyncCredentials(); await runFullSync({ credentials: creds }); }
});

// Hash sync: cada 6 horas (requiere DR Key unlocked)
SyncedCron.add({ name: 'DR: Sync password hashes', schedule: parser => parser.text('every 6 hours'),
  async job() { if (!isDrKeyUnlocked()) return; const creds = await getSyncCredentials(); await syncUserHashes({ credentials: creds }); }
});

// S3 backup: cada hora (verifica scheduleHours configurado)
SyncedCron.add({ name: 'DR: S3 Backup', schedule: parser => parser.text('every 1 hour'),
  async job() {
    const config = await SettingsCollection.findOneAsync({ key: 'backup.s3' });
    if (!config?.value?.enabled || !isDrKeyUnlocked()) return;
    const { runBackup } = require('../dr/drBackup');
    await runBackup({ includeMongo: config.includeMongoDump, includeSamba: config.includeSambaBackup });
  }
});
```

---

## 8. DC Replication via Environment Variables

La replicación de DCs se maneja completamente en el entrypoint de Docker (`docker/scripts/samba-setup.sh`):

### Variables de entorno:

| Variable | Default | Propósito |
|---|---|---|
| `SAMBA_REALM` | `SAMDOM.EXAMPLE.COM` | Realm Kerberos |
| `SAMBA_DOMAIN` | `SAMDOM` | NetBIOS domain |
| `SAMBA_ADMIN_PASSWORD` | (requerido) | Password del Administrator |
| `SAMBA_DNS_FORWARDER` | `8.8.8.8` | DNS forwarder |
| `SAMBA_SERVER_ROLE` | `dc` | Rol del server |
| `SAMBA_JOIN_AS_DC` | (unset) | `true` = join como replica |
| `SAMBA_PRIMARY_DC` | (unset) | Hostname del DC primario |
| `SAMBA_SITE` | (unset) | AD site name |
| `DATA_DIR` | `/data` | Directorio de datos persistentes |

### Lógica de join:

```bash
join_samba_domain() {
    if [ -f "$SAMBA_PROVISIONED" ] || [ -f "$SAMBA_JOINED" ]; then
        echo "Already provisioned/joined."
        return
    fi

    # Resolver IP del DC primario
    primary_ip=$(getent hosts "${SAMBA_PRIMARY_DC}" | awk '{print $1}')
    if [ -z "$primary_ip" ]; then
        primary_ip=$(dig +short "${SAMBA_PRIMARY_DC}" | head -1)
    fi

    # Apuntar DNS al primario
    echo "nameserver ${primary_ip}" > /etc/resolv.conf

    # Crear smb.conf mínimo antes del join
    cat > /etc/samba/smb.conf <<EOF
[global]
    workgroup = ${SAMBA_DOMAIN}
    realm = ${SAMBA_REALM}
    server role = active directory domain controller
    dns forwarder = ${SAMBA_DNS_FORWARDER}
    ad dc functional level = 2016
[sysvol]
    path = /var/lib/samba/sysvol
    read only = No
[netlogon]
    path = /var/lib/samba/sysvol/${realm_lower}/scripts
    read only = No
EOF

    # Esperar a que el primario esté listo (30 retries x 5s)
    while ! samba-tool domain info "${primary_ip}" >/dev/null 2>&1; do
        sleep 5
    done

    # Join como DC
    samba-tool domain join "${realm_lower}" DC \
        --server="${primary_ip}" \
        --dns-backend=SAMBA_INTERNAL \
        -U "Administrator%${SAMBA_ADMIN_PASSWORD}"

    touch "$SAMBA_JOINED"
}
```

### Docker Compose con replica:

```yaml
# docker/docker-compose.yml
services:
  samba-ad-dc:           # DC primario
    hostname: dc1
    environment:
      - SAMBA_REALM=SAMDOM.EXAMPLE.COM
      - SAMBA_ADMIN_PASSWORD=P@ssw0rd123!
    networks:
      samba-net:
        ipv4_address: 172.20.0.10

  samba-ad-dc-replica:   # DC replica (opcional)
    profiles: [replica]  # Solo con --profile replica
    hostname: dc2
    environment:
      - SAMBA_JOIN_AS_DC=true
      - SAMBA_PRIMARY_DC=dc1.samdom.example.com
      - SAMBA_ADMIN_PASSWORD=P@ssw0rd123!
    depends_on:
      - samba-ad-dc
    networks:
      samba-net:
        ipv4_address: 172.20.0.11
```

**Idempotencia:** los archivos `.provisioned` y `.joined` en `/data/samba/` marcan si ya se hizo la acción, evitando re-provisionar en restarts.

---

## 9. Theme Switching (Wine, Classic, Light)

### Implementación CSS (`client/main.css`):

Tailwind CSS 4 con `@theme` y CSS custom properties:

```css
@import 'tailwindcss';

@theme {
  --color-surface: var(--theme-surface);
  --color-surface-card: var(--theme-surface-card);
  --color-fg: var(--theme-fg);
  --color-fg-secondary: var(--theme-fg-secondary);
  --color-accent: var(--theme-accent);
  /* ... */
}

/* THEME: Wine (default) */
:root, .wine {
  --theme-surface: #1a0a14;
  --theme-surface-card: #2d1520;
  --theme-fg: #f5eff1;
  --theme-accent: #c45b7c;
}

/* THEME: Classic (gray) */
.classic {
  --theme-surface: #030712;
  --theme-surface-card: #111827;
  --theme-fg: #f3f4f6;
  --theme-accent: #2563eb;
}

/* THEME: Light */
.light {
  --theme-surface: #faf5f7;
  --theme-surface-card: #ffffff;
  --theme-fg: #1a0a14;
  --theme-accent: #9b2d50;
}
```

### Componente ThemeToggle (`app/components/ThemeToggle.js`):

```javascript
const THEMES = [
  { id: 'wine', label: 'Wine' },
  { id: 'classic', label: 'Classic' },
  { id: 'light', label: 'Light' },
];

function applyTheme({ theme }) {
  const html = document.documentElement;
  html.classList.remove('light', 'classic', 'wine');
  html.classList.add(theme);
}

export function ThemeToggle({ className = '' }) {
  const [theme, setTheme] = useState(getInitialTheme);
  const [open, setOpen] = useState(false);

  useEffect(() => { applyTheme({ theme }); }, [theme]);

  function handleSelect({ themeId }) {
    setTheme(themeId);
    localStorage.setItem('theme', themeId);
    setOpen(false);
  }
  // ... dropdown UI con preview de colores
}
```

**Mecanismo:** añade clase CSS al `<html>` element. La preferencia se guarda en `localStorage` (client-side, no persiste en servidor). No hay FOUC porque se aplica inmediatamente al cargar.

---

## 10. Docker Deployment

### Tres imágenes Docker:

| Imagen | Base | Contenido | Use case |
|---|---|---|---|
| **samba-ad-dc** | Fedora minimal 42 | Samba 4 AD DC solo | DC standalone |
| **web** | zcloud runtime | Meteor + samba-tool | Web UI contra DC externo |
| **all-in-one** | Fedora minimal + Node/Mongo | Samba + Web + MongoDB | Todo en un contenedor |

### All-in-One (`docker/all-in-one/Dockerfile`):

Multi-stage build:
1. **Builder:** `zcloudws/meteor-build:3.4` — compila la app Meteor
2. **Runtime:** `registry.fedoraproject.org/fedora-minimal:42` — Samba + Node + MongoDB + app

```dockerfile
# Stage 1: Build Meteor app
FROM zcloudws/meteor-build:3.4 AS builder
RUN meteor npm i --no-audit --legacy-peer-deps && meteor build --directory ../app-build

# Stage 2: Runtime (Fedora minimal)
FROM registry.fedoraproject.org/fedora-minimal:42
RUN microdnf install -y samba samba-dc samba-client samba-winbind krb5-workstation \
    bind-utils supervisor tdb-tools ldb-tools openssl && microdnf clean all

# Copiar Node.js y MongoDB del builder
COPY --from=builder /tmp/tools/.node /opt/nodejs/
COPY --from=builder /tmp/tools/.mongodb /opt/mongodb/
COPY --from=builder /app-source/app-build/bundle /opt/app

# Single volume
VOLUME ["/data"]
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
```

### Service management (Supervisor):

El entrypoint arma la config de supervisor condicionalmente:

```bash
# Samba siempre corre
cp "${TEMPLATES_DIR}/samba.conf" "${CONF_DIR}/"

if [ "${ENABLE_MONGODB}" = true ]; then
  cp "${TEMPLATES_DIR}/mongodb.conf" "${CONF_DIR}/"
fi

if [ "${ENABLE_WEBAPP}" = true ]; then
  cp "${TEMPLATES_DIR}/webapp.conf" "${CONF_DIR}/"
  # Generar METEOR_SETTINGS automáticamente si no viene
  if [ -z "${METEOR_SETTINGS}" ]; then
    METEOR_SETTINGS='{"samba":{"ldapUrl":"ldaps://127.0.0.1:636","baseDn":"...","realm":"...","tlsRejectUnauthorized":false}}'
  fi
fi

exec /usr/bin/supervisord -c /etc/supervisord.conf
```

### Seguridad Docker:

```yaml
# docker-compose.yml (all-in-one)
security_opt:
  - no-new-privileges:true
cap_drop:
  - ALL
cap_add:
  - NET_BIND_SERVICE  # Puertos privilegiados (53, 88, 389, 445)
  - SETUID            # Samba impersona usuarios AD
  - SETGID            # Samba impersona grupos AD
  - CHOWN             # Provisioning de sysvol/netlogon
  - FOWNER
  - DAC_OVERRIDE      # smbd accede archivos cross-user
```

---

## 11. Snippets Reutilizables

### Snippet 1: AES-256-GCM encrypt/decrypt para credenciales en memoria

```javascript
import crypto from 'crypto';

const ALGORITHM = 'aes-256-gcm';
const IV_LENGTH = 16;

const ENCRYPTION_KEY = process.env.CREDENTIAL_ENCRYPTION_KEY
  ? Buffer.from(process.env.CREDENTIAL_ENCRYPTION_KEY, 'hex')
  : crypto.randomBytes(32); // Random por boot

const store = new Map(); // userId → { encrypted, iv, tag, expiresAt }

function encrypt({ text }) {
  const iv = crypto.randomBytes(IV_LENGTH);
  const cipher = crypto.createCipheriv(ALGORITHM, ENCRYPTION_KEY, iv);
  const encrypted = Buffer.concat([cipher.update(text, 'utf8'), cipher.final()]);
  const tag = cipher.getAuthTag();
  return { encrypted, iv, tag };
}

function decrypt({ encrypted, iv, tag }) {
  const decipher = crypto.createDecipheriv(ALGORITHM, ENCRYPTION_KEY, iv);
  decipher.setAuthTag(tag);
  return Buffer.concat([decipher.update(encrypted), decipher.final()]).toString('utf8');
}

// TTL cleanup cada 60s
setInterval(() => {
  const now = Date.now();
  for (const [key, entry] of store) {
    if (now > entry.expiresAt) store.delete(key);
  }
}, 60 * 1000);
```

### Snippet 2: Wrapper seguro de samba-tool (execFile, no shell)

```javascript
import { execFile } from 'child_process';
import { promisify } from 'util';
const execFileAsync = promisify(execFile);

export async function runSambaTool({ args, credentials, timeout = 30000 }) {
  const authArgs = credentials
    ? [...args, '-U', `${credentials.username}%${credentials.password}`]
    : args;

  try {
    const { stdout, stderr } = await execFileAsync('/usr/bin/samba-tool', authArgs, { timeout });
    return { stdout: stdout.trim(), stderr: stderr.trim() };
  } catch (error) {
    // Sanitizar credenciales de errores
    const sanitized = (error.stderr || error.message).replace(/-U\s+\S+/g, '-U ***');
    throw new Error(`samba-tool error: ${sanitized}`);
  }
}
```

### Snippet 3: LDAP bind + search con ldapjs promisificado

```javascript
import ldap from 'ldapjs';

export function createLdapClient({ url, tlsRejectUnauthorized = true }) {
  const options = { url, connectTimeout: 10000, timeout: 30000 };
  if (url.startsWith('ldaps://')) {
    options.tlsOptions = { rejectUnauthorized: tlsRejectUnauthorized };
  }
  const client = ldap.createClient(options);
  client.on('error', (err) => console.error(`[LDAP] ${err.message}`));
  return client;
}

export function ldapSearch({ client, baseDn, filter, attributes = [] }) {
  return new Promise((resolve, reject) => {
    client.search(baseDn, { filter, scope: 'sub', attributes }, (error, res) => {
      if (error) return reject(error);
      const entries = [];
      res.on('searchEntry', (entry) => {
        const obj = {};
        entry.pojo.attributes.forEach((attr) => {
          obj[attr.type] = attr.values.length === 1 ? attr.values[0] : attr.values;
        });
        obj.dn = entry.pojo.objectName;
        entries.push(obj);
      });
      res.on('error', (err) => reject(err));
      res.on('end', () => resolve(entries));
    });
  });
}
```

### Snippet 4: DR Key con PBKDF2 + verificación sin almacenar la clave

```javascript
import crypto from 'crypto';

const PBKDF2_ITERATIONS = 100000;
const KEY_LENGTH = 32;

function deriveKey({ key, salt }) {
  return crypto.pbkdf2Sync(key, salt, PBKDF2_ITERATIONS, KEY_LENGTH, 'sha512');
}

function createVerificationHash({ key, salt }) {
  return crypto.createHmac('sha256', salt).update(key).digest('base64');
}

// Configurar: guardar salt + hash, no la clave
async function configureKey({ key }) {
  const salt = crypto.randomBytes(32);
  const verificationHash = createVerificationHash({ key, salt });
  const derivedKey = deriveKey({ key, salt });
  // Persistir solo salt + verificationHash
  await SettingsCollection.upsertAsync({ key: 'dr.config' }, { $set: {
    value: { salt: salt.toString('base64'), verificationHash }
  }});
  return derivedKey; // Solo en memoria
}

// Desbloquear: validar contra hash, derivar clave
async function unlockKey({ key }) {
  const config = await SettingsCollection.findOneAsync({ key: 'dr.config' });
  const salt = Buffer.from(config.value.salt, 'base64');
  if (createVerificationHash({ key, salt }) !== config.value.verificationHash) {
    throw new Error('Invalid key');
  }
  return deriveKey({ key, salt });
}
```

### Snippet 5: OAuth2 Authorization Code flow con página server-rendered

```javascript
// Página de autorización server-rendered (no React)
oauth2server.app.get('/oauth/authorize', function(req, res) {
  const { client_id, redirect_uri, response_type, state, scope } = req.query;

  const html = `<!DOCTYPE html><html>...
    <script>
      // Auto-login si hay sesión Meteor
      (async () => {
        const token = localStorage.getItem('Meteor.loginToken');
        if (token) {
          const res = await fetch('/oauth/check-token?token=' + token);
          if ((await res.json()).valid) {
            // Auto-submit authorization
            submitAuthorization(token);
          }
        }
      })();

      // Login manual via LDAP
      btn.onclick = async () => {
        const res = await fetch('/oauth/meteor-login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
          body: new URLSearchParams({ username, password })
        });
        submitAuthorization((await res.json()).token);
      };
    </script>`;

  res.writeHead(200, { 'Content-Type': 'text/html' });
  res.end(html);
});
```

### Snippet 6: S3 upload compatible con MinIO/Wasabi

```javascript
import { S3Client, PutObjectCommand } from '@aws-sdk/client-s3';
import fs from 'fs';

function createS3Client({ endpoint, region, accessKeyId, secretAccessKey }) {
  const config = { region, credentials: { accessKeyId, secretAccessKey } };
  if (endpoint) {
    config.endpoint = endpoint;
    config.forcePathStyle = true; // Requerido para MinIO
  }
  return new S3Client(config);
}

async function uploadToS3({ s3Client, bucket, key, filePath }) {
  const fileStream = fs.createReadStream(filePath);
  const stat = fs.statSync(filePath);
  await s3Client.send(new PutObjectCommand({
    Bucket: bucket, Key: key, Body: fileStream, ContentLength: stat.size,
  }));
  return { key, size: stat.size };
}
```

### Snippet 7: Docker entrypoint condicional con supervisor

```bash
#!/bin/bash
set -e

# Determinar servicios por env vars
ENABLE_WEBAPP=false; ENABLE_MONGODB=false
if [ -n "${ROOT_URL}" ]; then
  ENABLE_WEBAPP=true
  if [ -z "${MONGO_URL}" ]; then
    ENABLE_MONGODB=true
    MONGO_URL="mongodb://127.0.0.1:27017/samba-conductor"
  fi
fi

# Generar METEOR_SETTINGS si no viene
if [ -z "${METEOR_SETTINGS}" ]; then
  BASE_DN="DC=$(echo ${SAMBA_REALM} | tr '[:upper:]' '[:lower:]' | sed 's/\./,DC=/g')"
  METEOR_SETTINGS="{\"samba\":{\"ldapUrl\":\"ldaps://127.0.0.1:636\",\"baseDn\":\"${BASE_DN}\",\"realm\":\"${SAMBA_REALM}\"}}"
fi

# Ensamblar supervisor config
rm -f "${CONF_DIR}"/*.conf
cp "${TEMPLATES_DIR}/samba.conf" "${CONF_DIR}/"
[ "${ENABLE_MONGODB}" = true ] && cp "${TEMPLATES_DIR}/mongodb.conf" "${CONF_DIR}/"
[ "${ENABLE_WEBAPP}" = true ] && cp "${TEMPLATES_DIR}/webapp.conf" "${CONF_DIR}/"

exec /usr/bin/supervisord -c /etc/supervisord.conf
```

### Snippet 8: Sync account con password auto-generado y encriptado

```javascript
import crypto from 'crypto';

// Generar password fuerte (32 chars)
function generatePassword() {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%&*';
  const bytes = crypto.randomBytes(32);
  return Array.from(bytes).map((b) => chars[b % chars.length]).join('');
}

// Crear sync account en AD, encriptar password con DR Key
async function configureSyncAccount({ username, adminCredentials }) {
  const password = generatePassword();
  await createUser({ username, password, credentials: adminCredentials });

  // Agregar a Domain Admins
  await addGroupMember({ groupName: 'Domain Admins', memberName: username, credentials: adminCredentials });

  // Encriptar password con DR Key (persistente, el admin nunca lo ve)
  const encryptedPassword = drEncrypt({ text: password });
  await SettingsCollection.upsertAsync({ key: 'sync.account' }, { $set: {
    value: { configured: true, username, encryptedPassword }
  }});
}
```

### Snippet 9: Theme switching con CSS custom properties + localStorage

```javascript
// CSS: definir tokens semánticos que cambian por clase
:root, .wine { --theme-surface: #1a0a14; --theme-accent: #c45b7c; }
.classic { --theme-surface: #030712; --theme-accent: #2563eb; }
.light { --theme-surface: #faf5f7; --theme-accent: #9b2d50; }

// JS: toggle class en <html>, persistir en localStorage
function applyTheme({ theme }) {
  const html = document.documentElement;
  html.classList.remove('light', 'classic', 'wine');
  html.classList.add(theme);
}

const [theme, setTheme] = useState(() => localStorage.getItem('theme') || 'wine');
useEffect(() => { applyTheme({ theme }); localStorage.setItem('theme', theme); }, [theme]);
```

### Snippet 10: Samba DC join como replica via env vars

```bash
# Variables: SAMBA_JOIN_AS_DC=true, SAMBA_PRIMARY_DC=dc1.samdom.example.com

# Resolver IP del primario
primary_ip=$(getent hosts "${SAMBA_PRIMARY_DC}" | awk '{print $1}')
echo "nameserver ${primary_ip}" > /etc/resolv.conf

# Crear smb.conf mínimo
cat > /etc/samba/smb.conf <<EOF
[global]
    workgroup = ${SAMBA_DOMAIN}
    realm = ${SAMBA_REALM}
    server role = active directory domain controller
    ad dc functional level = 2016
EOF

# Esperar primario (30 retries)
while ! samba-tool domain info "${primary_ip}" >/dev/null 2>&1; do sleep 5; done

# Join como DC
samba-tool domain join "${realm_lower}" DC \
    --server="${primary_ip}" \
    --dns-backend=SAMBA_INTERNAL \
    -U "Administrator%${SAMBA_ADMIN_PASSWORD}"

touch "$SAMBA_JOINED"  # Idempotencia
```

---

## 12. Limitaciones

1. **MongoDB obligatorio** — A diferencia de un stack puro Node+Express, Meteor requiere MongoDB. No hay soporte para PostgreSQL u otras DBs. Esto añade complejidad operacional.

2. **Single-instance** — El credentialStore usa un `Map` en memoria. No funciona en multi-instancia sin `CREDENTIAL_ENCRYPTION_KEY` compartida, y aun así cada instancia tiene su propio Map. No hay sticky sessions ni Redis para compartir estado de sesión.

3. **No hay tests unitarios** — `package.json` tiene `"test": "exit 0"`. Solo hay tests E2E con Playwright. No hay cobertura de código ni tests de integración aislados.

4. **DR Key se pierde al reiniciar** — Si no se setea `DR_KEY` como env var, el key store queda locked después de un restart y requiere desbloqueo manual via UI. Esto puede ser problemático en despliegues automatizados.

5. **Sync account requiere Domain Admins** — La cuenta de sync se añade a `Domain Admins` para poder modificar perfiles de usuarios. Esto es excesivamente permisivo desde una perspectiva de seguridad (principio de mínimo privilegio).

6. **TLS con `rejectUnauthorized: false`** — En el modo all-in-one, la configuración por defecto desactiva la verificación TLS (`tlsRejectUnauthorized: false`). Aceptable para self-signed en localhost, pero peligroso si se cambia a red externa sin ajustar.

7. **No hay rate limiting** — Los endpoints de OAuth2 (login, token) no tienen rate limiting. Vulnerable a brute force.

8. **OAuth2 server package es un fork** — `leaonline:oauth2-server` es un fork con fixes RFC 6749. La viabilidad a largo plazo depende del mantenimiento del fork.

9. **No hay soporte multi-tenant / multi-dominio** — La configuración de Samba (`getSambaConfig()`) lee de `Meteor.settings.samba` que es estática. No se pueden administrar múltiples dominios Samba desde una sola instancia.

10. **Backup S3 no encripta los archivos en disco** — El `mongodump` y `samba-tool domain backup` se generan sin encriptar en `/tmp` antes de subir a S3. Si el filesystem es comprometido entre la generación y el upload, los datos están expuestos. La encriptación con DR Key se aplica solo a las credentials y hashes en MongoDB, no a los archivos de backup en sí.

11. **Password hash restore es experimental** — La inyección de hashes (`injectPasswordHash`) usa un enfoque no documentado de samba-tool. No hay garantía de compatibilidad entre versiones de Samba.

12. **No hay MFA/2FA** — El login es username+password contra LDAP. No hay soporte para multi-factor authentication.

13. **Meteor como dependencia pesada** — Meteor es un framework que ha perdido tracción en la comunidad. Su ecosistema de packages es pequeño y el build tooling es específico (no usa webpack/vite estándar). Esto crea vendor lock-in.

---

## Resumen Ejecutivo

Samba Conductor es una plataforma web para administrar Samba4 AD DC construida sobre **Meteor 3.4 + React 19 + Tailwind CSS 4**, con deployment Docker (Fedora minimal). Sus features más distintivas son:

- **Zero stored credentials:** Las credenciales de usuario se encriptan con AES-256-GCM y se guardan solo en un `Map` en memoria del servidor, con TTL de 30 minutos. Nada se persiste a disco. Separación read/write: reads pueden usar una "sync account" como fallback, writes requieren sesión activa.

- **OAuth2 server:** Authorization Code flow completo con página de consentimiento server-rendered, UserInfo endpoint OIDC-style, gestión de clients y realms desde admin UI. Usa un fork de `leaonline:oauth2-server` con fixes RFC 6749.

- **Disaster Recovery:** Sync periódico de metadata AD + password hashes a MongoDB (encriptados con DR Key via PBKDF2), backup de MongoDB + Samba a S3-compatible storage con retención configurable. Restore de usuarios y grupos desde snapshots.

- **DC Replication:** Join de DCs replica vía environment variables en Docker, idempotente con archivos `.provisioned`/`.joined`.

- **3 temas (Wine/Classic/Light):** CSS custom properties en Tailwind 4, toggle con localStorage, sin FOUC.

Las principales limitaciones para SambaForge: dependencia de MongoDB (via Meteor), single-instance sin HA, ausencia de tests unitarios, sync account sobre-permisiva (Domain Admins), y el lock-in al ecosistema Meteor.