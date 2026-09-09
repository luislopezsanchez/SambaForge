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
4. **Tarea 0.2.9 — Análisis de código fuente** 🟡 En progreso (4 subagentes corriendo en paralelo)
   - Subagente 0: Análisis de Vexa (Go+React) — corriendo
   - Subagente 1: Análisis de go-samba4 (Go+Echo+LDAP+Kerberos) — corriendo
   - Subagente 2: Análisis de Samba Conductor (React+Docker) — corriendo
   - Subagente 3: Análisis de cockpit-samba-ad-dc (GSoC 2020) — corriendo

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

**Archivos creados:**
- `docs/reference/go-libraries-evaluation.md` — Evaluación completa de 7 librerías Go con snippets
- `docs/reference/ui-style-guide.md` — Guía de estilo, paleta, tipografía, 6 wireframes, i18n
- `docs/adr/adr-001-to-009.md` — 9 ADRs cerrados

**Archivos en progreso (subagentes):**
- `docs/reference/code-analysis-vexa.md` — Subagente escribiendo
- `docs/reference/code-analysis-go-samba4.md` — Subagente escribiendo
- `docs/reference/code-analysis-samba-conductor.md` — Subagente escribiendo
- `docs/reference/code-analysis-cockpit-samba.md` — Subagente escribiendo

**Próximos pasos (Sesión 3):**
- Revisar los 4 análisis de código fuente cuando los subagentes terminen
- Tarea 0.1.5: Verificar soporte --json de subcomandos samba-tool (requiere servidor con Samba)
- Actualizar plan con hallazgos de los análisis de código
- Commit y push de todos los documentos nuevos
- **Gate de Fase 0:** Validar que todos los entregables están completos antes de pasar a Fase 1

**Riesgos identificados nuevos:**
- Samba AD no soporta clear-text LDAP binds — SambaForge debe usar STARTTLS o GSSAPI siempre
- `samba-tool` scrub de password en process title tiene race condition — usar PASSWD env var es obligatorio
- Subcomandos sin --json requerirán parseo de texto con regex — mapear cuáles en Fase 1

**Contexto para retomar:**
- Fase 0 casi completa. Solo falta que terminen los 4 subagentes de análisis de código fuente.
- 9 ADRs cerrados, 7 librerías Go evaluadas, 6 wireframes creados, guía de estilo definida.
- Total dependencias Go: 6 externas + stdlib, sin CGO.
- Total dependencias frontend: 8 (React, Vite, TS, Tailwind, shadcn/ui, Zustand, TanStack Query, react-i18next).
- GitHub repo: https://github.com/luislopezsanchez/SambaForge
- Token de GitHub en credential helper (~/.git-credentials)
- Usuario GitHub: luislopezsanchez
- Una vez que los subagentes terminen, hacer commit + push de todos los docs nuevos y revisar el gate de Fase 0.

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