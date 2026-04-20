<script setup lang="ts">
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'

withDefaults(
  defineProps<{
    loading?: boolean
    error?: string
    empty?: boolean
    emptyTitle?: string
    emptyDescription?: string
  }>(),
  {
    loading: false,
    error: '',
    empty: false,
    emptyTitle: '暂无数据',
    emptyDescription: '',
  }
)
</script>

<template>
  <div class="awd-status-shell">
    <div
      v-if="loading"
      class="awd-status-shell__state"
    >
      <AppLoading>正在加载 AWD 页面...</AppLoading>
    </div>

    <AppEmpty
      v-else-if="error"
      icon="AlertTriangle"
      title="加载失败"
      :description="error"
    />

    <AppEmpty
      v-else-if="empty"
      icon="Inbox"
      :title="emptyTitle"
      :description="emptyDescription"
    />

    <div
      v-else
      class="awd-status-shell__content"
    >
      <slot />
    </div>
  </div>
</template>

<style scoped>
.awd-status-shell,
.awd-status-shell__content {
  min-height: 100%;
}

.awd-status-shell__state {
  display: flex;
  min-height: 220px;
  align-items: center;
  justify-content: center;
  padding: 2rem 1rem;
}
</style>
