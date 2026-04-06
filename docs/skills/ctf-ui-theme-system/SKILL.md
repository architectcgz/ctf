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
- Use shared column tracks for header and rows (`grid-template-columns` must come from one source, e.g. CSS variables).
- Never mix column alignment rules: header cell and body cell in the same column must keep the same horizontal alignment.
- Avoid loose `auto` tracks on key columns (`category/status/actions`); prefer stable `clamp(...)` or tokenized fixed tracks.
- For action columns, choose one mode per table:
  - mode A: header + row actions all left-aligned
  - mode B: header + row actions all right-aligned

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
- Action column uses clear primary/secondary button hierarchy and accessible focus/group semantics.
- Copy has no design-brief style narration.
- Theme tokens are used; no accidental hardcoded drift.
- Keyboard semantics are complete for tabs/collapses/forms.
