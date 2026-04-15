<script setup lang="ts">
import { computed, onMounted, ref, toRef } from 'vue'
import { Plus, RefreshCw } from 'lucide-vue-next'

import {
  createAdminContestChallenge,
  deleteAdminContestChallenge,
  getChallenges,
  listAdminContestChallenges,
  updateAdminContestChallenge,
} from '@/api/admin'
import type { AdminChallengeListItem, AdminContestChallengeData, ContestDetailData } from '@/api/contracts'
import { ApiError } from '@/api/request'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import { useAwdCheckResultPresentation } from '@/composables/useAwdCheckResultPresentation'
import { useContestChallengePool } from '@/composables/useContestChallengePool'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useToast } from '@/composables/useToast'

import ContestChallengeEditorDialog from './ContestChallengeEditorDialog.vue'

const props = defineProps<{
  contestId: string
  contestMode: ContestDetailData['mode']
  challengeLinks?: AdminContestChallengeData[]
  loadingExternal?: boolean
}>()

const emit = defineEmits<{
  'open:awd-config': [challenge: AdminContestChallengeData]
  updated: []
}>()

const toast = useToast()
const CHALLENGE_CATALOG_PAGE_SIZE = 100
const loading = ref(true)
const saving = ref(false)
const loadingChallengeCatalog = ref(false)
const localChallengeLinks = ref<AdminContestChallengeData[]>([])
const challengeCatalog = ref<AdminChallengeListItem[]>([])
const dialogOpen = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const editingChallenge = ref<AdminContestChallengeData | null>(null)
const removingChallengeId = ref<string | null>(null)
const usingExternalChallengeLinks = computed(() => props.challengeLinks !== undefined)
const currentChallengeLinks = computed(() => props.challengeLinks ?? localChallengeLinks.value)
const panelLoading = computed(() => (usingExternalChallengeLinks.value ? Boolean(props.loadingExternal) : loading.value))

const {
  visibleItems,
  summaryItems,
  filterItems,
  activeFilter,
  isAwdContest,
  setFilter,
} = useContestChallengePool(currentChallengeLinks, toRef(props, 'contestMode'))

const panelCopy = computed(() =>
  isAwdContest.value
    ? '先在这里维护统一题目池，完成题目关联、顺序、分值和可见性；AWD 深度配置在下一阶段完成。'
    : '在这里维护统一题目池，安排题目顺序、基础分值和可见状态。'
)
const emptyState = computed(() =>
  isAwdContest.value && activeFilter.value !== 'all'
    ? {
        title: '当前筛选条件下没有匹配题目',
        description: '可以切换筛选查看其它题目，或继续补齐 AWD 配置后再返回这里检查。',
      }
    : {
        title: '当前竞赛还没有关联题目',
        description: '先从题库里关联题目，再安排顺序、分值和可见状态。',
      }
)

const existingChallengeIds = computed(() => currentChallengeLinks.value.map((item) => item.challenge_id))
const listTitle = computed(() => (isAwdContest.value ? '统一题目池' : '已关联题目'))

