# Modal and Drawer Patterns

Read this file when building modals, side-over drawers, or any overlay that holds forms, configs, or logs.

## Side-over drawer (the "slide-over" contract)

- Use for complex forms, extensive configurations, or real-time logs.
- Component: base on `SlideOverDrawer.vue`.
- **Layout contract**: must use CSS variables to drive layout.
  - `--modal-shell-justify: flex-end` (right alignment)
  - `--modal-shell-padding: 0`
- **Visuals**: aside body must be **fully opaque** solid color (`var(--color-bg-surface)`).
  Overlay shell must use strong blur (`backdrop-filter: blur(12px)`).
- A specialized business drawer must not inherit the generic modal visual template wholesale;
  see `feedback/2026-05-08-specialized-drawer-should-not-inherit-modal-template.md`.
- A shared modal's visible border should be owned by the content/slot root, not re-declared per page;
  see `feedback/2026-05-11-shared-modal-visible-border-owned-by-slot-root.md`.
- Shared drawer spacing is configured through the component contract, not page-local margins;
  see `feedback/2026-05-06-shared-drawer-spacing-contract.md`.

## Modal safety (scrollability)

- Every modal or drawer must have a safety guard for overflow.
- **Constraints**: set `max-height: calc(100vh - 4rem)` on the container.
- **Scroll**: use `overflow-y: auto` and `flex-grow: 1` on the body to keep footer actions visible
  and accessible on small screens.
