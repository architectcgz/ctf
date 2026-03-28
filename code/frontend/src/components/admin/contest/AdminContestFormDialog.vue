<script setup lang="ts">
import { reactive, watch } from 'vue'

import { getStatusLabel } from '@/utils/contest'
import type { ContestFormDraft } from '@/composables/useAdminContests'

interface ContestFieldLocks {
  mode: boolean
  starts_at: boolean
  ends_at: boolean
}

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

const localDraft = reactive<ContestFormDraft>({
  title: '',
  description: '',
  mode: 'jeopardy',
  starts_at: '',
  ends_at: '',
  status: 'draft',
})

const fieldErrors = reactive<Partial<Record<keyof ContestFormDraft, string>>>({})

watch(
  () => props.draft,
  (draft) => {
    Object.assign(localDraft, draft)
    clearErrors()
  },
  { immediate: true, deep: true }
)

function clearErrors() {
  fieldErrors.title = ''
  fieldErrors.starts_at = ''
  fieldErrors.ends_at = ''
}

function closeDialog() {
  emit('update:open', false)
}

function validate(): boolean {
  clearErrors()

  if (!localDraft.title.trim()) {
    fieldErrors.title = '请填写竞赛标题'
  }

  if (!localDraft.starts_at) {
    fieldErrors.starts_at = '请填写开始时间'
  }

  if (!localDraft.ends_at) {
    fieldErrors.ends_at = '请填写结束时间'
  }

  if (localDraft.starts_at && localDraft.ends_at && new Date(localDraft.ends_at) <= new Date(localDraft.starts_at)) {
    fieldErrors.ends_at = '结束时间必须晚于开始时间'
  }

  return !fieldErrors.title && !fieldErrors.starts_at && !fieldErrors.ends_at
}

function handleSubmit() {
  if (!validate()) {
    return
  }

  emit('save', {
    title: localDraft.title,
    description: localDraft.description,
    mode: localDraft.mode,
    starts_at: localDraft.starts_at,
    ends_at: localDraft.ends_at,
    status: localDraft.status,
  })
}
</script>

<template>
  <ElDialog
    :model-value="open"
    :title="mode === 'create' ? '创建竞赛' : '编辑竞赛'"
    width="640px"
    @close="closeDialog"
    @update:model-value="emit('update:open', $event)"
  >
    <form class="space-y-5" @submit.prevent="handleSubmit">
      <div class="space-y-2">
        <label class="text-sm font-medium text-[var(--color-text-primary)]" for="contest-title">竞赛标题</label>
        <input
          id="contest-title"
          v-model="localDraft.title"
          type="text"
          class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          placeholder="例如：2026 春季校园 CTF"
        >
        <p v-if="fieldErrors.title" class="text-xs text-[var(--color-danger)]">{{ fieldErrors.title }}</p>
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <label class="text-sm font-medium text-[var(--color-text-primary)]" for="contest-mode">竞赛模式</label>
          <select
            id="contest-mode"
            v-model="localDraft.mode"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="fieldLocks.mode"
          >
            <option value="jeopardy">Jeopardy</option>
            <option value="awd">AWD</option>
          </select>
          <p v-if="fieldLocks.mode" class="text-xs text-[var(--color-text-muted)]">竞赛进入 draft 之后不再允许修改模式。</p>
        </div>

        <div v-if="mode === 'edit'" class="space-y-2">
          <label class="text-sm font-medium text-[var(--color-text-primary)]" for="contest-status">状态</label>
          <select
            id="contest-status"
            v-model="localDraft.status"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          >
            <option
              v-for="option in statusOptions"
              :key="option.value"
              :value="option.value"
            >
              {{ getStatusLabel(option.value) }}
            </option>
          </select>
        </div>
      </div>

      <div class="space-y-2">
        <label class="text-sm font-medium text-[var(--color-text-primary)]" for="contest-description">竞赛描述</label>
        <textarea
          id="contest-description"
          v-model="localDraft.description"
          rows="4"
          class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          placeholder="描述赛制、参赛范围或报名说明。"
        />
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <label class="text-sm font-medium text-[var(--color-text-primary)]" for="contest-starts-at">开始时间</label>
          <input
            id="contest-starts-at"
            v-model="localDraft.starts_at"
            type="datetime-local"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="fieldLocks.starts_at"
          >
          <p v-if="fieldErrors.starts_at" class="text-xs text-[var(--color-danger)]">{{ fieldErrors.starts_at }}</p>
          <p v-else-if="fieldLocks.starts_at" class="text-xs text-[var(--color-text-muted)]">报名中、进行中、已结束状态禁止修改开始时间。</p>
        </div>

        <div class="space-y-2">
          <label class="text-sm font-medium text-[var(--color-text-primary)]" for="contest-ends-at">结束时间</label>
          <input
            id="contest-ends-at"
            v-model="localDraft.ends_at"
            type="datetime-local"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="fieldLocks.ends_at"
          >
          <p v-if="fieldErrors.ends_at" class="text-xs text-[var(--color-danger)]">{{ fieldErrors.ends_at }}</p>
          <p v-else-if="fieldLocks.ends_at" class="text-xs text-[var(--color-text-muted)]">进行中、已结束状态禁止修改结束时间。</p>
        </div>
      </div>
    </form>

    <template #footer>
      <div class="flex items-center justify-between gap-3">
        <p class="text-xs text-[var(--color-text-muted)]">
          当前未接入删除接口。若需下线竞赛，请通过状态流转控制访问窗口。
        </p>
        <div class="flex items-center gap-2">
          <button
            type="button"
            class="rounded-xl border border-border px-4 py-2 text-sm text-[var(--color-text-primary)] transition hover:border-primary"
            @click="closeDialog"
          >
            取消
          </button>
          <button
            type="button"
            class="rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="saving"
            @click="handleSubmit"
          >
            {{ saving ? '保存中...' : '保存竞赛' }}
          </button>
        </div>
      </div>
    </template>
  </ElDialog>
</template>
