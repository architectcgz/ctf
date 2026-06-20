# Typography Rules

Read this file when the task touches page titles, eyebrow labels, font stacks, or text scale.

## Font stacks

- Sans: `'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif`
- Mono accents only for ids/stats/code-like data:
  `'IBM Plex Mono', 'JetBrains Mono', 'SFMono-Regular', 'Consolas', monospace`

## Primary page title

- Use semantic `h1` for each page main title.
- Match approved challenge workspace title style:
  - `font-size: clamp(32px, 4vw, 46px)`
  - `line-height: 1.02`
  - `letter-spacing: -0.04em`
- Avoid forcing extra heavy weight unless truly needed; width/weight must not feel bloated.
- Reuse shared workspace page-header styles/tokens instead of duplicating the same title typography inside page-local scoped CSS.
- If a page needs a different intro width, keep that as a local layout override like `max-width`; do not re-declare the shared `font-size` / `line-height` / `letter-spacing` metrics per page.

## Eyebrow labels

- Keep structural eyebrow labels (English short labels are acceptable).
- Eyebrow style: uppercase, small size, increased letter spacing.
- Never add decorative top lines above eyebrow text.
- Non-active tab panels must not repeat overview-only eyebrows (see `layout-rules.md`).
