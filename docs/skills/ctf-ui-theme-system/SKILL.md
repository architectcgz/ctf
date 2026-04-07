---
name: ctf-ui-theme-system
description: Use when building or refactoring CTF frontend pages in this repo to keep layout, typography, color, copy tone, and interaction patterns aligned with the approved academy/challenges/admin workspace style.
---

# CTF UI Theme System

## Overview
This skill defines the approved UI language for this CTF project.
It captures decisions validated through repeated page refactors across `academy`, `challenges`, and `platform`.

Core outcome: **professional, restrained, technical, high-readability workspace UI**.

## Product Priorities
- Student workflow first: discover challenge -> run environment -> submit flag -> review solution.
- Teacher/admin pages should feel analytical and operational, not enterprise OA dashboards.
- Keep cognitive load low: structure before decoration.

## Brand And Mood
- Keywords: technical, professional, restrained, reliable.
- Default preference: light-first, dark supported.
- No anime/game-shop/neon/cyberpunk/OA-template tone.

## Typography Rules

### Font stacks
- Sans: `'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif`
- Mono accents only for ids/stats/code-like data:
  `'IBM Plex Mono', 'JetBrains Mono', 'SFMono-Regular', 'Consolas', monospace`

### Primary page title
- Use semantic `h1` for each page main title.
- Match approved challenge workspace title style:
  - `font-size: clamp(32px, 4vw, 46px)`
  - `line-height: 1.02`
  - `letter-spacing: -0.04em`
- Avoid forcing extra heavy weight unless truly needed; width/weight must not feel bloated.

### Eyebrow labels
- Keep structural eyebrow labels (English short labels are acceptable).
- Eyebrow style:
  - uppercase
  - small size
  - increased letter spacing
- Never add decorative top lines above eyebrow text.

## Layout Rules

### Root workspace shell
- Use one dominant workspace surface per page.
- Typical root pattern:
  - `section`
  - `min-h-full`
  - `flex-1`
  - rounded border + subtle gradient + soft shadow
- The page should fill main area; avoid half-height surfaces leaving empty voids.

### Flattening principle
- Prefer separators, spacing, and hierarchy over stacked cards.
- Internal sections should feel embedded in the same workspace.
- Replace mini-card stacks with:
  - section heads
  - line dividers
  - directory/list rows
  - left accent rails for key notes

### Directory/list pattern
- Provide list header row (`head`) + flat data rows.
- Rows use bottom borders, minimal background effects.
- Avoid repeated rounded containers per row.
- Admin/teacher workspace lists should use shared vertical spacing rules instead of per-page magic numbers:
  - list section head -> list body: `0.75rem`
  - list body -> bottom pagination: `0.5rem`
  - prefer shared utility classes/tokens for this rhythm; do not reintroduce page-local `mt-*`, ad-hoc `margin-top`, or `!important` overrides for the same spacing job
- Admin/teacher workspace pagination should default to a flat continuation of the list:
  - no extra top divider by default
  - keep pagination spacing compact and subordinate to the list content, not visually separated like a new card/section
- Use shared column tracks for header and rows (`grid-template-columns` must come from one source, e.g. CSS variables).
- One column should carry one primary responsibility. Do not stack unrelated fields into the title column just because space is tight.
- If a field already has an explicit column header, render it in that column instead of duplicating it inside the main title block.
- Never mix column alignment rules: header cell and body cell in the same column must keep the same horizontal alignment.
- Avoid loose `auto` tracks on key columns (`category/status/actions`); prefer stable `clamp(...)` or tokenized fixed tracks.
- For action columns, choose one mode per table:
  - mode A: header + row actions all left-aligned
  - mode B: header + row actions all right-aligned
- Treat long text handling as a default list requirement, not a later polish item:
  - primary title/name fields in rows should default to single-line ellipsis
  - secondary copy/description/access info should default to 2-line clamp unless the page explicitly needs full wrapping
  - row text containers inside grid/flex tracks must set `min-width: 0` before ellipsis/clamp rules
  - preserve full value via `title` or an equivalent accessible full-text reveal when content is truncated
