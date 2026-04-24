<script setup lang="ts">
import { computed, onMounted, ref, toRef } from 'vue'
import { MoreHorizontal, Plus, RefreshCw, Zap, Edit, Trash, Boxes, AlertTriangle } from 'lucide-vue-next'

import {
  createContestAWDService,
  createAdminContestChallenge,
  listAdminAwdServiceTemplates,
  listAdminContestChallenges,
  listContestAWDServices,
  deleteContestAWDService,
  deleteAdminContestChallenge,
  getChallenges,
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
import CActionMenu from '@/components/common/menus/CActionMenu.vue'
import { useAwdCheckResultPresentation } from '@/composables/useAwdCheckResultPresentation'
import { useContestChallengePool } from '@/composables/useContestChallengePool'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useToast } from '@/composables/useToast'
import {
  mapPlatformContestAwdServicesToChallengeLinks,
  mergePlatformContestChallengesWithAwdServices,
} from '@/utils/platformContestAwdChallengeLinks'

import ContestChallengeEditorDialog from './ContestChallengeEditorDialog.vue'
import ContestChallengeSummaryStrip from './ContestChallengeSummaryStrip.vue'

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
const openActionMenuId = ref<string | null>(null)
const usingExternalChallengeLinks = computed(() => props.challengeLinks !== undefined)
const currentChallengeLinks = computed(() => props.challengeLinks ?? localChallengeLinks.value)
const panelLoading = computed(() => (usingExternalChallengeLinks.value ? Boolean(props.loadingExternal) : loading.value))
const panelLoadError = computed(() =>
  usingExternalChallengeLinks.value ? props.loadErrorExternal?.trim() ?? '' : localLoadError.value
)

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
    ? '维护统一题目池，从 AWD 题库模板选题并完成比赛级分值编排。'
    : '维护统一题目池，安排题目顺序、分值和可见状态。'
)
const emptyState = computed(() =>
  isAwdContest.value && activeFilter.value !== 'all'
    ? {
        title: '没有匹配题目',
        description: '切换筛选或补齐 AWD 配置。',
      }
    : {
        title: '暂无关联题目',
        description: '先从题库里关联题目，再安排顺序。',
      }
)

const existingChallengeIdSet = computed(
  () => new Set(currentChallengeLinks.value.map((item) => String(item.challenge_id)))
)
const existingChallengeIds = computed(() => Array.from(existingChallengeIdSet.value))
const dialogChallengeOptions = computed(() =>
  dialogMode.value === 'edit'
    ? challengeCatalog.value
    : challengeCatalog.value.filter((item) => !existingChallengeIdSet.value.has(String(item.id)))
)

function formatDateTime(value?: string): string {
  if (!value) return '未记录'
  return new Date(value).toLocaleString('zh-CN', {
    month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit',
  })
}

const { getCheckerTypeLabel, getValidationStateLabel } = useAwdCheckResultPresentation({
  formatDateTime,
})

function getChallengeTitle(item: AdminContestChallengeViewData): string {
  return item.title?.trim() || `Challenge #${item.challenge_id}`
}

function getChallengeActionKey(item: AdminContestChallengeViewData): string {
  return item.challenge_id
}

function getCheckerLabel(item: AdminContestChallengeViewData): string {
  return getCheckerTypeLabel(item.awd_checker_type) || '未配置'
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
  if (error instanceof ApiError && error.message.trim()) return error.message
  return (error as Error).message || fallback
}

async function refresh() {
  if (usingExternalChallengeLinks.value) {
    emit('updated')
    return
  }
  loading.value = true
  try {
    if (props.contestMode === 'awd') {
      const [nextChallengeLinks, nextAwdServices] = await Promise.all([
        listAdminContestChallenges(props.contestId),
        listContestAWDServices(props.contestId),
      ])
      localChallengeLinks.value =
        nextChallengeLinks.length > 0
          ? mergePlatformContestChallengesWithAwdServices(nextChallengeLinks, nextAwdServices)
          : mapPlatformContestAwdServicesToChallengeLinks(nextAwdServices)
    } else {
      localChallengeLinks.value = await listAdminContestChallenges(props.contestId)
    }
    localLoadError.value = ''
  } catch (error) {
    localLoadError.value = humanizeRequestError(error, '加载失败')
    toast.error(localLoadError.value)
  } finally {
    loading.value = false
  }
}