function formatDateTime(value?: string): string {
  if (!value) {
    return '未记录'
  }
  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const { getCheckerTypeLabel, getValidationStateLabel } = useAwdCheckResultPresentation({
  formatDateTime,
})

function getChallengeTitle(item: AdminContestChallengeData): string {
  return item.title?.trim() || `Challenge #${item.challenge_id}`
}

function getCheckerLabel(item: AdminContestChallengeData): string {
  return getCheckerTypeLabel(item.awd_checker_type) || '未配置 AWD'
}

function getValidationSummary(item: AdminContestChallengeData): string {
  return getValidationStateLabel(item.awd_checker_validation_state) || '未验证'
}

function getAwdScoreSummary(item: AdminContestChallengeData): string {
  return `SLA ${item.awd_sla_score ?? 0} / 防守 ${item.awd_defense_score ?? 0}`
}

function getPreviewSummary(item: AdminContestChallengeData): string {
  return formatDateTime(item.awd_checker_last_preview_at)
}

function humanizeRequestError(error: unknown, fallback: string): string {
  if (error instanceof ApiError && error.message.trim()) {
    return error.message
  }
  if (error instanceof Error && error.message.trim()) {
    return error.message
  }
  return fallback
}

async function refresh() {
  if (usingExternalChallengeLinks.value) {
    emit('updated')
    return
  }

  loading.value = true
  try {
    localChallengeLinks.value = await listAdminContestChallenges(props.contestId)
  } catch (error) {
    toast.error(humanizeRequestError(error, '赛事题目加载失败'))
  } finally {
    loading.value = false
  }
}

async function ensureChallengeCatalogLoaded() {
  if (loadingChallengeCatalog.value || challengeCatalog.value.length > 0) {
    return
  }

  loadingChallengeCatalog.value = true
  try {
    const list: AdminChallengeListItem[] = []
    let page = 1
    let total = 0

    do {
      const result = await getChallenges({
        page,
        page_size: CHALLENGE_CATALOG_PAGE_SIZE,
        status: 'published',
      })
      list.push(...result.list)
      total = result.total
      page += 1
    } while (list.length < total)

    challengeCatalog.value = list
  } catch (error) {
    toast.error(humanizeRequestError(error, '题目目录加载失败'))
  } finally {
    loadingChallengeCatalog.value = false
  }
}

function openCreateDialog() {
  dialogMode.value = 'create'
  editingChallenge.value = null
  dialogOpen.value = true
  void ensureChallengeCatalogLoaded()
}

function openEditDialog(challenge: AdminContestChallengeData) {
  dialogMode.value = 'edit'
  editingChallenge.value = challenge
  dialogOpen.value = true
}

function closeDialog() {
  dialogOpen.value = false
  editingChallenge.value = null
}

async function handleSave(payload: {
  challenge_id: number
  points: number
  order: number
  is_visible: boolean
}) {
  saving.value = true
  try {
    if (dialogMode.value === 'create') {
      await createAdminContestChallenge(props.contestId, payload)
      toast.success('赛事题目已关联')
    } else if (editingChallenge.value) {
      await updateAdminContestChallenge(props.contestId, editingChallenge.value.challenge_id, {
        points: payload.points,
        order: payload.order,
        is_visible: payload.is_visible,
      })
      toast.success('赛事题目已更新')
    }

    closeDialog()
    if (usingExternalChallengeLinks.value) {
      emit('updated')
    } else {
      await refresh()
    }
  } catch (error) {
    toast.error(humanizeRequestError(error, dialogMode.value === 'create' ? '关联题目失败' : '更新题目失败'))
  } finally {
    saving.value = false
  }
}

async function handleRemove(challenge: AdminContestChallengeData) {
  const confirmed = await confirmDestructiveAction({
    title: '移除赛事题目',
    confirmButtonText: '确认移除',
    message: `确认将题目“${getChallengeTitle(challenge)}”从当前竞赛中移除吗？`,
  })
  if (!confirmed) {
    return
  }

  removingChallengeId.value = challenge.id
  try {
    await deleteAdminContestChallenge(props.contestId, challenge.challenge_id)
    toast.success('赛事题目已移除')
    if (usingExternalChallengeLinks.value) {
      emit('updated')
    } else {
      await refresh()
    }
  } catch (error) {
    toast.error(humanizeRequestError(error, '移除题目失败'))
  } finally {
    removingChallengeId.value = null
  }
}

onMounted(() => {
  if (!usingExternalChallengeLinks.value) {
    void refresh()
  }
})
</script>

<template>
  <section class="contest-challenge-panel">
    <header class="contest-challenge-panel__header">
      <div class="workspace-tab-heading__main">
        <div class="workspace-overline">Challenge Pool</div>
        <h1 class="workspace-page-title">题目池</h1>
        <p class="workspace-page-copy">
          {{ panelCopy }}
        </p>
      </div>

      <div class="contest-challenge-panel__actions">
        <button type="button" class="admin-btn admin-btn-ghost" @click="refresh">
          <RefreshCw class="h-4 w-4" />
          刷新列表
        </button>
        <button id="contest-challenge-add" type="button" class="admin-btn admin-btn-primary" @click="openCreateDialog">
          <Plus class="h-4 w-4" />
          关联题目
        </button>
      </div>
    </header>

    <div class="metric-panel-grid metric-panel-default-surface contest-challenge-panel__summary">
      <article v-for="item in summaryItems" :key="item.key" class="journal-note metric-panel-card">
        <div class="journal-note-label metric-panel-label">{{ item.label }}</div>
        <div class="journal-note-value metric-panel-value">{{ item.value }}</div>
        <div class="journal-note-helper metric-panel-helper">{{ item.hint }}</div>
      </article>
    </div>

    <section class="workspace-directory-section contest-challenge-directory">
      <header class="list-heading">
        <div>
          <div class="journal-note-label">Challenge Directory</div>
          <h2 class="list-heading__title">{{ listTitle }}</h2>
        </div>
        <div class="contest-section-meta">共 {{ currentChallengeLinks.length }} 道题目</div>
      </header>

      <div v-if="isAwdContest && filterItems.length > 0" class="contest-challenge-filters">
        <button
          v-for="filter in filterItems"
          :id="`contest-challenge-filter-${filter.key}`"
          :key="filter.key"
          type="button"
          class="contest-challenge-filter"
          :class="{ 'contest-challenge-filter--active': activeFilter === filter.key }"
          @click="setFilter(filter.key)"
        >
          <span class="contest-challenge-filter__label">{{ filter.label }}</span>
          <span class="contest-challenge-filter__count">{{ filter.count }}</span>
          <span class="contest-challenge-filter__hint">{{ filter.hint }}</span>
        </button>
      </div>

      <div
        v-if="panelLoading"
        class="contest-challenge-directory__loading"
      >
        <AppLoading>正在同步赛事题目...</AppLoading>
      </div>

      <AppEmpty
        v-else-if="visibleItems.length === 0"
        :title="emptyState.title"
        :description="emptyState.description"
        icon="FileChartColumnIncreasing"
      />

      <template v-else>
        <div
          class="contest-challenge-directory__head"
          :class="{ 'contest-challenge-directory__head--awd': isAwdContest }"
          aria-hidden="true"
        >
          <span>题目</span>
          <span>可见性</span>
          <span>分值</span>
          <span>顺序</span>
          <template v-if="isAwdContest">
            <span>Checker</span>
            <span>验证状态</span>
            <span>SLA / 防守分</span>
            <span>最近试跑</span>
          </template>
          <span class="contest-challenge-directory__actions-label">操作</span>
        </div>

        <article
          v-for="challenge in visibleItems"
          :key="challenge.id"
          class="contest-challenge-row"
          :class="{ 'contest-challenge-row--awd': isAwdContest }"
        >
          <div class="contest-challenge-row__identity">
            <h3 class="contest-challenge-row__title">{{ getChallengeTitle(challenge) }}</h3>
            <p class="contest-challenge-row__meta">
              {{ challenge.category || '未分类' }} · {{ challenge.difficulty || '未标记难度' }}
            </p>
          </div>
          <div class="contest-challenge-row__visibility">
            {{ challenge.is_visible ? '可见' : '隐藏' }}
          </div>
          <div class="contest-challenge-row__score">{{ challenge.points }} 分</div>
          <div class="contest-challenge-row__order">第 {{ challenge.order }} 位</div>
          <template v-if="isAwdContest">
            <div class="contest-challenge-row__awd-cell">
              {{ getCheckerLabel(challenge) }}
            </div>
            <div class="contest-challenge-row__awd-cell">
              {{ getValidationSummary(challenge) }}
            </div>
            <div class="contest-challenge-row__awd-cell">
              {{ getAwdScoreSummary(challenge) }}
            </div>
            <div class="contest-challenge-row__awd-cell">
              {{ getPreviewSummary(challenge) }}
            </div>
          </template>
          <div
            class="contest-challenge-row__actions"
            role="group"
            :aria-label="`题目 ${getChallengeTitle(challenge)} 操作`"
          >
            <button
              v-if="isAwdContest"
              :id="`contest-challenge-open-awd-config-${challenge.id}`"
              type="button"
              class="contest-challenge-row__button contest-challenge-row__button--ghost"
              @click="emit('open:awd-config', challenge)"
            >
              补 AWD 配置
            </button>
            <button
              :id="`contest-challenge-edit-${challenge.id}`"
              type="button"
              class="contest-challenge-row__button contest-challenge-row__button--primary"
              @click="openEditDialog(challenge)"
            >
              编辑
            </button>
            <button
              :id="`contest-challenge-remove-${challenge.id}`"
              type="button"
              class="contest-challenge-row__button contest-challenge-row__button--danger"
              :disabled="removingChallengeId === challenge.id"
              @click="handleRemove(challenge)"
            >
              {{ removingChallengeId === challenge.id ? '移除中...' : '移除' }}
            </button>
          </div>
        </article>
      </template>
    </section>

    <ContestChallengeEditorDialog
      :open="dialogOpen"
      :mode="dialogMode"
      :contest-mode="contestMode"
      :challenge-options="challengeCatalog"
      :existing-challenge-ids="existingChallengeIds"
      :draft="editingChallenge"
      :loading-challenge-catalog="loadingChallengeCatalog"
      :saving="saving"
      @update:open="dialogOpen = $event"
      @save="handleSave"
    />
  </section>
</template>

<style scoped>
.contest-challenge-panel {
  display: grid;
  gap: var(--space-5);
}

.contest-challenge-panel__header {
  display: grid;
  gap: var(--space-4);
}

.contest-challenge-panel__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
}

