<template>
  <div class="contest-page-shell" :style="contestAccentStyle">
    <section
      class="workspace-shell journal-shell journal-shell-user journal-hero contest-detail-view flex min-h-full flex-1 flex-col"
    >
      <main v-if="loading" class="content-pane">
        <div class="contest-loading">
          <div class="contest-loading__spinner" />
          <div class="contest-loading__text">正在同步竞赛详情...</div>
        </div>
      </main>

      <template v-else-if="contest && contestAccessible">
        <ContestAnnouncementRealtimeBridge :contest-id="contest.id" @updated="refreshAnnouncements" />

        <div class="workspace-tabbar top-tabs" role="tablist" aria-label="竞赛页面主切换">
          <button
            v-for="(tab, index) in workspaceTabs"
            :id="`contest-workspace-tab-${tab.id}`"
            :key="tab.id"
            :ref="(element) => setTabButtonRef(tab.id, element as HTMLButtonElement | null)"
            type="button"
            role="tab"
            class="workspace-tab top-tab"
            :class="{ active: activeWorkspaceTab === tab.id }"
            :aria-selected="activeWorkspaceTab === tab.id"
            :aria-controls="`contest-workspace-panel-${tab.id}`"
            :tabindex="activeWorkspaceTab === tab.id ? 0 : -1"
            @click="selectWorkspaceTab(tab.id)"
            @keydown="handleWorkspaceTabKeydown($event, index)"
          >
            {{ tab.label }}
          </button>
        </div>

        <main class="content-pane">
          <ContestOverviewPanel
            v-if="activeWorkspaceTab === 'overview'"
            :contest="contest"
            :countdown="countdown"
            :total-points="totalPoints"
            :solved-count="solvedCount"
            :member-count="memberCount"
            :challenge-count="challenges.length"
            :announcements="announcements"
            :announcements-error="announcementsError"
          />

          <ContestAnnouncementsWorkspaceSection
            v-else-if="activeWorkspaceTab === 'announcements'"
            panel-id="contest-workspace-panel-announcements"
            tab-id="contest-workspace-tab-announcements"
            :announcements="announcements"
            :announcements-error="announcementsError"
          />

          <section
            v-else-if="activeWorkspaceTab === 'challenges'"
            id="contest-workspace-panel-challenges"
            class="workspace-panel"
            role="tabpanel"
            aria-labelledby="contest-workspace-tab-challenges"
          >
            <section class="contest-section">
              <div class="contest-section__head workspace-tab-heading">
                <div class="workspace-tab-heading__main">
                  <div class="workspace-overline">
                    {{ contest.mode === 'awd' ? 'Battlefield' : 'Challenges' }}
                  </div>
                  <h2 class="contest-section__title workspace-tab-heading__title">
                    {{ contest.mode === 'awd' ? '攻防战场' : '题目' }}
                  </h2>
                </div>
                <div class="contest-section__hint">
                  {{
                    contest.mode === 'awd'
                      ? `${challenges.length} 题`
                      : `${solvedCount} / ${challenges.length} 已解`
                  }}
                </div>
              </div>

              <ContestAWDWorkspacePanel
                v-if="contest.mode === 'awd'"
                :contest="contest"
                :challenges="challenges"
              />

              <ContestChallengeWorkspacePanel
                v-else
                :challenges="challenges"
                :selected-challenge="selectedChallenge"
                :flag-input="flagInput"
                :submitting="submitting"
                :submit-result="submitResult"
                @select-challenge="selectChallenge"
                @update:flag-input="setFlagInput"
                @submit-flag="submitFlagAction"
              />
            </section>
          </section>

          <ContestTeamWorkspaceSection
            v-else
            panel-id="contest-workspace-panel-team"
            tab-id="contest-workspace-tab-team"
            :team="team"
            :member-count="memberCount"
            :is-captain="isCaptain"
            @create-team="openCreateTeam"
            @join-team="openJoinTeam"
            @kick-member="kickMember"
          />
        </main>
      </template>

      <main v-else-if="contest" class="content-pane">
        <div class="contest-not-found">
          <AppEmpty
            icon="Flag"
            title="当前竞赛暂未开放"
            description="该竞赛还处于筹备阶段，暂不对学生开放查看或报名。"
          >
            <template #action>
              <RouterLink class="ui-btn ui-btn--primary" to="/contests">
                <Trophy class="h-4 w-4" />
                返回竞赛中心
              </RouterLink>
            </template>
          </AppEmpty>
        </div>
      </main>

      <main v-else class="content-pane">
        <div class="contest-not-found">
          <AppEmpty
            icon="AlertTriangle"
            title="竞赛不存在或暂不可用"
            description="请返回竞赛中心重新选择竞赛，或稍后再试。"
          >
            <template #action>
              <RouterLink class="ui-btn ui-btn--primary" to="/contests">
                <Trophy class="h-4 w-4" />
                返回竞赛中心
              </RouterLink>
            </template>
          </AppEmpty>
        </div>
      </main>
    </section>

    <ContestTeamDialogs
      :show-create-team="showCreateTeam"
      :show-join-team="showJoinTeam"
      :team-name="teamName"
      :team-id-input="teamIdInput"
      :creating-team="creatingTeam"
      :joining-team="joiningTeam"
      @update:show-create-team="setShowCreateTeam"
      @update:show-join-team="setShowJoinTeam"
      @update:team-name="setTeamName"
      @update:team-id-input="setTeamIdInput"
      @close-create-team="closeCreateTeam"
      @close-join-team="closeJoinTeam"
      @create-team="createTeamAction"
      @join-team="joinTeamAction"
    />
  </div>
