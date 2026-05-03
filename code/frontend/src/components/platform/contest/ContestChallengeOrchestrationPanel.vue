<script setup lang="ts">
import { toRef } from 'vue'
import { RouterLink } from 'vue-router'
import { Plus, RefreshCw, Trash, Boxes, AlertTriangle, MoreHorizontal } from 'lucide-vue-next'

import type {
  AdminContestChallengeViewData,
  ContestDetailData,
} from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import CActionMenu from '@/components/common/menus/CActionMenu.vue'
import { useContestChallengeOrchestration } from '@/features/contest-workbench'

import ContestChallengeEditorDialog from './ContestChallengeEditorDialog.vue'
import ContestChallengeFilterStrip from './ContestChallengeFilterStrip.vue'
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

const {
  visibleItems,
  summaryItems,
  filterItems,
  activeFilter,
  isAwdContest,
  setFilter,
  showAwdChallengeFilters,
  showChallengeOverflowMenu,
  panelCopy,
  panelLoading,
  panelLoadError,
  currentChallengeLinks,
  emptyState,
  existingChallengeIds,
  awdChallengeFilters,
  awdChallengeCatalog,
  awdChallengePage,
  awdChallengePageSize,
  awdChallengeTotal,
  loadingAwdChallengeCatalog,
  awdChallengeLoadError,
  refreshAwdChallengeCatalog,
  changeAwdChallengePage,
  setAwdChallengeKeyword,
  setAwdChallengeServiceType,
  setAwdChallengeDeploymentMode,
  setAwdChallengeReadiness,
  dialogChallengeOptions,
  dialogOpen,
  dialogMode,
  editingChallenge,
  loadingChallengeCatalog,
  saving,
  removingChallengeId,
  openActionMenuId,
  refresh,
  handleCreateAction,
  openEditDialog,
  handleSave,
  handleRemove,
} = useContestChallengeOrchestration({
  contestId: toRef(props, 'contestId'),
  contestMode: toRef(props, 'contestMode'),
  challengeLinks: toRef(props, 'challengeLinks'),
  loadingExternal: toRef(props, 'loadingExternal'),
  loadErrorExternal: toRef(props, 'loadErrorExternal'),
  createDialogRequestKey: toRef(props, 'createDialogRequestKey'),
  onUpdated: () => emit('updated'),
})

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

    <ContestChallengeFilterStrip
      v-if="showAwdChallengeFilters && isAwdContest && filterItems.length > 0"
      :filter-items="filterItems"
      :active-filter="activeFilter"
      @select="setFilter"
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
                    <CActionMenu
                      v-if="showChallengeOverflowMenu"
                      :open="openActionMenuId === challenge.id"
                      title="Challenge Actions"
                      menu-label="题目更多操作"
                      @update:open="openActionMenuId = $event ? challenge.id : null"
                    >
                      <template #trigger="{ open, toggle, setTriggerRef }">
                        <button
                          :ref="setTriggerRef"
                          type="button"
                          class="c-action-menu__trigger c-action-menu__trigger--icon"
                          :aria-expanded="open ? 'true' : 'false'"
                          aria-haspopup="menu"
                          aria-label="题目更多操作"
                          @click.stop="toggle"
                        >
                          <MoreHorizontal class="h-3.5 w-3.5" />
                        </button>
                      </template>

                      <template #default="{ close }">
                        <button
                          :id="`contest-challenge-remove-${getChallengeActionKey(challenge)}`"
                          type="button"
                          class="c-action-menu__item c-action-menu__item--danger"
                          :disabled="removingChallengeId === challenge.id"
                          @click="close(); void handleRemove(challenge)"
                        >
                          <Trash class="h-3.5 w-3.5" />
                          移除
                        </button>
                      </template>
                    </CActionMenu>
                    <button
                      :id="`contest-challenge-remove-${getChallengeActionKey(challenge)}`"
                      type="button"
                      class="ui-btn ui-btn--sm ui-btn--danger"
                      :disabled="removingChallengeId === challenge.id"
                      @click="handleRemove(challenge)"
                    >
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
