<script setup lang="ts">
import { computed, onMounted, ref, toRef, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { Plus, RefreshCw, Trash, Boxes, AlertTriangle } from 'lucide-vue-next'

import {
  createContestAWDService,
  createAdminContestChallenge,
  listAdminContestChallenges,
  listContestAWDServices,
  deleteContestAWDService,
  deleteAdminContestChallenge,
  updateContestAWDService,
  updateAdminContestChallenge,
} from '@/api/admin/contests'
import { getChallenges } from '@/api/admin/authoring'
import type {
  AdminAwdChallengeData,
  AdminChallengeListItem,
  AdminContestChallengeViewData,
  ContestDetailData,
} from '@/api/contracts'
import { ApiError } from '@/api/request'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import { useContestAwdChallengePicker } from '@/composables/useContestAwdChallengePicker'
import { useContestChallengePool } from '@/composables/useContestChallengePool'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useToast } from '@/composables/useToast'
import {
  mapPlatformContestAwdServicesToChallengeLinks,
} from '@/utils/platformContestAwdChallengeLinks'

import ContestChallengeEditorDialog from './ContestChallengeEditorDialog.vue'
import ContestChallengeSummaryStrip from './ContestChallengeSummaryStrip.vue'

const props = defineProps<{
  contestId: string
  contestMode: ContestDetailData['mode']
  challengeLinks?: AdminContestChallengeViewData[]
  loadingExternal?: boolean
  loadErrorExternal?: string
  createDialogRequestKey?: number
}>()

const emit = defineEmits<{
  updated: []
}>()

const toast = useToast()
const CHALLENGE_CATALOG_PAGE_SIZE = 100
const AWD_CHALLENGE_PAGE_SIZE = 20
const loading = ref(true)
const saving = ref(false)
const loadingChallengeCatalog = ref(false)
const localChallengeLinks = ref<AdminContestChallengeViewData[]>([])
const localLoadError = ref('')
const challengeCatalog = ref<AdminChallengeListItem[]>([])
const dialogOpen = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const editingChallenge = ref<AdminContestChallengeViewData | null>(null)
const removingChallengeId = ref<string | null>(null)
const usingExternalChallengeLinks = computed(() => props.challengeLinks !== undefined)
const currentChallengeLinks = computed(() => props.challengeLinks ?? localChallengeLinks.value)
const panelLoading = computed(() => (usingExternalChallengeLinks.value ? Boolean(props.loadingExternal) : loading.value))
const panelLoadError = computed(() =>
  usingExternalChallengeLinks.value ? props.loadErrorExternal?.trim() ?? '' : localLoadError.value
)

const {
  visibleItems,
  summaryItems,
  isAwdContest,
} = useContestChallengePool(currentChallengeLinks, toRef(props, 'contestMode'))

const panelCopy = computed(() =>
  isAwdContest.value
    ? '维护统一题目池，从 AWD 题库选题并完成比赛级分值编排。'
    : '维护统一题目池，安排题目顺序、分值和可见状态。'
)
const emptyState = computed(() => ({
  title: '暂无关联题目',
  description: '先从题库里关联题目，再安排顺序。',
}))

const existingChallengeIdSet = computed(
  () => new Set(currentChallengeLinks.value.map((item) => String(item.challenge_id)))
)
const existingChallengeIds = computed(() => Array.from(existingChallengeIdSet.value))
const {
  filters: awdChallengeFilters,
  list: awdChallengeCatalog,
  total: awdChallengeTotal,
  page: awdChallengePage,
  pageSize: awdChallengePageSize,
  loading: loadingAwdChallengeCatalog,
  loadError: awdChallengeLoadError,
  refresh: refreshAwdChallengeCatalog,
  changePage: changeAwdChallengePage,
  setKeyword: setAwdChallengeKeyword,
  setServiceType: setAwdChallengeServiceType,
  setDeploymentMode: setAwdChallengeDeploymentMode,
  setReadinessStatus: setAwdChallengeReadiness,
} = useContestAwdChallengePicker({
  existingChallengeIds,
  pageSize: AWD_CHALLENGE_PAGE_SIZE,
})
const dialogChallengeOptions = computed(() =>
  dialogMode.value === 'edit'
    ? challengeCatalog.value
    : challengeCatalog.value.filter((item) => !existingChallengeIdSet.value.has(String(item.id)))
)

function getChallengeTitle(item: AdminContestChallengeViewData): string {
  return item.title?.trim() || `Challenge #${item.challenge_id}`
}