async function ensureChallengeCatalogLoaded() {
  if (loadingChallengeCatalog.value || challengeCatalog.value.length > 0) return
  loadingChallengeCatalog.value = true
  try {
    const result = await getChallenges({
      page: 1,
      page_size: CHALLENGE_CATALOG_PAGE_SIZE,
      status: 'published',
    })
    challengeCatalog.value = result.list
  } catch (error) {
    toast.error(humanizeRequestError(error, '题库加载失败'))
  } finally {
    loadingChallengeCatalog.value = false
  }
}

async function ensureTemplateCatalogLoaded() {
  if (loadingTemplateCatalog.value || templateCatalog.value.length > 0) return
  loadingTemplateCatalog.value = true
  try {
    const result = await listAdminAwdServiceTemplates({ page: 1, page_size: 100, status: 'published' })
    templateCatalog.value = result.list
  } finally {
    loadingTemplateCatalog.value = false
  }
}

function openCreateDialog() {
  dialogMode.value = 'create'
  editingChallenge.value = null
  dialogOpen.value = true
  void ensureChallengeCatalogLoaded()
  if (isAwdContest.value) void ensureTemplateCatalogLoaded()
}

function openEditDialog(challenge: AdminContestChallengeViewData) {
  dialogMode.value = 'edit'
  editingChallenge.value = challenge
  dialogOpen.value = true
  if (isAwdContest.value) void ensureTemplateCatalogLoaded()
}

function closeDialog() {
  dialogOpen.value = false
  editingChallenge.value = null
}

function setActionMenuOpen(challengeId: string, nextOpen: boolean): void {
  openActionMenuId.value = nextOpen ? challengeId : null
}

interface ContestOrchestrationSavePayload {
  challenge_id?: number
  template_id?: number
  points: number
  order: number
  is_visible: boolean
  awd_checker_type?: AdminContestChallengeViewData['awd_checker_type']
  awd_checker_config?: Record<string, unknown>
  awd_sla_score?: number
  awd_defense_score?: number
  awd_checker_preview_token?: string
}

async function handleSave(payload: ContestOrchestrationSavePayload) {
  saving.value = true
  try {
    if (isAwdContest.value) {
      if (!payload.template_id || !payload.challenge_id) {
        toast.error('请选择题目和服务模板')
        return
      }
      if (dialogMode.value === 'create') {
        await createContestAWDService(props.contestId, {
          challenge_id: payload.challenge_id,
          template_id: payload.template_id,
          order: payload.order,
          is_visible: payload.is_visible,
        } as Parameters<typeof createContestAWDService>[1] & { challenge_id: number })
        await updateAdminContestChallenge(props.contestId, String(payload.challenge_id), {
          points: payload.points,
        })
      } else if (editingChallenge.value) {
        await updateContestAWDService(
          props.contestId,
          editingChallenge.value.awd_service_id!,
          {
            template_id: payload.template_id,
            order: payload.order,
            is_visible: payload.is_visible,
          }
        )
        await updateAdminContestChallenge(props.contestId, editingChallenge.value.challenge_id, {
          points: payload.points,
        })
      }
    } else if (dialogMode.value === 'create') {
      await createAdminContestChallenge(props.contestId, {
        challenge_id: payload.challenge_id!,
        points: payload.points,
        order: payload.order,
        is_visible: payload.is_visible,
      })
    } else if (editingChallenge.value) {
      await updateAdminContestChallenge(props.contestId, editingChallenge.value.challenge_id, {
        points: payload.points,
        order: payload.order,
        is_visible: payload.is_visible,
      })
    }
    toast.success('题目已保存')
    closeDialog()
    emit('updated')
    if (!usingExternalChallengeLinks.value) await refresh()
  } catch (error) {
    toast.error(humanizeRequestError(error, '保存失败'))
  } finally {
    saving.value = false
  }
}

