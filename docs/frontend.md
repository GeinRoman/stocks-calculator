# Frontend Agent Instructions

Applies to frontend work in `frontend/` (paths below are relative to that directory). Follow the repository's root `AGENTS.md` too. Inspect the actual code and configuration before changing anything; this document describes current conventions, not a substitute for reading the implementation.

## Stack and boundaries

- Use SvelteKit, TypeScript, Tailwind CSS, and shadcn-svelte. Preserve the configured package manager, paths, aliases, and versions; inspect `package.json` and `components.json` before running tools.
- Implement only the requested functionality. Avoid unrelated refactors, redesigns, or additional UI libraries without approval.
- Go remains authoritative for portfolio valuation, rebalancing, authorization, and persistence. The frontend uses typed API contracts and displays backend results; temporary mock calculations are not production business logic.
- Treat mock fixtures and services as temporary. Do not present simulated values as live data or claim changes persist when they are in-memory only.

## Design system and components

- Shared design tokens (colors, spacing, typography, radii, borders, and shadows) live in `src/routes/layout.css`, imported by `src/routes/+layout.svelte`. Reuse existing tokens and semantic Tailwind utilities such as `bg-primary`, `text-foreground`, and `border-border`; avoid one-off hardcoded styles. Do not create a second global theme file.
- Look in `src/lib/components/ui/` before adding UI primitives. Reuse existing shadcn-svelte components and their variants/sizes; extend a shared component when a reusable variant is needed.
- If a required shadcn-svelte primitive is missing, install it from the frontend root using the project's CLI configuration, for example `npx shadcn-svelte@latest add button`. Check the resulting files/dependencies and never overwrite customizations blindly. The configured destination is `src/lib/components/ui/`.
- Put application-specific views in `src/lib/components/`; use `src/lib/assets/` for imported images/fonts/icons and `static/` for directly served assets. Do not put component or theme source files in `static/`.
- Follow supplied images/wireframes for appearance and written requirements for behavior. Maintain responsive layout, accessible controls, keyboard/focus states, and meaningful loading, empty, error, and expired-session states. Show currency, units, and data freshness clearly.

## Code organization

- `src/lib/components/tabs/`: portfolio, profile, stock, bond, editor, and dialog views. Keep route files primarily for navigation, composition, and shared state.
- `src/lib/logic/tabs/`: shared tab types, mock data, service boundaries, and reusable validation or draft logic. Keep local state, markup, and trivial event handlers in the relevant `.svelte` component.
- `src/lib/components/ui/`: shadcn-svelte primitives and their supporting source; keep shared `src/lib/utils.ts` in place.
- Use descriptive kebab-case filenames and the configured `#lib` import alias. Preserve the project's `.js` suffix convention for TypeScript module imports. When moving files, update imports and tests; don't leave unnecessary duplicate or forwarding files.
- Before editing a feature, locate its current coordinator, UI components, and service/logic files rather than relying on a fixed file inventory in this document.

## Portfolio editing contracts

### Stocks and profile state

- Keep read-only snapshots distinct from editing drafts. Confirming a composition applies the draft; cancelling discards it. Confirming a group deletion only changes the draft until the overall composition is saved.
- Weight editing has separate save/cancel behavior: `null` means uninitialized; `0` is explicitly assigned. Intermediate drafts may total less than 100, but saving requires all weights initialized and a total of exactly 100. The slider uses a fixed 0–100 scale; limit changes to the remaining budget without automatically moving other weights.
- Stock quantities are lots. A stock code may belong to only one group in a profile; adding it again to that group merges lot counts. Keep draft validation in `src/lib/logic/tabs/stock-editor.ts`.
- Editing components use the asynchronous `searchStocks`, `saveComposition`, and `saveWeights` boundary in `src/lib/logic/tabs/portfolio-service.ts`, not fixture data directly. These currently simulate backend behavior.
- Preserve separate saved snapshots per profile. Disable profile switching during composition or weight edits. Current mock edits are in-memory and reset on reload.
- Real API integration must account for uninitialized weights and atomic composition saves; never convert `null` to zero silently or present independent requests as atomic.

### Bonds and shared editors

- Stocks and bonds have separate holdings, groups, and weights per profile. Matching group names across asset classes are permitted; allocation percentages are relative to each asset class, not the combined portfolio.
- Reuse the common group editor, instrument search dialog, and group weights dialog where applicable. Preserve draft confirmation/cancellation, asynchronous search state (including stale responses), and integer-quantity validation.
- A bond quantity counts individual bonds (minimum 1), not stock lots. Zero price/coupon values are valid; `null` metadata means unavailable.
- Show current/redemption price and remaining/lifetime coupon counts as paired values. Bond market value is current price × quantity; do not include redemption principal or future coupons in this value.
- Mock coupon schedules cover the instrument's lifetime. Only payments strictly after a snapshot's `asOf` count as remaining. Sum actual future amounts (including variable coupons); unknown amounts mean unavailable income, while no future payments mean zero. Multiply per-bond remaining income by quantity. Show snapshot freshness in UTC.
- Keep valuation and coupon projection outputs behind the mock service/snapshot boundary; views display them rather than recalculating them. Refresh mock snapshots on load and after confirmed saves.
- Bond services currently simulate data without persistence. Do not route bond requests to stock endpoints: the backend does not yet support the corresponding bond features.

## Working procedure

1. Inspect relevant components, logic, tests, theme tokens, and existing UI primitives.
2. Reuse the design system and existing API/service boundary; install missing shadcn-svelte primitives only when needed.
3. Make a focused change and add or update regression tests for altered behavior.
4. Run relevant checks/tests/build commands from the existing `package.json` scripts when feasible. Avoid assuming scripts exist.
5. Report changed files, added components/dependencies, checks actually run, and remaining limitations. Never claim verification that did not occur.
