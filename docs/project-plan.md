# SambaForge — Plan de desarrollo

## Resumen ejecutivo

Plataforma web open-source para instalar, administrar y gestionar un controlador de dominio Active Directory basado en Samba4 en Linux. Diseñada desde cero tomando lo mejor de las soluciones existentes (cockpit-samba-ad-dc, Samba Conductor, Vexa, go-samba4, samba4-manager, Zentyal, NethServer, UCS) pero superando sus limitaciones: despliegue 100% vía web, UI moderna y amigable, seguridad robusta, y arquitectura extensible.

---

## Stack tecnológico

| Capa | Tecnología | Justificación |
|---|---|---|
| **Backend** | **Go 1.23+** (Echo o Gin) | Binario único sin runtime externo — crítico en un DC. Arranque instantáneo, bajo consumo de RAM. Tipado fuerte, memory safety. Librerías maduras: `go-ldap/ldap`, `gokrb5` (Kerberos), `os/exec` para `samba-tool`. Menor superficie de ataque. Usado por los proyectos más serios (Vexa, go-samba4). |
| **Frontend** | **React 19 + Vite + TypeScript + Tailwind CSS** | Estándar de facto. Componentes con shadcn/ui. Soporte i18n nativo (es/en/pt desde día 1). |
| **DB interna** | **SQLite** (con WAL) | Para sesión, logs de auditoría, configuración de la plataforma. Sin servicio extra. |
| **Autenticación** | **LDAP bind + Kerberos + 2FA TOTP** | Autenticación real contra el dominio. Credenciales por sesión, nunca persistidas. |
| **Contenedor** | **Docker** (opcional) + binario nativo | Deployment dual: binario estático o Docker. El DC corre Samba nativo. |
| **Reverse proxy** | **Caddy** o **nginx** | TLS automático (Caddy) o configuración manual. Solo expone 443. |

---

## Fases de desarrollo

### Fase 0 — Investigación y análisis técnico (1-2 semanas)

**Objetivo:** Estudiar a fondo Samba, `samba-tool`, los proyectos existentes y las librerías Go antes de escribir una línea de código. Salir de esta fase con un documento de arquitectura validado, un mapa completo de subcomandos, y las decisiones técnicas cerradas.

> **Principio:** "Sin investigación no hay desarrollo." Toda la información recopilada aquí es la base sólida que permite avanzar rápido y sin errores en las fases siguientes.

#### 0.1 — Estudio de Samba y su documentación oficial

| # | Tarea | Entregable | Estado |
|---|---|---|---|
| 0.1.1 | Revisar y documentar versiones de Samba activas (4.24.x stable, 4.23.x maintenance) | Tabla de versiones + recomendación de soporte | ✅ Hecho — ver `docs/reference/samba-reference.md` |
| 0.1.2 | Estudiar changelog de 4.22, 4.23 y 4.24 — identificar breaking changes y nuevos features relevantes | Nota técnica de cambios que afectan a SambaForge | ✅ Hecho — ver `docs/reference/samba-reference.md` |
| 0.1.3 | Mapear **todos** los subcomandos de `samba-tool` y clasificarlos por módulo de SambaForge | Tabla subcomando → módulo | ✅ Hecho — ver `docs/reference/samba-reference.md` |
| 0.1.4 | Estudiar el flujo completo de provisioning (SambaWiki: Setting up Samba as AD DC) | Checklist de pre-requisitos + pasos + post-provision | ✅ Hecho — ver `docs/reference/samba-reference.md` |
| 0.1.5 | Identificar qué subcomandos de `samba-tool` soportan `--json` (output parseable) | Lista de subcomandos con soporte JSON | ⏳ Pendiente — ejecutar `samba-tool <cmd> --help` en servidor |
| 0.1.6 | Documentar CVEs recientes (2026) y versiones mínimas seguras | Tabla CVE + versión mínima recomendada | ✅ Hecho — ver `docs/reference/samba-reference.md` |
| 0.1.7 | Estudiar paths de Samba (self-compiled vs distro packages) y cómo detectarlos | Nota técnica de detección de paths | ✅ Hecho — ver `docs/reference/samba-reference.md` |
| 0.1.8 | Estudiar el manejo seguro de credenciales en `samba-tool` (no pasar passwords por CLI) | Nota técnica con patrones seguros | ✅ Hecho — ver `docs/reference/samba-reference.md` |

