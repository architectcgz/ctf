<script setup lang="ts">
import './platformContestFormPanel.css'

import { reactive, ref, watch } from 'vue'

import type { ContestFieldLocks, ContestFormDraft } from '../model'
import PlatformContestFormActions from './PlatformContestFormActions.vue'
import PlatformContestIdentitySection from './PlatformContestIdentitySection.vue'
import PlatformContestRulesSection from './PlatformContestRulesSection.vue'
import PlatformContestTimelineSection from './PlatformContestTimelineSection.vue'

const props = withDefaults(
  defineProps<{
    mode: 'create' | 'edit'
    draft: ContestFormDraft
    saving: boolean
    statusOptions?: Array<{ label: string; value: ContestFormDraft['status'] }>
    fieldLocks: ContestFieldLocks
    showCancel?: boolean
    note?: string
  }>(),
  {
    statusOptions: () => [],
    showCancel: true,
    note: '',
  }
)

const emit = defineEmits<{
  cancel: []
  save: [value: ContestFormDraft]
  'update:draft': [value: ContestFormDraft]
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
const syncingFromProps = ref(false)

watch(
  () => props.draft,
  (draft) => {
    syncingFromProps.value = true
    Object.assign(localDraft, draft)
    fieldErrors.title = ''
    fieldErrors.starts_at = ''
    fieldErrors.ends_at = ''
    syncingFromProps.value = false
  },
  { immediate: true, deep: true }
)

watch(
  localDraft,
  (draft) => {
    if (syncingFromProps.value) return
    emit('update:draft', { ...draft })
  },
  { deep: true }
)

function validate(): boolean {
  fieldErrors.title = ''
  fieldErrors.starts_at = ''
  fieldErrors.ends_at = ''

  if (!localDraft.title.trim()) fieldErrors.title = '请填写竞赛标题'
  if (!localDraft.starts_at) fieldErrors.starts_at = '请填写开始时间'
  if (!localDraft.ends_at) fieldErrors.ends_at = '请填写结束时间'

  if (
    localDraft.starts_at &&
    localDraft.ends_at &&
    new Date(localDraft.ends_at) <= new Date(localDraft.starts_at)
  ) {
    fieldErrors.ends_at = '结束时间必须晚于开始时间'
  }
  return !fieldErrors.title && !fieldErrors.starts_at && !fieldErrors.ends_at
}

function handleSubmit() {
  if (!validate()) return
  emit('save', { ...localDraft })
}
</script>

<template>
  <form
    class="contest-form-layout"
    @submit.prevent="handleSubmit"
  >
    <PlatformContestIdentitySection
      class="contest-form-section--identity"
      :title-value="localDraft.title"
      :description-value="localDraft.description"
      :title-error="fieldErrors.title ?? ''"
      @update:title="localDraft.title = $event"
      @update:description="localDraft.description = $event"
    />

    <PlatformContestRulesSection
      class="contest-form-section--rules"
      :mode="mode"
      :contest-mode="localDraft.mode"
      :contest-status="localDraft.status"
      :status-options="statusOptions"
      :field-locks="fieldLocks"
      @update:mode="localDraft.mode = $event"
      @update:status="localDraft.status = $event"
    />

    <PlatformContestTimelineSection
      class="contest-form-section--timeline"
      :starts-at="localDraft.starts_at"
      :ends-at="localDraft.ends_at"
      :field-locks="fieldLocks"
      :starts-at-error="fieldErrors.starts_at ?? ''"
      :ends-at-error="fieldErrors.ends_at ?? ''"
      @update:starts-at="localDraft.starts_at = $event"
      @update:ends-at="localDraft.ends_at = $event"
    />

    <PlatformContestFormActions
      :mode="mode"
      :saving="saving"
      :show-cancel="showCancel"
      @cancel="emit('cancel')"
    />
  </form>
</template>