function getChallengePreviewRoute(item: AdminContestChallengeViewData) {
  return {
    name: 'PlatformChallengeDetail',
    params: { id: item.challenge_id },
  }
}

function getChallengeActionKey(item: AdminContestChallengeViewData): string {
  return item.challenge_id
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
      const nextAwdServices = await listContestAWDServices(props.contestId)
      localChallengeLinks.value = mapPlatformContestAwdServicesToChallengeLinks(nextAwdServices)
    } else {
      localChallengeLinks.value = await listAdminContestChallenges(props.contestId)
    }
    localLoadError.value = ''
  } catch (error) {
    localLoadError.value = humanizeRequestError(error, '加载失败')
  } finally {
    loading.value = false
  }
}

async function ensureChallengeCatalogLoaded() {
  if (loadingChallengeCatalog.value || challengeCatalog.value.length > 0) return
  loadingChallengeCatalog.value = true
  try {
    const result = await getChallenges(
      {
        page: 1,
        page_size: CHALLENGE_CATALOG_PAGE_SIZE,
        status: 'published',
      }
    )
    challengeCatalog.value = result.list
  } catch (error) {
    toast.error(humanizeRequestError(error, '题库加载失败'))
  } finally {
    loadingChallengeCatalog.value = false
  }
}

function openCreateDialog() {
  dialogMode.value = 'create'
  editingChallenge.value = null
  dialogOpen.value = true
  if (isAwdContest.value) {
    void changeAwdChallengePage(1)
  } else {
    void ensureChallengeCatalogLoaded()
  }
}

function handleCreateAction() {
  openCreateDialog()
}

function openEditDialog(challenge: AdminContestChallengeViewData) {
  dialogMode.value = 'edit'
  editingChallenge.value = challenge
  dialogOpen.value = true
  if (isAwdContest.value) void refreshAwdChallengeCatalog()
}

function closeDialog() {
  dialogOpen.value = false
  editingChallenge.value = null
}

interface ContestOrchestrationSavePayload {
  challenge_id?: number
  awd_challenge_id?: number
  awd_challenge_ids?: number[]
  points: number
  order: number
  is_visible: boolean
  awd_checker_type?: AdminContestChallengeViewData['awd_checker_type']
  awd_checker_config?: Record<string, unknown>
  awd_sla_score?: number
  awd_defense_score?: number
  awd_checker_preview_token?: string
}

function summarizeAwdChallengeFailures(awdChallengeIds: number[]): string {
  const failedNames = awdChallengeIds
    .map((awdChallengeId) => awdChallengeCatalog.value.find((item) => Number(item.id) === awdChallengeId)?.name || `AWD #${awdChallengeId}`)

  return `部分 AWD 题目关联失败：${failedNames.join('、')}`
}

function buildAwdServiceCreatePayload(
  awdChallengeId: number,
  payload: ContestOrchestrationSavePayload,
  order: number
) {
  const awdChallenge = awdChallengeCatalog.value.find((item) => Number(item.id) === awdChallengeId)
  const checkerConfig =
    awdChallenge?.checker_config && typeof awdChallenge.checker_config === 'object'
      ? awdChallenge.checker_config
      : undefined

  return {
    awd_challenge_id: awdChallengeId,
    points: payload.points,
    order,
    is_visible: payload.is_visible,
    ...(awdChallenge?.checker_type ? { checker_type: awdChallenge.checker_type } : {}),
    ...(checkerConfig ? { checker_config: checkerConfig } : {}),
  }
}

