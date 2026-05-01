<script setup lang="ts">
import { computed } from 'vue'

import PlatformContestFormPanel from '@/components/platform/contest/PlatformContestFormPanel.vue'
import AdminSurfaceModal from '@/components/common/modal-templates/AdminSurfaceModal.vue'
import type { ContestFieldLocks, ContestFormDraft } from '@/features/platform-contests'

const props = defineProps<{
  open: boolean
  mode: 'create' | 'edit'
  draft: ContestFormDraft
  saving: boolean
  statusOptions: Array<{ label: string; value: ContestFormDraft['status'] }>
  fieldLocks: ContestFieldLocks
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  save: [value: ContestFormDraft]
}>()

const dialogTitle = computed(() => (props.mode === 'create' ? '创建竞赛' : '编辑竞赛'))
const dialogCopy = computed(() =>
  props.mode === 'create'
    ? '填写赛事基础信息和时间窗口，保存后即可在赛事目录里继续编排。'
    : '更新赛事窗口、赛制信息和状态，让目录与运维视图保持同步。'
)

function closeDialog() {
  emit('update:open', false)
}
</script>

<template>
  <AdminSurfaceModal
    :open="open"
    class="contest-form-dialog"
    :title="dialogTitle"
    :subtitle="dialogCopy"
    eyebrow="Contest Workspace"
    width="45rem"
    @close="closeDialog"
    @update:open="emit('update:open', $event)"
  >
    <PlatformContestFormPanel
      :mode="mode"
      :draft="draft"
      :saving="saving"
      :status-options="statusOptions"
      :field-locks="fieldLocks"
      @cancel="closeDialog"
      @save="emit('save', $event)"
    />
  </AdminSurfaceModal>
</template>

<style scoped>
:deep(.contest-form-dialog .modal-template-panel--classic) {
  border-color: color-mix(
    in srgb,
    var(--journal-border, var(--color-border-default)) 82%,
    transparent
  );
  background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--journal-accent, var(--color-primary)) 10%, transparent),
      transparent 18rem
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 98%, var(--color-bg-base)),
      var(--journal-surface, var(--color-bg-surface))
    );
  box-shadow: 0 24px 60px var(--color-shadow-soft);
}

:deep(.contest-form-dialog .modal-template-classic__header) {
  border-bottom-color: color-mix(
    in srgb,
    var(--journal-border, var(--color-border-default)) 70%,
    transparent
  );
  background: transparent;
}

:deep(.contest-form-dialog .modal-template-classic__body) {
  padding: 1rem 1.5rem 0;
}

@media (max-width: 767px) {
  :deep(.contest-form-dialog .modal-template-classic__header),
  :deep(.contest-form-dialog .modal-template-classic__body) {
    padding-left: 1rem;
    padding-right: 1rem;
  }
}
</style>
