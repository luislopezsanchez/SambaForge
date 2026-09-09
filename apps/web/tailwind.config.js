/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        bg: {
          base: '#0a0a0b',
          card: '#141416',
          elevated: '#1a1a1d',
          input: '#1e1e21',
        },
        border: {
          DEFAULT: '#27272a',
          strong: '#3f3f46',
        },
        accent: {
          DEFAULT: '#6366f1',
          hover: '#5558e0',
        },
      },
    },
  },
  plugins: [],
}