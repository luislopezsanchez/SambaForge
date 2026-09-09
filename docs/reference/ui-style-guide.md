# SambaForge — Guía de estilo y wireframes

> Documento de la Fase 0.4. Define la dirección visual, componentes y layout de SambaForge.

---

## Filosofía de diseño

**"Cockpit for Samba AD"** — la herramienta que un sysadmin dejaría abierta en una pestaña del navegador todo el día, junto a Grafana y su monitoring.

### Principios
1. **Densidad informativa sin saturación** — un admin de dominio necesita ver mucho datos a la vez (usuarios, grupos, DNS, estado del DC). Pero no queremos una hoja de cálculo.
2. **Acciones rápidas, no wizards lentos** — todo lo que es de uso frecuente (crear usuario, reset password, añadir DNS record) debe ser 1-2 clicks. Los wizards solo para provisioning inicial.
3. **Feedback inmediato** — toda operación contra `samba-tool` muestra progreso y resultado. Errores son claros y accionables.
4. **Bilingüe nativo** — español e inglés al mismo nivel. No traducción de segundo orden.
5. **Dark mode first** — los sysadmin trabajan en salas de servidores y terminales. Dark mode es el default.

### Referencias visuales
- **Linear** — densidad, teclas rápidas, transiciones suaves
- **Grafana** — dashboards, cards de estado, tablas densas
- **Cockpit** — layouts funcionales, no decorativos
- **shadcn/ui** — componentes base (botones, modales, tablas, forms)

---

## Paleta de colores

### Dark mode (default)

| Token | Color | Uso |
|---|---|---|
| `--bg-base` | `#0a0a0b` | Fondo principal |
| `--bg-card` | `#141416` | Cards, panels |
| `--bg-elevated` | `#1a1a1d` | Hover, elevación |
| `--bg-input` | `#1e1e21` | Inputs, selects |
| `--border` | `#27272a` | Bordes sutiles |
| `--border-strong` | `#3f3f46` | Bordes activos |
| `--text-primary` | `#fafafa` | Texto principal |
| `--text-secondary` | `#a1a1aa` | Texto secundario |
| `--text-muted` | `#71717a` | Texto terciario, hints |
| `--accent` | `#6366f1` (indigo) | Acciones primarias, links |
| `--accent-hover` | `#5558e0` | Hover de accent |
| `--success` | `#22c55e` | DC activo, operación OK |
| `--warning` | `#f59e0b` | DC degradado, warnings |
| `--danger` | `#ef4444` | DC caído, errores |
| `--info` | `#3b82f6` | Info, notas |

### Light mode

| Token | Color | Uso |
|---|---|---|
| `--bg-base` | `#fafafa` | Fondo principal |
| `--bg-card` | `#ffffff` | Cards, panels |
| `--bg-elevated` | `#f4f4f5` | Hover |
| `--bg-input` | `#ffffff` | Inputs |
| `--border` | `#e4e4e7` | Bordes |
| `--border-strong` | `#d4d4d8` | Bordes activos |
| `--text-primary` | `#18181b` | Texto principal |
| `--text-secondary` | `#52525b` | Texto secundario |
| `--text-muted` | `#a1a1aa` | Texto terciario |
| `--accent` | `#6366f1` | Mismo accent |
| `--success` | `#16a34a` | |
| `--warning` | `#d97706` | |
| `--danger` | `#dc2626` | |

### Acento de marca
- Color principal: **Indigo** `#6366f1` — no es azul corporativo de Microsoft ni verde de Linux. Es propio.
- El logo de SambaForge usará indigo + un yunque (forge) estilizado.

---

## Tipografía

| Uso | Fuente | Tamaño | Peso |
|---|---|---|---|
| Display/Títulos | Inter | 20px | 600 |
| Body | Inter | 14px | 400 |
| Labels/Caption | Inter | 12px | 500 |
| Monospace (comandos, DNs) | JetBrains Mono | 13px | 400 |
| Tablas | Inter | 13px | 400 |

Inter se carga desde Google Fonts. JetBrains Mono para mostrar DNs, comandos samba-tool, rutas de archivos.

---

## Layout

### Estructura general

```
┌─────────────────────────────────────────────────────────────┐
│  Topbar: Logo  |  Domain selector  |  User  |  Theme  |  ⚙  │  56px
├──────────┬──────────────────────────────────────────────────┤
│          │                                                   │
│ Sidebar  │              Main content area                     │
│          │                                                   │
│ Dashboard│  ┌─────────────────────────────────────────────┐  │
│ Users    │  │  Page header: Title + breadcrumb + actions   │  │
│ Groups   │  ├─────────────────────────────────────────────┤  │
│ OUs      │  │                                              │  │
│ Computers│  │  Content (tables, forms, cards, wizards)     │  │
│ DNS      │  │                                              │  │
│ GPO      │  │                                              │  │
│ Settings │  └─────────────────────────────────────────────┘  │
│          │                                                   │
│ ──────── │                                                   │
│ Backup   │                                                   │
│ Audit    │                                                   │
│          │                                                   │
├──────────┴──────────────────────────────────────────────────┤
│  Status bar: DC status ● Active | Samba 4.24.6 | 42 users   │  32px
└─────────────────────────────────────────────────────────────┘
```

