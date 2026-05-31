export const RADAR_GRID_STROKE = 'color-mix(in srgb, var(--color-text-muted) 18%, transparent)'
export const RADAR_AREA_FILL = 'color-mix(in srgb, var(--color-primary) 42%, transparent)'
export const RADAR_AREA_STROKE = 'var(--color-primary)'
export const RADAR_POINT_FILL =
  'color-mix(in srgb, var(--color-primary-hover) 82%, var(--color-primary))'
export const RADAR_LABEL_FILL = 'var(--color-text-muted)'

const RADAR_CANVAS_AREA_ALPHA = '5c'
const RADAR_CANVAS_SPLIT_ALPHA = '12'

function cssVar(name: string): string {
  if (typeof window === 'undefined') {
    return ''
  }
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
}

function withAlpha(color: string, alpha: string): string {
  if (!color || color.startsWith('color-mix(') || color.startsWith('var(')) {
    return color
  }
  return `${color}${alpha}`
}

export function resolveRadarCanvasVisuals(semanticAreaFill = RADAR_AREA_FILL) {
  const primary = cssVar('--color-primary')
  const primaryHover = cssVar('--color-primary-hover') || primary
  const axisLabelColor =
    cssVar('--color-text-primary') || cssVar('--color-text-secondary') || primary

  return {
    primary,
    pointFill: primaryHover,
    axisLabelColor,
    splitAreaFill: primary ? withAlpha(primary, RADAR_CANVAS_SPLIT_ALPHA) : RADAR_GRID_STROKE,
    areaFill: primary ? withAlpha(primary, RADAR_CANVAS_AREA_ALPHA) : semanticAreaFill,
  }
}
