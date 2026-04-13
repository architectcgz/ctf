<script setup lang="ts">
import { ArrowLeft, Gauge, Home } from 'lucide-vue-next'

import ErrorStatusShell from '@/components/errors/ErrorStatusShell.vue'
</script>

<template>
  <ErrorStatusShell
    status-code="429"
    kicker="429 · Too Many Requests"
    title="请求过于频繁"
    description="系统在短时间内收到了过多请求。建议稍后再试，并先返回上一页避免重复触发当前操作；如果没有可返回记录，再回工作台。"
    accent="warning"
    :icon="Gauge"
    :primary-icon="ArrowLeft"
    :secondary-icon="Home"
    primary-action="back"
    secondary-to-home
  />
</template>
