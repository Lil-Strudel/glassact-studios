import js from "@eslint/js";
import solid from "eslint-plugin-solid";
import globals from "globals";
import tseslint, { ConfigWithExtends } from "typescript-eslint";
import { defineConfig } from "eslint/config";

export default defineConfig([
  {
    files: ["**/*.{js,mjs,cjs,ts,mts,cts}"],
    plugins: { js },
    extends: ["js/recommended"],
    languageOptions: { globals: globals.browser },
  },
  tseslint.configs.recommended,
  {
    files: ["**/*.{ts,tsx}"],
    ...solid.configs["flat/typescript"],
    languageOptions: {
      parser: tseslint.parser,
      parserOptions: {
        project: "tsconfig.json",
      },
    },
  } as unknown as ConfigWithExtends[],
]);
