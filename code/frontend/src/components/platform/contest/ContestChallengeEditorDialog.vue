<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

import SlideOverDrawer from '@/components/common/modal-templates/SlideOverDrawer.vue'
import type {
  AdminAwdServiceTemplateData,
  AdminChallengeListItem,
  AdminContestChallengeViewData,
  ContestDetailData,
} from '@/api/contracts'

type DialogMode = 'create' | 'edit'

const props = defineProps<{
  open: boolean
  mode: DialogMode
  contestMode: ContestDetailData['mode']
  challengeOptions: AdminChallengeListItem[]
  templateOptions?: AdminAwdServiceTemplateData[]
  existingChallengeIds: string[]
  draft?: AdminContestChallengeViewData | null
  loadingChallengeCatalog: boolean
  loadingTemplateCatalog?: boolean
  saving: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  save: [
    value: {
      challenge_id?: number
      template_id?: number
      points: number
      order: number
      is_visible: boolean
    },
  ]
}>()

const form = reactive({
  challenge_id: '',
  template_id: '',
  points: '100',
  order: '0',
  is_visible: 'true',
})

const fieldErrors = reactive({
  challenge_id: '',
  template_id: '',
  points: '',
  order: '',
})

const dialogTitle = computed(() =>
  props.mode === 'create'
    ? isAwdContest.value
      ? '关联 AWD 题库题目'
      : '关联赛事题目'
    : '编辑赛事题目'
)

const selectableChallenges = computed(() =>
  props.challengeOptions.filter(
    (item) => props.mode === 'edit' || !props.existingChallengeIds.includes(item.id)
  )
)
const selectableTemplates = computed(() => props.templateOptions ?? [])

const isAwdContest = computed(() => props.contestMode === 'awd')
const selectedTemplate = computed<AdminAwdServiceTemplateData | null>(() =>
  selectableTemplates.value.find((item) => item.id === form.template_id) || null
)
const selectedTemplateSummary = computed(() => {
  if (!selectedTemplate.value) {
    return ''
  }
  return [
    selectedTemplate.value.category.toUpperCase(),
    selectedTemplate.value.difficulty,
    selectedTemplate.value.service_type,
    selectedTemplate.value.deployment_mode,
  ].join(' · ')
})
const selectedTemplateSections = computed(() => {
  if (!selectedTemplate.value) {
    return []
  }
  return [
    {
      key: 'flag',
      label: 'Flag',
      value: selectedTemplate.value.flag_config,
    },
    {
      key: 'access',
      label: '访问入口',
      value: selectedTemplate.value.access_config,
    },
    {
      key: 'runtime',
      label: '运行参数',
      value: selectedTemplate.value.runtime_config,
    },
  ]
})

watch(
  () => [props.open, props.mode, props.draft, selectableChallenges.value, selectableTemplates.value] as const,
  ([open]) => {
    if (!open) {
      return
    }
    form.challenge_id =
      props.mode === 'edit'
        ? props.draft?.challenge_id || ''
        : isAwdContest.value
          ? selectableTemplates.value[0]?.id || ''
        : selectableChallenges.value[0]?.id || ''
    form.template_id = isAwdContest.value
      ? props.draft?.awd_template_id || selectableTemplates.value[0]?.id || ''
      : ''
    form.points = String(props.draft?.points ?? 100)
    form.order = String(props.draft?.order ?? 0)
    form.is_visible = props.draft?.is_visible === false ? 'false' : 'true'
    clearErrors()
  },
  { immediate: true, deep: true }
)

function clearErrors() {
  fieldErrors.challenge_id = ''
  fieldErrors.template_id = ''
  fieldErrors.points = ''
  fieldErrors.order = ''
}

function closeDialog() {
  emit('update:open', false)
}

function formatTemplateJSON(value?: Record<string, unknown>): string {
  if (!value || Object.keys(value).length === 0) {
    return '{}'
  }
  return JSON.stringify(value, null, 2)
}

