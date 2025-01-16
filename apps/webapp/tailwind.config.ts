import baseConfig from "@glassact/ui/tailwind.config.ts";
import type { Config } from "tailwindcss";

export default {
  content: [...baseConfig.content, "../../libs/ui/src/*.{ts,tsx}"],
  presets: [baseConfig],
} satisfies Config;
