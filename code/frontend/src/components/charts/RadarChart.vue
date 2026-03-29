<script setup lang="ts">
import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { use } from 'echarts/core'
import { RadarChart as EChartsRadarChart } from 'echarts/charts'
import { CanvasRenderer } from 'echarts/renderers'
import { LegendComponent, RadarComponent, TooltipComponent } from 'echarts/components'
import VChart from 'vue-echarts'

use([RadarComponent, TooltipComponent, LegendComponent, EChartsRadarChart, CanvasRenderer])

interface Indicator {
  name: string
  max?: number
}

const props = withDefaults(defineProps<{
  indicators: Indicator[]
  values: number[]
  name?: string
}>(), {
  name: '能力画像',
})

function cssVar(name: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
}

const option = computed<EChartsOption>(() => {
  const primary = cssVar('--color-primary')
  const primaryHover = cssVar('--color-primary-hover')
  return {
    tooltip: { trigger: 'item' },
    radar: {
      indicator: props.indicators.map((indicator) => ({
        name: indicator.name,
        max: indicator.max ?? 100,
      })),
      splitArea: {
        areaStyle: {
          color: [`${primary}0a`],
        },
      },
    },
    series: [
      {
        name: props.name,
        type: 'radar',
        data: [
          {
            value: props.values,
            name: props.name,
            areaStyle: { color: `${primary}33` },
            lineStyle: { color: primary },
            itemStyle: { color: primaryHover },
          },
        ],
      },
    ],
  }
})
</script>

<template>
  <VChart class="h-80 w-full" :option="option" autoresize />
</template>