function submit() {
  clearErrors()

  if (!isAwdContest.value && !form.challenge_id.trim()) {
    fieldErrors.challenge_id = '请选择题目'
  }
  if (isAwdContest.value && !form.template_id.trim()) {
    fieldErrors.template_id = '请选择服务模板'
  }

  const points = Number(form.points)
  if (!Number.isFinite(points) || points < 1) {
    fieldErrors.points = '分值至少为 1'
  }

  const order = Number(form.order)
  if (!Number.isFinite(order) || order < 0) {
    fieldErrors.order = '顺序不能小于 0'
  }

  if (
    fieldErrors.challenge_id ||
    fieldErrors.template_id ||
    fieldErrors.points ||
    fieldErrors.order
  ) {
    return
  }

  emit('save', {
    challenge_id: isAwdContest.value ? undefined : Number(form.challenge_id),
    template_id: isAwdContest.value ? Number(form.template_id) : undefined,
    points,
    order,
    is_visible: form.is_visible === 'true',
  })
}
</script>

<template>
  <SlideOverDrawer
    :open="open"
    :title="dialogTitle"
    :subtitle="
      isAwdContest
        ? '在题目池里从 AWD 题库模板选题，先继承模板里的入口与攻防定义，再编排顺序、分值和可见性。'
        : '维护赛事题目的关联关系、顺序、分值和可见性。'
    "
    eyebrow="Contest Orchestration"
    width="42rem"
    @close="closeDialog"
    @update:open="emit('update:open', $event)"
  >
    <form
      class="contest-challenge-dialog"
      @submit.prevent="submit"
    >
      <p
        v-if="isAwdContest"
        class="contest-challenge-dialog__hint"
      >
        这里先从 AWD 题库模板选题并继承端口、入口和 flag 策略；Checker 与预检细节继续在 AWD 配置页补充。
      </p>

      <label
        v-if="!isAwdContest"
        class="ui-field contest-challenge-dialog__field"
        for="contest-challenge-select"
      >
        <span class="ui-field__label contest-challenge-dialog__label">题目</span>
        <template v-if="mode === 'create'">
          <span
            class="ui-control-wrap"
            :class="{ 'is-disabled': loadingChallengeCatalog || selectableChallenges.length === 0, 'is-error': !!fieldErrors.challenge_id }"
          >
            <select
              id="contest-challenge-select"
              v-model="form.challenge_id"
              class="ui-control contest-challenge-dialog__control"
              :disabled="loadingChallengeCatalog || selectableChallenges.length === 0"
            >
              <option
                value=""
                disabled
              >
                {{ loadingChallengeCatalog ? '正在加载题目目录...' : '请选择题目' }}
              </option>
              <option
                v-for="challenge in selectableChallenges"
                :key="challenge.id"
                :value="challenge.id"
              >
                {{ challenge.title }}
              </option>
            </select>
          </span>
        </template>
        <template v-else>
          <span class="ui-control-wrap contest-challenge-dialog__readonly">
            <span class="ui-control contest-challenge-dialog__control">
              {{ draft?.title || `Challenge #${draft?.challenge_id || ''}` }}
            </span>
          </span>
        </template>
        <span
          v-if="fieldErrors.challenge_id"
          class="ui-field__error contest-challenge-dialog__error"
        >
          {{ fieldErrors.challenge_id }}
        </span>
      </label>

      <label
        v-if="isAwdContest"
        class="ui-field contest-challenge-dialog__field"
        for="contest-challenge-template"
      >
        <span class="ui-field__label contest-challenge-dialog__label">AWD 题库模板</span>
        <span
          class="ui-control-wrap"
          :class="{
            'is-disabled': loadingTemplateCatalog || selectableTemplates.length === 0,
            'is-error': !!fieldErrors.template_id,
          }"
        >
          <select
            id="contest-challenge-template"
            v-model="form.template_id"
            class="ui-control contest-challenge-dialog__control"
            :disabled="loadingTemplateCatalog || selectableTemplates.length === 0"
          >
            <option
              value=""
              disabled
            >
              {{ loadingTemplateCatalog ? '正在加载 AWD 题库模板...' : '请选择 AWD 题库模板' }}
            </option>
            <option
              v-for="template in selectableTemplates"
              :key="template.id"
              :value="template.id"
            >
              {{ template.name }}
            </option>
          </select>
        </span>
        <span
          v-if="fieldErrors.template_id"
          class="ui-field__error contest-challenge-dialog__error"
        >
          {{ fieldErrors.template_id }}
        </span>
      </label>

      <section
        v-if="isAwdContest && selectedTemplate"
        class="contest-challenge-dialog__template-card"
      >
        <header class="contest-challenge-dialog__template-head">
          <div>
            <div class="journal-note-label">
              Template Snapshot
            </div>
            <h3 class="contest-challenge-dialog__template-title">
              题库模板快照
            </h3>
          </div>
          <p class="contest-challenge-dialog__template-copy">
            这部分内容来自 AWD 题库定义，比赛里的服务会以它为基础生成。
          </p>
        </header>

        <div class="contest-challenge-dialog__template-chips">
          <div class="contest-challenge-dialog__template-chip">
            <span class="contest-challenge-dialog__template-chip-label">模板</span>
            <span class="contest-challenge-dialog__template-chip-value">{{ selectedTemplate.name }}</span>
          </div>
          <div class="contest-challenge-dialog__template-chip">
            <span class="contest-challenge-dialog__template-chip-label">摘要</span>
            <span class="contest-challenge-dialog__template-chip-value">{{ selectedTemplateSummary }}</span>
          </div>
          <div class="contest-challenge-dialog__template-chip">
            <span class="contest-challenge-dialog__template-chip-label">防守入口</span>
            <span class="contest-challenge-dialog__template-chip-value">{{ selectedTemplate.defense_entry_mode || '未定义' }}</span>
          </div>
          <div class="contest-challenge-dialog__template-chip">
            <span class="contest-challenge-dialog__template-chip-label">Flag 模式</span>
            <span class="contest-challenge-dialog__template-chip-value">{{ selectedTemplate.flag_mode || '未定义' }}</span>
          </div>
        </div>

        <div class="contest-challenge-dialog__template-grid">
          <article
            v-for="section in selectedTemplateSections"
            :key="section.key"
            class="contest-challenge-dialog__template-panel"
          >
            <h4 class="contest-challenge-dialog__template-panel-title">
              {{ section.label }}
            </h4>
            <pre class="contest-challenge-dialog__template-json">{{ formatTemplateJSON(section.value) }}</pre>
          </article>
        </div>
      </section>

      <div class="contest-challenge-dialog__grid">
        <label
          class="ui-field contest-challenge-dialog__field"
          for="contest-challenge-points"
        >
          <span class="ui-field__label contest-challenge-dialog__label">分值</span>
          <span
            class="ui-control-wrap"
            :class="{ 'is-error': !!fieldErrors.points }"
          >
            <input
              id="contest-challenge-points"
              v-model="form.points"
              type="number"
              min="1"
              step="1"
              class="ui-control contest-challenge-dialog__control"
            >
          </span>
          <span
            v-if="fieldErrors.points"
            class="ui-field__error contest-challenge-dialog__error"
          >
            {{ fieldErrors.points }}
          </span>
        </label>

        <label
          class="ui-field contest-challenge-dialog__field"
          for="contest-challenge-order"
        >
          <span class="ui-field__label contest-challenge-dialog__label">顺序</span>
          <span
            class="ui-control-wrap"
            :class="{ 'is-error': !!fieldErrors.order }"
          >
            <input
              id="contest-challenge-order"
              v-model="form.order"
              type="number"
              min="0"
              step="1"
              class="ui-control contest-challenge-dialog__control"
            >
          </span>
          <span
            v-if="fieldErrors.order"
            class="ui-field__error contest-challenge-dialog__error"
          >
            {{ fieldErrors.order }}
          </span>
        </label>
      </div>

      <label
        class="ui-field contest-challenge-dialog__field"
        for="contest-challenge-visibility"
      >
        <span class="ui-field__label contest-challenge-dialog__label">可见性</span>
        <span class="ui-control-wrap">
          <select
            id="contest-challenge-visibility"
            v-model="form.is_visible"
            class="ui-control contest-challenge-dialog__control"
          >
            <option value="true">可见</option>
            <option value="false">隐藏</option>
          </select>
        </span>
      </label>
    </form>

    <template #footer>
      <div class="contest-challenge-dialog__footer">
        <button
          type="button"
          class="ui-btn ui-btn--secondary contest-challenge-dialog__button"
          @click="closeDialog"
        >
          取消
        </button>
        <button
          id="contest-challenge-dialog-submit"
          type="button"
          class="ui-btn ui-btn--primary contest-challenge-dialog__button"
          :disabled="saving"
          @click="submit"
        >
          {{ saving ? '保存中...' : mode === 'create' ? '关联题目' : '保存变更' }}
        </button>
      </div>
    </template>
  </SlideOverDrawer>