#### 0.2 — Análisis de proyectos existentes

| # | Tarea | Entregable | Estado |
|---|---|---|---|
| 0.2.1 | Clonar y revisar `cockpit-samba-ad-dc` — mapeo completo de subcomandos cubiertos | Análisis de cobertura | ✅ Hecho — ver `docs/reference/existing-projects-analysis.md` |
| 0.2.2 | Clonar y revisar `Samba Conductor` — arquitectura, zero-cred pattern, OAuth2, DR | Análisis de patrones reutilizables | ✅ Hecho |
| 0.2.3 | Clonar y revisar `Vexa` — stack Go+React, provisioning wizard, PAM auth, bootstrap script | Análisis de arquitectura Go+React | ✅ Hecho |
| 0.2.4 | Clonar y revisar `go-samba4` — go-ldap, gokrb5, 2FA TOTP, i18n, Cobra CLI | Análisis de librerías Go | ✅ Hecho |
| 0.2.5 | Clonar y revisar `samba4-manager` — LDAP bind auth pattern | Análisis de auth pattern | ✅ Hecho |
| 0.2.6 | Revisar `Zentyal` y `NethServer` — visión integral, provisioning modes, RBAC | Análisis comparativo | ✅ Hecho |
| 0.2.7 | Revisar `UCS` — S4 Connector pattern, App Center | Análisis de patrones de integración | ✅ Hecho |
| 0.2.8 | Generar matriz comparativa de features | Tabla feature × proyecto | ✅ Hecho — ver `docs/reference/existing-projects-analysis.md` |
| 0.2.9 | **Leer código fuente** de los 4 proyectos principales (cockpit, Conductor, Vexa, go-samba4) y documentar patrones concretos de implementación | Nota técnica por proyecto con snippets de código | ⏳ Pendiente — requiere clonar repos |

#### 0.3 — Estudio de librerías Go

| # | Tarea | Entregable | Estado |
|---|---|---|---|
| 0.3.1 | Evaluar `go-ldap/ldap` — API, ejemplos, madurez | Nota técnica | ⏳ Pendiente |
| 0.3.2 | Evaluar `gokrb5` — Kerberos client, SPNEGO, GSSAPI en Go | Nota técnica | ⏳ Pendiente |
| 0.3.3 | Evaluar `os/exec` — patrones para envolver `samba-tool` de forma segura | Nota técnica con ejemplos | ⏳ Pendiente |
| 0.3.4 | Evaluar Echo vs Gin — benchmark, features, comunidad | ADR-005: Framework HTTP | ⏳ Pendiente |
| 0.3.5 | Evaluar SQLite drivers Go (`mattn/go-sqlite3` vs `modernc.org/sqlite`) | ADR-006: SQLite driver | ⏳ Pendiente |
| 0.3.6 | Evaluar librerías TOTP Go (`pquerna/otp`) | Nota técnica | ⏳ Pendiente |
| 0.3.7 | Evaluar librerías JWT Go (`golang-jwt/jwt`) | Nota técnica | ⏳ Pendiente |

#### 0.4 — Estudio de frontend y UI/UX

| # | Tarea | Entregable | Estado |
|---|---|---|---|
| 0.4.1 | Revisar shadcn/ui — componentes disponibles, accesibilidad, tematización | Nota técnica | ⏳ Pendiente |
| 0.4.2 | Estudiar i18n en React (react-i18next vs FormatJS) | ADR-007: Librería i18n | ⏳ Pendiente |
| 0.4.3 | Diseñar wireframes de pantallas principales (login, dashboard, users, groups, DNS, GPO) | Mockups en Excalidraw o HTML | ⏳ Pendiente |
| 0.4.4 | Definir paleta de colores y tema (claro/oscuro) | Guía de estilo | ⏳ Pendiente |

#### 0.5 — Decisiones de arquitectura (ADRs)

