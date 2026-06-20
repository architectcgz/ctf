# Color And Surface Tokens

Read this file when the task touches colors, surfaces, theme variables, category/difficulty pills, or warning/danger styling.

## Token-first rules (hard constraints)

- Use semantic tokens/variables; avoid ad-hoc hardcoded palette drift.
- Accent is theme-driven (support user theme preferences).
- **Prohibit hardcoded colors**: NEVER use hex/rgb values like `#ffffff` or `#0f172a` directly in `<style>` blocks. Always use CSS variables (e.g., `var(--color-bg-base)`).
- **Prohibit hardcoded font-sizes**: DO NOT use pixel values for font sizes. Use standard typography variables (e.g., `var(--font-size-14)`).
- **Prohibit hardcoded spacing**: DO NOT use pixel values for margins, paddings, or gaps. Use standard spacing variables (e.g., `var(--space-4)`).
- **Prohibit !important**: DO NOT use `!important` in component styles. Resolve specificity via higher-specificity selectors or proper inheritance. Exception: accessibility (reduced-motion) and absolute layering (z-index).
- **Theme adaptivity**: use `color-mix(in srgb, var(--variable) X%, transparent)` for secondary backgrounds or borders so they work on both light and dark themes.
- Keep backgrounds solid or subtly layered; avoid foggy overlays.
- No glassmorphism/backdrop blur in primary workspaces.

## Challenge category pills

- Must use explicit category-pill semantics. Use variables named like
  `--challenge-category-pill-web` / `--challenge-category-pill-crypto` for
  `web/pwn/reverse/crypto/misc/forensics` pills.
- Do not name these variables `level` (level means difficulty), and avoid generic `tone`.
- Keep category pill colors separate from broader category/status palettes such as `--color-cat-*`;
  shared challenge-category pill UI should consume `--challenge-category-pill-*` directly or via the challenge entity presentation helpers.

## Challenge difficulty pills

- Must use explicit difficulty-pill semantics. Use variables named like
  `--challenge-difficulty-pill-easy` / `--challenge-difficulty-pill-hard` for
  `beginner/easy/medium/hard/insane`.
- Do not use ambiguous names such as `level`, `diff`, or `tone` for the exported pill token.

## Warning / danger styling

- Derive soft risk backgrounds, borders, and hover states from `--color-warning`, `--color-danger`, or approved soft tokens.
- Prefer `color-mix(...)` or a local semantic bridge variable over bespoke amber/red overrides.
- Do not maintain separate light-mode and dark-mode warning palettes inside page-local CSS.
