<script setup>
import { computed } from 'vue'

const props = defineProps({
  scores: {
    type: Array,
    required: true,
  },
})

const center = 140
const radius = 96
const rings = [0.25, 0.5, 0.75, 1]
const gridStroke = 'color-mix(in srgb, var(--color-text-muted) 18%, transparent)'
const polygonFill = 'color-mix(in srgb, var(--color-primary) 22%, transparent)'
const polygonStroke = 'var(--color-primary)'
const centerFill = 'color-mix(in srgb, var(--color-primary-hover) 82%, var(--color-primary))'
const labelFill = 'var(--color-text-muted)'

const points = computed(() => {
  const total = props.scores.length

  return props.scores.map((item, index) => {
    const angle = (-Math.PI / 2) + (Math.PI * 2 * index) / total
    const x = center + Math.cos(angle) * radius * (item.value / 100)
    const y = center + Math.sin(angle) * radius * (item.value / 100)
    const labelX = center + Math.cos(angle) * (radius + 28)
    const labelY = center + Math.sin(angle) * (radius + 28)

    return {
      ...item,
      x,
      y,
      labelX,
      labelY,
    }
  })
})

const polygon = computed(() => points.value.map((point) => `${point.x},${point.y}`).join(' '))

function ringPoints(scale) {
  return points.value
    .map((point, index) => {
      const angle = (-Math.PI / 2) + (Math.PI * 2 * index) / points.value.length
      return `${center + Math.cos(angle) * radius * scale},${center + Math.sin(angle) * radius * scale}`
    })
    .join(' ')
}
</script>

<template>
  <div class="rounded-2xl border border-[var(--color-primary)]/10 bg-[var(--color-bg-surface)] p-4">
    <svg viewBox="0 0 280 280" class="h-[280px] w-full">
      <polygon
        v-for="scale in rings"
        :key="scale"
        :points="ringPoints(scale)"
        fill="none"
        :stroke="gridStroke"
        stroke-width="1"
      />
      <line
        v-for="point in points"
        :key="`${point.name}-axis`"
        :x1="center"
        :y1="center"
        :x2="point.labelX - ((point.labelX - center) * 0.12)"
        :y2="point.labelY - ((point.labelY - center) * 0.12)"
        :stroke="gridStroke"
      />
      <polygon :points="polygon" :fill="polygonFill" :stroke="polygonStroke" stroke-width="2" />
      <circle cx="140" cy="140" r="4" :fill="centerFill" />
      <g v-for="point in points" :key="point.name">
        <circle :cx="point.x" :cy="point.y" r="4" :fill="point.color" />
        <text
          :x="point.labelX"
          :y="point.labelY"
          :fill="labelFill"
          font-size="12"
          text-anchor="middle"
        >
          {{ point.name }}
        </text>
      </g>
    </svg>
  </div>
</template>