| # | Tarea | Entregable | Estado |
|---|---|---|---|
| 0.5.1 | ADR-001: Stack Go+React | Documento de decisión | ✅ Cerrado |
| 0.5.2 | ADR-002: Arquitectura local-only (sin agente) | Documento de decisión | ✅ Cerrado |
| 0.5.3 | ADR-003: Auth model (LDAP bind + JWT + 2FA TOTP) | Documento de decisión | ✅ Cerrado |
| 0.5.4 | ADR-004: Seguridad (TLS, zero stored creds, audit log, RBAC) | Documento de decisión | ✅ Cerrado |
| 0.5.5 | ADR-005: Echo vs Gin | Documento de decisión | ⏳ Pendiente |
| 0.5.6 | ADR-006: SQLite driver | Documento de decisión | ⏳ Pendiente |
| 0.5.7 | ADR-007: Librería i18n frontend | Documento de decisión | ⏳ Pendiente |
| 0.5.8 | ADR-008: Reverse proxy (Caddy vs nginx) | Documento de decisión | ⏳ Pendiente |
| 0.5.9 | ADR-009: Cómo ejecutar samba-tool (root vs sudoers restringido) | Documento de decisión | ⏳ Pendiente |

**Gate de Fase 0:** Todos los ADRs cerrados. Documentos de referencia completos. Matriz de features validada. Wireframes de UI aprobados. Librerías Go evaluadas y seleccionadas.

---

### Fase 1 — Fundaciones del proyecto (1-2 semanas)

**Objetivo:** Crear el repositorio, scaffold del monorepo, CI/CD, y entorno de desarrollo funcional.

| # | Tarea | Entregable |
|---|---|---|
| 1.1 | Crear repo GitHub, LICENSE (MIT), README, CONTRIBUTING, SECURITY | Repo público |
| 1.2 | Scaffold monorepo: `apps/api` (Go), `apps/web` (React), `deploy/`, `docs/` | Estructura de directorios |
| 1.3 | Go module init, Echo/Gin setup, health endpoint, middleware base | API con `/api/health` |
| 1.4 | React+Vite+TS+Tailwind+shadcn/ui scaffold | Frontend con landing page |
| 1.5 | CI: golangci-lint, eslint, tests automáticos, build de binario | GitHub Actions |
| 1.6 | Dockerfile multi-stage (build Go + build React + runtime mínimo) | Imagen Docker funcional |
| 1.7 | docker-compose de desarrollo (API + Web + hot reload) | Entorno dev funcional |
| 1.8 | Makefile con targets: `make dev`, `make build`, `make test`, `make lint` | Build system |

**Gate de Fase 1:** `docker compose up` levanta backend+frontend. CI pasa. `make build` produce binario.

---

### Fase 2 — Spike técnico: Provisioning end-to-end (1-2 semanas)

**Objetivo:** Validar el caso más crítico — provisionar un dominio AD completo desde la API, sin UI.

> Este es el riesgo más alto del proyecto. Si no podemos provisionar un dominio desde la web, nada más importa.

| # | Tarea | Entregable |
|---|---|---|
| 2.1 | Implementar `preflight` — detectar Samba instalado, versión, paths, hostname, IP, DNS | `GET /api/server/preflight` |
| 2.2 | Implementar wrapper `samba-tool domain provision` (no-interactivo, password via env) | `POST /api/domain/provision` |
| 2.3 | Post-provision: configurar resolv.conf, krb5.conf, iniciar servicio samba | Lógica de post-provision |
| 2.4 | Verificación automática: `samba-tool domain level show`, `kinit`, DNS records SRV | `GET /api/domain/health` |
| 2.5 | Test E2E en VPS limpio (Ubuntu 24.04 / Debian 12) | Dominio funcionando, Windows puede join |
| 2.6 | Documentar limitaciones y workarounds encontrados | Nota técnica |

**Gate de Fase 2:** VPS limpio + `curl POST /api/domain/provision` → dominio AD funcional.

---

### Fase 3 — MVP: Gestión básica del dominio (3-4 semanas)

**Objetivo:** Plataforma usable con login, dashboard, usuarios, grupos y DNS básico.

| # | Módulo | Features |
|---|---|---|
| 3.1 | **Auth** | Login con credenciales del dominio (LDAP bind como `Administrator`). Sesión JWT. Logout. Credenciales por sesión, nunca en disco. |
| 3.2 | **Dashboard** | Estado del DC (servicio activo, versión Samba, hostname, dominio, DNS). Contadores: usuarios, grupos, computers, OUs. Health check. |
| 3.3 | **Usuarios** | Listar (paginado, buscar, `--json`). Crear. Editar. Eliminar. Activar/desactivar. Reset password. Forzar cambio. |
| 3.4 | **Grupos** | Listar. Crear. Eliminar. Añadir/quitar miembros. Tipos. Anidados. |
| 3.5 | **DNS básico** | Zonas. Registros A, CNAME, MX. Forwarders. Estado del DNS interno. |
| 3.6 | **i18n** | Español (default), Inglés, Portugués desde el inicio. |
| 3.7 | **UI/UX** | Layout responsive. Tema claro/oscuro. Sidebar. shadcn/ui. |

