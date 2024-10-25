/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./internal/public/**/*.{html,js}"],
  theme: {
    extend: {
      gridTemplateColumns: {
        '16': 'repeat(16, minmax(0, 1fr))',
      },
    },
  },
  plugins: [require("@tailwindcss/typography"),require('daisyui')],
  daisyui: {
    themes: ["light", "dark"],
  },
};