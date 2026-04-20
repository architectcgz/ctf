<script setup lang="ts">
import { computed, onMounted, ref, toRef } from 'vue'
import { Plus, RefreshCw } from 'lucide-vue-next'

import {
  createContestAWDService,
  createAdminContestChallenge,
  listAdminAwdServiceTemplates,
  listContestAWDServices,
  deleteContestAWDService,
  deleteAdminContestChallenge,
  getChallenges,
  listAdminContestChallenges,
  updateContestAWDService,
  updateAdminContestChallenge,
} from '@/api/admin'
import type {
  AdminAwdServiceTemplateData,
  AdminChallengeListItem,
  AdminContestChallengeViewData,
  ContestDetailData,
} from '@/api/contracts'
import { ApiError } from '@/api/request'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import { useAwdCheckResultPresentation } from '@/composables/useAwdCheckResultPresentation'
import { useContestChallengePool } from '@/composables/useContestChallengePool'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useToast } from '@/composables/useToast'
import { mergePlatformContestChallengesWithAwdServices } from '@/utils/platformContestAwdChallengeLinks'

import ContestChallengeEditorDialog from './ContestChallengeEditorDialog.vue'

const props = defineProps<{
  contestId: string
  contestMode: ContestDetailData['mode']
  challengeLinks?: AdminContestChallengeViewData[]
  loadingExternal?: boolean
  loadErrorExternal?: string
}>()

const emit = defineEmits<{
  'open:awd-config': [challenge: AdminContestChallengeViewData]
  updated: []
}>()

const toast = useToast()
const CHALLENGE_CATALOG_PAGE_SIZE = 100
const loading = ref(true)
const saving = ref(false)
const loadingChallengeCatalog = ref(false)
const loadingTemplateCatalog = ref(false)
const localChallengeLinks = ref<AdminContestChallengeViewData[]>([])
const localLoadError = ref('')
const challengeCatalog = ref<AdminChallengeListItem[]>([])
const templateCatalog = ref<AdminAwdServiceTemplateData[]>([])
const dialogOpen = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const editingChallenge = ref<AdminContestChallengeViewData | null>(null)
const removingChallengeId = ref<string | null>(null)
const quickActionKey = ref<string | null>(null)
const usingExternalChallengeLinks = computed(() => props.challengeLinks !== undefined)
const currentChallengeLinks = computed(() => props.challengeLinks ?? localChallengeLinks.value)
const panelLoading = computed(() => (usingExternalChallengeLinks.value ? Boolean(props.loadingExternal) : loading.value))
const panelLoadError = computed(() =>
  usingExternalChallengeLinks.value ? props.loadErrorExternal?.trim() ?? '' : localLoadError.value
)

const {
  sortedItems,
  visibleItems,
  summaryItems,
  filterItems,
  activeFilter,
  isAwdContest,
  setFilter,
} = useContestChallengePool(currentChallengeLinks, toRef(props, 'contestMode'))

