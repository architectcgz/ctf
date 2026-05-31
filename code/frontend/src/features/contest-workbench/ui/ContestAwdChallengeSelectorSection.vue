<script setup lang="ts">
import { computed } from 'vue'

import WorkspaceDataTable from '@/shared/ui/common/WorkspaceDataTable.vue'
import type { AdminAwdChallengeData } from '@/api/contracts'

const props = defineProps<{
  awdChallengeOptions: AdminAwdChallengeData[]
  awdChallengePage: number
  awdChallengeTotalPages: number
  awdChallengeKeyword?: string
  awdChallengeServiceType?: AdminAwdChallengeData['service_type'] | ''
  awdChallengeDeploymentMode?: AdminAwdChallengeData['deployment_mode'] | ''
  awdChallengeReadiness?: AdminAwdChallengeData['readiness_status'] | ''
  awdChallengeLoadError?: string
  loadingAwdChallengeCatalog: boolean
  selectedAwdChallengeIds: string[]
  fieldError?: string
}>()

const emit = defineEmits<{
  select: [awdChallengeId: string]
  'update-awd-challenge-keyword': [value: string]
  'update-awd-challenge-service-type': [value: AdminAwdChallengeData['service_type'] | '']
  'update-awd-challenge-deployment-mode': [value: AdminAwdChallengeData['deployment_mode'] | '']
  'update-awd-challenge-readiness': [value: AdminAwdChallengeData['readiness_status'] | '']
  'change-awd-challenge-page': [page: number]
  'refresh-awd-challenge-catalog': []
}>()

const hasAwdChallengeFilters = computed(() =>
  Boolean(
    (props.awdChallengeKeyword ?? '').trim() ||
      props.awdChallengeServiceType ||
      props.awdChallengeDeploymentMode ||
      props.awdChallengeReadiness
  )
)

const canGoToPreviousAwdChallengePage = computed(() => props.awdChallengePage > 1)
const canGoToNextAwdChallengePage = computed(
  () => props.awdChallengePage < props.awdChallengeTotalPages
)

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

function isAwdChallengeSelected(awdChallengeId: string): boolean {
  return props.selectedAwdChallengeIds.includes(awdChallengeId)
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
</script>

<template>
  <section
    id="contest-awd-challenge-list"
    class="contest-awd-challenge-list"
    :class="{ 'is-error': !!fieldError }"
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
            @input="
              emit('update-awd-challenge-keyword', ($event.target as HTMLInputElement).value)
            "
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
            @change="
              emit(
                'update-awd-challenge-service-type',
                ($event.target as HTMLSelectElement).value as AdminAwdChallengeData['service_type'] | ''
              )
            "
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
            @change="
              emit(
                'update-awd-challenge-deployment-mode',
                ($event.target as HTMLSelectElement).value as AdminAwdChallengeData['deployment_mode'] | ''
              )
            "
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
            @change="
              emit(
                'update-awd-challenge-readiness',
                ($event.target as HTMLSelectElement).value as AdminAwdChallengeData['readiness_status'] | ''
              )
            "
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
      v-if="awdChallengeLoadError"
      class="contest-awd-challenge-list__error"
    >
      <span>{{ awdChallengeLoadError }}</span>
      <button
        type="button"
        class="ui-btn ui-btn--ghost"
        @click="emit('refresh-awd-challenge-catalog')"
      >
        重试
      </button>
    </div>
    <div
      v-if="awdChallengeOptions.length > 0"
      class="contest-awd-challenge-list__table workspace-directory-list"
    >
      <WorkspaceDataTable
        :columns="awdChallengeTableColumns"
        :rows="awdChallengeOptions"
        row-key="id"
        row-class="contest-awd-challenge-table-row"
      >
        <template #cell-name="{ row }">
          <button
            :id="`contest-awd-challenge-name-${(row as AdminAwdChallengeData).id}`"
            type="button"
            class="contest-awd-challenge-table__name-button"
            :aria-pressed="isAwdChallengeSelected((row as AdminAwdChallengeData).id)"
            @click="emit('select', (row as AdminAwdChallengeData).id)"
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
            @click="emit('select', (row as AdminAwdChallengeData).id)"
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
        @click="emit('change-awd-challenge-page', awdChallengePage - 1)"
      >
        上一页
      </button>
      <button
        id="contest-awd-challenge-next-page"
        type="button"
        class="ui-btn ui-btn--ghost"
        :disabled="!canGoToNextAwdChallengePage"
        @click="emit('change-awd-challenge-page', awdChallengePage + 1)"
      >
        下一页
      </button>
    </div>
    <span
      v-if="fieldError"
      class="ui-field__error contest-challenge-dialog__error"
    >
      {{ fieldError }}
    </span>
  </section>
</template>

<style scoped>
.contest-challenge-dialog__field {
  --ui-field-gap: var(--space-2);
}

.contest-challenge-dialog__label {
  font-size: var(--font-size-0-875);
}

.contest-challenge-dialog__control {
  min-height: 2.75rem;
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

@media (max-width: 767px) {
  .contest-awd-challenge-list__filters {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
