<script setup lang="ts">
import { BellRing, Home, ServerOff } from 'lucide-vue-next'

import ErrorStatusShell from '@/components/errors/ErrorStatusShell.vue'
</script>

<template>
  <ErrorStatusShell
    status-code="503"
    kicker="503 · Service Unavailable"
    title="服务暂时不可用"
    description="系统正在维护或暂时承压，当前无法完成请求。建议稍后再试，或先返回工作台继续其他操作。"
    accent="warning"
    :icon="ServerOff"
    :primary-icon="Home"
    :secondary-icon="BellRing"
    primary-to="/dashboard"
    primary-label="返回工作台"
  />
</template>