const panelCopy = computed(() =>
  isAwdContest.value
    ? '先在这里维护统一题目池，完成题目关联、服务模板、顺序、分值和可见性；Checker 等检查配置继续留在 AWD 工作台。'
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

function getChallengeTitle(item: AdminContestChallengeViewData): string {
  return item.title?.trim() || `Challenge #${item.challenge_id}`
}

function getCheckerLabel(item: AdminContestChallengeViewData): string {
  return getCheckerTypeLabel(item.awd_checker_type) || '未配置 AWD'
}

function getValidationSummary(item: AdminContestChallengeViewData): string {
  return getValidationStateLabel(item.awd_checker_validation_state) || '未验证'
}

function getAwdScoreSummary(item: AdminContestChallengeViewData): string {
  return `SLA ${item.awd_sla_score ?? 0} / 防守 ${item.awd_defense_score ?? 0}`
}

function getPreviewSummary(item: AdminContestChallengeViewData): string {
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
    const nextChallengeLinks = await listAdminContestChallenges(props.contestId)
    const nextAwdServices =
      props.contestMode === 'awd' ? await listContestAWDServices(props.contestId) : []
    localChallengeLinks.value = mergePlatformContestChallengesWithAwdServices(
      nextChallengeLinks,
      nextAwdServices
    )
    localLoadError.value = ''
  } catch (error) {
    localLoadError.value = humanizeRequestError(error, '赛事题目加载失败')
    toast.error(localLoadError.value)
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

async function ensureTemplateCatalogLoaded() {
  if (loadingTemplateCatalog.value || templateCatalog.value.length > 0) {
    return
  }

  loadingTemplateCatalog.value = true
  try {
    const list: AdminAwdServiceTemplateData[] = []
    let page = 1
    let total = 0

    do {
      const result = await listAdminAwdServiceTemplates({
        page,
        page_size: 100,
        status: 'published',
      })
      list.push(...result.list)
      total = result.total
      page += 1
    } while (list.length < total)

    templateCatalog.value = list
  } catch (error) {
    toast.error(humanizeRequestError(error, '服务模板目录加载失败'))
  } finally {
    loadingTemplateCatalog.value = false
  }
}

function openCreateDialog() {
  dialogMode.value = 'create'
  editingChallenge.value = null
  dialogOpen.value = true
  void ensureChallengeCatalogLoaded()
  if (isAwdContest.value) {
    void ensureTemplateCatalogLoaded()
  }
}

function openEditDialog(challenge: AdminContestChallengeViewData) {
  dialogMode.value = 'edit'
  editingChallenge.value = challenge
  dialogOpen.value = true
  if (isAwdContest.value) {
    void ensureTemplateCatalogLoaded()
  }
}

function closeDialog() {
  dialogOpen.value = false
  editingChallenge.value = null
}

async function handleSave(payload: {
  challenge_id: number
  template_id?: number
  points: number
  order: number
  is_visible: boolean
}) {
  const templateId = payload.template_id
  if (isAwdContest.value && (!templateId || templateId < 1)) {
    toast.error('请选择服务模板')
    return
  }
  const ensuredTemplateId = templateId as number | undefined

  saving.value = true
  try {
    if (isAwdContest.value) {
      if (dialogMode.value === 'create') {
        await createContestAWDService(props.contestId, {
          challenge_id: payload.challenge_id,
          template_id: ensuredTemplateId as number,
          order: payload.order,
          is_visible: payload.is_visible,
        })
        await updateAdminContestChallenge(props.contestId, String(payload.challenge_id), {
          points: payload.points,
        })
        toast.success('AWD 题目已关联')
      } else if (editingChallenge.value) {
        if (editingChallenge.value.awd_service_id) {
          await updateContestAWDService(props.contestId, editingChallenge.value.awd_service_id, {
            template_id: ensuredTemplateId as number,
            order: payload.order,
            is_visible: payload.is_visible,
          })
        } else {
          await createContestAWDService(props.contestId, {
            challenge_id: Number(editingChallenge.value.challenge_id),
            template_id: ensuredTemplateId as number,
            order: payload.order,
            is_visible: payload.is_visible,
          })
        }
        await updateAdminContestChallenge(props.contestId, editingChallenge.value.challenge_id, {
          points: payload.points,
        })
        toast.success('AWD 题目已更新')
      }
    } else if (dialogMode.value === 'create') {
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
    toast.error(
      humanizeRequestError(error, dialogMode.value === 'create' ? '关联题目失败' : '更新题目失败')
    )
  } finally {
    saving.value = false
  }
}

async function handleRemove(challenge: AdminContestChallengeViewData) {
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
    if (props.contestMode === 'awd' && challenge.awd_service_id) {
      await deleteContestAWDService(props.contestId, challenge.awd_service_id)
    } else {
      await deleteAdminContestChallenge(props.contestId, challenge.challenge_id)
    }
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

function createQuickActionKey(action: string, challengeId: string): string {
  return `${action}:${challengeId}`
}

function isQuickActionPending(action: string, challengeId: string): boolean {
  return quickActionKey.value === createQuickActionKey(action, challengeId)
}

function canMoveChallenge(challenge: AdminContestChallengeViewData, offset: -1 | 1): boolean {
  const currentIndex = sortedItems.value.findIndex((item) => item.id === challenge.id)
  return currentIndex >= 0 && currentIndex + offset >= 0 && currentIndex + offset < sortedItems.value.length
}

async function refreshAfterMutation() {
  if (usingExternalChallengeLinks.value) {
    emit('updated')
    return
  }

  await refresh()
}

async function updateChallengeVisibility(
  challenge: AdminContestChallengeViewData,
  nextVisibility: boolean
) {
  if (props.contestMode === 'awd' && challenge.awd_service_id) {
    await updateContestAWDService(props.contestId, challenge.awd_service_id, {
      is_visible: nextVisibility,
    })
    return
  }

  await updateAdminContestChallenge(props.contestId, challenge.challenge_id, {
    is_visible: nextVisibility,
  })
}

async function updateChallengeOrder(challenge: AdminContestChallengeViewData, nextOrder: number) {
  if (props.contestMode === 'awd' && challenge.awd_service_id) {
    await updateContestAWDService(props.contestId, challenge.awd_service_id, {
      order: nextOrder,
    })
    return
  }

  await updateAdminContestChallenge(props.contestId, challenge.challenge_id, {
    order: nextOrder,
  })
}

async function handleToggleVisibility(challenge: AdminContestChallengeViewData) {
  const nextVisibility = !challenge.is_visible
  const actionKey = createQuickActionKey('toggle-visibility', challenge.id)
  if (quickActionKey.value) {
    return
  }

  quickActionKey.value = actionKey
  try {
    await updateChallengeVisibility(challenge, nextVisibility)
    toast.success(nextVisibility ? '题目已设为可见' : '题目已隐藏')
    await refreshAfterMutation()
  } catch (error) {
    toast.error(humanizeRequestError(error, nextVisibility ? '显示题目失败' : '隐藏题目失败'))
  } finally {
    quickActionKey.value = null
  }
}

async function handleMoveChallenge(challenge: AdminContestChallengeViewData, offset: -1 | 1) {
  if (quickActionKey.value) {
    return
  }

  const currentIndex = sortedItems.value.findIndex((item) => item.id === challenge.id)
  if (currentIndex < 0) {
    return
  }

  const adjacentChallenge = sortedItems.value[currentIndex + offset]
  if (!adjacentChallenge) {
    return
  }

  const actionKey = createQuickActionKey(offset < 0 ? 'move-up' : 'move-down', challenge.id)
  quickActionKey.value = actionKey
  try {
    await updateChallengeOrder(challenge, adjacentChallenge.order)
    await updateChallengeOrder(adjacentChallenge, challenge.order)
    toast.success(offset < 0 ? '题目已上移' : '题目已下移')
    await refreshAfterMutation()
  } catch (error) {
    toast.error(humanizeRequestError(error, offset < 0 ? '题目上移失败' : '题目下移失败'))
  } finally {
    quickActionKey.value = null
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
        <button type="button" class="ui-btn ui-btn--ghost" @click="refresh">
          <RefreshCw class="h-4 w-4" />
          刷新列表
        </button>
        <button
          id="contest-challenge-add"
          type="button"
          class="ui-btn ui-btn--primary"
          @click="openCreateDialog"
        >
          <Plus class="h-4 w-4" />
          关联题目
        </button>
      </div>
    </header>

    <AppEmpty
      v-if="panelLoadError && currentChallengeLinks.length === 0"
      title="赛事题目暂时不可用"
      :description="panelLoadError"
      icon="AlertTriangle"
    >
      <template #action>
        <button type="button" class="ui-btn ui-btn--ghost" @click="refresh">
          重试加载
        </button>
      </template>
    </AppEmpty>

    <template v-else>
      <p v-if="panelLoadError" class="contest-challenge-panel__warning" role="status">
        题目池刷新失败，当前显示上次成功同步的数据。{{ panelLoadError }}
      </p>

      <div
        class="progress-strip metric-panel-grid metric-panel-default-surface contest-challenge-panel__summary"
      >
        <article
          v-for="item in summaryItems"
          :key="item.key"
          class="journal-note progress-card metric-panel-card"
        >
          <div class="journal-note-label progress-card-label metric-panel-label">
            {{ item.label }}
          </div>
          <div class="journal-note-value progress-card-value metric-panel-value">
            {{ item.value }}
          </div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">
            {{ item.hint }}
          </div>
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
              class="ui-row-actions contest-challenge-row__actions"
              role="group"
              :aria-label="`题目 ${getChallengeTitle(challenge)} 操作`"
            >
              <button
                v-if="isAwdContest"
                :id="`contest-challenge-open-awd-config-${challenge.id}`"
                type="button"
                class="ui-btn ui-btn--sm ui-btn--secondary contest-challenge-row__button"
                @click="emit('open:awd-config', challenge)"
              >
                补 AWD 配置
              </button>
              <button
                :id="`contest-challenge-move-up-${challenge.id}`"
                type="button"
                class="ui-btn ui-btn--sm ui-btn--ghost contest-challenge-row__button"
                :disabled="!canMoveChallenge(challenge, -1) || quickActionKey !== null"
                @click="handleMoveChallenge(challenge, -1)"
              >
                {{ isQuickActionPending('move-up', challenge.id) ? '上移中...' : '上移' }}
              </button>
              <button
                :id="`contest-challenge-move-down-${challenge.id}`"
                type="button"
                class="ui-btn ui-btn--sm ui-btn--ghost contest-challenge-row__button"
                :disabled="!canMoveChallenge(challenge, 1) || quickActionKey !== null"
                @click="handleMoveChallenge(challenge, 1)"
              >
                {{ isQuickActionPending('move-down', challenge.id) ? '下移中...' : '下移' }}
              </button>
              <button
                :id="`contest-challenge-toggle-visibility-${challenge.id}`"
                type="button"
                class="ui-btn ui-btn--sm ui-btn--ghost contest-challenge-row__button"
                :disabled="quickActionKey !== null"
                @click="handleToggleVisibility(challenge)"
              >
                {{
                  isQuickActionPending('toggle-visibility', challenge.id)
                    ? challenge.is_visible
                      ? '隐藏中...'
                      : '显示中...'
                    : challenge.is_visible
                      ? '隐藏'
                      : '显示'
                }}
              </button>
              <button
                :id="`contest-challenge-edit-${challenge.id}`"
                type="button"
                class="ui-btn ui-btn--sm ui-btn--primary contest-challenge-row__button"
                :disabled="quickActionKey !== null"
                @click="openEditDialog(challenge)"
              >
                编辑
              </button>
              <button
                :id="`contest-challenge-remove-${challenge.id}`"
                type="button"
                class="ui-btn ui-btn--sm ui-btn--danger contest-challenge-row__button"
                :disabled="removingChallengeId === challenge.id || quickActionKey !== null"
                @click="handleRemove(challenge)"
              >
                {{ removingChallengeId === challenge.id ? '移除中...' : '移除' }}
              </button>
            </div>
          </article>
        </template>
      </section>
    </template>

    <ContestChallengeEditorDialog
      :open="dialogOpen"
      :mode="dialogMode"
      :contest-mode="contestMode"
      :challenge-options="challengeCatalog"
      :template-options="templateCatalog"
      :existing-challenge-ids="existingChallengeIds"
      :draft="editingChallenge"
      :loading-challenge-catalog="loadingChallengeCatalog"
      :loading-template-catalog="loadingTemplateCatalog"
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

.contest-challenge-panel__summary {
  --admin-summary-grid-columns: repeat(auto-fit, minmax(11rem, 1fr));
}

.contest-challenge-panel__warning {
  margin: 0;
  border: 1px solid color-mix(in srgb, var(--journal-danger, #d9594c) 32%, transparent);
  border-radius: 1rem;
  padding: var(--space-3) var(--space-4);
  background: color-mix(in srgb, var(--journal-danger, #d9594c) 12%, transparent);
  color: var(--journal-ink);
  font-size: var(--font-size-0-875);
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
  grid-template-columns: minmax(16rem, 1.7fr) minmax(6rem, 0.7fr) minmax(5rem, 0.55fr) minmax(
      5rem,
      0.55fr
    ) auto;
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
  justify-content: flex-end;
}

.contest-challenge-row__button {
  --ui-btn-radius: 0.7rem;
  --ui-btn-font-size: var(--font-size-0-82);
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