- Before shipping any new or refactored list page, check extreme cases with overly long:
  - challenge titles
  - contest names
  - image refs/tags
  - usernames/class names
  - notification bodies
  - access URLs or runtime descriptions
- Do not expose internal ids/codes by default in visible list UI unless they are needed for user decision-making, search, copy, or support workflows.
- If an id is only implementation-facing, keep it out of the row and out of the header.

### Operational list decisions
- Refresh strategy must be explicit:
  - default to manual refresh for admin/teacher operational lists
  - only introduce polling when the status is truly time-sensitive and users benefit from passive updates
  - if polling is used, scope it narrowly (for example: only while active jobs exist) and keep the interval readable in UI copy
- Avoid low-signal columns or chips that do not change user decisions.
- If a tag/status merely restates obvious context or remains empty/default most of the time, remove it instead of filling the row with noise.
- List state handling must be explicit and consistent:
  - distinguish first-load empty from filtered/search no-result
  - loading, empty, error, and no-result states should each have different copy and recovery action
  - destructive or operational pages should prefer retry-in-place instead of pushing users away from the current workspace
- Search/filter/pagination rules should be decided before implementation:
  - changing filters/search resets pagination to page 1 unless there is a strong product reason not to
  - if filters are meaningful to revisit/share, persist them in route query
  - row refresh, tab switch, and back-navigation should not silently lose the user’s current filter context
- Async result ordering must be defended:
  - search/filter pages should ignore stale responses from earlier requests
  - polling/manual refresh must not overwrite newer user-triggered results with older payloads
  - do not assume requests resolve in issue order

### Row action density
- Row actions should stay compact and scannable.
- If a row needs more than 2 primary-visible actions, collapse secondary/destructive actions into a `更多`/`More` menu.
- Keep the visible action set decision-oriented:
  - one default entry action
  - optionally one secondary high-frequency action
  - everything else moves to overflow
- Do not let action buttons force content columns to collapse before the actual responsive breakpoint.

### Responsive downgrade order
- When a directory/list starts feeling cramped:
  1. tighten column tracks and gaps
  2. simplify visible row actions
  3. hide header and switch to stacked row layout only at narrower breakpoints
- Do not jump from desktop multi-column layout to single-column mobile layout too early.
- Prefer preserving table semantics around medium desktop widths instead of collapsing at the first sign of pressure.

### Action controls pattern
- Action area buttons must be hierarchical:
  - primary operation (`View`/`Manage`) gets accent-tinted style
  - secondary operation (`Solution`/`More`) stays neutral-outline style
- Keep compact and consistent sizing in list rows:
  - button min-height: `34px` desktop, `36px` mobile minimum touch comfort
  - button radius: `10px` to fit flat workspace rhythm
  - button gap: `6px` to `8px`
- Do not render all row actions with identical neutral style; it weakens scanability.
- Row action groups should expose semantics (`role="group"` + `aria-label`).
- Always keep visible keyboard focus (`:focus-visible` outline/ring); never remove default focus without replacement.

## Color And Surface Rules
- Use semantic tokens/variables; avoid ad-hoc hardcoded palette drift.
- Accent is theme-driven (support user theme preferences).
- Keep backgrounds solid or subtly layered; avoid foggy overlays.
- No glassmorphism/backdrop blur in primary workspaces.

## Copy Rules
- Keep structural labels and functional hints.
- Remove “design presentation” text that explains layout intent.
- Copy must be task-facing, short, and operational.
- Do not convert all labels to Chinese only; preserve existing bilingual structure when established.
- UI must contain only end-user product copy.
- Never render implementation/design/meta narration in visible UI, including:
  - mock/demo/prototype notes
  - option-comparison labels such as `方案 A/B/C`
  - layout-explanation text (for example: “顶部 tabs + 平铺列表 + 右侧信息轨道”)
  - process guidance text aimed at developers/reviewers
