<script setup lang="ts">
import { MoreHorizontal, Trash } from 'lucide-vue-next'

import type { AdminContestChallengeViewData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import CActionMenu from '@/components/common/menus/CActionMenu.vue'
import {
  ChallengeCategoryPill,
  ChallengeDifficultyText,
  toChallengeCategory,
  toChallengeDifficulty,
} from '@/entities/challenge'

defineProps<{
  items: AdminContestChallengeViewData[]
  loading: boolean
  loadError: string
  emptyState: {
    title: string
    description: string
  }
  showChallengeOverflowMenu: boolean
  openActionMenuId: string | null
  removingChallengeId: string | null
}>()

const emit = defineEmits<{
  refresh: []
  edit: [challenge: AdminContestChallengeViewData]
  remove: [challenge: AdminContestChallengeViewData]
  'update:openActionMenuId': [value: string | null]
}>()

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
  <div class="studio-directory-canvas">
    <AppEmpty
      v-if="loadError && items.length === 0"
      title="赛事题目暂时不可用"
      :description="loadError"
      icon="AlertTriangle"
      class="py-20"
    >
      <template #action>
        <button
          type="button"
          class="ui-btn ui-btn--ghost"
          @click="emit('refresh')"
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
        v-if="loading"
        class="flex justify-center py-24"
      >
        <AppLoading>同步中...</AppLoading>
      </div>
      <AppEmpty
        v-else-if="items.length === 0"
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
              v-for="challenge in items"
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
                    <ChallengeCategoryPill
                      v-if="toChallengeCategory(challenge.category)"
                      :category="toChallengeCategory(challenge.category)!"
                    />
                    <span v-else>{{ challenge.category || '通用' }}</span>
                    <ChallengeDifficultyText
                      v-if="toChallengeDifficulty(challenge.difficulty)"
                      :difficulty="toChallengeDifficulty(challenge.difficulty)!"
                    />
                    <span v-else>{{ challenge.difficulty || '常规' }}</span>
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
                    @click="emit('edit', challenge)"
                  >
                    编辑
                  </button>
                  <CActionMenu
                    v-if="showChallengeOverflowMenu"
                    :open="openActionMenuId === challenge.id"
                    title="Challenge Actions"
                    menu-label="题目更多操作"
                    @update:open="emit('update:openActionMenuId', $event ? challenge.id : null)"
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
                        @click="close(); emit('remove', challenge)"
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
                    @click="emit('remove', challenge)"
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
</template>

<style scoped>
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

.challenge-identity {
  min-width: 0;
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
  display: inline-flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2);
  margin-top: var(--space-1);
  font-size: var(--font-size-13);
  color: var(--color-text-muted);
}

.contest-points-cell {
  font-family: var(--font-family-mono);
  font-weight: 900;
  color: color-mix(in srgb, var(--journal-ink) 82%, var(--journal-muted));
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
</style>
