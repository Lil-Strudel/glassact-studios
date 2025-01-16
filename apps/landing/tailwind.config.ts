import type { Config } from "tailwindcss";
import baseConfig from "@glassact/ui/tailwind.config.ts";

export default {
  content: [...baseConfig.content, "../../libs/ui/src/*.{ts,tsx}"],
  presets: [baseConfig],
} satisfies Config;
