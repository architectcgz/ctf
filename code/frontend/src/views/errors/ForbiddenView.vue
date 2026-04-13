<script setup lang="ts">
import { ArrowLeft, Home, LockKeyhole } from 'lucide-vue-next'

import ErrorStatusShell from '@/components/errors/ErrorStatusShell.vue'
</script>

<template>
  <ErrorStatusShell
    status-code="403"
    kicker="403 · Access Restricted"
    title="这个区域需要更高权限"
    description="你当前没有访问这个区域的权限。可能是账号角色不匹配，或该入口仅对特定身份开放。建议先返回上一页从正式入口重新进入；如果没有可返回记录，再回工作台。"
    accent="danger"
    :icon="LockKeyhole"
    :primary-icon="ArrowLeft"
    :secondary-icon="Home"
    primary-action="back"
    secondary-to-home
  />
</template>
