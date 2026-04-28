<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

import AdminSurfaceModal from '@/components/common/modal-templates/AdminSurfaceModal.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
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
      ? '关联 AWD 题库服务'
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
const isAwdCreateMode = computed(() => isAwdContest.value && props.mode === 'create')
const dialogWidth = computed(() =>
  isAwdCreateMode.value ? 'min(60rem, calc(100vw - (var(--space-4) * 2)))' : '42rem'
)
const showContestSelector = computed(() => !isAwdContest.value || props.mode === 'edit')
const showContestSettings = computed(() => !isAwdCreateMode.value)
const templateTableColumns = [
  {
    key: 'name',
    label: '名称',
    widthClass: 'w-[30%] min-w-[14rem]',
    cellClass: 'contest-template-table__name-cell',
  },
  {
    key: 'category',
    label: '分类',
    align: 'center' as const,
    widthClass: 'w-[12%] min-w-[6rem]',
  },
  {
    key: 'difficulty',
    label: '难度',
    align: 'center' as const,
    widthClass: 'w-[12%] min-w-[6rem]',
  },
  {
    key: 'service_type',
    label: '服务类型',
    align: 'center' as const,
    widthClass: 'w-[16%] min-w-[8rem]',
  },
  {
    key: 'deployment_mode',
    label: '部署方式',
    align: 'center' as const,
    widthClass: 'w-[16%] min-w-[8rem]',
  },
  {
    key: 'actions',
    label: '选择',
    align: 'right' as const,
    widthClass: 'w-[7rem]',
    cellClass: 'contest-template-table__actions-cell',
  },
]

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
          ? ''
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

function selectTemplate(templateId: string) {
  form.template_id = templateId
}

function getServiceTypeLabel(value: AdminAwdServiceTemplateData['service_type']): string {
  switch (value) {
    case 'binary_tcp':
      return 'Binary TCP'
    case 'multi_container':
      return 'Multi Container'
    case 'web_http':
    default:
      return 'Web HTTP'
  }
}

function getDeploymentModeLabel(value: AdminAwdServiceTemplateData['deployment_mode']): string {
  switch (value) {
    case 'topology':
      return '拓扑'
    case 'single_container':
    default:
      return '单容器'
  }
}

function submit() {
  if (props.saving) {
    return
  }

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
    challenge_id: isAwdContest.value
      ? undefined
      : form.challenge_id
        ? Number(form.challenge_id)
        : undefined,
    template_id: isAwdContest.value ? Number(form.template_id) : undefined,
    points,
    order,
    is_visible: form.is_visible === 'true',
  })
}
</script>

