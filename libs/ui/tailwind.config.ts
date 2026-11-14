import type { Config } from "tailwindcss";
import tailwindAnimate from "tailwindcss-animate";

const token = (name: string) => `hsl(var(--${name}) / <alpha-value>)`;

/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ["class", '[data-kb-theme="dark"]'],
  content: ["./src/**/*.{astro,html,js,jsx,md,mdx,svelte,ts,tsx,vue}"],
  theme: {
    extend: {
      container: {
        center: true,
        padding: "2rem",
        screens: {
          "2xl": "1400px",
        },
      },
      colors: {
        came: token("came"),
        border: token("border"),
        input: token("input"),
        ring: token("ring"),
        background: token("background"),
        foreground: token("foreground"),
        faint: token("faint"),
        primary: {
          DEFAULT: token("primary"),
          foreground: token("primary-foreground"),
        },
        crimson: {
          deep: token("crimson-deep"),
          DEFAULT: token("primary"),
          light: token("crimson-light"),
        },
        gilt: {
          DEFAULT: token("gilt"),
          foreground: token("gilt-foreground"),
        },
        sage: token("sage"),
        plaque: token("plaque"),
        secondary: {
          DEFAULT: token("secondary"),
          foreground: token("secondary-foreground"),
        },
        destructive: {
          DEFAULT: token("destructive"),
          foreground: token("destructive-foreground"),
        },
        info: {
          DEFAULT: token("info"),
          foreground: token("info-foreground"),
        },
        success: {
          DEFAULT: token("success"),
          foreground: token("success-foreground"),
        },
        warning: {
          DEFAULT: token("warning"),
          foreground: token("warning-foreground"),
        },
        error: {
          DEFAULT: token("error"),
          foreground: token("error-foreground"),
        },
        status: {
          draft: token("status-draft"),
          ordered: token("status-ordered"),
          production: token("status-production"),
          shipped: token("status-shipped"),
          invoiced: token("status-invoiced"),
          completed: token("status-completed"),
          cancelled: token("status-cancelled"),
        },
        muted: {
          DEFAULT: token("muted"),
          foreground: token("muted-foreground"),
        },
        accent: {
          DEFAULT: token("accent"),
          foreground: token("accent-foreground"),
        },
        popover: {
          DEFAULT: token("popover"),
          foreground: token("popover-foreground"),
        },
        card: {
          DEFAULT: token("card"),
          foreground: token("card-foreground"),
        },
      },
      borderRadius: {
        xl: "calc(var(--radius) + 4px)",
        lg: `var(--radius)`,
        md: `calc(var(--radius) - 2px)`,
        sm: "calc(var(--radius) - 4px)",
        arch: "50% 50% var(--radius) var(--radius) / 30% 30% var(--radius) var(--radius)",
      },
      fontFamily: {
        display: ["Marcellus", "Optima", "Georgia", "serif"],
        body: ["Spectral", "Georgia", "serif"],
      },
      letterSpacing: {
        display: "0.01em",
        label: "0.16em",
      },
      keyframes: {
        "accordion-down": {
          from: { height: "0" },
          to: { height: "var(--kb-accordion-content-height)" },
        },
        "accordion-up": {
          from: { height: "var(--kb-accordion-content-height)" },
          to: { height: "0" },
        },
        "collapsible-down": {
          from: { height: "0" },
          to: { height: "var(--kb-collapsible-content-height)" },
        },
        "collapsible-up": {
          from: { height: "var(--kb-collapsible-content-height)" },
          to: { height: "0" },
        },
        "content-show": {
          from: { opacity: "0", transform: "scale(0.96)" },
          to: { opacity: "1", transform: "scale(1)" },
        },
        "content-hide": {
          from: { opacity: "1", transform: "scale(1)" },
          to: { opacity: "0", transform: "scale(0.96)" },
        },
      },
      animation: {
        "accordion-down": "accordion-down 0.2s ease-out",
        "accordion-up": "accordion-up 0.2s ease-out",
        "collapsible-down": "collapsible-down 0.2s ease-out",
        "collapsible-up": "collapsible-up 0.2s ease-out",
        "content-show": "content-show 0.2s ease-out",
        "content-hide": "content-hide 0.2s ease-out",
      },
    },
  },
  plugins: [tailwindAnimate],
} satisfies Config;
