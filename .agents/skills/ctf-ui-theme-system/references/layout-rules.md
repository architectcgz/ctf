# Layout Rules

Read this file when the task changes shell structure, panel organization, spacing, rails, lists, or action layout.

## Root workspace shell

- Use one dominant workspace surface per page.
- Preferred structure:
  - `workspace-shell`
  - `workspace-topbar`
  - `top-tabs` when the route contains multiple page-level views
  - `content-pane`
  - optional `context-rail`
- The page should fill the main area; avoid half-height shells that leave large dead space.

## Flattening principle

- Prefer separators, spacing, and hierarchy over stacked cards.
- Internal sections should feel embedded in the same workspace.
- Replace repeated mini-card stacks with section heads, dividers, flat rows, or compact rails.

## Module boundaries

- Use spacing first between modules.
- Keep only one explicit divider for each module boundary.
- Do not stack a previous block's `border-bottom` and the next block's `border-top` on the same boundary.

## Top tab rail pattern

- Route-level tabs sit directly below `workspace-topbar` and before the main content region.
- Route-level tabs synchronize with a stable `?panel=` query key.
- If a header or action group only applies to one panel, move it inside that `tabpanel`.
- Non-active panels must not repeat overview-only headers, copy, or actions.
- Do not override shared tab visibility by forcing every `.tab-panel` visible in local CSS.
- Do not use tabs to split "overview" metrics from the directory they describe.
  For backend-style management pages, prefer the SaaS workbench pattern in `references/saas-workbench-pattern.md`.

## Content pane and context rail

- `content-pane` holds the active task content.
- `context-rail` holds switching, navigation shortcuts, export entry points, or compact utility notes.
- The rail must not duplicate the main task body.

## Directory and list pattern

- Use one header row plus flat data rows.
- Prefer shared column tracks between header and rows.
- Avoid loose `auto` tracks on key columns such as category, status, or actions.
- Keep pagination attached to the list section, directly after the last row.
- Keep the section heading, filter bar, and list body inside the same `workspace-directory-section`.
- Prefer a light `list-heading` over a separate filter card header when the section purpose is already obvious from the list title.
- When the page also needs KPIs, place a metric strip directly above the directory section instead of moving the list into a sibling tab.

## Filter bars

- List and directory filter bars should auto-apply instead of relying on an explicit submit button.
- Text input filters should use a short debounce, around `250ms`, before refreshing results.
- Status, date, and toggle filters should join the same auto-apply flow rather than introducing a separate "应用筛选" action.
- Keep only actions that are not part of filtering itself, such as reset, refresh, export, or create.
- Do not render explanatory copy, “激活筛选 X 项”, or a dedicated filter card title when the surrounding section already makes the context clear.
- Order high-frequency filters first, then keyword search, then non-filter actions such as clear or export.
- When filter count grows, keep high-frequency filters visible and collapse low-frequency filters behind a compact “更多筛选 / 收起筛选” toggle.
- Expanded low-frequency filters should reuse the same control width as the primary select columns instead of switching to a different visual scale.
- A reset or clear action should live in the same action row as the collapse toggle rather than occupying a full extra row by itself.

## Action controls

- Primary row action gets accent-tinted styling.
- Secondary row action stays neutral-outline.
- Keep row actions compact and consistent:
  - min-height `34px` desktop
  - min-height `36px` mobile
  - radius around `10px`
  - gap `6px` to `8px`
- Use visible `:focus-visible` states and `role="group"` for grouped row actions.
- If a shortcut applies to one concrete row, it belongs in the row action group, not in a floating summary block above the list.

## Workspace hero contract

- When a workspace page needs both narrative context and operational entry actions, use the standard `workspace-hero` two-column intro instead of inventing a page-local header.
- Left column owns: eyebrow / structural label, semantic `h1`, one concise summary paragraph.
- Right column owns: quick actions, compact meta/status badges, bottom-aligned operational controls.
- Default hero structure:
  - `grid-template-columns: minmax(0, 1fr) auto`
  - shared page-title typography variables (see `typography-rules.md`)
  - bottom divider + spacing that visually connects the hero to the same workspace surface
