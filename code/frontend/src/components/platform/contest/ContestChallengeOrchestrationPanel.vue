<script setup lang="ts">
import { computed, onMounted, ref, toRef } from 'vue'
import { MoreHorizontal, Plus, RefreshCw, Zap, Edit, Trash, Boxes, AlertTriangle } from 'lucide-vue-next'

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
import CActionMenu from '@/components/common/menus/CActionMenu.vue'
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
    ? '维护统一题目池，完成题目关联、服务模板及分值配置。'
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

const existingChallengeIds = computed(() => currentChallengeLinks.value.map((item) => item.challenge_id))

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

function getCheckerLabel(item: AdminContestChallengeViewData): string {
  return getCheckerTypeLabel(item.awd_checker_type) || '未配置'
}

function getValidationSummary(item: AdminContestChallengeViewData): string {
  return getValidationStateLabel(item.awd_checker_validation_state) || '未验证'
}

function getAwdScoreSummary(item: AdminContestChallengeViewData): string {
  return `S:${item.awd_sla_score ?? 0} D:${item.awd_defense_score ?? 0}`
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
    const nextChallengeLinks = await listAdminContestChallenges(props.contestId)
    const nextAwdServices = props.contestMode === 'awd' ? await listContestAWDServices(props.contestId) : []
    localChallengeLinks.value = mergePlatformContestChallengesWithAwdServices(nextChallengeLinks, nextAwdServices)
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
    const result = await getChallenges({ page: 1, page_size: 200, status: 'published' })
    challengeCatalog.value = result.list
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

async function handleSave(payload: any) {
  saving.value = true
  try {
    if (isAwdContest.value) {
      if (dialogMode.value === 'create') {
        await createContestAWDService(props.contestId, payload)
        await updateAdminContestChallenge(props.contestId, String(payload.challenge_id), { points: payload.points })
      } else if (editingChallenge.value) {
        if (editingChallenge.value.awd_service_id) {
          await updateContestAWDService(props.contestId, editingChallenge.value.awd_service_id, payload)
        } else {
          await createContestAWDService(props.contestId, { ...payload, challenge_id: Number(editingChallenge.value.challenge_id) })
        }
        await updateAdminContestChallenge(props.contestId, editingChallenge.value.challenge_id, { points: payload.points })
      }
    } else if (dialogMode.value === 'create') {
      await createAdminContestChallenge(props.contestId, payload)
    } else if (editingChallenge.value) {
      await updateAdminContestChallenge(props.contestId, editingChallenge.value.challenge_id, payload)
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
    if (props.contestMode === 'awd' && challenge.awd_service_id) {
      await deleteContestAWDService(props.contestId, challenge.awd_service_id)
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
        <h1 class="pane-title">题目编排</h1>
        <p class="pane-description">{{ panelCopy }}</p>
      </div>
      <div class="header-actions">
        <button type="button" class="ops-btn ops-btn--neutral" @click="refresh">
          <RefreshCw class="h-3.5 w-3.5" :class="{ 'animate-spin': panelLoading }" />
          <span>同步数据</span>
        </button>
        <button type="button" class="ops-btn ops-btn--primary" @click="openCreateDialog">
          <Plus class="h-3.5 w-3.5" />
          <span>关联新题目</span>
        </button>
      </div>
    </header>

    <div v-if="summaryItems.length > 0" class="studio-metric-band">
      <div v-for="item in summaryItems" :key="item.key" class="metric-pill">
        <span class="metric-pill__label">{{ item.label }}</span>
        <span class="metric-pill__value">{{ item.value }}</span>
      </div>
    </div>

    <div class="studio-directory-canvas">
      <AppEmpty v-if="panelLoadError && currentChallengeLinks.length === 0" title="同步中断" :description="panelLoadError" icon="AlertTriangle" class="py-20">
        <template #action><button type="button" class="ops-btn ops-btn--neutral" @click="refresh">重试</button></template>
      </AppEmpty>

      <template v-else>
        <nav v-if="isAwdContest && filterItems.length > 0" class="studio-quick-nav">
          <button v-for="filter in filterItems" :key="filter.key" class="nav-pill" :class="{ active: activeFilter === filter.key }" @click="setFilter(filter.key)">
            <span class="nav-pill__label">{{ filter.label }}</span>
            <span class="nav-pill__count">{{ filter.count }}</span>
          </button>
        </nav>

        <div v-if="panelLoading" class="flex justify-center py-24"><AppLoading>同步中...</AppLoading></div>
        <AppEmpty v-else-if="visibleItems.length === 0" :title="emptyState.title" :description="emptyState.description" icon="Boxes" class="py-20" />

        <div v-else class="studio-table-wrap custom-scrollbar">
          <table class="studio-table">
            <thead>
              <tr>
                <th class="col-identity">题目资源</th>
                <th class="col-meta">可见性</th>
                <th class="col-meta">分值</th>
                <th class="col-meta">权重</th>
                <template v-if="isAwdContest">
                  <th class="col-awd">裁判引擎</th>
                  <th class="col-awd">就绪校验</th>
                  <th class="col-awd">A/D 权重</th>
                  <th class="col-awd">最后活动</th>
                </template>
                <th class="col-actions">管理</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="challenge in visibleItems" :key="challenge.id" class="studio-row">
                <td class="col-identity">
                  <div class="challenge-identity">
                    <div class="challenge-title">{{ getChallengeTitle(challenge) }}</div>
                    <div class="challenge-subtitle">{{ challenge.category || '通用' }} · {{ challenge.difficulty || '常规' }}</div>
                  </div>
                </td>
                <td class="col-meta"><span class="status-badge" :class="challenge.is_visible ? 'is-visible' : 'is-hidden'">{{ challenge.is_visible ? '公开' : '隐藏' }}</span></td>
                <td class="col-meta font-mono font-black text-slate-700">{{ challenge.points }} <small>PTS</small></td>
                <td class="col-meta"><div class="order-chip">RANK {{ challenge.order }}</div></td>
                <template v-if="isAwdContest">
                  <td class="col-awd"><div class="engine-tag">{{ getCheckerLabel(challenge) }}</div></td>
                  <td class="col-awd"><span class="validation-status" :class="challenge.awd_checker_validation_state">{{ getValidationSummary(challenge) }}</span></td>
                  <td class="col-awd font-mono text-[10px] text-slate-500">{{ getAwdScoreSummary(challenge) }}</td>
                  <td class="col-awd text-[10px] text-slate-400">{{ getPreviewSummary(challenge) }}</td>
                </template>
                <td class="col-actions">
                  <CActionMenu :open="openActionMenuId === challenge.id" @update:open="setActionMenuOpen(challenge.id, $event)">
                    <template #trigger="{ open, toggle, setTriggerRef }"><button :ref="setTriggerRef" class="action-trigger" @click.stop="toggle"><MoreHorizontal class="h-4 w-4" /></button></template>
                    <template #default="{ close }">
                      <button v-if="isAwdContest" class="menu-item" @click="handleOpenAwdConfig(challenge, close)"><Zap class="h-3.5 w-3.5 mr-2" /> 补 AWD 配置</button>
                      <button class="menu-item" @click="handleOpenEditDialog(challenge, close)"><Edit class="h-3.5 w-3.5 mr-2" /> 属性修改</button>
                      <div class="menu-divider"></div>
                      <button class="menu-item danger" @click="handleRemoveFromMenu(challenge, close)"><Trash class="h-3.5 w-3.5 mr-2" /> 移除题目</button>
                    </template>
                  </CActionMenu>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>

    <ContestChallengeEditorDialog
      :open="dialogOpen" :mode="dialogMode" :contest-mode="contestMode"
      :challenge-options="challengeCatalog" :template-options="templateCatalog"
      :existing-challenge-ids="existingChallengeIds" :draft="editingChallenge"
      :loading-challenge-catalog="loadingChallengeCatalog" :loading-template-catalog="loadingTemplateCatalog"
      :saving="saving" @update:open="dialogOpen = $event" @save="handleSave"
    />
  </section>
</template>

<style scoped>
.studio-orchestration { display: flex; flex-direction: column; gap: 1.5rem; padding: 1.5rem 2rem; background: #fdfdfd; }
.studio-pane-header { display: flex; justify-content: space-between; align-items: flex-end; }
.pane-title { font-size: 1.25rem; font-weight: 900; color: #0f172a; margin: 0; }
.pane-description { font-size: 13px; color: #64748b; margin: 0.5rem 0 0; max-width: 40rem; }
.header-actions { display: flex; gap: 0.75rem; }
.ops-btn { display: inline-flex; align-items: center; gap: 0.5rem; height: 2.25rem; padding: 0 1rem; border-radius: 0.75rem; font-size: 12px; font-weight: 700; cursor: pointer; transition: all 0.2s ease; }
.ops-btn--neutral { background: white; border: 1px solid #e2e8f0; color: #475569; }
.ops-btn--primary { background: #2563eb; color: white; border: none; box-shadow: 0 4px 12px rgba(37,99,235,0.15); }
.studio-metric-band { display: flex; gap: 0.75rem; padding: 1rem; background: #f1f5f9; border: 1px solid #e2e8f0; border-radius: 1rem; }
.metric-pill { background: white; border: 1px solid #e2e8f0; padding: 0.5rem 1rem; border-radius: 0.75rem; display: flex; align-items: baseline; gap: 0.75rem; }
.metric-pill__label { font-size: 9px; font-weight: 800; text-transform: uppercase; color: #64748b; letter-spacing: 0.05em; }
.metric-pill__value { font-size: 14px; font-weight: 900; color: #0f172a; font-family: var(--font-family-mono); }
.studio-quick-nav { display: flex; gap: 0.5rem; margin-bottom: 0.5rem; }
.nav-pill { padding: 0.45rem 1rem; border-radius: 999px; background: white; border: 1px solid #e2e8f0; font-size: 12px; font-weight: 700; color: #64748b; cursor: pointer; display: flex; align-items: center; gap: 0.5rem; }
.nav-pill.active { background: #eff6ff; border-color: #3b82f6; color: #2563eb; }
.nav-pill__count { background: rgba(0,0,0,0.05); padding: 0 0.35rem; border-radius: 4px; font-size: 10px; }
.studio-table-wrap { border: none; border-radius: 0; background: transparent; overflow-x: auto; }
.studio-table { width: 100%; border-collapse: collapse; background: white; }
.studio-table th { background: #f8fafc; padding: 0.75rem 1rem; text-align: left; font-size: 10px; font-weight: 800; text-transform: uppercase; color: #94a3b8; border-bottom: 1px solid #e2e8f0; border-top: 1px solid #e2e8f0; }
.studio-table td { padding: 1rem; border-bottom: 1px solid #f1f5f9; }
.studio-row:hover { background: #f8fafc; }
.challenge-title { font-size: 14px; font-weight: 800; color: #1e293b; }
.challenge-subtitle { font-size: 11px; color: #94a3b8; margin-top: 0.15rem; }
.status-badge { padding: 0.15rem 0.5rem; border-radius: 6px; font-size: 10px; font-weight: 800; }
.is-visible { background: #f0fdf4; color: #166534; }
.is-hidden { background: #f1f5f9; color: #475569; }
.order-chip { font-size: 10px; font-weight: 900; color: #3b82f6; background: #eff6ff; padding: 0.25rem 0.5rem; border-radius: 4px; display: inline-block; }
.engine-tag { font-size: 11px; font-weight: 700; color: #475569; }
.validation-status { font-size: 10px; font-weight: 700; }
.validation-status.valid { color: #16a34a; }
.validation-status.invalid { color: #dc2626; }
.validation-status.pending { color: #d97706; }
.action-trigger { width: 2rem; height: 2rem; display: flex; align-items: center; justify-content: center; border-radius: 0.5rem; color: #94a3b8; transition: all 0.2s ease; cursor: pointer; }
.action-trigger:hover { background: #e2e8f0; color: #0f172a; }
.menu-item { width: 100%; padding: 0.5rem 1rem; display: flex; align-items: center; font-size: 12px; font-weight: 600; color: #475569; background: transparent; cursor: pointer; }
.menu-item:hover { background: #f1f5f9; }
.menu-item.danger { color: #ef4444; }
.menu-divider { height: 1px; background: #e2e8f0; margin: 0.25rem 0; }
</style>
