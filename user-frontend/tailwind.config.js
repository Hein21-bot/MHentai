export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{vue,js,ts}'],
  theme: {
    extend: {
      colors: {
        primary: { DEFAULT: '#e53e3e', 600: '#c53030', 700: '#9b2c2c' },
        dark: { bg: '#0f0f0f', surface: '#1a1a1a', card: '#242424', border: '#2a2a2a', hover: '#2f2f2f' }
      },
      fontSize: {
        '2xs': ['0.65rem', { lineHeight: '1rem' }]
      },
      screens: {
        xs: '420px'
      }
    }
  },
  plugins: []
}
