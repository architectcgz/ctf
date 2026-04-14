<script setup lang="ts">
import { computed } from 'vue'

import AdminContestFormPanel from '@/components/admin/contest/AdminContestFormPanel.vue'
import type { ContestFieldLocks, ContestFormDraft } from '@/composables/useAdminContests'

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
  <ElDialog
    class="contest-form-dialog"
    :model-value="open"
    width="720px"
    @close="closeDialog"
    @update:model-value="emit('update:open', $event)"
  >
    <template #header>
      <div class="contest-form-dialog__header">
        <div class="journal-note-label">Contest Workspace</div>
        <h2 class="contest-form-dialog__title">{{ dialogTitle }}</h2>
        <p class="contest-form-dialog__copy">{{ dialogCopy }}</p>
      </div>
    </template>

    <AdminContestFormPanel
      :mode="mode"
      :draft="draft"
      :saving="saving"
      :status-options="statusOptions"
      :field-locks="fieldLocks"
      @cancel="closeDialog"
      @save="emit('save', $event)"
    />
  </ElDialog>
</template>

<style scoped>
.contest-form-dialog__header {
  display: grid;
  gap: var(--space-2);
}

.contest-form-dialog__title {
  margin: 0;
  font-size: clamp(1.5rem, 2vw, 1.9rem);
  font-weight: 700;
  color: var(--journal-ink, var(--color-text-primary));
}

.contest-form-dialog__copy {
  margin: 0;
  line-height: 1.7;
  color: var(--journal-muted, var(--color-text-secondary));
}

:deep(.contest-form-dialog .el-dialog) {
  border: 1px solid
    color-mix(in srgb, var(--journal-border, var(--color-border-default)) 82%, transparent);
  border-radius: 24px;
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

:deep(.contest-form-dialog .el-dialog__header) {
  margin-right: 0;
  padding: 1.5rem 1.5rem 0;
}

:deep(.contest-form-dialog .el-dialog__body) {
  padding: 1rem 1.5rem 0;
}

:deep(.contest-form-dialog .el-dialog__footer) {
  padding: 1rem 1.5rem 1.5rem;
  border-top: 1px solid
    color-mix(in srgb, var(--journal-border, var(--color-border-default)) 70%, transparent);
}

:deep(.contest-form-dialog .el-dialog__headerbtn) {
  top: 1.15rem;
  right: 1.15rem;
}

:deep(.contest-form-dialog .el-dialog__close) {
  color: var(--journal-muted, var(--color-text-secondary));
}

@media (max-width: 767px) {
  :deep(.contest-form-dialog .el-dialog) {
    width: min(720px, calc(100vw - 1.5rem)) !important;
    margin-top: 4vh !important;
  }

  :deep(.contest-form-dialog .el-dialog__header),
  :deep(.contest-form-dialog .el-dialog__body),
  :deep(.contest-form-dialog .el-dialog__footer) {
    padding-left: 1rem;
    padding-right: 1rem;
  }
}
</style>
