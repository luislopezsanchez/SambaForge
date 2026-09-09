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

**Decisiones cerradas:**
- D-001: Stack Go (backend) + React/TS (frontend) — Go por resiliencia, seguridad, performance
- D-002: Arquitectura local-only — SambaForge corre en el mismo servidor que el DC
- D-003: Single-DC primero, multi-DC en Fase 8
- D-004: Nombre: SambaForge
- D-005: Licencia: MIT
- D-006: `samba-tool` como motor, LDAP para lecturas/escrituras directas

**Decisiones pendientes:**
- P-001: Echo vs Gin (framework HTTP Go)
- P-002: SQLite driver Go (mattn vs modernc)
- P-003: shadcn/ui vs otra librería de componentes
- P-004: Caddy vs nginx
- P-005: i18n lib React (react-i18next vs FormatJS)
- P-006: Cómo ejecutar samba-tool (root vs sudoers restringido)

**Archivos creados:**
- `docs/project-plan.md` — Plan completo de 10 fases
- `docs/reference/samba-reference.md` — Referencia técnica de Samba (versiones, samba-tool, provisioning, CVEs)
- `docs/reference/existing-projects-analysis.md` — Análisis de 10 proyectos existentes con matriz comparativa
- `README.md` — Descripción del proyecto
- `LICENSE` — MIT
- `.gitignore`

**Próximos pasos (Sesión 2):**
- Tarea 0.1.5: Verificar qué subcomandos de `samba-tool` soportan `--json` (requiere acceso a servidor con Samba)
- Tarea 0.2.9: Clonar y leer código fuente de cockpit-samba-ad-dc, Samba Conductor, Vexa, go-samba4
- Tarea 0.3.x: Evaluar librerías Go (go-ldap, gokrb5, os/exec, Echo vs Gin, SQLite drivers, TOTP, JWT)
- Tarea 0.4.x: Wireframes UI y guía de estilo
- Cerrar ADRs pendientes (P-001 a P-006)

**Riesgos identificados:**
- `samba-tool` puede cambiar flags entre versiones → detectar versión y adaptar
- GPOs limitadas en Samba → plantillas preconfiguradas cubren 80%
- Replicación SYSVOL sin DFS-R → rsync + cron
- Seguridad web UI en DC → usuario dedicado, TLS, 2FA, audit log
- Algunos subcomandos pueden no tener `--json` → parseo de texto con fallback

**Contexto para retomar:**
- El proyecto está en Fase 0 (investigación). No hay código todavía.
- Los documentos de referencia ya tienen el estudio de Samba y de proyectos existentes.
- Falta: lectura de código fuente de los 4 repos principales, evaluación de librerías Go, wireframes UI, y cerrar los 6 ADRs pendientes.
- GitHub repo: https://github.com/luislopezsanchez/SambaForge (público, MIT)
- Token de GitHub configurado en credential helper para futuros pushes
- Usuario GitHub: luislopezsanchez

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