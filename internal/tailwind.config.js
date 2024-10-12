/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./internal/public/**/*.{html,js}"],
  theme: {
    extend: {
      fontFamily: {},
    },
  },
  plugins: [require("@tailwindcss/typography"),require('daisyui')],
  daisyui: {
    themes: ["light", "dark"],
  },
};