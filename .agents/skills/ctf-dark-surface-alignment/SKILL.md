---
name: ctf-dark-surface-alignment
description: >
  Use when aligning CTF frontend pages in this repo to the established
  low-contrast dark surface system, especially when Vue pages still use
  hardcoded light colors, Element Plus defaults leak through, or a user asks to
  match the surface style of pages like /academy/instances, /academy/overview,
  or /notifications. Activate on "这个页面暗色不对 / 还是亮的 / ElTable·ElDialog
  漏白底 / 这个节点(XPath)颜色不对".
---

# CTF Dark Surface Alignment

Apply the repo's established low-contrast dark surface system instead of inventing a new theme.

## Use When

- A CTF page still looks light in dark theme
- Element Plus wrappers leak default light backgrounds
- A page only partially follows the established academy or notification surface style
- The user points to a specific DOM node or XPath that is still visually wrong

## Do Not Use

- Layout redesign or information architecture changes
- Broad design-system work that should follow `ctf-ui-theme-system`

## Workflow

1. Read the route view first, then find the real rendering component under `frontend/src/components/...`.
2. Compare against the nearest reference page such as `/academy/instances`, `/academy/overview`, or `/notifications` (teacher backend routes live under `/academy/*`, not `/teacher/*`).
3. Load only the reference files that match the problem.
4. Fix shared tokens, shared selectors, or Element Plus wrappers before adding page-local overrides.
5. If the user gave an exact selector or XPath, verify that exact leaking layer changed.

## Reference Map

- `references/surface-tokens.md`
  Read for approved dark surface tokens, shell treatment, and dense-text contrast.
- `references/element-plus-overrides.md`
  Read when `ElTable`, `ElDialog`, inputs, textareas, or other Element Plus wrappers leak light backgrounds.
- `references/selector-ownership.md`
  Read when the target mixes shared utility classes, the user gives an XPath, or metric-panel ownership is involved.
- `references/verification.md`
  Read before closing the task.

## Output Expectations (✓Check before closing)

- ✓Check：触达路径上是否还有 hardcoded light surface 残留？
- ✓Check：共享暗色面是否比 raw black+white 更低对比，没有破坏既有节奏？
- ✓Check：共享 selector 边界是否完好，没有误伤其它页面？
- ✓Check：verification 是否确认了**确切**泄漏节点，而不仅是它的父壳？
