<script setup lang="ts">
import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { use } from 'echarts/core'
import { RadarChart as EChartsRadarChart } from 'echarts/charts'
import { CanvasRenderer } from 'echarts/renderers'
import { LegendComponent, RadarComponent, TooltipComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { RADAR_AREA_FILL, resolveRadarCanvasVisuals } from '@/components/charts/radarVisuals'

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

const option = computed<EChartsOption>(() => {
  const visuals = resolveRadarCanvasVisuals(RADAR_AREA_FILL)
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
        color: visuals.axisLabelColor,
        fontSize: props.labelFontSize,
        fontWeight: 600,
      },
      axisNameGap: props.axisNameGap,
      splitArea: {
        areaStyle: {
          color: [visuals.splitAreaFill],
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
            areaStyle: { color: visuals.areaFill },
            lineStyle: { color: visuals.primary },
            itemStyle: { color: visuals.pointFill },
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
