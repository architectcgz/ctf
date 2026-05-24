<script setup lang="ts">
import AppEmpty from '@/components/common/AppEmpty.vue'

defineProps<{
  loading: boolean
  error: string | null
  hasContests: boolean
}>()

const emit = defineEmits<{
  reload: []
}>()
</script>

<template>
  <div
    v-if="loading"
    class="teacher-skeleton-list workspace-directory-loading"
  >
    <div
      v-for="index in 3"
      :key="index"
      class="h-28 animate-pulse rounded-[22px] bg-[color-mix(in_srgb,var(--journal-surface-subtle)_92%,transparent)]"
    />
  </div>

  <AppEmpty
    v-else-if="error"
    class="teacher-empty-state workspace-directory-empty"
    icon="AlertTriangle"
    title="AWD复盘目录加载失败"
    :description="error"
  >
    <template #action>
      <button
        type="button"
        class="header-btn header-btn--primary"
        @click="emit('reload')"
      >
        重新加载
      </button>
    </template>
  </AppEmpty>

  <AppEmpty
    v-else-if="!hasContests"
    class="teacher-empty-state workspace-directory-empty"
    icon="Waypoints"
    title="暂无 AWD 赛事"
    description="当前还没有可进入复盘的 AWD 赛事。"
  />

  <slot v-else />
</template>

<style scoped>
.teacher-skeleton-list {
  display: grid;
  gap: var(--space-3);
}
</style>
