<script setup lang="ts">
defineProps<{
  statusLabel: string
  exportReady: boolean
}>()
</script>

<template>
  <div class="teacher-directory-row-tags">
    <span
      class="teacher-directory-chip"
      :class="['workspace-directory-status-pill', 'workspace-directory-status-pill--primary']"
    >
      {{ statusLabel }}
    </span>
    <span
      class="teacher-directory-chip"
      :class="
        exportReady
          ? 'workspace-directory-status-pill workspace-directory-status-pill--success'
          : 'workspace-directory-status-pill workspace-directory-status-pill--muted teacher-directory-chip-muted'
      "
    >
      {{ exportReady ? '可导出' : '实时复盘' }}
    </span>
  </div>
</template>

<style scoped>
.teacher-directory-row-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}
</style>