- **Sidebar:** 240px fijo, colapsable a 64px (solo iconos)
- **Topbar:** 56px — logo, domain/realm selector, usuario, theme toggle, settings
- **Status bar:** 32px — estado del DC, versión de Samba, contadores rápidos
- **Content:** Flexible, max-width 1400px, padding 24px

### Responsive
- **Desktop >1024px:** Sidebar visible, contenido a la derecha
- **Tablet 768-1024px:** Sidebar colapsable (overlay)
- **Mobile <768px:** Bottom navigation, contenido full-width

---

## Componentes principales

### Navigation
- Sidebar con secciones agrupadas:
  - **Dominio:** Dashboard, Users, Groups, OUs, Computers, Contacts
  - **Infraestructura:** DNS, GPOs
  - **Seguridad:** Password Policy, Audit Log, RBAC
  - **Sistema:** Backup, Multi-DC, Settings

### Tablas de datos
- shadcn/ui DataTable base
- Columnas: sortable, filterable, hideable
- Búsqueda global (topbar)
- Acciones por fila: inline (edit, delete, disable)
- Bulk actions: checkbox + barra de acciones
- Paginación: 25/50/100 por página
- Empty state: ilustración + texto + CTA

### Forms
- shadcn/ui Form (react-hook-form + zod)
- Validación en tiempo real
- Campos requeridos marcados
- Error messages claros
- Submit con loading state
- Cancelar siempre disponible

### Modals/Dialogs
- Para operaciones rápidas: crear usuario, reset password, añadir DNS record
- No para flows complejos — esos van a páginas dedicadas

### Toast notifications
- Success: verde, 3s auto-dismiss
- Error: rojo, persistente hasta dismiss
- Warning: amarillo, 5s
- Info: azul, 3s

### Status indicators
- Punto verde: servicio activo
- Punto amarillo: servicio degradado
- Punto rojo: servicio caído
- Spinner: operación en progreso

---

## Wireframes — Pantallas principales

### WF-01: Login

```
┌─────────────────────────────────────┐
│                                     │
│                                     │
│          🔨 SambaForge              │
│                                     │
│    ┌─────────────────────────┐      │
│    │  Usuario                 │      │
│    │  ┌─────────────────────┐ │      │
│    │  │ admin@domain.com    │ │      │
│    │  └─────────────────────┘ │      │
│    │                          │      │
│    │  Contraseña              │      │
│    │  ┌─────────────────────┐ │      │
│    │  │ ••••••••••••        │ │      │
│    │  └─────────────────────┘ │      │
│    │                          │      │
│    │  ☐ Recordar sesión       │      │
│    │                          │      │
│    │  ┌─────────────────────┐ │      │
│    │  │    Iniciar sesión    │ │      │
│    │  └─────────────────────┘ │      │
│    └─────────────────────────┘      │
│                                     │
│  Si el dominio no está provisionado │
│  → ir al Wizard de Instalación      │
│                                     │
└─────────────────────────────────────┘
```

- Login con credenciales del dominio (LDAP bind)
- Si no hay dominio provisionado → redirige a Wizard
- Después del login: si 2FA está activado, pide código TOTP

### WF-02: Dashboard

```
┌─────────────────────────────────────────────────────────────┐
│ SambaForge │ samdom.example.com ▾ │ admin │ 🌙 │ ⚙ │       │
├──────────┬──────────────────────────────────────────────────┤
│          │ Dashboard                                        │
│ ●Dashboard│ Inicio > Dashboard                               │
│ 👥 Users │                                                   │
│ 📁 Groups│ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐    │
│ 🏢 OUs   │ │ DC   │ │Users │ │Groups│ │Comput│ │ GPOs │    │
│ 💻 PCs   │ │● ON  │ │  42  │ │  15  │ │  28  │ │   8  │    │
│ 🌐 DNS   │ │4.24.6│ │      │ │      │ │      │ │      │    │
│ 📋 GPO   │ └──────┘ └──────┘ └──────┘ └──────┘ └──────┘    │
│          │                                                   │
│ ──────── │ Estado del Controlador de Dominio                 │
│ 💾 Backup│ ┌─────────────────────────────────────────────┐   │
│ 📊 Audit │ │ Samba 4.24.6  │  Uptime 23d 14h  │  FSMO ✓  │   │
│ ⚙ Settings│ │ DNS ✓          │  Kerberos ✓      │  LDAP ✓  │   │
│          │ │ Replication N/A│  SYSVOL ✓        │  Fsmo ✓  │   │
│          │ └─────────────────────────────────────────────┘   │
│          │                                                   │
│          │ Actividad reciente                               │
│          │ ┌─────────────────────────────────────────────┐   │
│          │ │ 10:32  Usuario 'jdoe' creado por admin       │   │
│          │ │ 09:15  Password reset para 'asmith'          │   │
│          │ │ 08:45  DNS record A 'wiki' → 10.0.0.15       │   │
│          │ │ 08:30  Grupo 'dev-team' modificado           │   │
│          │ │ ayer   Backup automático completado          │   │
│          │ └─────────────────────────────────────────────┘   │
├──────────┴──────────────────────────────────────────────────┤
│ ● Active │ Samba 4.24.6 │ 42 users │ 15 groups │ 28 computers│
└─────────────────────────────────────────────────────────────┘
```