</template>

<style scoped>
.contest-page-shell {
  --contest-accent: var(--color-primary);
  --journal-shell-accent: var(--contest-accent);
  --journal-shell-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-shell-surface: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --journal-shell-surface-subtle: color-mix(
    in srgb,
    var(--color-bg-surface) 76%,
    var(--color-bg-base)
  );
  --page-top-tabs-gap: 0.35rem;
  --page-top-tabs-margin: 0;
  --page-top-tabs-padding: 0;
  --page-top-tabs-border: color-mix(in srgb, var(--color-border-default) 86%, transparent);
  --page-top-tab-min-height: 3rem;
  --page-top-tab-padding: 0.4rem 0.1rem 0.65rem;
  --page-top-tab-font-size: var(--font-size-0-90);
  --page-top-tab-font-weight: 600;
  --page-top-tab-color: var(--color-text-secondary);
  --page-top-tab-active-color: var(--color-text-primary);
  --page-top-tab-active-border: color-mix(in srgb, var(--contest-accent) 72%, transparent);
  --workspace-panel-padding-top: 0;
  --journal-shell-hero-radial-strength: 10%;
  --journal-shell-hero-radial-size: 18rem;
  --journal-shell-hero-top-strength: 97%;
  --journal-shell-hero-end: color-mix(
    in srgb,
    var(--color-bg-elevated) 95%,
    var(--color-bg-base)
  );
  --journal-shell-hero-shadow: none;
  flex: 1 1 auto;
}

.content-pane {
  padding: 1.5rem;
}

.contest-loading,
.contest-not-found {
  min-height: 18rem;
}

.contest-loading {
  display: grid;
  justify-items: center;
  gap: 0.7rem;
  align-content: center;
}

.contest-loading__spinner {
  width: 2rem;
  height: 2rem;
  border-radius: 999px;
  border: 3px solid color-mix(in srgb, var(--color-border-default) 88%, transparent);
  border-top-color: var(--contest-accent);
  animation: contestDetailSpin 0.9s linear infinite;
}

.contest-loading__text {
  font-size: var(--font-size-0-88);
  color: var(--color-text-secondary);
}

.contest-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 86%, transparent);
  padding: 0.36rem 0.78rem;
  font-size: var(--font-size-0-76);
  font-weight: 700;
  color: var(--color-text-secondary);
}

.contest-chip--status,
.contest-chip--accent,
.contest-chip--success {
  color: color-mix(in srgb, var(--contest-accent) 88%, var(--color-text-primary));
}

.contest-chip--status {
  border-color: color-mix(in srgb, var(--contest-accent) 28%, transparent);
  background: color-mix(in srgb, var(--contest-accent) 11%, transparent);
}

