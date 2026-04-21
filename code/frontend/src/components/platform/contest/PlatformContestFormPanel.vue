<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { 
  FileText, 
  Settings, 
  Clock, 
  Swords, 
  Trophy,
} from 'lucide-vue-next'

import type { ContestFieldLocks, ContestFormDraft } from '@/composables/usePlatformContests'
import { getStatusLabel } from '@/utils/contest'

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
  
  if (localDraft.starts_at && localDraft.ends_at && new Date(localDraft.ends_at) <= new Date(localDraft.starts_at)) {
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
    class="studio-settings-layout"
    @submit.prevent="handleSubmit"
  >
    <!-- Section: Identity -->
    <section class="settings-group">
      <div class="settings-group__info">
        <div class="info-icon">
          <FileText class="h-4 w-4" />
        </div>
        <h3 class="info-title">
          核心标识
        </h3>
        <p class="info-desc">
          定义竞赛在平台展示的基础信息与访问权限。
        </p>
      </div>
      
      <div class="settings-group__content">
        <div class="settings-row">
          <label class="row-label">竞赛标题</label>
          <div class="row-control">
            <div
              class="control-wrap"
              :class="{ 'is-error': !!fieldErrors.title }"
            >
              <input
                id="contest-title"
                v-model="localDraft.title"
                type="text"
                class="studio-input"
                placeholder="输入竞赛标题..."
              >
            </div>
            <p
              v-if="fieldErrors.title"
              class="field-error"
            >
              {{ fieldErrors.title }}
            </p>
            <p class="field-hint">
              请控制在 40 个字以内，建议包含年份与赛季信息。
            </p>
          </div>
        </div>

        <div class="settings-row">
          <label class="row-label">竞赛描述</label>
          <div class="row-control">
            <div class="control-wrap">
              <textarea
                id="contest-description"
                v-model="localDraft.description"
                rows="4"
                class="studio-textarea"
                placeholder="描述竞赛的背景、赛制及对参赛者的要求..."
              />
            </div>
            <p class="field-hint">
              支持 Markdown 语法，将展示在竞赛详情页。
            </p>
          </div>
        </div>
      </div>
    </section>

    <!-- Section: Configuration -->
    <section class="settings-group">
      <div class="settings-group__info">
        <div class="info-icon">
          <Settings class="h-4 w-4" />
        </div>
        <h3 class="info-title">
          赛制与状态
        </h3>
        <p class="info-desc">
          控制竞赛的底层逻辑模式与全平台生命周期。
        </p>
      </div>
      
      <div class="settings-group__content">
        <div class="settings-row">
          <label class="row-label">竞技模式</label>
          <div class="row-control">
            <div class="flex gap-4">
              <button 
                type="button" 
                class="mode-card" 
                :class="{ active: localDraft.mode === 'jeopardy', disabled: fieldLocks.mode }"
                @click="!fieldLocks.mode && (localDraft.mode = 'jeopardy')"
              >
                <Trophy class="h-5 w-5 mb-2" />
                <span class="mode-label">Jeopardy</span>
                <span class="mode-desc">经典夺旗解题赛</span>
              </button>
              <button 
                type="button" 
                class="mode-card" 
                :class="{ active: localDraft.mode === 'awd', disabled: fieldLocks.mode }"
                @click="!fieldLocks.mode && (localDraft.mode = 'awd')"
              >
                <Swords class="h-5 w-5 mb-2" />
                <span class="mode-label">AWD</span>
                <span class="mode-desc">实时攻防对抗赛</span>
              </button>
            </div>
            <p
              v-if="fieldLocks.mode"
              class="field-hint text-orange-500 mt-3 font-bold"
            >
              竞赛已生效，模式锁定不可更改。
            </p>
          </div>
        </div>

        <div
          v-if="mode === 'edit'"
          class="settings-row"
        >
          <label class="row-label">运行阶段</label>
          <div class="row-control">
            <div class="control-wrap">
              <select
                id="contest-status"
                v-model="localDraft.status"
                class="studio-select"
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
            <p class="field-hint">
              手动控制竞赛在前端的可见性与交互状态。
            </p>
          </div>
        </div>
      </div>
    </section>

    <!-- Section: Timeline -->
    <section class="settings-group">
      <div class="settings-group__info">
        <div class="info-icon">
          <Clock class="h-4 w-4" />
        </div>
        <h3 class="info-title">
          时间窗口
        </h3>
        <p class="info-desc">
          精确配置比赛的启停节点，系统将按此时钟自动调度。
        </p>
      </div>
      
      <div class="settings-group__content">
        <div class="settings-row">
          <label class="row-label">赛程时间轴</label>
          <div class="row-control">
            <div class="flex items-center gap-4">
              <div class="flex-1">
                <div
                  class="control-wrap"
                  :class="{ 'is-disabled': fieldLocks.starts_at }"
                >
                  <input
                    id="contest-starts-at"
                    v-model="localDraft.starts_at"
                    type="datetime-local"
                    class="studio-input"
                    :disabled="fieldLocks.starts_at"
                  >
                </div>
                <p class="field-hint mt-1">
                  开始时间
                </p>
              </div>
              <div class="text-slate-300">
                ——
              </div>
              <div class="flex-1">
                <div
                  class="control-wrap"
                  :class="{ 'is-disabled': fieldLocks.ends_at }"
                >
                  <input
                    id="contest-ends-at"
                    v-model="localDraft.ends_at"
                    type="datetime-local"
                    class="studio-input"
                    :disabled="fieldLocks.ends_at"
                  >
                </div>
                <p class="field-hint mt-1">
                  结束时间
                </p>
              </div>
            </div>
            <p
              v-if="fieldErrors.starts_at || fieldErrors.ends_at"
              class="field-error mt-2"
            >
              {{ fieldErrors.starts_at || fieldErrors.ends_at }}
            </p>
          </div>
        </div>
      </div>
    </section>

    <div class="contest-form-actions">
      <button
        v-if="showCancel"
        type="button"
        class="ui-btn ui-btn--ghost contest-form-button contest-form-button--secondary"
        @click="emit('cancel')"
      >
        取消
      </button>
      <button
        type="submit"
        class="ui-btn ui-btn--primary contest-form-button contest-form-button--primary"
        :disabled="saving"
      >
        {{ saving ? '保存中...' : mode === 'create' ? '创建竞赛' : '保存变更' }}
      </button>
    </div>
  </form>
</template>

<style scoped>
.studio-settings-layout {
  width: 100%;
  max-width: 64rem;
  margin: 0 auto;
  padding: 4rem 2rem;
  display: flex;
  flex-direction: column;
  gap: 5rem;
}

.settings-group {
  display: grid;
  grid-template-columns: 18rem 1fr;
  gap: 4rem;
}

.settings-group__info {
  display: flex;
  flex-direction: column;
}

.info-icon {
  width: 2.25rem; height: 2.25rem; border-radius: 0.75rem;
  background: #eff6ff; color: #3b82f6;
  display: flex; align-items: center; justify-content: center;
  margin-bottom: 1.25rem;
}

.info-title { font-size: 1rem; font-weight: 900; color: #0f172a; margin: 0; }
.info-desc { font-size: 13px; color: #64748b; line-height: 1.6; margin-top: 0.5rem; }

.settings-group__content {
  display: flex;
  flex-direction: column;
  gap: 2.5rem;
}

.settings-row {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.row-label { font-size: 13px; font-weight: 800; color: #1e293b; }
.row-control { width: 100%; max-width: 36rem; }

.control-wrap {
  width: 100%; border-radius: 0.6rem; border: 1px solid #e2e8f0;
  background: white; transition: all 0.2s ease;
  overflow: hidden;
}
.control-wrap:focus-within { border-color: #3b82f6; box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.06); }
.control-wrap.is-error { border-color: #ef4444; }
.control-wrap.is-disabled { background: #f8fafc; opacity: 0.7; cursor: not-allowed; }

.studio-input, .studio-select, .studio-textarea {
  width: 100%; border: none; background: transparent; padding: 0.65rem 0.85rem;
  font-size: 14px; font-weight: 600; color: #1e293b; outline: none;
}
.studio-textarea { min-height: 7rem; resize: vertical; line-height: 1.6; }

.field-hint { font-size: 12px; color: #94a3b8; margin-top: 0.45rem; font-weight: 500; }
.field-error { font-size: 12px; color: #ef4444; font-weight: 700; }

.contest-form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: -1rem;
}

/* Mode Cards */
.mode-card {
  flex: 1; padding: 1.5rem; border-radius: 1rem; border: 1px solid #e2e8f0;
  background: white; display: flex; flex-direction: column; align-items: center;
  cursor: pointer; transition: all 0.2s ease; color: #64748b;
}
.mode-card:hover:not(.disabled) { border-color: #3b82f6; background: #f8fafc; }
.mode-card.active { border-color: #3b82f6; background: #eff6ff; color: #2563eb; box-shadow: 0 4px 12px rgba(37,99,235,0.1); }
.mode-card.active .mode-label { color: #1e293b; }
.mode-card.disabled { opacity: 0.6; cursor: not-allowed; }

.mode-label { font-size: 14px; font-weight: 900; margin-bottom: 0.25rem; }
.mode-desc { font-size: 11px; opacity: 0.8; }

@media (max-width: 1024px) {
  .settings-group { grid-template-columns: 1fr; gap: 1.5rem; }
  .studio-settings-layout { padding: 2rem 1.5rem; gap: 3rem; }
}
</style>
