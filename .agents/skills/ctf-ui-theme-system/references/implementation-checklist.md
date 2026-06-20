# Implementation Checklist

Read this file before closing any CTF UI page build or refactor. Each item is a ✓Check.

## Shell and structure
- Page has one dominant workspace shell.
- Internal areas are flattened (divider/list/rail first), not stacked cards.
- Cards are used only for metrics or high-priority callouts (see `card-usage-rules.md`).

## Typography
- Main title uses `h1` and approved title metrics.
- Page-title and intro-copy typography come from shared workspace styles; local CSS keeps only necessary layout overrides.

## Lists and tables
- Header and data share track + alignment; key columns avoid loose `auto`.
- Each visible column has a clear responsibility; unrelated fields are not stuffed into the title column.
- Internal ids are hidden unless they provide real user value.
- Long row text hardened by default (single-line ellipsis for key titles, 2-line clamp for secondary copy, full text still recoverable).
- Admin/teacher list spacing follows shared rhythm (`0.75rem` head→body, `0.5rem` body→pagination).
- Pagination is visually attached to the list by default; no extra top divider unless justified.
- Row actions are compact enough to not break column layout at desktop widths.
- Action column uses clear primary/secondary hierarchy and accessible focus/group semantics.

## Behavior
- Refresh behavior is deliberate (manual by default, polling only when justified and scoped).
- Loading/empty/error/no-result states are intentionally differentiated.
- Search/filter/pagination behavior is stable and does not unexpectedly reset user context.
- Stale async responses are ignored instead of overwriting newer list state.
- Responsive downgrade order is controlled (compress first, stack later).

## Copy, color, semantics
- Copy has no design/meta narration (no mock/proposal/process text in visible UI).
- Machine values have an intentional read/copy/mobile-access strategy.
- Time, numbers, and status colors/labels are semantically consistent with sibling pages.
- Theme tokens are used; no accidental hardcoded drift.
- Keyboard semantics are complete for tabs/collapses/forms.
- If shared selectors changed, add or tighten a focused regression test asserting the exact selector boundary.
