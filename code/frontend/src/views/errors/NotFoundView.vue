<script setup lang="ts">
import { ArrowLeft, Home, MapPinned } from 'lucide-vue-next'

import ErrorStatusShell from '@/components/errors/ErrorStatusShell.vue'
</script>

<template>
  <ErrorStatusShell
    status-code="404"
    kicker="404 · Not Found"
    title="页面不存在"
    description="你访问的地址不存在，可能已被移动或删除。建议先返回上一页确认入口；如果没有可返回记录，再回当前工作台继续浏览。"
    accent="warning"
    :icon="MapPinned"
    :primary-icon="ArrowLeft"
    :secondary-icon="Home"
    primary-action="back"
    secondary-to-home
  />
</template>
