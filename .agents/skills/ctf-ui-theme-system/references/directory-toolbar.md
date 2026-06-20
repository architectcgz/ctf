# Directory Toolbar Pattern

Read this file when building or refactoring a reusable list/directory toolbar (search, filter, sort, count controls), especially `WorkspaceDirectoryToolbar`.

## Variable-bridged primitive

- Treat reusable list toolbars as a variable-bridged workspace primitive, not as one-off styling per page.
- Toolbar styling should expose component-local semantic variables (for example `--workspace-toolbar-*`) that map back to global tokens for:
  - surface/background
  - border strength
  - text and muted text
  - shadow/elevation
  - font sizes
  - spacing
  - control height and radius

## One control contract

- Search, filter, sort, and count controls should share one control contract:
  - same control height and radius
  - same border/background/shadow language
  - hover strengthens border before introducing stronger color shifts
  - focus uses a primary border plus a low-opacity `color-mix(...)` ring, not a custom theme branch
- Menus and filter panels should share the same menu surface/border/elevation variables instead of duplicating separate popup palettes.

## Theme adaptation

- Do not append component-tail `:root[data-theme='light']` / `:root[data-theme='dark']` override blocks just to repair the toolbar.
  Drive theme adaptation through variables and semantic tokens from the start.

## Mobile downgrade

- Mobile downgrade should stay structural, not cosmetic:
  - wrap toolbar groups
  - let search expand to full width
  - preserve the same control styling contract after wrapping

## Related

- Filter-bar behavior (auto-apply, debounce, collapse) lives in `layout-rules.md` § Filter bars.
- Admin floating-island toolbar language lives in `admin-design-system.md` and `saas-workbench-pattern.md`.