**Gate de Fase 3:** Administrador gestiona usuarios, grupos y DNS básico desde el navegador.

---

### Fase 4 — Gestión avanzada de objetos (2-3 semanas)

| # | Módulo | Features |
|---|---|---|
| 4.1 | **OUs** | Árbol jerárquico. Crear, mover, eliminar. Arrastrar objetos entre OUs. |
| 4.2 | **Computadoras** | Listar máquinas unidas. Eliminar. Mover a OU. Última conexión. |
| 4.3 | **Políticas de contraseña** | Ver/editar política del dominio. PSO (fine-grained). |
| 4.4 | **Cuentas de servicio** | gMSA. Crear, gestionar KDS root keys. |
| 4.5 | **Atributos extendidos** | Editor LDAP arbitrario (modo avanzado). Búsqueda por atributo. |
| 4.6 | **Importación masiva** | CSV de usuarios/grupos. Validación. Preview. Ejecución. |

**Gate de Fase 4:** Gestión completa equivalente a ADUC de RSAT.

---

### Fase 5 — DNS avanzado y GPOs (3-4 semanas)

| # | Módulo | Features |
|---|---|---|
| 5.1 | **DNS avanzado** | Zonas reversas. SRV, TXT, PTR. Zone transfers. Scavenging. DNS dinámico seguro. |
| 5.2 | **GPO — Gestión** | Listar. Crear vacío. Eliminar. Linkear/deslinkear. Backup/restore. |
| 5.3 | **GPO — Editor básico** | Plantillas preconfiguradas: password policy, drive maps, registry, wallpaper, logon script. |
| 5.4 | **GPO — Importar/exportar** | Importar/exportar GPO desde/hacia archivo. |

> **Nota:** La edición ADMX/ADML completa está fuera del alcance inicial. Se cubre con plantillas preconfiguradas (80% de casos reales).

**Gate de Fase 5:** DNS completo + GPOs con plantillas sin tocar consola.

---

### Fase 6 — Seguridad y auditoría (2-3 semanas)

| # | Módulo | Features |
|---|---|---|
| 6.1 | **2FA** | TOTP (Google Authenticator). Obligatorio para admin. Opcional para self-service. |
| 6.2 | **RBAC** | Roles: Domain Admin, Helpdesk, Read-only, Self-service. Permisos granulares. |
| 6.3 | **Audit log** | Registro persistente: quién, qué, cuándo, IP. Inmutable (append-only). Exportable. |
| 6.4 | **Session mgmt** | Timeout configurable. Sesiones activas. Forzar logout remoto. |
| 6.5 | **Hardening** | TLS obligatorio. Rate limiting. CSP, HSTS. CSRF tokens. |
| 6.6 | **Certificados** | Let's Encrypt (via Caddy) o custom. |

**Gate de Fase 6:** Cumple requisitos de auditoría. 2FA + RBAC + audit log.

---

### Fase 7 — Backup, restore y disaster recovery (2-3 semanas)

| # | Módulo | Features |
|---|---|---|
| 7.1 | **Backup AD** | `samba-tool domain backup` online y offline. |
| 7.2 | **Backup SYSVOL** | Políticas y scripts de logon. |
| 7.3 | **Restore** | Selectivo (tombstone) y completo. |
| 7.4 | **Programación** | Automático (cron). Retención. Notificaciones. |
| 7.5 | **Destinos** | Local, S3-compatible, SSH remoto. |
| 7.6 | **Health checks** | Verificación de integridad. Test de restore periódico. |

**Gate de Fase 7:** Backups automáticos + restore verificado.

---

### Fase 8 — Multi-DC (3-4 semanas)

| # | Módulo | Features |
|---|---|---|
| 8.1 | **Join DC** | Asistente para unir segundo DC (`samba-tool domain join`). |
| 8.2 | **Replicación** | Monitoreo. Estado de sincronización. Forzar replicación. |
| 8.3 | **FSMO roles** | Ver, transferir, seize. |
| 8.4 | **SYSVOL replication** | rsync basado en cron. Documentación. |
| 8.5 | **Demotion** | Eliminar DC de forma segura. |
| 8.6 | **Trusts** | Crear/eliminar trust. Validar. |

