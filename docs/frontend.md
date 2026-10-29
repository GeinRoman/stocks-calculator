# Frontend Agent Instructions

These rules apply to all frontend work. Follow the root AGENTS.md as well. Paths below are relative to the "frontend/" folder in root unless explicitly stated otherwise.

## Stack and scope
- Use SvelteKit, TypeScript, Tailwind CSS, and shadcn-svelte.
- Preserve the existing project configuration, conventions, and package manager; check package.json before running commands.
- Keep portfolio valuation, rebalancing, authorization, and persistence authoritative in the Go backend. Use typed API requests/responses; do not duplicate backend business rules in the browser.
- Implement only the requested screen or component. Avoid unrelated redesigns and additional UI libraries without approval.

## Design system and asset locations
Treat the assets/ directory as the frontend's design-system source of truth:

assets/components/    # Installed/local shadcn-svelte UI components
assets/theme/         # Color palette, semantic tokens, typography, radii, spacing, shadows
assets/...            # Images, icons, fonts, and other visual assets

- Inspect assets/theme/ first for existing palettes and tokens. Reuse them; do not introduce arbitrary hex colors, spacing scales, radii, or typography in individual screens.
- Prefer semantic Tailwind utilities such as bg-primary, text-foreground, and border-border over literal colors like bg-blue-600 when implementing application UI.
- Add or modify design tokens centrally in assets/theme/; ensure they are actually imported by the app's global stylesheet and exposed to Tailwind. Follow the existing Tailwind version and theme configuration.
- Reuse fonts, imagery, and icons already in assets/ when appropriate. Do not introduce duplicate assets.
- Do not assume assets/ is a public/static directory. Svelte components and theme sources must remain importable; configure and verify SvelteKit aliases and stylesheet imports as needed. Use the framework's static/public asset path only for files that need to be served directly.
Components and shadcn-svelte
- Before implementing any UI primitive, check assets/components/ for an existing component (button, input, card, dialog, table, etc.). Reuse it and its supported variants/sizes.
- If a required shadcn-svelte component is missing, add it using the project's shadcn-svelte CLI rather than hand-writing a duplicate. For example, from the frontend root:
  npx shadcn-svelte@latest add button
- Make sure the project's components.json and SvelteKit aliases direct generated UI components to assets/components/, not a second, default components directory. If this setup is absent, establish/verify it before adding components. The CLI may install dependencies; review resulting changes.
- Do not overwrite an existing or customized component without inspecting the diff and confirming it is needed.
- Edit shared component definitions when a new reusable variant or size is genuinely required; do not create one-off button/input styles across pages.
- Keep application-specific components (for example, portfolio views) separate from generic shadcn-svelte primitives, following the project's existing structure.
- Prefer accessible, semantic controls and preserve keyboard, focus, disabled, loading, and error states.

## Visual consistency
- Follow provided screenshots or wireframes for layout and visual structure, and written requirements for behavior and functionality.
- Use the existing theme for consistent typography, spacing, borders, corners, shadows, and colors; don't invent a new style for each page.
- Use existing component variants to express intent (primary, secondary, outline, ghost, destructive, etc.).
- Design for desktop and responsive layouts. Cover meaningful loading, empty, failure, and expired-session states.
- Display currencies, units, and price freshness clearly. Do not present stale or unavailable prices as current.

## Workflow and verification
1. Inspect relevant frontend code, assets/theme/, assets/components/, and any references provided with the task.
2. Reuse an existing primitive or install a missing shadcn-svelte component to the configured location.
3. Implement the smallest requested change using shared components and semantic tokens.
4. Run the existing frontend check/test/build scripts as appropriate; inspect package.json rather than assuming command names.
5. Report changed files, newly installed components or dependencies, checks actually run, and any limitations.
Do not modify generated code blindly, invent non-existent scripts, or claim verification succeeded when it did not run.
