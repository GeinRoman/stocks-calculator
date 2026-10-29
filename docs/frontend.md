# Frontend Agent Instructions

These rules apply to all frontend work. Follow the root AGENTS.md as well. Paths below are relative to the "frontend/" folder in root unless explicitly stated otherwise.

## Stack and scope
- Use SvelteKit, TypeScript, Tailwind CSS, and shadcn-svelte.
- Preserve the existing project configuration, conventions, and package manager; check package.json before running commands.
- Keep portfolio valuation, rebalancing, authorization, and persistence authoritative in the Go backend. Use typed API requests/responses; do not duplicate backend business rules in the browser.
- Implement only the requested screen or component. Avoid unrelated redesigns and additional UI libraries without approval.

## Design system and asset locations
Use the following SvelteKit, shadcn-svelte, and application-specific locations:

```text
src/lib/components/ui/  # Installed/local shadcn-svelte UI components
src/lib/components/     # Application-specific UI, grouped by purpose
src/lib/components/tabs/ # Profile, stocks, and general-information tab views
src/lib/logic/          # Application component types, data, and reusable logic
src/lib/logic/tabs/     # Types, data, and reusable logic used by tab views
src/lib/assets/         # Importable images, icons, fonts, and other assets
src/routes/layout.css   # Global stylesheet and shared design tokens
static/                 # Files served directly without imports
```

- Inspect src/routes/layout.css first for existing palettes and tokens. Reuse them; do not introduce arbitrary hex colors, spacing scales, radii, or typography in individual screens.
- Prefer semantic Tailwind utilities such as bg-primary, text-foreground, and border-border over literal colors like bg-blue-600 when implementing application UI.
- Add or modify design tokens centrally in src/routes/layout.css; ensure it is imported by src/routes/+layout.svelte and the tokens are exposed to Tailwind. Follow the existing Tailwind version and theme configuration. If the generated project uses a different global stylesheet name, preserve that path rather than creating a second stylesheet.
- Reuse fonts, imagery, and icons already in src/lib/assets/ or static/ when appropriate. Do not introduce duplicate assets.
- Use the project's configured #lib alias for importable components, logic, and assets. Use static/ only for files that need to be served directly; do not place Svelte components or theme sources there.

## File organization

Group application files by purpose, with matching directories for UI and its supporting TypeScript code:

```text
src/lib/
  components/
    tabs/
      general-information.svelte
      profile.svelte
      stocks.svelte
    ui/                       # Shared shadcn-svelte primitives
  logic/
    tabs/
      portfolio.ts            # Shared tab data interfaces
      demo-portfolio.ts       # Placeholder profiles and portfolio snapshots
  assets/
  utils.ts                    # Shared UI utilities required by shadcn-svelte
  index.ts                    # Library entrypoint
src/routes/
  +page.svelte                # Workspace composition and shared UI state
  +layout.svelte              # Application-wide layout setup
  layout.css                  # Theme and global styling
tests/                        # Frontend regression tests
```

- Place tab view components in src/lib/components/tabs/, not directly in components/ or routes/. Keep routes focused on routing, composing views, and coordinating their shared state.
- Place application-specific TypeScript files in src/lib/logic/ under the same purpose grouping as their components. For example tab interfaces, demo data, and extracted reusable behavior belong in logic/tabs/. Add other purpose directories only when needed.
- Keep component-specific markup, local state, derived display values, and event handlers together in their .svelte component. Do not extract trivial local behavior solely to create a separate logic file.
- Keep generic UI primitives and their own supporting files inside components/ui/. Do not move generated shadcn-svelte TypeScript files into logic/. Keep shared framework utilities such as utils.ts and the library entrypoint in their configured locations.
- Use kebab-case filenames and descriptive names. Shared tab data types belong in portfolio.ts; placeholder values belong in demo-portfolio.ts, separate from rendering and interaction code.
- Import application files through #lib, for example #lib/components/tabs/stocks.svelte and #lib/logic/tabs/portfolio.js. Preserve the project's .js extension convention for TypeScript module imports.
- When moving a file, update every import and test source path. Do not leave duplicate files or forwarding modules at old locations.
- This organization does not move financial business logic into the browser. Authoritative valuation and allocation calculations remain in the Go backend.

## Components and shadcn-svelte
- Before implementing any UI primitive, check src/lib/components/ui/ for an existing component (button, input, card, dialog, table, etc.). Reuse it and its supported variants/sizes.
- If a required shadcn-svelte component is missing, add it using the project's shadcn-svelte CLI rather than hand-writing a duplicate. For example, from the frontend root:
  npx shadcn-svelte@latest add button
- Keep generated UI components in the default src/lib/components/ui/ directory. The CLI may install dependencies; review resulting changes.
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
1. Inspect relevant frontend code, the global stylesheet, src/lib/components/ui/, and any references provided with the task.
2. Reuse an existing primitive or install a missing shadcn-svelte component to the configured location.
3. Implement the smallest requested change using shared components and semantic tokens.
4. Run the existing frontend check/test/build scripts as appropriate; inspect package.json rather than assuming command names.
5. Report changed files, newly installed components or dependencies, checks actually run, and any limitations.
Do not modify generated code blindly, invent non-existent scripts, or claim verification succeeded when it did not run.
