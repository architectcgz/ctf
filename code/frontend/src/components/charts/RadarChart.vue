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

const props = withDefaults(
  defineProps<{
    indicators: Indicator[]
    values: number[]
    name?: string
    heightClass?: string
    labelFontSize?: number
    axisNameGap?: number
    radius?: string | number
    centerY?: string
  }>(),
  {
    name: '能力画像',
    heightClass: 'h-80',
    labelFontSize: 14,
    axisNameGap: 18,
    radius: '68%',
    centerY: '50%',
  }
)

function cssVar(name: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
}

const option = computed<EChartsOption>(() => {
  const primary = cssVar('--color-primary')
  const primaryHover = cssVar('--color-primary-hover')
  const axisLabelColor =
    cssVar('--color-text-primary') || cssVar('--color-text-secondary') || primary
  return {
    tooltip: { trigger: 'item' },
    radar: {
      center: ['50%', props.centerY],
      radius: props.radius,
      indicator: props.indicators.map((indicator) => ({
        name: indicator.name,
        max: indicator.max ?? 100,
      })),
      axisName: {
        color: axisLabelColor,
        fontSize: props.labelFontSize,
        fontWeight: 600,
      },
      axisNameGap: props.axisNameGap,
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
  <VChart :class="[props.heightClass, 'w-full']" :option="option" autoresize />
</template>