.contest-chip--accent {
  border-color: color-mix(in srgb, var(--contest-accent) 24%, transparent);
  background: color-mix(in srgb, var(--contest-accent) 8%, transparent);
}

.contest-chip--success {
  border-color: color-mix(in srgb, var(--color-success) 28%, transparent);
  background: color-mix(in srgb, var(--color-success) 10%, transparent);
  color: color-mix(in srgb, var(--color-success) 88%, var(--color-text-primary));
}

.contest-divider {
  margin-block: var(--contest-divider-gap, var(--space-divider-gap));
  border-top: 1px solid
    var(
      --contest-divider-border,
      var(--journal-divider, color-mix(in srgb, var(--color-border-default) 86%, transparent))
    );
}

.contest-divider--compact {
  margin-block: var(--space-divider-gap-compact);
}

.contest-section__head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-3);
}

.contest-section__title:not(.workspace-tab-heading__title) {
  margin-top: var(--space-1-5);
  font-size: var(--font-size-1-10);
  font-weight: 700;
  color: var(--color-text-primary);
}

.contest-section__hint {
  font-size: var(--font-size-0-82);
  color: var(--color-text-secondary);
}

.contest-not-found {
  display: grid;
  align-content: center;
}

@keyframes contestDetailSpin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 860px) {
  .journal-shell {
    padding: 1rem;
  }
}
</style>

<script setup lang="ts">
import { Trophy } from 'lucide-vue-next'
import { RouterLink } from 'vue-router'

import type {
  ContestAnnouncement,
  ContestChallengeItem,
  ContestDetailData,
  SubmitFlagData,
  TeamData,
} from '@/api/contracts'
import AppEmpty from '@/shared/ui/common/AppEmpty.vue'
import { ContestAnnouncementRealtimeBridge } from '@/features/contest-announcements'
import { ContestAWDWorkspacePanel } from '@/features/contest-awd-workspace'
import {
  ContestAnnouncementsWorkspaceSection,
  ContestChallengeWorkspacePanel,
  ContestOverviewPanel,
  ContestTeamDialogs,
  ContestTeamWorkspaceSection,
} from '@/features/contest-detail'

type ContestDetailWorkspaceTab = 'overview' | 'announcements' | 'challenges' | 'team'

defineProps<{
  contest: ContestDetailData | null
  team: TeamData | null
  challenges: ContestChallengeItem[]
  announcements: ContestAnnouncement[]
  announcementsError: string
  loading: boolean
  countdown: string
  selectedChallenge: ContestChallengeItem | null
  flagInput: string
  submitting: boolean
  submitResult: SubmitFlagData | null
  showCreateTeam: boolean
  showJoinTeam: boolean
  teamName: string
  teamIdInput: string
  creatingTeam: boolean
  joiningTeam: boolean
  isCaptain: boolean
  activeWorkspaceTab: ContestDetailWorkspaceTab
  workspaceTabs: Array<{ id: ContestDetailWorkspaceTab; label: string }>
  solvedCount: number
  totalPoints: number
  memberCount: number
  contestAccentStyle?: Record<string, string>
  contestAccessible: boolean
  setTabButtonRef: (tabId: ContestDetailWorkspaceTab, element: HTMLButtonElement | null) => void
  selectWorkspaceTab: (tabId: ContestDetailWorkspaceTab) => void
  handleWorkspaceTabKeydown: (event: KeyboardEvent, index: number) => void
  selectChallenge: (challenge: ContestChallengeItem) => void
  setFlagInput: (value: string) => void
  submitFlagAction: () => void | Promise<void>
  setShowCreateTeam: (value: boolean) => void
  setShowJoinTeam: (value: boolean) => void
  setTeamName: (value: string) => void
  setTeamIdInput: (value: string) => void
  openCreateTeam: () => void
  closeCreateTeam: () => void
  createTeamAction: () => void | Promise<void>
  openJoinTeam: () => void
  closeJoinTeam: () => void
  joinTeamAction: () => void | Promise<void>
  kickMember: (userId: string) => void | Promise<void>
  refreshAnnouncements: () => void | Promise<void>
}>()
</script>
