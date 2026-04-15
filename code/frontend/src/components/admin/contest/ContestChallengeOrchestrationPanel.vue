<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Plus, RefreshCw } from 'lucide-vue-next'

import {
  createAdminContestChallenge,
  deleteAdminContestChallenge,
  getChallenges,
  listAdminContestChallenges,
  updateAdminContestChallenge,
} from '@/api/admin'
import type {
  AdminChallengeListItem,
  AdminContestChallengeData,
  ContestDetailData,
} from '@/api/contracts'
import { ApiError } from '@/api/request'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useToast } from '@/composables/useToast'

import ContestChallengeEditorDialog from './ContestChallengeEditorDialog.vue'

const props = defineProps<{
  contestId: string
  contestMode: ContestDetailData['mode']
}>()

const toast = useToast()
const CHALLENGE_CATALOG_PAGE_SIZE = 100
const loading = ref(true)
const saving = ref(false)
const loadingChallengeCatalog = ref(false)
const challengeLinks = ref<AdminContestChallengeData[]>([])
const challengeCatalog = ref<AdminChallengeListItem[]>([])
const dialogOpen = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const editingChallenge = ref<AdminContestChallengeData | null>(null)
const removingChallengeId = ref<string | null>(null)

const sortedChallengeLinks = computed(() =>
  [...challengeLinks.value].sort(
    (left, right) => left.order - right.order || left.challenge_id.localeCompare(right.challenge_id)
  )
)

const summaryItems = computed(() => [
  {
    key: 'total',
    label: '已关联题目',
    value: String(challengeLinks.value.length),
    hint: '当前竞赛中已经挂载的题目数量',
  },
  {
    key: 'visible',
    label: '对选手可见',
    value: String(challengeLinks.value.filter((item) => item.is_visible).length),
    hint: '进入竞赛详情后默认可见的题目数量',
  },
  {
    key: 'hidden',
    label: '暂时隐藏',
    value: String(challengeLinks.value.filter((item) => !item.is_visible).length),
    hint: '已关联但不会直接展示给选手的题目数量',
  },
])

const panelCopy = computed(() =>
  props.contestMode === 'awd'
    ? '先在这里关联赛事题目、整理顺序和基础分值；AWD 的 Checker、SLA 与防守分继续在 AWD 运维面板维护。'
    : '在这里关联竞赛题目，安排展示顺序、基础分值和可见状态。'
)

const existingChallengeIds = computed(() => challengeLinks.value.map((item) => item.challenge_id))

function getChallengeTitle(item: AdminContestChallengeData): string {
  return item.title?.trim() || `Challenge #${item.challenge_id}`
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
  loading.value = true
  try {
    challengeLinks.value = await listAdminContestChallenges(props.contestId)
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
    await refresh()
  } catch (error) {
    toast.error(
      humanizeRequestError(error, dialogMode.value === 'create' ? '关联题目失败' : '更新题目失败')
    )
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
    await refresh()
  } catch (error) {
    toast.error(humanizeRequestError(error, '移除题目失败'))
  } finally {
    removingChallengeId.value = null
  }
}

onMounted(() => {
  void refresh()
})
</script>

<template>
  <section class="contest-challenge-panel">
    <header class="contest-challenge-panel__header">
      <div class="workspace-tab-heading__main">
        <div class="workspace-overline">Contest Challenges</div>
        <h1 class="workspace-page-title">题目编排</h1>
        <p class="workspace-page-copy">
          {{ panelCopy }}
        </p>
      </div>

      <div class="contest-challenge-panel__actions">
        <button type="button" class="admin-btn admin-btn-ghost" @click="refresh">
          <RefreshCw class="h-4 w-4" />
          刷新列表
        </button>
        <button
          id="contest-challenge-add"
          type="button"
          class="admin-btn admin-btn-primary"
          @click="openCreateDialog"
        >
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
          <h2 class="list-heading__title">已关联题目</h2>
        </div>
        <div class="contest-section-meta">共 {{ challengeLinks.length }} 道题目</div>
      </header>

      <div v-if="loading" class="contest-challenge-directory__loading">
        <AppLoading>正在同步赛事题目...</AppLoading>
      </div>

      <AppEmpty
        v-else-if="sortedChallengeLinks.length === 0"
        title="当前竞赛还没有关联题目"
        description="先从题库里关联题目，再安排顺序、分值和可见状态。"
        icon="FileChartColumnIncreasing"
      />

      <template v-else>
        <div class="contest-challenge-directory__head" aria-hidden="true">
          <span>题目</span>
          <span>可见性</span>
          <span>分值</span>
          <span>顺序</span>
          <span class="contest-challenge-directory__actions-label">操作</span>
        </div>

        <article
          v-for="challenge in sortedChallengeLinks"
          :key="challenge.id"
          class="contest-challenge-row"
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
          <div
            class="contest-challenge-row__actions"
            role="group"
            :aria-label="`题目 ${getChallengeTitle(challenge)} 操作`"
          >
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
  --admin-summary-grid-columns: repeat(3, minmax(0, 1fr));
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

.contest-challenge-directory__head,
.contest-challenge-row {
  display: grid;
  gap: var(--space-3);
  grid-template-columns: minmax(16rem, 1.7fr) minmax(6rem, 0.7fr) minmax(5rem, 0.55fr) minmax(
      5rem,
      0.55fr
    ) auto;
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
.contest-challenge-row__order {
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
