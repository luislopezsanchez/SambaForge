import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'

i18n.use(initReactI18next).init({
  resources: {
    es: {
      translation: {
        login: {
          title: 'SambaForge',
          username: 'Usuario',
          password: 'Contraseña',
          submit: 'Iniciar sesión',
          remember: 'Recordar sesión',
          noDomain: 'Si el dominio no está provisionado, ir al asistente de instalación',
        },
        dashboard: {
          title: 'Panel de control',
          dcStatus: 'Estado del DC',
          users: 'Usuarios',
          groups: 'Grupos',
          computers: 'Equipos',
          gpos: 'GPOs',
        },
      },
    },
    en: {
      translation: {
        login: {
          title: 'SambaForge',
          username: 'Username',
          password: 'Password',
          submit: 'Sign in',
          remember: 'Remember session',
          noDomain: 'If the domain is not provisioned, go to the installation wizard',
        },
        dashboard: {
          title: 'Dashboard',
          dcStatus: 'DC Status',
          users: 'Users',
          groups: 'Groups',
          computers: 'Computers',
          gpos: 'GPOs',
        },
      },
    },
    pt: {
      translation: {
        login: {
          title: 'SambaForge',
          username: 'Usuário',
          password: 'Senha',
          submit: 'Entrar',
          remember: 'Lembrar sessão',
          noDomain: 'Se o domínio não está provisionado, ir para o assistente de instalação',
        },
        dashboard: {
          title: 'Painel de controle',
          dcStatus: 'Status do DC',
          users: 'Usuários',
          groups: 'Grupos',
          computers: 'Computadores',
          gpos: 'GPOs',
        },
      },
    },
  },
  lng: 'es',
  fallbackLng: 'en',
  interpolation: { escapeValue: false },
})

export default i18n