<script setup lang="ts">
import { computed, nextTick, ref, watch, type ComponentPublicInstance } from 'vue'
import { MoreHorizontal, Swords } from 'lucide-vue-next'

import type { ContestDetailData, ContestStatus } from '@/api/contracts'
import AdminPaginationControls from '@/components/admin/AdminPaginationControls.vue'
import { getModeLabel, getStatusLabel } from '@/utils/contest'

const props = defineProps<{
  contests: ContestDetailData[]
  page: number
  pageSize: number
  total: number
}>()

const emit = defineEmits<{
  edit: [contest: ContestDetailData]
  export: [contest: ContestDetailData]
  workbench: [contest: ContestDetailData]
  changePage: [page: number]
}>()

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const openActionMenuId = ref<string | null>(null)
const actionMenuPanelRef = ref<HTMLElement | null>(null)
const actionMenuStyle = ref<Record<string, string>>({})
const actionMenuButtonRefs = new Map<string, HTMLButtonElement>()
const activeActionContest = computed(
  () => props.contests.find((contest) => contest.id === openActionMenuId.value) ?? null
)

function formatTime(value: string): string {
  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function getStatusPillClass(status: ContestStatus): string {
  if (status === 'running') return 'contest-status-pill--running'
  if (status === 'registering') return 'contest-status-pill--registering'
  if (status === 'draft' || status === 'published') return 'contest-status-pill--draft'
  if (status === 'frozen') return 'contest-status-pill--frozen'
  if (status === 'ended' || status === 'archived') return 'contest-status-pill--ended'
  if (status === 'cancelled') return 'contest-status-pill--cancelled'
  return 'contest-status-pill--neutral'
}

function canEnterWorkbench(contest: ContestDetailData): boolean {
  return contest.mode === 'awd' && (contest.status === 'running' || contest.status === 'frozen')
}

function setActionMenuButtonRef(
  contestId: string,
  element: Element | ComponentPublicInstance | null
): void {
  if (element instanceof HTMLButtonElement) {
    actionMenuButtonRefs.set(contestId, element)
    return
  }

  actionMenuButtonRefs.delete(contestId)
}

function closeActionMenu(): void {
  openActionMenuId.value = null
}

function updateActionMenuPosition(): void {
  if (!openActionMenuId.value) {
    return
  }

  const trigger = actionMenuButtonRefs.get(openActionMenuId.value)
  if (!trigger) {
    return
  }

  const rect = trigger.getBoundingClientRect()
  const viewportPadding = 12
  const gap = 8
  const panelWidth = actionMenuPanelRef.value?.offsetWidth ?? 176
  const panelHeight = actionMenuPanelRef.value?.offsetHeight ?? 132
  const maxLeft = Math.max(viewportPadding, window.innerWidth - panelWidth - viewportPadding)
  const left = Math.min(Math.max(viewportPadding, rect.right - panelWidth), maxLeft)
  const spaceBelow = window.innerHeight - rect.bottom - viewportPadding
  const spaceAbove = rect.top - viewportPadding
  const shouldOpenUpward = spaceBelow < panelHeight + gap && spaceAbove > spaceBelow
  const maxTop = Math.max(viewportPadding, window.innerHeight - panelHeight - viewportPadding)
  const top = shouldOpenUpward
    ? Math.max(viewportPadding, rect.top - panelHeight - gap)
    : Math.min(rect.bottom + gap, maxTop)

  actionMenuStyle.value = {
    top: `${top}px`,
    left: `${left}px`,
    width: `${panelWidth}px`,
  }
}

async function toggleActionMenu(contestId: string): Promise<void> {
  const shouldOpen = openActionMenuId.value !== contestId
  openActionMenuId.value = shouldOpen ? contestId : null

  if (!shouldOpen) {
    actionMenuStyle.value = {}
    return
  }

  await nextTick()
  updateActionMenuPosition()
}

function handleEdit(contest: ContestDetailData): void {
  closeActionMenu()
  emit('edit', contest)
}

function handleExport(contest: ContestDetailData): void {
  closeActionMenu()
  emit('export', contest)
}

watch(openActionMenuId, async (contestId, _previousId, onCleanup) => {
  if (!contestId) {
    actionMenuStyle.value = {}
    return
  }

  await nextTick()
  updateActionMenuPosition()

  const handleViewportChange = () => {
    updateActionMenuPosition()
  }
  const handleEscape = (event: KeyboardEvent) => {
    if (event.key === 'Escape') {
      closeActionMenu()
    }
  }

  window.addEventListener('resize', handleViewportChange)
  window.addEventListener('scroll', handleViewportChange, true)
  window.addEventListener('keydown', handleEscape)

  onCleanup(() => {
    window.removeEventListener('resize', handleViewportChange)
    window.removeEventListener('scroll', handleViewportChange, true)
    window.removeEventListener('keydown', handleEscape)
  })
})
</script>

<template>
  <div class="space-y-5">
    <div class="contest-directory workspace-directory-list">
      <div class="contest-directory-head" aria-hidden="true">
        <span>竞赛</span>
        <span>模式</span>
        <span>状态</span>
        <span>开始时间</span>
        <span>结束时间</span>
        <span class="contest-directory-head__actions">操作</span>
      </div>

      <article v-for="contest in contests" :key="contest.id" class="contest-row">
        <div class="contest-row__identity">
          <h3 class="contest-row__title" :title="contest.title">{{ contest.title }}</h3>
          <p class="contest-row__description">
            {{ contest.description || '当前未填写竞赛描述。' }}
          </p>
        </div>

        <div class="contest-row__mode">{{ getModeLabel(contest.mode) }}</div>

        <div class="contest-row__status">
          <span class="ui-badge contest-status-pill" :class="getStatusPillClass(contest.status)">
            {{ getStatusLabel(contest.status) }}
          </span>
        </div>

        <div class="contest-row__starts-at">
          <p>{{ formatTime(contest.starts_at) }}</p>
        </div>

        <div class="contest-row__ends-at">
          <p>{{ formatTime(contest.ends_at) }}</p>
        </div>

        <div class="ui-row-actions contest-row__actions" role="group" aria-label="竞赛操作">
          <button
            v-if="canEnterWorkbench(contest)"
            :id="`contest-open-workbench-${contest.id}`"
            type="button"
            class="ui-btn ui-btn--sm ui-btn--primary contest-action contest-action--workbench"
            @click="emit('workbench', contest)"
          >
            <Swords class="h-3.5 w-3.5" />
            进入 AWD 赛区
          </button>
          <button
            :id="`contest-row-more-${contest.id}`"
            :ref="(element) => setActionMenuButtonRef(contest.id, element)"
            type="button"
            class="contest-row-menu-button"
            :aria-expanded="openActionMenuId === contest.id ? 'true' : 'false'"
            aria-haspopup="menu"
            aria-label="更多竞赛操作"
            :class="{ 'contest-row-menu-button--active': openActionMenuId === contest.id }"
            @click.stop="void toggleActionMenu(contest.id)"
          >
            <MoreHorizontal class="h-3.5 w-3.5" />
          </button>
        </div>
      </article>
    </div>

    <Teleport to="body">
      <div v-if="activeActionContest" class="contest-row-menu-layer" @click="closeActionMenu">
        <div
          ref="actionMenuPanelRef"
          class="contest-row-menu"
          :style="actionMenuStyle"
          role="menu"
          aria-label="更多竞赛操作"
          @click.stop
        >
          <div class="contest-row-menu__title">Contest Actions</div>
          <button
            :id="`contest-row-menu-edit-${activeActionContest.id}`"
            type="button"
            class="contest-row-menu__item"
            role="menuitem"
            @click="handleEdit(activeActionContest)"
          >
            编辑
          </button>
          <button
            :id="`contest-row-menu-export-${activeActionContest.id}`"
            type="button"
            class="contest-row-menu__item"
            role="menuitem"
            @click="handleExport(activeActionContest)"
          >
            导出结果
          </button>
        </div>
      </div>
    </Teleport>

    <div
      class="admin-pagination workspace-directory-pagination text-sm text-[var(--color-text-muted)]"
    >
      <AdminPaginationControls
        :page="page"
        :total-pages="totalPages"
        :total="total"
        :total-label="`共 ${total} 场竞赛`"
        @change-page="emit('changePage', $event)"
      />
    </div>
  </div>
</template>

<style scoped>
.contest-directory {
  --contest-directory-columns: minmax(17rem, 1.46fr) minmax(6rem, 0.54fr) minmax(7rem, 0.68fr)
    minmax(9.5rem, 0.78fr) minmax(9.5rem, 0.78fr) minmax(11rem, 11rem);
  display: grid;
  gap: 0;
}

.contest-directory-head {
  display: grid;
  grid-template-columns: var(--contest-directory-columns);
  gap: var(--space-4);
  padding: 0 0 var(--space-3);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.contest-directory-head > span {
  min-width: 0;
}

.contest-directory-head__actions {
  text-align: right;
}

.contest-row {
  display: grid;
  grid-template-columns: var(--contest-directory-columns);
  gap: var(--space-4);
  align-items: start;
  padding: var(--space-4) 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.contest-row > div {
  min-width: 0;
}

.contest-row__identity {
  display: grid;
  gap: var(--space-1-5);
}

.contest-row__title {
  min-width: 0;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-1-00);
  font-weight: 600;
  color: var(--journal-ink);
}

.contest-row__description {
  margin: 0;
  display: -webkit-box;
  overflow: hidden;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  font-size: var(--font-size-0-875);
  line-height: 1.55;
  color: var(--journal-muted);
}

.contest-row__mode,
.contest-row__starts-at,
.contest-row__ends-at {
  font-size: var(--font-size-0-90);
  color: var(--journal-muted);
}

.contest-row__starts-at p,
.contest-row__ends-at p {
  margin: 0;
  line-height: 1.45;
}

.contest-row__starts-at p {
  color: color-mix(in srgb, var(--journal-ink) 84%, var(--journal-muted));
}

.contest-row__ends-at p {
  color: color-mix(in srgb, var(--journal-muted) 88%, var(--journal-ink));
}

.contest-status-pill {
  --ui-badge-radius: 999px;
  --ui-badge-padding: 0.35rem 0.75rem;
  --ui-badge-size: var(--font-size-0-78);
  --ui-badge-spacing: 0.02em;
  line-height: 1;
}

.contest-status-pill--running {
  --ui-badge-border: color-mix(in srgb, #22d3ee 38%, transparent);
  --ui-badge-background: color-mix(in srgb, #22d3ee 16%, var(--journal-surface));
  --ui-badge-color: #67e8f9;
}

.contest-status-pill--registering {
  --ui-badge-border: color-mix(in srgb, #f59e0b 34%, transparent);
  --ui-badge-background: color-mix(in srgb, #f59e0b 15%, var(--journal-surface));
  --ui-badge-color: #fbbf24;
}

.contest-status-pill--draft {
  --ui-badge-border: color-mix(in srgb, #a78bfa 28%, transparent);
  --ui-badge-background: color-mix(in srgb, #a78bfa 12%, var(--journal-surface));
  --ui-badge-color: #c4b5fd;
}

.contest-status-pill--frozen {
  --ui-badge-border: color-mix(in srgb, #60a5fa 30%, transparent);
  --ui-badge-background: color-mix(in srgb, #60a5fa 13%, var(--journal-surface));
  --ui-badge-color: #93c5fd;
}

.contest-status-pill--ended {
  --ui-badge-border: color-mix(in srgb, #34d399 28%, transparent);
  --ui-badge-background: color-mix(in srgb, #34d399 12%, var(--journal-surface));
  --ui-badge-color: #6ee7b7;
}

.contest-status-pill--cancelled,
.contest-status-pill--neutral {
  --ui-badge-border: color-mix(in srgb, var(--journal-border) 84%, transparent);
  --ui-badge-background: color-mix(in srgb, var(--journal-surface) 92%, transparent);
  --ui-badge-color: color-mix(in srgb, var(--journal-muted) 92%, var(--journal-ink));
}

.contest-row__actions {
  justify-content: flex-end;
  flex-wrap: nowrap;
}

.contest-action {
  min-width: 5.25rem;
}

.contest-action--workbench {
  --ui-btn-primary-bg: color-mix(in srgb, var(--color-success) 78%, var(--journal-ink));
  --ui-btn-primary-border: color-mix(in srgb, var(--color-success) 56%, transparent);
  --ui-btn-primary-color: white;
  box-shadow: 0 10px 24px color-mix(in srgb, var(--color-success) 18%, transparent);
}

.contest-row-menu-button {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  width: 1.95rem;
  height: 1.95rem;
  border: 1px solid color-mix(in srgb, var(--journal-border) 84%, transparent);
  border-radius: 0.75rem;
  background: color-mix(in srgb, var(--journal-surface) 94%, white);
  color: var(--journal-muted);
  transition:
    background-color 160ms ease,
    border-color 160ms ease,
    color 160ms ease,
    box-shadow 160ms ease;
}

.contest-row-menu-button:hover,
.contest-row-menu-button--active {
  border-color: color-mix(in srgb, var(--workspace-brand) 26%, var(--journal-border));
  background: color-mix(in srgb, var(--workspace-brand) 8%, white);
  color: color-mix(in srgb, var(--workspace-brand) 86%, var(--journal-ink));
  box-shadow: 0 12px 26px color-mix(in srgb, var(--workspace-brand) 10%, transparent);
}

.contest-row-menu-layer {
  position: fixed;
  inset: 0;
  z-index: 120;
}

.contest-row-menu {
  position: fixed;
  z-index: 130;
  width: 11rem;
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--journal-border) 86%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, white 96%, var(--journal-surface));
  box-shadow:
    0 24px 60px rgba(15, 23, 42, 0.16),
    0 10px 24px rgba(15, 23, 42, 0.1);
}

.contest-row-menu__title {
  padding: 0.78rem 1rem 0.55rem;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 68%, transparent);
  background: color-mix(in srgb, var(--workspace-brand) 4%, white);
  font-size: 0.62rem;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.contest-row-menu__item {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: flex-start;
  padding: 0.72rem 1rem;
  font-size: var(--font-size-0-82);
  font-weight: 600;
  color: color-mix(in srgb, var(--journal-ink) 84%, var(--journal-muted));
  transition:
    background-color 160ms ease,
    color 160ms ease;
}

.contest-row-menu__item:hover {
  background: color-mix(in srgb, var(--workspace-brand) 6%, white);
  color: color-mix(in srgb, var(--workspace-brand) 90%, var(--journal-ink));
}

@media (max-width: 1023px) {
  .contest-directory-head {
    display: none;
  }

  .contest-row {
    grid-template-columns: 1fr;
    gap: var(--space-2-5);
    padding: var(--space-4) 0;
  }

  .contest-row__actions {
    justify-content: flex-start;
  }
}
</style>