async function handleRemove(challenge: AdminContestChallengeViewData) {
  const confirmed = await confirmDestructiveAction({
    title: '移除题目',
    message: `确认将“${getChallengeTitle(challenge)}”从竞赛中移除吗？`,
  })
  if (!confirmed) return
  removingChallengeId.value = challenge.id
  try {
    if (props.contestMode === 'awd') {
      await deleteContestAWDService(props.contestId, challenge.awd_service_id!)
    } else {
      await deleteAdminContestChallenge(props.contestId, challenge.challenge_id)
    }
    toast.success('题目已移除')
    emit('updated')
    if (!usingExternalChallengeLinks.value) await refresh()
  } catch (error) {
    toast.error(humanizeRequestError(error, '移除失败'))
  } finally {
    removingChallengeId.value = null
  }
}

function handleOpenAwdConfig(challenge: AdminContestChallengeViewData, close: () => void) {
  close(); emit('open:awd-config', challenge)
}

function handleOpenEditDialog(challenge: AdminContestChallengeViewData, close: () => void) {
  close(); openEditDialog(challenge)
}

async function handleRemoveFromMenu(challenge: AdminContestChallengeViewData, close: () => void) {
  close(); await handleRemove(challenge)
}

onMounted(() => {
  if (!usingExternalChallengeLinks.value) void refresh()
})
</script>