.admin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: 2.75rem;
  border-radius: 1rem;
  padding: var(--space-2-5) var(--space-4);
  font-size: var(--font-size-0-875);
  font-weight: 600;
  transition: all 150ms ease;
}

.admin-btn-primary {
  border: 1px solid transparent;
  background: var(--color-primary);
  color: #fff;
}

.admin-btn-primary:hover {
  opacity: 0.92;
}

.admin-btn-ghost {
  border: 1px solid color-mix(in srgb, var(--journal-border) 76%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  color: var(--journal-ink);
}

.admin-btn-ghost:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  color: var(--journal-accent);
}

.contest-challenge-panel__summary {
  --admin-summary-grid-columns: repeat(auto-fit, minmax(11rem, 1fr));
}

.contest-challenge-directory {
  display: grid;
  gap: var(--space-4);
  padding: var(--space-5) var(--space-5-5);
}

.contest-challenge-directory__loading {
  display: flex;
  justify-content: center;
  padding: var(--space-6) 0;
}

.list-heading {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-3);
}

.list-heading__title {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-1-20);
  font-weight: 700;
  color: var(--journal-ink);
}

.contest-section-meta {
  font-size: var(--font-size-0-82);
  color: var(--journal-muted);
}

.contest-challenge-filters {
  display: grid;
  gap: var(--space-3);
  grid-template-columns: repeat(auto-fit, minmax(10.5rem, 1fr));
}