- Do not wrap the hero in another card, stat shelf, or inset panel just to place action buttons.
- Collapse to one column only when the right-side action rail would otherwise crush the title/summary block.
- Local page CSS may tune width, gaps, or action alignment, but must not re-declare a separate hero typography system when shared workspace title tokens already exist.

## Directory/list — column responsibility and long text

- One column carries one primary responsibility. Do not stack unrelated fields into the title column because space is tight.
- If a field already has an explicit column header, render it in that column instead of duplicating it inside the main title block.
- Never mix column alignment: header cell and body cell in the same column keep the same horizontal alignment.
- Avoid loose `auto` tracks on key columns (`category/status/actions`); prefer stable `clamp(...)` or tokenized fixed tracks.
- For action columns, choose one mode per table: all left-aligned, or all right-aligned (header + rows together).
- Treat long-text handling as a default requirement, not later polish:
  - primary title/name fields default to single-line ellipsis
  - secondary copy/description/access info default to 2-line clamp unless the page needs full wrapping
  - row text containers inside grid/flex tracks must set `min-width: 0` before ellipsis/clamp
  - preserve full value via `title` or an equivalent accessible full-text reveal when truncated
- Before shipping a new/refactored list page, check extreme cases with overly long: challenge titles, contest names, image refs/tags, usernames/class names, notification bodies, access URLs, runtime descriptions.
- Do not expose internal ids/codes by default unless needed for user decisions, search, copy, or support (see `data-display.md`).
- Investigation/review/audit lists prefer action-first directory rows over stacks of alert cards:
  - keep all rows inside one shared list surface
  - make the whole row clickable when the dominant action is to open detail/audit context
  - reveal a lightweight affordance on hover/focus (`ArrowRight` or equivalent) instead of one visible action button per row
  - guide via title color shift, subtle surface lift, action reveal; avoid aggressive transforms or floating-card theatrics
  - keep badges/time/context compact and subordinate to the row title + reason
- If a row carries warning/risk semantics, keep the row structure neutral and push urgency into semantic badges or inline notes instead of tinting the whole list area.

## Operational list decisions

- Refresh strategy must be explicit:
  - default to manual refresh for admin/teacher operational lists
  - introduce polling only when status is truly time-sensitive and users benefit from passive updates
  - if polling is used, scope it narrowly (e.g. only while active jobs exist) and keep the interval readable in UI copy
- Avoid low-signal columns or chips that do not change user decisions; if a tag/status restates obvious context or stays empty/default most of the time, remove it.
- List state handling must be explicit and consistent:
  - distinguish first-load empty from filtered/search no-result
  - loading, empty, error, and no-result states each have different copy and recovery action
  - destructive/operational pages prefer retry-in-place instead of pushing users away from the workspace
- Search/filter/pagination rules decided before implementation:
  - changing filters/search resets pagination to page 1 unless there is a strong product reason not to
  - if filters are meaningful to revisit/share, persist them in route query
  - row refresh, tab switch, and back-navigation should not silently lose the current filter context
- Async result ordering must be defended:
  - ignore stale responses from earlier requests
  - polling/manual refresh must not overwrite newer user-triggered results with older payloads
  - do not assume requests resolve in issue order

## Row action density

- Row actions stay compact and scannable.
- If a row has only one or two actions, show them directly; do not hide them behind `更多`/`More`.
- A two-action row usually renders the default action and the destructive/secondary action side by side, with the destructive action using the danger button treatment and confirmation flow.
- If a row needs more than 2 primary-visible actions, collapse secondary/destructive actions into a `更多`/`More` menu.
- Keep the visible action set decision-oriented: one default entry action, optionally one secondary high-frequency action, everything else in overflow.
- Do not let action buttons force content columns to collapse before the actual responsive breakpoint.

## Responsive downgrade order

- When a directory/list starts feeling cramped:
  1. tighten column tracks and gaps
  2. simplify visible row actions
  3. hide header and switch to stacked row layout only at narrower breakpoints
- Do not jump from desktop multi-column layout to single-column mobile layout too early.
- Prefer preserving table semantics around medium desktop widths instead of collapsing at the first sign of pressure.
