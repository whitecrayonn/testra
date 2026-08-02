import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: "class",
  content: [
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./features/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          50: "#eef2ff",
          100: "#e0e7ff",
          200: "#c7d2fe",
          300: "#a5b4fc",
          400: "#818cf8",
          500: "#6366f1",
          600: "#4f46e5",
          700: "#4338ca",
          800: "#3730a3",
          900: "#312e81",
        },
        bg: "var(--bg)",
        panel: "var(--panel)",
        "panel-hi": "var(--panel-hi)",
        glass: "var(--glass)",
        hair: "var(--hair)",
        "hair-hi": "var(--hair-hi)",
        fg: "var(--fg)",
        fg2: "var(--fg2)",
        fg3: "var(--fg3)",
        acc: "var(--acc)",
        acc2: "var(--acc2)",
        "acc-soft": "var(--acc-soft)",
        pass: "var(--pass)",
        "pass-soft": "var(--pass-soft)",
        fail: "var(--fail)",
        "fail-soft": "var(--fail-soft)",
        warn: "var(--warn)",
        "warn-soft": "var(--warn-soft)",
        info: "var(--info)",
      },
      fontFamily: {
        sans: ["var(--font-instrument-sans)", "ui-sans-serif", "system-ui", "sans-serif"],
        mono: ["var(--font-jetbrains-mono)", "ui-monospace", "monospace"],
      },
      boxShadow: {
        glass: "var(--shadow)",
      },
      keyframes: {
        rise: { from: { opacity: "0", transform: "translateY(16px)" }, to: { opacity: "1", transform: "none" } },
        riseSm: { from: { opacity: "0", transform: "translateY(8px)" }, to: { opacity: "1", transform: "none" } },
        fadeIn: { from: { opacity: "0" }, to: { opacity: "1" } },
        pop: {
          "0%": { opacity: "0", transform: "scale(.94) translateY(10px)" },
          "60%": { opacity: "1", transform: "scale(1.012) translateY(0)" },
          "100%": { transform: "scale(1)" },
        },
        floatY: { "0%, 100%": { transform: "translateY(0)" }, "50%": { transform: "translateY(-4px)" } },
        dotPulse: { "0%, 100%": { opacity: "1", transform: "scale(1)" }, "50%": { opacity: ".35", transform: "scale(.72)" } },
        sheen: { from: { transform: "translateX(-120%)" }, to: { transform: "translateX(320%)" } },
        bootBar: { "0%, 100%": { transform: "scaleY(.28)" }, "50%": { transform: "scaleY(1)" } },
      },
      animation: {
        rise: "rise .55s cubic-bezier(.2,.8,.2,1) both",
        "rise-sm": "riseSm .5s ease both",
        "fade-in": "fadeIn .2s ease both",
        pop: "pop .55s cubic-bezier(.2,.8,.2,1) both",
        "float-y": "floatY 3.4s ease-in-out infinite",
        "dot-pulse": "dotPulse 1.6s ease-in-out infinite",
        sheen: "sheen 2.2s ease-in-out infinite",
        "boot-bar": "bootBar 1.15s ease-in-out infinite",
      },
    },
  },
  plugins: [],
};

export default config;
