<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

import AdminSurfaceModal from '@/components/common/modal-templates/AdminSurfaceModal.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import type {
  AdminAwdChallengeData,
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
  awdChallengeOptions?: AdminAwdChallengeData[]
  awdChallengePage?: number
  awdChallengePageSize?: number
  awdChallengeTotal?: number
  awdChallengeKeyword?: string
  awdChallengeServiceType?: AdminAwdChallengeData['service_type'] | ''
  awdChallengeDeploymentMode?: AdminAwdChallengeData['deployment_mode'] | ''
  awdChallengeReadiness?: AdminAwdChallengeData['readiness_status'] | ''
  awdChallengeLoadError?: string
  existingChallengeIds: string[]
  draft?: AdminContestChallengeViewData | null
  loadingChallengeCatalog: boolean
  loadingAwdChallengeCatalog?: boolean
  saving: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  save: [
    value: {
      challenge_id?: number
      awd_challenge_id?: number
      awd_challenge_ids?: number[]
      points: number
      order: number
      is_visible: boolean
    },
  ]
  'update-awd-challenge-keyword': [value: string]
  'update-awd-challenge-service-type': [value: AdminAwdChallengeData['service_type'] | '']
  'update-awd-challenge-deployment-mode': [value: AdminAwdChallengeData['deployment_mode'] | '']
  'update-awd-challenge-readiness': [value: AdminAwdChallengeData['readiness_status'] | '']
  'change-awd-challenge-page': [page: number]
  'refresh-awd-challenge-catalog': []
}>()

const form = reactive({
  challenge_id: '',
  awd_challenge_id: '',
  awd_challenge_ids: [] as string[],
  points: '100',
  order: '0',
  is_visible: 'true',
})

const fieldErrors = reactive({
  challenge_id: '',
  awd_challenge_id: '',
  points: '',
  order: '',
})

const dialogTitle = computed(() =>
  props.mode === 'create'
    ? isAwdContest.value
      ? '关联 AWD 题目'
      : '关联赛事题目'
    : '编辑赛事题目'
)

const selectableChallenges = computed(() =>
  props.challengeOptions.filter(
    (item) => props.mode === 'edit' || !props.existingChallengeIds.includes(item.id)
  )
)
const selectableAwdChallenges = computed(() =>
  (props.awdChallengeOptions ?? []).filter(
    (item) => props.mode === 'edit' || !props.existingChallengeIds.includes(item.id)
  )
)

const isAwdContest = computed(() => props.contestMode === 'awd')
const isAwdCreateMode = computed(() => isAwdContest.value && props.mode === 'create')
const dialogWidth = computed(() =>
  isAwdCreateMode.value ? 'min(60rem, calc(100vw - (var(--space-4) * 2)))' : '42rem'
)
const showContestSelector = computed(() => !isAwdContest.value || props.mode === 'edit')
const showContestSettings = computed(() => !isAwdCreateMode.value)
const showAwdChallengeSelector = computed(() => isAwdCreateMode.value)
const awdChallengePage = computed(() => props.awdChallengePage ?? 1)
const awdChallengePageSize = computed(() => props.awdChallengePageSize ?? 20)
const awdChallengeTotal = computed(() => props.awdChallengeTotal ?? selectableAwdChallenges.value.length)
const awdChallengeTotalPages = computed(() =>
  Math.max(1, Math.ceil(awdChallengeTotal.value / awdChallengePageSize.value))
)
const hasAwdChallengeFilters = computed(() =>
  Boolean(
    (props.awdChallengeKeyword ?? '').trim() ||
      props.awdChallengeServiceType ||
      props.awdChallengeDeploymentMode ||
      props.awdChallengeReadiness
  )
)
const canGoToPreviousAwdChallengePage = computed(() => awdChallengePage.value > 1)
const canGoToNextAwdChallengePage = computed(() => awdChallengePage.value < awdChallengeTotalPages.value)
const awdChallengeTableColumns = [
  {
    key: 'name',
    label: '名称',
    widthClass: 'w-[22%] min-w-[13rem]',
    cellClass: 'contest-awd-challenge-table__name-cell',
  },
  {
    key: 'slug',
    label: '标识',
    widthClass: 'w-[12%] min-w-[8rem]',
  },
  {
    key: 'category',
    label: '分类',
    align: 'center' as const,
    widthClass: 'w-[10%] min-w-[6rem]',
  },
  {
    key: 'difficulty',
    label: '难度',
    align: 'center' as const,
    widthClass: 'w-[10%] min-w-[6rem]',
  },
  {
    key: 'service_type',
    label: '服务类型',
    align: 'center' as const,
    widthClass: 'w-[14%] min-w-[8rem]',
  },
  {
    key: 'deployment_mode',
    label: '部署方式',
    align: 'center' as const,
    widthClass: 'w-[14%] min-w-[8rem]',
  },
  {
    key: 'readiness_status',
    label: '就绪状态',
    align: 'center' as const,
    widthClass: 'w-[12%] min-w-[7rem]',
  },
  {
    key: 'last_verified_at',
    label: '最近验证',
    align: 'center' as const,
    widthClass: 'w-[13%] min-w-[8rem]',
  },
  {
    key: 'actions',
    label: '选择',
    align: 'right' as const,
    widthClass: 'w-[7rem]',
    cellClass: 'contest-awd-challenge-table__actions-cell',
  },
]