async function handleSave(payload: ContestOrchestrationSavePayload) {
  saving.value = true
  try {
    if (isAwdContest.value) {
      const awdChallengeIds =
        dialogMode.value === 'create' && payload.awd_challenge_ids?.length
          ? payload.awd_challenge_ids
          : payload.awd_challenge_id
            ? [payload.awd_challenge_id]
            : []

      if (awdChallengeIds.length === 0) {
        toast.error('请选择 AWD 题目')
        return
      }
      if (dialogMode.value === 'create') {
        const results = await Promise.allSettled(
          awdChallengeIds.map((awdChallengeId, index) =>
            createContestAWDService(
              props.contestId,
              buildAwdServiceCreatePayload(awdChallengeId, payload, payload.order + index)
            )
          )
        )
        const failedResults = results.flatMap((result, index) =>
          result.status === 'rejected'
            ? [
                {
                  awdChallengeId: awdChallengeIds[index],
                  error: result.reason,
                },
              ]
            : []
        )

        if (failedResults.length > 0) {
          const failedIds = failedResults.map(({ awdChallengeId }) => awdChallengeId)
          const failureMessage = summarizeAwdChallengeFailures(failedIds)

          if (failedResults.length === awdChallengeIds.length) {
            toast.error(failureMessage)
            return
          }

          toast.warning(failureMessage)
          emit('updated')
          if (!usingExternalChallengeLinks.value) {
            await refresh()
          }
          return
        }
      } else if (editingChallenge.value) {
        await updateContestAWDService(
          props.contestId,
          editingChallenge.value.awd_service_id!,
          {
            awd_challenge_id: awdChallengeIds[0],
            points: payload.points,
            order: payload.order,
            is_visible: payload.is_visible,
          }
        )
      }
    } else if (dialogMode.value === 'create') {
      await createAdminContestChallenge(
        props.contestId,
        {
          challenge_id: payload.challenge_id!,
          points: payload.points,
          order: payload.order,
          is_visible: payload.is_visible,
        }
      )
    } else if (editingChallenge.value) {
      await updateAdminContestChallenge(
        props.contestId,
        editingChallenge.value.challenge_id,
        {
          points: payload.points,
          order: payload.order,
          is_visible: payload.is_visible,
        }
      )
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

onMounted(() => {
  if (!usingExternalChallengeLinks.value) void refresh()
})

watch(
  () => props.createDialogRequestKey,
  (requestKey, previousRequestKey) => {
    if (!requestKey || requestKey === previousRequestKey) return
    handleCreateAction()
  },
  { immediate: true }
)

watch(awdChallengeLoadError, (error, previousError) => {
  if (!error || error === previousError) {
    return
  }
  toast.error(error)
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
          @click="handleCreateAction"
        >
          <Plus class="h-3.5 w-3.5" />
          <span>{{ isAwdContest ? '新增服务' : '关联新题目' }}</span>
        </button>
      </div>
    </header>

    <ContestChallengeSummaryStrip
      v-if="!isAwdContest && summaryItems.length > 0"
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

      <div
        v-else
        class="studio-directory-stack"
      >
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
                  顺序
                </th>
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
                    <RouterLink
                      :id="`contest-challenge-preview-${getChallengeActionKey(challenge)}`"
                      class="challenge-title challenge-title-link"
                      :to="getChallengePreviewRoute(challenge)"
                      :title="`打开题目详情：${getChallengeTitle(challenge)}`"
                    >
                      {{ getChallengeTitle(challenge) }}
                    </RouterLink>
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
                <td class="col-actions">
                  <div
                    class="ui-row-actions contest-challenge-row__actions"
                    role="group"
                    aria-label="题目编排操作"
                  >
                    <button
                      :id="`contest-challenge-edit-${getChallengeActionKey(challenge)}`"
                      type="button"
                      class="ui-btn ui-btn--sm ui-btn--secondary ui-row-action--default"
                      @click="openEditDialog(challenge)"
                    >
                      编辑
                    </button>
                    <button
                      :id="`contest-challenge-remove-${getChallengeActionKey(challenge)}`"
                      type="button"
                      class="ui-btn ui-btn--sm ui-btn--danger"
                      :disabled="removingChallengeId === challenge.id"
                      @click="handleRemove(challenge)"
                    >
                      <Trash class="h-3.5 w-3.5" />
                      移除
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <ContestChallengeEditorDialog
      :key="`${dialogMode}:${existingChallengeIds.join(',')}`"
      :open="dialogOpen"
      :mode="dialogMode"
      :contest-mode="contestMode"
      :challenge-options="dialogChallengeOptions"
      :awd-challenge-options="awdChallengeCatalog"
      :awd-challenge-page="awdChallengePage"
      :awd-challenge-page-size="awdChallengePageSize"
      :awd-challenge-total="awdChallengeTotal"
      :awd-challenge-keyword="awdChallengeFilters.keyword"
      :awd-challenge-service-type="awdChallengeFilters.serviceType"
      :awd-challenge-deployment-mode="awdChallengeFilters.deploymentMode"
      :awd-challenge-readiness="awdChallengeFilters.readinessStatus"
      :awd-challenge-load-error="awdChallengeLoadError"
      :existing-challenge-ids="existingChallengeIds"
      :draft="editingChallenge"
      :loading-challenge-catalog="loadingChallengeCatalog"
      :loading-awd-challenge-catalog="loadingAwdChallengeCatalog"
      :saving="saving"
      @update:open="dialogOpen = $event"
      @update-awd-challenge-keyword="setAwdChallengeKeyword"
      @update-awd-challenge-service-type="setAwdChallengeServiceType"
      @update-awd-challenge-deployment-mode="setAwdChallengeDeploymentMode"
      @update-awd-challenge-readiness="setAwdChallengeReadiness"
      @change-awd-challenge-page="changeAwdChallengePage"
      @refresh-awd-challenge-catalog="refreshAwdChallengeCatalog"
      @save="handleSave"
    />
  </section>
</template>

<style scoped>
.studio-orchestration {
  display: flex;
  flex-direction: column;
  gap: var(--space-section-gap);
  background: transparent;
  padding: var(--space-6) var(--space-8);
}
.studio-pane-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-4);
}

.pane-title {
  margin: 0;
  font-size: var(--font-size-20);
  font-weight: 900;
  color: var(--color-text-primary);
}

.pane-description {
  margin: var(--space-2) 0 0;
  max-width: var(--ui-selector-width-lg);
  font-size: var(--font-size-13);
  color: var(--color-text-secondary);
}

.header-actions {
  display: flex;
  gap: var(--space-3);
}

.studio-directory-stack {
  display: flex;
  flex-direction: column;
  gap: var(--space-section-gap-compact);
}

.studio-table-wrap {
  overflow-x: auto;
  border: 1px solid color-mix(in srgb, var(--workspace-line-soft) 86%, transparent);
  border-radius: var(--ui-control-radius-lg);
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--color-bg-surface) 94%, var(--color-bg-base)),
      color-mix(in srgb, var(--color-bg-surface) 84%, var(--color-bg-base))
    );
  box-shadow: 0 var(--space-2) var(--space-5)
    color-mix(in srgb, var(--color-shadow-soft) 24%, transparent);
}