<template>
  <AdminSurfaceModal
    :open="open"
    :title="dialogTitle"
    :subtitle="
      isAwdContest
        ? '从 AWD 题库选择服务模板。'
        : '维护赛事题目的关联关系、顺序、分值和可见性。'
    "
    eyebrow="Contest Orchestration"
    :width="dialogWidth"
    @close="closeDialog"
    @update:open="emit('update:open', $event)"
  >
    <form
      class="contest-challenge-dialog"
      @submit.prevent="submit"
    >
      <label
        v-if="showContestSelector"
        class="ui-field contest-challenge-dialog__field"
        for="contest-challenge-select"
      >
        <span class="ui-field__label contest-challenge-dialog__label">{{ isAwdContest ? 'AWD 服务' : '题目' }}</span>
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

      <section
        v-if="isAwdContest"
        class="contest-template-list"
        :class="{ 'is-error': !!fieldErrors.template_id }"
      >
        <div class="contest-template-list__head">
          <span class="ui-field__label contest-challenge-dialog__label">AWD 题库模板</span>
          <span class="contest-template-list__count">
            {{ loadingTemplateCatalog ? '加载中' : `${selectableTemplates.length} 个可选` }}
          </span>
        </div>
        <div
          v-if="selectableTemplates.length > 0"
          class="contest-template-list__table workspace-directory-list"
        >
          <WorkspaceDataTable
            :columns="templateTableColumns"
            :rows="selectableTemplates"
            row-key="id"
            row-class="contest-template-table-row"
          >
            <template #cell-name="{ row }">
              <span class="contest-template-table__name">
                {{ (row as AdminAwdServiceTemplateData).name }}
              </span>
            </template>
            <template #cell-service_type="{ row }">
              <span class="contest-template-table__mono">
                {{ getServiceTypeLabel((row as AdminAwdServiceTemplateData).service_type) }}
              </span>
            </template>
            <template #cell-deployment_mode="{ row }">
              <span class="contest-template-table__text">
                {{ getDeploymentModeLabel((row as AdminAwdServiceTemplateData).deployment_mode) }}
              </span>
            </template>
            <template #cell-actions="{ row }">
              <button
                :id="`contest-template-option-${(row as AdminAwdServiceTemplateData).id}`"
                type="button"
                class="contest-template-option"
                :class="{ 'is-selected': form.template_id === (row as AdminAwdServiceTemplateData).id }"
                :aria-pressed="form.template_id === (row as AdminAwdServiceTemplateData).id"
                @click="selectTemplate((row as AdminAwdServiceTemplateData).id)"
              >
                {{ form.template_id === (row as AdminAwdServiceTemplateData).id ? '已选择' : '选择' }}
              </button>
            </template>
          </WorkspaceDataTable>
        </div>
        <div
          v-else
          class="contest-template-list__empty"
        >
          {{ loadingTemplateCatalog ? '正在加载 AWD 题库模板...' : '暂无可选 AWD 题库模板' }}
        </div>
        <span
          v-if="fieldErrors.template_id"
          class="ui-field__error contest-challenge-dialog__error"
        >
          {{ fieldErrors.template_id }}
        </span>
      </section>

      <div
        v-if="showContestSettings"
        class="contest-challenge-dialog__grid"
      >
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
        v-if="showContestSettings"
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
          {{ saving ? '保存中...' : mode === 'create' ? (isAwdContest ? '关联服务' : '关联题目') : '保存变更' }}
        </button>
      </div>
    </template>
  </AdminSurfaceModal>
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

.contest-template-list {
  display: grid;
  gap: var(--space-3);
}

.contest-template-list__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.contest-template-list__count {
  font-size: var(--font-size-0-75);
  color: var(--journal-muted);
}

.contest-template-list__empty {
  border: 1px solid var(--color-border-default);
  border-radius: var(--ui-control-radius);
  background: var(--color-bg-surface);
  padding: var(--space-4);
  color: var(--journal-muted);
  font-size: var(--font-size-0-875);
}

.contest-template-list__table {
  max-height: clamp(12rem, calc(100dvh - 18rem), 30rem);
  overflow: auto;
}

.contest-template-list__table :deep(.workspace-data-table) {
  min-width: 48rem;
}

.contest-template-list__table :deep(.workspace-data-table__head-cell) {
  position: sticky;
  top: 0;
  z-index: 1;
  background: var(--color-bg-surface);
}

.contest-template-table__name {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-0-875);
  font-weight: 800;
  color: var(--color-text-primary);
}

.contest-template-table__mono,
.contest-template-table__text {
  font-size: var(--font-size-0-75);
  font-weight: 700;
  color: var(--color-text-secondary);
}

.contest-template-table__mono {
  font-family: var(--font-family-mono);
}

.contest-template-option {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: var(--ui-control-height-sm);
  min-width: 4.5rem;
  border: 1px solid var(--color-border-default);
  border-radius: var(--ui-control-radius-md);
  background: var(--color-bg-surface);
  padding: 0 var(--space-3);
  font-size: var(--font-size-0-75);
  font-weight: 700;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition:
    background var(--ui-motion-fast),
    border-color var(--ui-motion-fast),
    color var(--ui-motion-fast);
}

.contest-template-option:hover,
.contest-template-option:focus-visible {
  border-color: color-mix(in srgb, var(--color-primary) 62%, var(--color-border-default));
  background: color-mix(in srgb, var(--color-primary) 8%, var(--color-bg-surface));
  color: var(--color-primary);
}

.contest-template-option.is-selected {
  border-color: color-mix(in srgb, var(--color-primary) 70%, var(--color-border-default));
  background: var(--color-primary-soft);
  color: var(--color-primary);
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
