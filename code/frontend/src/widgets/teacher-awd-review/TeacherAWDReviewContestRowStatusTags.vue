<script setup lang="ts">
defineProps<{
  statusLabel: string
  exportReady: boolean
}>()
</script>

<template>
  <div class="teacher-directory-row-tags">
    <span class="teacher-directory-chip">
      {{ statusLabel }}
    </span>
    <span
      class="teacher-directory-chip"
      :class="exportReady ? '' : 'teacher-directory-chip-muted'"
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

.teacher-directory-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.7rem;
  padding: 0 var(--space-2-5);
  border-radius: 0.5rem;
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  font-size: var(--font-size-0-75);
  font-weight: 600;
  color: var(--journal-accent-strong);
}

.teacher-directory-chip-muted {
  background: color-mix(in srgb, var(--journal-muted) 10%, transparent);
  color: var(--journal-muted);
}
</style>