- If explanation is needed, keep it in docs/PR/assistant response, not in page content.
- Machine-value fields need a deliberate display policy:
  - define whether the value is for reading, copying, debugging, or navigation
  - if a field is frequently copied (URL, tag, slug, username, image ref), consider explicit copy affordance instead of hover-only recovery
  - mobile users cannot rely on `title` hover, so critical full values need a tap/copy/detail path

## Data Display Semantics
- Time formatting must be consistent across similar pages:
  - choose absolute time, relative time, or a paired strategy deliberately
  - use one locale/timezone policy per product area; do not mix raw timestamps with localized strings in sibling pages
  - countdown/remaining time should stay obviously different from calendar timestamps
- Numeric formatting should be stable:
  - points, counts, attempts, solve totals, and pagination totals should use the same wording/order across sibling pages
- Status semantics should be shared rather than page-local:
  - the same conceptual state (`draft/pending/running/succeeded/failed/archived`) should keep the same label tone and color meaning across pages
  - do not remap similar statuses into conflicting colors just for local visual variety
- If a status or metric does not change what the user should do next, it is a candidate for removal from the default row layout

## Interaction And Accessibility
- Keyboard operable by default.
- Tabs and collapses require proper semantics:
  - `tablist/tab/tabpanel`
  - `aria-selected`, `aria-controls`
  - `aria-expanded` on collapses
- Inputs need explicit labels, not placeholder-only labeling.
- Touch targets in mobile should not be cramped.

## Visual Anti-Patterns (Do Not Use)
- Heavy card grids inside already-carded page shells.
- Random pill overuse for every control/badge.
- Generic AI-style indigo gradients everywhere.
- Decorative explanatory paragraphs about UI organization.
- Global class names that cause style leakage (example: generic `.overline`).

## Page Pattern Presets

### Challenge workspace
- Top tabs in one bar: question/solutions/submissions/review.
- Show challenge base info in question tab only.
- Hints belong under question statement, not under environment area.
- Non-question tabs should hide right-side flag tools.

### Admin platform list pages
- `platform/challenges`, `platform/images`: flat directory rows with operational actions.
- Keep import/manage actions clear and concise.

### Environment template workspace
- Single workspace surface.
- Flat tab rail + flat template directory rows.
- Side notes and boundary status as inline/rail blocks, not nested cards.

## Implementation Checklist
- Page has one dominant workspace shell.
- Main title uses `h1` and approved title metrics.
- Internal areas are flattened (divider/list/rail first).
- List/table columns are strictly aligned (header and data share track + alignment).
- Each visible column has a clear responsibility; unrelated fields are not stuffed back into the title column.
- Internal ids are hidden unless they provide real user value.
- Long row text is hardened by default (single-line ellipsis for key titles, 2-line clamp for secondary copy, full text still recoverable).
- Admin/teacher list spacing follows the shared rhythm (`0.75rem` head->body, `0.5rem` body->pagination) instead of page-local tuning.
- Pagination is visually attached to the list by default and does not add an extra top divider unless the page has a strong product reason.
- Refresh behavior is deliberate (manual by default, polling only when justified and scoped).
- Loading/empty/error/no-result states are intentionally differentiated.
- Search/filter/pagination behavior is stable and does not unexpectedly reset user context.
- Stale async responses are ignored instead of overwriting newer list state.
- Row actions are compact enough that they do not break column layout at desktop widths.
- Responsive downgrade order is controlled (compress first, stack later).
- Action column uses clear primary/secondary button hierarchy and accessible focus/group semantics.
- Copy has no design/meta narration (no mock/proposal/process text in visible UI).
- Machine values have an intentional read/copy/mobile-access strategy.
- Time, numbers, and status colors/labels are semantically consistent with sibling pages.
- Theme tokens are used; no accidental hardcoded drift.
- Keyboard semantics are complete for tabs/collapses/forms.