<template>
  <section class="studio-orchestration">
    <header class="studio-pane-header">
      <div class="header-main">
        <h1 class="pane-title">
          题目编排
        </h1>
        <p class="pane-description">
          {{ panelCopy }}
        </p>
      </div>
      <div class="header-actions">
        <button
          type="button"
          class="ui-btn ui-btn--ghost"
          @click="refresh"
        >
          <RefreshCw
            class="h-3.5 w-3.5"
            :class="{ 'animate-spin': panelLoading }"
          />
          <span>同步数据</span>
        </button>
        <button
          id="contest-challenge-add"
          type="button"
          class="ui-btn ui-btn--primary"
          @click="openCreateDialog"
        >
          <Plus class="h-3.5 w-3.5" />
          <span>{{ isAwdContest ? '新增服务' : '关联新题目' }}</span>
        </button>
      </div>
    </header>

    <ContestChallengeSummaryStrip
      v-if="summaryItems.length > 0"
      :summary-items="summaryItems"
    />

    <div class="studio-directory-canvas">
      <AppEmpty
        v-if="panelLoadError && currentChallengeLinks.length === 0"
        title="赛事题目暂时不可用"
        :description="panelLoadError"
        icon="AlertTriangle"
        class="py-20"
      >
        <template #action>
          <button
            type="button"
            class="ui-btn ui-btn--ghost"
            @click="refresh"
          >
            重试
          </button>
        </template>
      </AppEmpty>

      <template v-else>
        <nav
          v-if="isAwdContest && filterItems.length > 0"
          class="studio-quick-nav"
        >
          <button
            v-for="filter in filterItems"
            :key="filter.key"
            :id="`contest-challenge-filter-${filter.key}`"
            class="nav-pill"
            :class="{ active: activeFilter === filter.key }"
            @click="setFilter(filter.key)"
          >
            <span class="nav-pill__label">{{ filter.label }}</span>
            <span class="nav-pill__count">{{ filter.count }}</span>
          </button>
        </nav>

        <div
          v-if="panelLoading"
          class="flex justify-center py-24"
        >
          <AppLoading>同步中...</AppLoading>
        </div>
        <AppEmpty
          v-else-if="visibleItems.length === 0"
          :title="emptyState.title"
          :description="emptyState.description"
          icon="Boxes"
          class="py-20"
        />

        <div
          v-else
          class="studio-table-wrap custom-scrollbar"
        >
          <table class="studio-table">
            <thead>
              <tr>
                <th class="col-identity">
                  题目资源
                </th>
                <th class="col-meta">
                  可见性
                </th>
                <th class="col-meta">
                  分值
                </th>
                <th class="col-meta">
                  权重
                </th>
                <template v-if="isAwdContest">
                  <th class="col-awd">
                    Checker
                  </th>
                  <th class="col-awd">
                    验证状态
                  </th>
                  <th class="col-awd">
                    SLA / 防守分
                  </th>
                  <th class="col-awd">
                    最近试跑
                  </th>
                </template>
                <th class="col-actions">
                  管理
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="challenge in visibleItems"
                :key="challenge.id"
                class="studio-row"
              >
                <td class="col-identity">
                  <div class="challenge-identity">
                    <div class="challenge-title">
                      {{ getChallengeTitle(challenge) }}
                    </div>
                    <div class="challenge-subtitle">
                      {{ challenge.category || '通用' }} · {{ challenge.difficulty || '常规' }}
                    </div>
                  </div>
                </td>
                <td class="col-meta">
                  <span
                    class="status-badge"
                    :class="challenge.is_visible ? 'is-visible' : 'is-hidden'"
                  >{{ challenge.is_visible ? '公开' : '隐藏' }}</span>
                </td>
                <td class="col-meta contest-points-cell">
                  {{ challenge.points }} <small>PTS</small>
                </td>
                <td class="col-meta">
                  <div class="order-chip">
                    第 {{ challenge.order }} 位
                  </div>
                </td>
                <template v-if="isAwdContest">
                  <td class="col-awd">
                    <div class="engine-tag">
                      {{ getCheckerLabel(challenge) }}
                    </div>
                  </td>
                  <td class="col-awd">
                    <span
                      class="validation-status"
                      :class="challenge.awd_checker_validation_state"
                    >{{ getValidationSummary(challenge) }}</span>
                  </td>
                  <td class="col-awd contest-awd-score">
                    {{ getAwdScoreSummary(challenge) }}
                  </td>
                  <td class="col-awd contest-awd-preview">
                    {{ getPreviewSummary(challenge) }}
                  </td>
                </template>
                <td class="col-actions">
                  <div class="ui-row-actions contest-challenge-row__actions">
                    <CActionMenu
                      :open="openActionMenuId === getChallengeActionKey(challenge)"
                      @update:open="setActionMenuOpen(getChallengeActionKey(challenge), $event)"
                    >
                      <template #trigger="{ open, toggle, setTriggerRef }">
                        <button
                          :id="`contest-challenge-actions-${getChallengeActionKey(challenge)}`"
                          :ref="setTriggerRef"
                          class="c-action-menu__trigger c-action-menu__trigger--icon"
                          @click.stop="toggle"
                        >
                          <MoreHorizontal class="h-4 w-4" />
                        </button>
                      </template>
                      <template #default="{ close }">
                        <button
                          v-if="isAwdContest"
                          :id="`contest-challenge-open-awd-config-${getChallengeActionKey(challenge)}`"
                          class="c-action-menu__item"
                          @click="handleOpenAwdConfig(challenge, close)"
                        >
                          <Zap class="h-3.5 w-3.5 mr-2" /> 补 AWD 配置
                        </button>
                        <button
                          :id="`contest-challenge-edit-${getChallengeActionKey(challenge)}`"
                          class="c-action-menu__item"
                          @click="handleOpenEditDialog(challenge, close)"
                        >
                          <Edit class="h-3.5 w-3.5 mr-2" /> 属性修改
                        </button>
                        <div class="menu-divider" />
                        <button
                          :id="`contest-challenge-remove-${getChallengeActionKey(challenge)}`"
                          class="c-action-menu__item c-action-menu__item--danger"
                          @click="handleRemoveFromMenu(challenge, close)"
                        >
                          <Trash class="h-3.5 w-3.5 mr-2" /> 移除题目
                        </button>
                      </template>
                    </CActionMenu>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>

    <ContestChallengeEditorDialog
      :key="`${dialogMode}:${existingChallengeIds.join(',')}`"
      :open="dialogOpen"
      :mode="dialogMode"
      :contest-mode="contestMode"
      :challenge-options="dialogChallengeOptions"
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
.studio-orchestration { display: flex; flex-direction: column; gap: 1.5rem; padding: 1.5rem 2rem; background: var(--color-bg-base); }
.studio-pane-header { display: flex; justify-content: space-between; align-items: flex-end; }
.pane-title { font-size: 1.25rem; font-weight: 900; color: var(--color-text-primary); margin: 0; }
.pane-description { font-size: 13px; color: var(--color-text-secondary); margin: 0.5rem 0 0; max-width: 40rem; }
.header-actions { display: flex; gap: 0.75rem; }
/* Metric Band - Flattened and Scaled Up */
.studio-metric-band {
  display: flex;
  gap: var(--space-4);
  background: transparent;
  padding: 0;
  border: none;
  border-radius: 0;
}
.metric-pill {
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-default);
  padding: var(--space-4) var(--space-6);
  border-radius: 1rem;
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  min-width: 9rem;
  box-shadow: var(--color-shadow-soft);
}
.metric-pill__label {
  font-size: var(--font-size-11);
  font-weight: 800;
  text-transform: uppercase;
  color: var(--color-text-muted);
  letter-spacing: 0.1em;
}
.metric-pill__value {
  font-size: var(--font-size-20);
  font-weight: 900;
  color: var(--color-primary);
  font-family: var(--font-family-mono);
  line-height: 1.1;
}
.studio-quick-nav { display: flex; gap: 0.5rem; margin-bottom: 0.5rem; }
.nav-pill { padding: 0.45rem 1rem; border-radius: 999px; background: var(--color-bg-surface); border: 1px solid var(--color-border-default); font-size: 12px; font-weight: 700; color: var(--color-text-secondary); cursor: pointer; display: flex; align-items: center; gap: 0.5rem; }
.nav-pill.active { background: var(--color-primary-soft); border-color: var(--color-primary); color: var(--color-primary); }
.nav-pill__count {
  background: color-mix(in srgb, var(--journal-ink) 5%, transparent);
  padding: 0 0.35rem;
  border-radius: 4px;
  font-size: 10px;
}
.studio-table-wrap { border: none; border-radius: 0; background: transparent; overflow-x: auto; }
.studio-table { width: 100%; border-collapse: collapse; background: var(--color-bg-surface); }
.studio-table th { background: var(--color-bg-elevated); padding: 0.75rem 1rem; text-align: left; font-size: 10px; font-weight: 800; text-transform: uppercase; color: var(--color-text-muted); border-bottom: 1px solid var(--color-border-default); border-top: 1px solid var(--color-border-default); }
.studio-table td { padding: 1rem; border-bottom: 1px solid var(--color-border-subtle); }
.studio-row:hover { background: var(--color-bg-elevated); }
.challenge-title { font-size: 14px; font-weight: 800; color: var(--color-text-primary); }
.challenge-subtitle { font-size: 11px; color: var(--color-text-muted); margin-top: 0.15rem; }
.contest-points-cell {
  font-family: var(--font-family-mono);
  font-weight: 900;
  color: color-mix(in srgb, var(--journal-ink) 82%, var(--journal-muted));
}
.contest-awd-score {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-11);
  color: var(--journal-muted);
}
.contest-awd-preview {
  font-size: var(--font-size-11);
  color: color-mix(in srgb, var(--journal-muted) 84%, var(--journal-ink));
}
.status-badge { padding: 0.15rem 0.5rem; border-radius: 6px; font-size: 10px; font-weight: 800; }
.is-visible { background: var(--color-success-soft); color: var(--color-success); }
.is-hidden { background: var(--color-bg-elevated); color: var(--color-text-secondary); }
.order-chip { font-size: 10px; font-weight: 900; color: var(--color-primary); background: var(--color-primary-soft); padding: 0.25rem 0.5rem; border-radius: 4px; display: inline-block; }
.engine-tag { font-size: 11px; font-weight: 700; color: var(--color-text-secondary); }
.validation-status { font-size: 10px; font-weight: 700; }
.validation-status.valid { color: var(--color-success); }
.validation-status.invalid { color: var(--color-danger); }
.validation-status.pending { color: var(--color-warning); }
.action-trigger { width: 2rem; height: 2rem; display: flex; align-items: center; justify-content: center; border-radius: 0.5rem; color: var(--color-text-muted); transition: all 0.2s ease; cursor: pointer; }
.action-trigger:hover { background: var(--color-bg-elevated); color: var(--color-text-primary); }
.menu-item { width: 100%; padding: 0.5rem 1rem; display: flex; align-items: center; font-size: 12px; font-weight: 600; color: var(--color-text-secondary); background: transparent; cursor: pointer; }
.menu-item:hover { background: var(--color-bg-elevated); }
.menu-item.danger { color: var(--color-danger); }
.menu-divider { height: 1px; background: var(--color-border-default); margin: 0.25rem 0; }
</style>