**Gate de Fase 8:** Segundo DC + replicación + FSMO desde la web.

---

### Fase 9 — Features diferenciadoras (2-3 semanas)

| # | Módulo | Features |
|---|---|---|
| 9.1 | **Self-service portal** | Cambio de password. Edición de atributos propios. |
| 9.2 | **API REST documentada** | OpenAPI/Swagger. API key para integraciones. |
| 9.3 | **OAuth2 server** | Authorization Code flow. SSO. |
| 9.4 | **Notificaciones** | Email (SMTP) + webhooks: usuario creado, backup fallido, DC caído. |
| 9.5 | **Instalador web** | Asistente: descargar binario, configurar, levantar. Una URL, un comando. |
| 9.6 | **Docker image oficial** | `docker pull sambaforge/sambaforge` + compose. |

**Gate de Fase 9:** Supera a las existentes en UX, features y deployment.

---

## No-goals explícitos (fuera del alcance inicial)

- **Editor ADMX/ADML completo** — Plantillas preconfiguradas cubren 80%.
- **Gestión de file shares** — SambaForge gestiona el DC, no file servers.
- **Servicios integrados (mail, VPN, firewall)** — Eso es Zentyal/UCS.
- **Migración desde Windows Server AD** — Se evalúa después.
- **NT4-style PDC** — Solo AD DC (Samba4+).
- **Arquitectura agente+servidor remoto** — Local-only en el DC.

---

## Estimación temporal

| Fase | Duración | Acumulado |
|---|---|---|
| Fase 0 — Investigación y análisis | 1-2 sem | 2 sem |
| Fase 1 — Fundaciones del proyecto | 1-2 sem | 4 sem |
| Fase 2 — Spike provisioning | 1-2 sem | 6 sem |
| Fase 3 — MVP gestión básica | 3-4 sem | 10 sem |
| Fase 4 — Gestión avanzada | 2-3 sem | 13 sem |
| Fase 5 — DNS avanzado + GPO | 3-4 sem | 17 sem |
| Fase 6 — Seguridad y auditoría | 2-3 sem | 20 sem |
| Fase 7 — Backup y DR | 2-3 sem | 23 sem |
| Fase 8 — Multi-DC | 3-4 sem | 27 sem |
| Fase 9 — Diferenciadores | 2-3 sem | 30 sem |

**MVP funcional (Fase 0-3): ~2.5 meses.** Plataforma completa: ~7-8 meses.

---

## Riesgos identificados

| Riesgo | Probabilidad | Impacto | Mitigación |
|---|---|---|---|
| `samba-tool` cambia flags entre versiones | Media | Alto | Detectar versión, adaptar wrapper. Tests E2E por versión. |
| GPOs limitadas en Samba vs Windows Server | Alta | Medio | Plantillas preconfiguradas. Documentar limitaciones honestamente. |
| Replicación SYSVOL sin DFS-R | Alta | Medio | rsync + cron. Documentar. |
| Seguridad: web UI con acceso al DC | Media | Crítico | Usuario dedicado, sudoers restringido, TLS, 2FA, audit log. |
| Subcomandos sin `--json` | Alta | Medio | Parseo de texto con fallback. Probar cada uno en Fase 0. |
| CVEs de Samba requieren actualizaciones urgentes | Media | Alto | Detector de versión + alerta. Documentar procedimiento de upgrade. |

---

## Documentos de referencia

| Documento | Ubicación | Estado |
|---|---|---|
| Referencia técnica de Samba | `docs/reference/samba-reference.md` | ✅ Completo |
| Análisis de proyectos existentes | `docs/reference/existing-projects-analysis.md` | ✅ Completo |
| ADRs | `docs/adr/` | ⏳ Pendiente |
| Wireframes UI | `docs/design/` | ⏳ Pendiente |

---

## Reglas para actualizar este plan

1. Al inicio de cada sesión, leer este documento primero.
2. Al final de cada sesión, actualizar el estado de las tareas (⏳ → ✅).
3. Cuando se cierra un ADR, marcarlo y crear el documento en `docs/adr/`.
4. Cuando se completa un gate de fase, revisar el plan con el usuario antes de continuar.
5. Toda decisión nueva se documenta como ADR antes de implementar.