watch(
  () => [props.open, props.mode, props.draft, selectableChallenges.value, selectableAwdChallenges.value] as const,
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
    form.awd_challenge_id = isAwdContest.value
      ? props.draft?.awd_challenge_id || selectableAwdChallenges.value[0]?.id || ''
      : ''
    form.awd_challenge_ids = isAwdCreateMode.value && form.awd_challenge_id ? [form.awd_challenge_id] : []
    form.points = String(props.draft?.points ?? 100)
    form.order = String(props.draft?.order ?? 0)
    form.is_visible = props.draft?.is_visible === false ? 'false' : 'true'
    clearErrors()
  },
  { immediate: true, deep: true }
)

function clearErrors() {
  fieldErrors.challenge_id = ''
  fieldErrors.awd_challenge_id = ''
  fieldErrors.points = ''
  fieldErrors.order = ''
}

function closeDialog() {
  emit('update:open', false)
}

function selectAwdChallenge(awdChallengeId: string) {
  if (isAwdCreateMode.value) {
    const selected = new Set(form.awd_challenge_ids)
    if (selected.has(awdChallengeId)) {
      if (selected.size > 1) selected.delete(awdChallengeId)
    } else {
      selected.add(awdChallengeId)
    }
    form.awd_challenge_ids = selectableAwdChallenges.value
      .map((item) => item.id)
      .filter((id) => selected.has(id))
    form.awd_challenge_id = form.awd_challenge_ids[0] || ''
    return
  }

  form.awd_challenge_id = awdChallengeId
}

function isAwdChallengeSelected(awdChallengeId: string): boolean {
  return isAwdCreateMode.value
    ? form.awd_challenge_ids.includes(awdChallengeId)
    : form.awd_challenge_id === awdChallengeId
}

