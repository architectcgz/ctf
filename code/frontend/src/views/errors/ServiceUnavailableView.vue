<script setup lang="ts">
import { ArrowLeft, LayoutDashboard, ServerOff } from 'lucide-vue-next'

import ErrorStatusShell from '@/components/errors/ErrorStatusShell.vue'
</script>

<template>
  <ErrorStatusShell
    status-code="503"
    kicker="503 · Service Unavailable"
    title="服务暂时不可用"
    description="系统正在维护或暂时承压，当前无法完成请求。建议先返回上一页稍后再试；若仍不可用，再回工作台继续其他操作。"
    accent="warning"
    :icon="ServerOff"
    :primary-icon="ArrowLeft"
    :secondary-icon="LayoutDashboard"
    primary-action="back"
    secondary-to-home
  />
</template>