</template>

<style scoped>
.contest-challenge-dialog {
  display: grid;
  gap: var(--space-4);
}

.contest-challenge-dialog__grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.contest-challenge-dialog__field {
  --ui-field-gap: var(--space-2);
}

.contest-challenge-dialog__label {
  font-size: var(--font-size-0-875);
}

.contest-challenge-dialog__control,
.contest-challenge-dialog__readonly {
  min-height: 2.75rem;
}

.contest-challenge-dialog__readonly {
  background: color-mix(in srgb, var(--journal-surface) 96%, transparent);
}

.contest-challenge-dialog__hint {
  margin: 0;
  font-size: var(--font-size-0-875);
  color: var(--journal-muted);
}

.contest-challenge-dialog__template-card {
  display: grid;
  gap: var(--space-4);
  padding: var(--space-4);
  border: 1px solid var(--color-border-default);
  border-radius: 1rem;
  background: var(--color-bg-surface);
}

.contest-challenge-dialog__template-head {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  gap: var(--space-3);
}

.contest-challenge-dialog__template-title {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-1-125);
  font-weight: 700;
  color: var(--color-text-primary);
}

.contest-challenge-dialog__template-copy {
  margin: 0;
  max-width: 22rem;
  font-size: var(--font-size-0-875);
  line-height: 1.6;
  color: var(--journal-muted);
}