.studio-table {
  width: 100%;
  border-collapse: collapse;
}

.studio-table th {
  border-bottom: 1px solid color-mix(in srgb, var(--workspace-line-soft) 86%, transparent);
  background: color-mix(in srgb, var(--color-bg-surface) 72%, var(--color-bg-base));
  padding: var(--space-4);
  text-align: left;
  font-size: var(--font-size-11);
  font-weight: 800;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.studio-table td {
  border-bottom: 1px solid var(--color-border-subtle);
  padding: var(--space-5) var(--space-4);
}

.studio-table .col-actions {
  text-align: right;
}

.studio-table tbody tr:last-child td {
  border-bottom: 0;
}

.studio-row {
  transition: background var(--ui-motion-fast);
}

.studio-row:hover {
  background: color-mix(in srgb, var(--color-primary-soft) 24%, var(--color-bg-surface));
}

.challenge-title {
  display: inline-block;
  max-width: 100%;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-16);
  font-weight: 800;
  color: var(--color-text-primary);
}

.challenge-title-link {
  text-decoration: none;
  transition:
    color var(--ui-motion-fast),
    text-decoration-color var(--ui-motion-fast);
}

.challenge-title-link:hover {
  color: var(--color-primary);
  text-decoration: underline;
  text-decoration-thickness: var(--ui-focus-ring-width);
  text-underline-offset: var(--space-1);
}

.challenge-title-link:focus-visible {
  outline: var(--ui-focus-ring-width) solid
    color-mix(in srgb, var(--color-primary) 72%, transparent);
  outline-offset: var(--space-1);
  border-radius: var(--ui-control-radius-sm);
}

.challenge-subtitle {
  margin-top: var(--space-1);
  font-size: var(--font-size-13);
  color: var(--color-text-muted);
}

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

.status-badge {
  border-radius: var(--ui-badge-radius-soft);
  padding: var(--space-1) var(--space-2);
  font-size: var(--font-size-11);
  font-weight: 800;
}

.is-visible {
  background: var(--color-success-soft);
  color: var(--color-success);
}

.is-hidden {
  background: var(--color-bg-elevated);
  color: var(--color-text-secondary);
}

.order-chip {
  display: inline-block;
  border-radius: var(--ui-badge-radius-soft);
  background: var(--color-primary-soft);
  padding: var(--space-1) var(--space-2);
  font-size: var(--font-size-11);
  font-weight: 900;
  color: var(--color-primary);
}

.engine-tag {
  font-size: var(--font-size-13);
  font-weight: 700;
  color: var(--color-text-secondary);
}

.validation-status {
  font-size: var(--font-size-11);
  font-weight: 700;
}

.validation-status.valid {
  color: var(--color-success);
}

.validation-status.invalid {
  color: var(--color-danger);
}

.validation-status.pending {
  color: var(--color-warning);
}

.menu-divider {
  border-top: 1px solid var(--color-border-default);
  margin: var(--space-1) 0;
}

</style>
