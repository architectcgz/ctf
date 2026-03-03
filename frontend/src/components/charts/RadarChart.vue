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

const option = computed<EChartsOption>(() => ({
  tooltip: { trigger: 'item' },
  radar: {
    indicator: props.indicators.map((indicator) => ({
      name: indicator.name,
      max: indicator.max ?? 100,
    })),
    splitArea: {
      areaStyle: {
        color: ['rgba(8, 145, 178, 0.04)'],
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
          areaStyle: { color: 'rgba(8, 145, 178, 0.20)' },
          lineStyle: { color: '#0891b2' },
          itemStyle: { color: '#06b6d4' },
        },
      ],
    },
  ],
}))
</script>

<template>
  <VChart class="h-80 w-full" :option="option" autoresize />
</template>