.contest-challenge-filter {
  display: grid;
  gap: var(--space-1);
  justify-items: start;
  border: 1px solid color-mix(in srgb, var(--journal-border) 76%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  padding: var(--space-3);
  text-align: left;
  transition: all 150ms ease;
}

.contest-challenge-filter:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 28%, transparent);
}

.contest-challenge-filter--active {
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, var(--journal-surface));
}

.contest-challenge-filter__label,
.contest-challenge-filter__count {
  font-weight: 700;
  color: var(--journal-ink);
}

.contest-challenge-filter__count {
  font-size: var(--font-size-1-20);
}

.contest-challenge-filter__hint {
  font-size: var(--font-size-0-82);
  color: var(--journal-muted);
}

.contest-challenge-directory__head,
.contest-challenge-row {
  display: grid;
  gap: var(--space-3);
  grid-template-columns: minmax(16rem, 1.7fr) minmax(6rem, 0.7fr) minmax(5rem, 0.55fr) minmax(5rem, 0.55fr) auto;
}

.contest-challenge-directory__head--awd,
.contest-challenge-row--awd {
  grid-template-columns:
    minmax(14rem, 1.5fr)
    minmax(5rem, 0.55fr)
    minmax(4.5rem, 0.45fr)
    minmax(4.5rem, 0.45fr)
    minmax(8rem, 0.9fr)
    minmax(7rem, 0.75fr)
    minmax(8.5rem, 0.95fr)
    minmax(8.5rem, 0.95fr)
    auto;
}

.contest-challenge-directory__head {
  padding: 0 var(--space-3) var(--space-2);
  font-size: var(--font-size-0-82);
  font-weight: 600;
  color: var(--journal-muted);
}

.contest-challenge-directory__actions-label {
  text-align: right;
}

.contest-challenge-row {
  align-items: center;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 70%, transparent);
  padding: var(--space-4) var(--space-3);
}

.contest-challenge-row__title {
  margin: 0;
  font-size: var(--font-size-1-05);
  font-weight: 700;
  color: var(--journal-ink);
}

.contest-challenge-row__meta {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-0-82);
  color: var(--journal-muted);
}

.contest-challenge-row__visibility,
.contest-challenge-row__score,
.contest-challenge-row__order,
.contest-challenge-row__awd-cell {
  font-size: var(--font-size-0-875);
  color: var(--journal-ink);
}

.contest-challenge-row__actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-2);
}

.contest-challenge-row__button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.2rem;
  border-radius: 0.7rem;
  padding: 0.5rem 0.85rem;
  font-size: var(--font-size-0-82);
  font-weight: 600;
  transition: all 150ms ease;
}

.contest-challenge-row__button--primary {
  border: 1px solid color-mix(in srgb, var(--journal-accent) 24%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, var(--journal-surface));
  color: var(--journal-accent);
}

.contest-challenge-row__button--ghost {
  border: 1px solid color-mix(in srgb, var(--journal-border) 76%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  color: var(--journal-ink);
}

.contest-challenge-row__button--danger {
  border: 1px solid color-mix(in srgb, var(--color-danger) 28%, transparent);
  background: color-mix(in srgb, var(--color-danger) 8%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-danger) 82%, var(--journal-ink));
}

.contest-challenge-row__button:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

@media (max-width: 900px) {
  .contest-challenge-panel__summary {
    --admin-summary-grid-columns: minmax(0, 1fr);
  }

  .contest-challenge-directory__head {
    display: none;
  }

  .contest-challenge-row {
    grid-template-columns: minmax(0, 1fr);
  }

  .contest-challenge-row__actions {
    justify-content: flex-start;
  }
}
</style>