### WF-03: Users list

```
┌──────────┬────────────────────────────────────────────────────┐
│          │ Usuarios                        [+ Nuevo Usuario]  │
│ ●Dashboard│ Inicio > Usuarios                                │
│ 👥 Users │                                                   │
│ 📁 Groups│ [🔍Buscar...]  [Estado: Todos ▾] [OU: Todas ▾]    │
│ 🏢 OUs   │                                                   │
│ 💻 PCs   │ ☐ │ Usuario      │ Nombre        │ OU     │ Estado│
│ 🌐 DNS   │───┼──────────────┼───────────────┼────────┼───────│
│ 📋 GPO   │ ☐ │ admin        │ Administrator │ Users  │ ●     │
│          │ ☐ │ jdoe         │ John Doe      │ Dev    │ ●     │
│          │ ☐ │ asmith       │ Alice Smith   │ IT     │ ◐     │
│          │ ☐ │ mbrown       │ Mike Brown    │ Sales  │ ●     │
│          │ ☐ │ lgarcia      │ Luis García   │ Sales  │ ●     │
│          │ ☐ │ ...          │ ...           │ ...    │ ...   │
│          │                                                   │
│          │ Mostrando 1-25 de 42    ‹ 1 2 3 ›  [25/pág ▾]    │
│          │                                                   │
│          │ ── Bulk actions (3 seleccionados) ──              │
│          │ [Deshabilitar] [Mover a OU...] [Exportar CSV]     │
└──────────┴──────────────────────────────────────────────────┘
```

### WF-04: Provisioning Wizard

