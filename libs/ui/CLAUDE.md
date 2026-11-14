# CLAUDE.md — UI Library

SolidJS component library published as `@glassact/ui`. Same conventions as the webapp — see `apps/webapp/CLAUDE.md` for SolidJS patterns (prop reactivity, signals, `createMemo`, `splitProps`), and the root `CLAUDE.md` for repo context.

Library-specific notes:
- This package is consumed by both `apps/webapp` and `apps/landing` (Astro). Keep components framework-agnostic within SolidJS — no webapp-specific assumptions (router, query client) unless the dependency is in `package.json`.
- Styling lives here: `globals.css` (design tokens, fonts, base styles, motif classes) and `tailwind.config.ts` are exported for consumers. How to use them is in `docs/styleguide.md`.
- Built via `pnpm --filter @glassact/ui build` (plain `tsc`). Run `pnpm libs:build` from the root before apps consume the `dist/` output.
