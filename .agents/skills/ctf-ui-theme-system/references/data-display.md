# Data Display and Naming

Read this file when the task touches identifiers, machine values, time/number/status formatting, or de-cluttering decisions.

## Identifiers and types

- **Clean identifiers**: render usernames, slugs, and IDs as raw text. Do not add decorative
  prefixes like `@` (use `john_doe`, not `@john_doe`) unless the design strictly mandates a shell/terminal aesthetic.
- **Type safety**: NEVER use `any` for DTOs or form payloads. Always define explicit interfaces
  (e.g., `interface CreateContestPayload`) to ensure data integrity.

## Machine-value display policy

- Machine-value fields need a deliberate display policy:
  - define whether the value is for reading, copying, debugging, or navigation
  - if a field is frequently copied (URL, tag, slug, username, image ref), provide an explicit copy affordance instead of hover-only recovery
  - mobile users cannot rely on `title` hover, so critical full values need a tap/copy/detail path
- Do not expose internal ids/codes by default unless needed for user decisions, search, copy, or support.
  If an id is only implementation-facing, keep it out of the row and the header.

## Time formatting

- Must be consistent across similar pages:
  - choose absolute time, relative time, or a paired strategy deliberately
  - use one locale/timezone policy per product area; do not mix raw timestamps with localized strings in sibling pages
  - countdown/remaining time should stay obviously different from calendar timestamps

## Numeric and status formatting

- Numeric formatting should be stable: points, counts, attempts, solve totals, and pagination totals
  use the same wording/order across sibling pages.
- Status semantics should be shared, not page-local:
  - the same conceptual state (`draft/pending/running/succeeded/failed/archived`) keeps the same label tone and color meaning across pages
  - do not remap similar statuses into conflicting colors for local visual variety
- If a status or metric does not change what the user should do next, it is a candidate for removal from the default row layout.

## Complexity management (de-cluttering)

- **Eliminate redundant focus states**: avoid heavy "focus shelves" or secondary highlight bars that
  duplicate the function of the main list row. Use row-level active states instead.
- **Action hierarchy**: only show high-frequency actions by default; hide management/destructive
  operations behind overflow menus or secondary drawers.