function getServiceTypeLabel(value: AdminAwdChallengeData['service_type']): string {
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

function getDeploymentModeLabel(value: AdminAwdChallengeData['deployment_mode']): string {
  switch (value) {
    case 'topology':
      return '拓扑'
    case 'single_container':
    default:
      return '单容器'
  }
}

function getReadinessLabel(value?: AdminAwdChallengeData['readiness_status']): string {
  switch (value) {
    case 'passed':
      return '已就绪'
    case 'failed':
      return '未通过'
    case 'pending':
    default:
      return '待验证'
  }
}

function emitAwdChallengeKeyword(value: string) {
  emit('update-awd-challenge-keyword', value)
}

function emitAwdChallengeServiceType(value: string) {
  emit('update-awd-challenge-service-type', value as AdminAwdChallengeData['service_type'] | '')
}

function emitAwdChallengeDeploymentMode(value: string) {
  emit('update-awd-challenge-deployment-mode', value as AdminAwdChallengeData['deployment_mode'] | '')
}

function emitAwdChallengeReadiness(value: string) {
  emit('update-awd-challenge-readiness', value as AdminAwdChallengeData['readiness_status'] | '')
}

function changeAwdChallengePage(nextPage: number) {
  if (nextPage < 1 || nextPage > awdChallengeTotalPages.value || nextPage === awdChallengePage.value) {
    return
  }
  emit('change-awd-challenge-page', nextPage)
}

function formatLastVerifiedAt(value?: string): string {
  if (!value) {
    return '未验证'
  }

  return new Date(value).toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function submit() {
  if (props.saving) {
    return
  }

  clearErrors()

  if (!isAwdContest.value && !form.challenge_id.trim()) {
    fieldErrors.challenge_id = '请选择题目'
  }
  if (
    isAwdContest.value &&
    (isAwdCreateMode.value ? form.awd_challenge_ids.length === 0 : !form.awd_challenge_id.trim())
  ) {
    fieldErrors.awd_challenge_id = '请选择 AWD 题目'
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
    fieldErrors.awd_challenge_id ||
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
    awd_challenge_id: isAwdContest.value ? Number(form.awd_challenge_id) : undefined,
    awd_challenge_ids: isAwdCreateMode.value
      ? form.awd_challenge_ids.map((id) => Number(id))
      : undefined,
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
          ? '从 AWD 题库选择题目'
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
        v-if="showAwdChallengeSelector"
        id="contest-awd-challenge-list"
        class="contest-awd-challenge-list"
        :class="{ 'is-error': !!fieldErrors.awd_challenge_id }"
      >
        <div class="contest-awd-challenge-list__head">
          <span class="ui-field__label contest-challenge-dialog__label">AWD 题目</span>
          <span class="contest-awd-challenge-list__count">
            {{ loadingAwdChallengeCatalog ? '加载中' : `第 ${awdChallengePage} 页 / 共 ${awdChallengeTotalPages} 页` }}
          </span>
        </div>
        <div class="contest-awd-challenge-list__filters">
          <label class="ui-field contest-challenge-dialog__field">
            <span class="ui-field__label contest-challenge-dialog__label">关键词</span>
            <span class="ui-control-wrap">
              <input
                id="contest-awd-challenge-keyword"
                :value="awdChallengeKeyword ?? ''"
                type="text"
                class="ui-control contest-challenge-dialog__control"
                placeholder="搜索名称或 slug"
                @input="emitAwdChallengeKeyword(($event.target as HTMLInputElement).value)"
              >
            </span>
          </label>
          <label class="ui-field contest-challenge-dialog__field">
            <span class="ui-field__label contest-challenge-dialog__label">服务类型</span>
            <span class="ui-control-wrap">
              <select
                id="contest-awd-challenge-service-type"
                :value="awdChallengeServiceType ?? ''"
                class="ui-control contest-challenge-dialog__control"
                @change="emitAwdChallengeServiceType(($event.target as HTMLSelectElement).value)"
              >
                <option value="">全部</option>
                <option value="web_http">Web HTTP</option>
                <option value="binary_tcp">Binary TCP</option>
                <option value="multi_container">Multi Container</option>
              </select>
            </span>
          </label>
          <label class="ui-field contest-challenge-dialog__field">
            <span class="ui-field__label contest-challenge-dialog__label">部署方式</span>
            <span class="ui-control-wrap">
              <select
                id="contest-awd-challenge-deployment-mode"
                :value="awdChallengeDeploymentMode ?? ''"
                class="ui-control contest-challenge-dialog__control"
                @change="emitAwdChallengeDeploymentMode(($event.target as HTMLSelectElement).value)"
              >
                <option value="">全部</option>
                <option value="single_container">单容器</option>
                <option value="topology">拓扑</option>
              </select>
            </span>
          </label>
          <label class="ui-field contest-challenge-dialog__field">
            <span class="ui-field__label contest-challenge-dialog__label">就绪状态</span>
            <span class="ui-control-wrap">
              <select
                id="contest-awd-challenge-readiness"
                :value="awdChallengeReadiness ?? ''"
                class="ui-control contest-challenge-dialog__control"
                @change="emitAwdChallengeReadiness(($event.target as HTMLSelectElement).value)"
              >
                <option value="">全部</option>
                <option value="passed">已就绪</option>
                <option value="pending">待验证</option>
                <option value="failed">未通过</option>
              </select>
            </span>
          </label>
        </div>
        <div
          v-if="props.awdChallengeLoadError"
          class="contest-awd-challenge-list__error"
        >
          <span>{{ props.awdChallengeLoadError }}</span>
          <button
            type="button"
            class="ui-btn ui-btn--ghost"
            @click="emit('refresh-awd-challenge-catalog')"
          >
            重试
          </button>
        </div>
        <div
          v-if="selectableAwdChallenges.length > 0"
          class="contest-awd-challenge-list__table workspace-directory-list"
        >
          <WorkspaceDataTable
            :columns="awdChallengeTableColumns"
            :rows="selectableAwdChallenges"
            row-key="id"
            row-class="contest-awd-challenge-table-row"
          >
            <template #cell-name="{ row }">
              <button
                :id="`contest-awd-challenge-name-${(row as AdminAwdChallengeData).id}`"
                type="button"
                class="contest-awd-challenge-table__name-button"
                :aria-pressed="isAwdChallengeSelected((row as AdminAwdChallengeData).id)"
                @click="selectAwdChallenge((row as AdminAwdChallengeData).id)"
              >
                {{ (row as AdminAwdChallengeData).name }}
              </button>
            </template>
            <template #cell-slug="{ row }">
              <span class="contest-awd-challenge-table__slug">
                {{ (row as AdminAwdChallengeData).slug }}
              </span>
            </template>
            <template #cell-service_type="{ row }">
              <span class="contest-awd-challenge-table__mono">
                {{ getServiceTypeLabel((row as AdminAwdChallengeData).service_type) }}
              </span>
            </template>
            <template #cell-deployment_mode="{ row }">
              <span class="contest-awd-challenge-table__text">
                {{ getDeploymentModeLabel((row as AdminAwdChallengeData).deployment_mode) }}
              </span>
            </template>
            <template #cell-readiness_status="{ row }">
              <span class="contest-awd-challenge-table__readiness">
                {{ getReadinessLabel((row as AdminAwdChallengeData).readiness_status) }}
              </span>
            </template>
            <template #cell-last_verified_at="{ row }">
              <span class="contest-awd-challenge-table__text">
                {{ formatLastVerifiedAt((row as AdminAwdChallengeData).last_verified_at) }}
              </span>
            </template>
            <template #cell-actions="{ row }">
              <button
                :id="`contest-awd-challenge-option-${(row as AdminAwdChallengeData).id}`"
                type="button"
                class="contest-awd-challenge-option"
                :class="{ 'is-selected': isAwdChallengeSelected((row as AdminAwdChallengeData).id) }"
                :aria-pressed="isAwdChallengeSelected((row as AdminAwdChallengeData).id)"
                @click="selectAwdChallenge((row as AdminAwdChallengeData).id)"
              >
                {{ isAwdChallengeSelected((row as AdminAwdChallengeData).id) ? '已选择' : '选择' }}
              </button>
            </template>
          </WorkspaceDataTable>
        </div>
        <div
          v-else
          class="contest-awd-challenge-list__empty"
        >
          {{
            loadingAwdChallengeCatalog
              ? '正在加载 AWD 题目...'
              : hasAwdChallengeFilters
                ? '当前筛选条件下没有匹配的 AWD 题目'
                : '暂无可选 AWD 题目'
          }}
        </div>
        <div class="contest-awd-challenge-list__pagination">
          <button
            id="contest-awd-challenge-prev-page"
            type="button"
            class="ui-btn ui-btn--ghost"
            :disabled="!canGoToPreviousAwdChallengePage"
            @click="changeAwdChallengePage(awdChallengePage - 1)"
          >
            上一页
          </button>
          <button
            id="contest-awd-challenge-next-page"
            type="button"
            class="ui-btn ui-btn--ghost"
            :disabled="!canGoToNextAwdChallengePage"
            @click="changeAwdChallengePage(awdChallengePage + 1)"
          >
            下一页
          </button>
        </div>
        <span
          v-if="fieldErrors.awd_challenge_id"
          class="ui-field__error contest-challenge-dialog__error"
        >
          {{ fieldErrors.awd_challenge_id }}
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
          {{ saving ? '保存中...' : mode === 'create' ? (isAwdContest ? '关联题目' : '关联题目') : '保存变更' }}
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

.contest-awd-challenge-list {
  display: grid;
  gap: var(--space-3);
}

.contest-awd-challenge-list__head,
.contest-awd-challenge-list__pagination,
.contest-awd-challenge-list__error {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.contest-awd-challenge-list__count {
  font-size: var(--font-size-0-75);
  color: var(--journal-muted);
}

.contest-awd-challenge-list__filters {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.contest-awd-challenge-list__empty {
  border: 1px solid var(--color-border-default);
  border-radius: var(--ui-control-radius);
  background: var(--color-bg-surface);
  padding: var(--space-4);
  color: var(--journal-muted);
  font-size: var(--font-size-0-875);
}

.contest-awd-challenge-list__table {
  max-height: clamp(12rem, calc(100dvh - 18rem), 30rem);
  overflow: auto;
}

.contest-awd-challenge-list__table :deep(.workspace-data-table) {
  min-width: 48rem;
}

.contest-awd-challenge-list__table :deep(.workspace-data-table__head-cell) {
  position: sticky;
  top: 0;
  z-index: 1;
  background: var(--color-bg-surface);
}

.contest-awd-challenge-list__error {
  padding: var(--space-3) var(--space-4);
  border: 1px solid var(--color-border-default);
  border-radius: var(--ui-control-radius);
  background: color-mix(in srgb, var(--color-danger-soft) 24%, var(--color-bg-surface));
}

.contest-awd-challenge-table__name-button {
  display: block;
  overflow: hidden;
  width: 100%;
  border: 0;
  background: transparent;
  padding: 0;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-align: left;
  font-size: var(--font-size-0-875);
  font-weight: 800;
  color: var(--color-text-primary);
  cursor: pointer;
  transition: color var(--ui-motion-fast);
}

.contest-awd-challenge-table__name-button:hover,
.contest-awd-challenge-table__name-button:focus-visible {
  color: var(--color-primary);
}

.contest-awd-challenge-table__name-button:focus-visible {
  outline: var(--ui-focus-ring-width) solid
    color-mix(in srgb, var(--color-primary) 72%, transparent);
  outline-offset: var(--space-1);
  border-radius: var(--ui-control-radius-sm);
}

.contest-awd-challenge-table__slug {
  color: var(--journal-muted);
  font-size: var(--font-size-0-75);
  font-family: var(--font-family-mono);
}

.contest-awd-challenge-table__mono,
.contest-awd-challenge-table__text,
.contest-awd-challenge-table__readiness {
  font-size: var(--font-size-0-75);
  font-weight: 700;
  color: var(--color-text-secondary);
}

.contest-awd-challenge-table__mono {
  font-family: var(--font-family-mono);
}

.contest-awd-challenge-option {
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

.contest-awd-challenge-option:hover,
.contest-awd-challenge-option:focus-visible {
  border-color: color-mix(in srgb, var(--color-primary) 62%, var(--color-border-default));
  background: color-mix(in srgb, var(--color-primary) 8%, var(--color-bg-surface));
  color: var(--color-primary);
}

.contest-awd-challenge-option.is-selected {
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

  .contest-awd-challenge-list__filters {
    grid-template-columns: minmax(0, 1fr);
  }

  .contest-challenge-dialog__footer {
    flex-direction: column-reverse;
  }
}
</style>