.contest-challenge-dialog__template-chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
}

.contest-challenge-dialog__template-chip {
  display: grid;
  gap: 0.15rem;
  min-width: 10rem;
  padding: 0.8rem 0.9rem;
  border-radius: 0.85rem;
  border: 1px solid var(--color-border-subtle);
  background: var(--color-bg-elevated);
}

.contest-challenge-dialog__template-chip-label {
  font-size: var(--font-size-0-75);
  color: var(--journal-muted);
}

.contest-challenge-dialog__template-chip-value {
  font-weight: 600;
  color: var(--color-text-primary);
  word-break: break-word;
}

.contest-challenge-dialog__template-grid {
  display: grid;
  gap: var(--space-3);
}

.contest-challenge-dialog__template-panel {
  display: grid;
  gap: var(--space-2);
  padding: 0.85rem;
  border-radius: 0.85rem;
  border: 1px solid var(--color-border-subtle);
  background: var(--color-bg-elevated);
}

.contest-challenge-dialog__template-panel-title {
  margin: 0;
  font-size: var(--font-size-0-875);
  font-weight: 700;
  color: var(--color-text-primary);
}

.contest-challenge-dialog__template-json {
  margin: 0;
  font-family: var(--font-family-mono);
  font-size: var(--font-size-0-75);
  line-height: 1.6;
  color: var(--color-text-secondary);
  white-space: pre-wrap;
  word-break: break-word;
}

.contest-challenge-dialog__error {
  font-size: var(--font-size-0-75);
}

.contest-challenge-dialog__footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
}

.contest-challenge-dialog__button {
  min-width: 6rem;
}

@media (max-width: 767px) {
  .contest-challenge-dialog__grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .contest-challenge-dialog__footer {
    flex-direction: column-reverse;
  }
}
</style>
