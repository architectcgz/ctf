<script setup lang="ts">
import { Home, LifeBuoy, MapPinned } from 'lucide-vue-next'

import ErrorStatusShell from '@/components/errors/ErrorStatusShell.vue'
</script>

<template>
  <ErrorStatusShell
    status-code="404"
    kicker="404 · Not Found"
    title="页面不存在"
    description="你访问的地址不存在，可能已被移动或删除。建议先返回当前工作台，或去通知中心确认入口是否已更新。"
    accent="warning"
    :icon="MapPinned"
    :primary-icon="Home"
    :secondary-icon="LifeBuoy"
  />
</template>
