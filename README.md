# SambaForge

Plataforma web open-source para instalación, administración y gestión de controladores de dominio Active Directory basados en Samba4 en Linux.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 🎯 Objetivo

Llevar a una interfaz web moderna todo lo que hoy se hace manualmente por consola con `samba-tool` al instalar y administrar un controlador de dominio Samba4 AD DC en Linux — provisioning, usuarios, grupos, DNS, GPOs, backup, multi-DC y más.

## 🛠️ Stack

- **Backend:** Go (Echo/Gin) + go-ldap + gokrb5
- **Frontend:** React 19 + Vite + TypeScript + Tailwind CSS + shadcn/ui
- **DB interna:** SQLite (WAL)
- **Deployment:** Binario nativo + Docker opcional

## 📋 Estado

🟡 **Fase 0 — Investigación y análisis técnico**

## 📖 Documentación

- [Plan de desarrollo completo](docs/project-plan.md)
- [Referencia técnica de Samba](docs/reference/samba-reference.md)
- [Análisis de proyectos existentes](docs/reference/existing-projects-analysis.md)

## 🌐 Características planificadas

- ✅ Provisioning wizard 100% vía web
- ✅ Gestión de usuarios, grupos, OUs, computadoras
- ✅ DNS management completo
- ✅ GPOs con plantillas preconfiguradas
- ✅ Políticas de contraseña (dominio + PSO)
- ✅ Multi-DC (join, replicación, FSMO, trusts)
- ✅ Backup y disaster recovery
- ✅ RBAC + 2FA TOTP + audit log
- ✅ Self-service portal
- ✅ API REST + OAuth2 server
- ✅ i18n (es/en/pt)

## 📜 Licencia

MIT — ver [LICENSE](LICENSE)