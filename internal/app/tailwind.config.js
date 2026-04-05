/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './app/**/*.{js,ts,jsx,tsx}',
    './components/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        zinc: {
          950: '#030712',
          900: '#09090b',
          800: '#18181b',
          700: '#27272a',
          600: '#3f3f46',
        },
      },
    },
  },
  plugins: [],
}
