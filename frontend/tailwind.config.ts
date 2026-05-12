import type { Config } from 'tailwindcss';

// Design tokens lifted from project/styles.css (EcoCycle design bundle).
// All values match :root variables there so the EcoCycle look transfers
// 1:1 to React components.
const config: Config = {
  content: ['./src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        // Brand
        accent: {
          DEFAULT: '#2f7d52',
          deep: '#1f5d3a',
          soft: '#eaf3ec',
        },
        // Ink (text)
        ink: {
          DEFAULT: '#131815',
          2: '#2f3833',
          3: '#6e776f',
          4: '#a3aaa3',
        },
        // Surfaces
        paper: {
          DEFAULT: '#f6f6f3',
          2: '#efefea',
        },
        surface: '#ffffff',
        // Lines / dividers
        line: {
          DEFAULT: 'rgba(20,30,24,0.08)',
          2: 'rgba(20,30,24,0.05)',
        },
      },
      fontFamily: {
        sans: ['"Plus Jakarta Sans"', 'ui-sans-serif', 'system-ui', '-apple-system', 'BlinkMacSystemFont', 'sans-serif'],
      },
      borderRadius: {
        sm: '12px',
        DEFAULT: '18px',
        lg: '22px',
      },
      boxShadow: {
        glass: '0 6px 24px rgba(20,30,24,0.05), 0 1px 0 rgba(255,255,255,0.7) inset',
        card: '0 12px 32px rgba(20,30,24,0.08)',
      },
      backdropBlur: {
        glass: '18px',
      },
      letterSpacing: {
        tight: '-0.025em',
      },
      keyframes: {
        pageIn: {
          '0%': { opacity: '0', transform: 'translateY(8px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
      },
      animation: {
        pageIn: 'pageIn 0.35s cubic-bezier(0.2, 0.8, 0.2, 1)',
      },
    },
  },
  plugins: [],
};

export default config;