```
┌─────────────────────────────────────────────────────────────┐
│ SambaForge — Asistente de Instalación                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ●━━━━━━━○━━━━━━○━━━━━━○━━━━━━○━━━━━━○                     │
│  Pre-req  Dominio  DNS   Admin  Review  Instalar              │
│                                                             │
│  Paso 1 de 6: Verificación de Pre-requisitos                 │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐     │
│  │ ✓ Samba 4.24.6 instalado                            │     │
│  │ ✓ Hostname: dc1 (< 15 chars)                        │     │
│  │ ✓ IP estática: 10.0.0.5                             │     │
│  │ ✗ /etc/hosts: FQDN resuelve a 127.0.0.1 (corregir)  │     │
│  │ ✓ Sin smb.conf previo                               │     │
│  │ ✓ Sin DBs previos (*.tdb, *.ldb)                    │     │
│  │ ⚠ avahi-daemon está corriendo (recomendado detener) │     │
│  └─────────────────────────────────────────────────────┘     │
│                                                             │
│  [Corregir automáticamente]           [Continuar →]          │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### WF-05: DNS Management

```
┌──────────┬────────────────────────────────────────────────────┐
│          │ DNS                              [+ Nuevo Record] │
│ 🌐 DNS   │ Inicio > DNS                                     │
│          │                                                   │
│ Zonas:   │ Zona: samdom.example.com ▾                        │
│ ▸ samdom │                                                   │
│ ▸ 0.0.10 │ Tipo  │ Nombre    │ Valor          │ TTL  │ Acción│
│          │───────┼───────────┼────────────────┼──────┼───────│
│ Records: │ A     │ dc1       │ 10.0.0.5       │ 3600 │ ✎ 🗑  │
│ A (15)   │ A     │ wiki      │ 10.0.0.15      │ 3600 │ ✎ 🗑  │
│ CNAME(3) │ CNAME │ mail      │ dc1.samdom...  │ 3600 │ ✎ 🗑  │
│ MX (2)   │ MX     │ @         │ 10 mail.samdom │ 3600 │ ✎ 🗑  │
│ SRV (8)  │ SRV   │ _ldap._tcp │ 0 100 389 dc1  │ 3600 │ ✎ 🗑  │
│ TXT (1)  │ TXT   │ @         │ "v=spf1 ..."   │ 3600 │ ✎ 🗑  │
│          │                                                   │
│          │ Forwarders: 8.8.8.8, 8.8.4.4        [Editar]      │
│          │ Scavenging: Desactivado               [Activar]   │
└──────────┴──────────────────────────────────────────────────┘
```

### WF-06: GPO Management

```
┌──────────┬────────────────────────────────────────────────────┐
│          │ Group Policy Objects              [+ Nuevo GPO]   │
│ 📋 GPO   │ Inicio > GPO                                      │
│          │                                                   │
│          │ Nombre GPO       │ Linked To    │ Estado  │ Acción│
│          │──────────────────┼─────────────┼─────────┼───────│
│          │ Default Policy   │ Domain      │ Activo  │ ✎ 🗑  │
│          │ Password Policy  │ Domain      │ Activo  │ ✎ 🗑  │
│          │ Drive Maps       │ IT OU       │ Activo  │ ✎ 🗑  │
│          │ Wallpaper        │ Domain      │ Inactivo│ ✎ 🗑  │
│          │ Logon Script     │ Sales OU    │ Activo  │ ✎ 🗑  │
│          │                                                   │
│          │ Plantillas preconfiguradas:                       │
│          │ ┌────────────┐ ┌────────────┐ ┌────────────┐      │
│          │ │ Password   │ │ Drive Maps │ │ Wallpaper  │      │
│          │ │ Policy     │ │            │ │            │      │
│          │ │ Complejidad│ │ Mapear UN: │ │ Fondo      │      │
│          │ │ Longitud   │ │ a rutas    │ │ corporativo│      │
│          │ │ Historial  │ │ de red     │ │            │      │
│          │ │ Expiración │ │            │ │            │      │
│          │ └────────────┘ └────────────┘ └────────────┘      │
│          │ [Ver todas las plantillas →]                      │
└──────────┴──────────────────────────────────────────────────┘
```

---

## i18n — Internacionalización

### Librería: react-i18next

**Justificación:**
- Estándar de facto para i18n en React
- Soporta namespaces (separar traducciones por módulo)
- Lazy loading de traducciones
- Interpolación, pluralización
- Detección automática de idioma del browser

### Estructura de archivos

```
apps/web/src/locales/
├── es/
│   ├── common.json      # Botones, errores, labels globales
│   ├── dashboard.json   # Dashboard
│   ├── users.json       # Módulo usuarios
│   ├── groups.json      # Módulo grupos
│   ├── dns.json         # Módulo DNS
│   ├── gpo.json         # Módulo GPO
│   └── settings.json    # Configuración
├── en/
│   ├── common.json
│   ├── dashboard.json
│   └── ...
└── pt/
    ├── common.json
    ├── dashboard.json
    └── ...
```

### Ejemplo de traducción

```json
// es/users.json
{
  "title": "Usuarios",
  "new": "Nuevo Usuario",
  "search": "Buscar usuarios...",
  "filter": {
    "status": "Estado",
    "all": "Todos",
    "active": "Activos",
    "disabled": "Deshabilitados"
  },
  "columns": {
    "username": "Usuario",
    "name": "Nombre",
    "ou": "Unidad Organizativa",
    "status": "Estado"
  },
  "actions": {
    "edit": "Editar",
    "delete": "Eliminar",
    "disable": "Deshabilitar",
    "enable": "Habilitar",
    "resetPassword": "Resetear contraseña"
  },
  "create": {
    "title": "Crear nuevo usuario",
    "username": "Nombre de usuario",
    "password": "Contraseña",
    "confirm": "Confirmar contraseña",
    "firstName": "Nombre",
    "lastName": "Apellido",
    "email": "Correo electrónico",
    "ou": "Unidad Organizativa",
    "submit": "Crear usuario",
    "success": "Usuario {{name}} creado correctamente"
  }
}
```

### Idiomas soportados desde día 1
| Código | Idioma | Default |
|---|---|---|
| es | Español | ✅ |
| en | English | |
| pt | Português | |

El idioma se detecta del browser, se guarda en localStorage, y se puede cambiar desde el topbar.

---

## Decisiones de diseño cerradas

| ID | Decisión |
|---|---|
| UI-001 | Dark mode first, light mode opcional |
| UI-002 | Accent: Indigo #6366f1 |
| UI-003 | Tipografía: Inter + JetBrains Mono |
| UI-004 | shadcn/ui como base de componentes |
| UI-005 | react-i18next para i18n (es/en/pt) |
| UI-006 | Layout: Sidebar 240px + Topbar 56px + Status bar 32px |
| UI-007 | Zustand para state management (ligero, simple) |
| UI-008 | TanStack Query para data fetching + caching |
| UI-009 | react-hook-form + zod para forms |
| UI-010 | React Router para routing |