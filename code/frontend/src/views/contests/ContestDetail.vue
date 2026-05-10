<script setup lang="ts">
import { BellRing, CalendarRange, Clock3, Flag, Swords, Trophy, UsersRound } from 'lucide-vue-next'
import { RouterLink } from 'vue-router'

import AppEmpty from '@/components/common/AppEmpty.vue'
import CFocusedInputDialog from '@/components/common/modal-templates/CFocusedInputDialog.vue'
import ContestAnnouncementsPanel from '@/components/contests/ContestAnnouncementsPanel.vue'
import ContestAWDWorkspacePanel from '@/components/contests/ContestAWDWorkspacePanel.vue'
import ContestAnnouncementRealtimeBridge from '@/components/contests/ContestAnnouncementRealtimeBridge.vue'
import ContestChallengeWorkspacePanel from '@/components/contests/ContestChallengeWorkspacePanel.vue'
import ContestOverviewPanel from '@/components/contests/ContestOverviewPanel.vue'
import ContestTeamPanel from '@/components/contests/ContestTeamPanel.vue'
import { useContestDetailRoutePage } from '@/features/contest-detail'

const {
  router,
  contest,
  team,
  challenges,
  announcements,
  announcementsError,
  loading,
  countdown,
  selectedChallenge,
  flagInput,
  submitting,
  submitResult,
  showCreateTeam,
  showJoinTeam,
  teamName,
  teamIdInput,
  creatingTeam,
  joiningTeam,
  isCaptain,
  activeWorkspaceTab,
  setTabButtonRef,
  selectWorkspaceTab,
  handleWorkspaceTabKeydown,
  workspaceTabs,
  solvedCount,
  totalPoints,
  memberCount,
  contestAccentStyle,
  contestAccessible,
  selectChallenge,
  submitFlagAction,
  openCreateTeam,
  closeCreateTeam,
  createTeamAction,
  openJoinTeam,
  closeJoinTeam,
  joinTeamAction,
  kickMember,
  refreshAnnouncements,
} = useContestDetailRoutePage()
</script>

<template>
  <div
    class="contest-page-shell"
    :style="contestAccentStyle"
  >
    <section
      class="workspace-shell journal-shell journal-shell-user journal-hero contest-detail-view flex min-h-full flex-1 flex-col"
    >
      <main
        v-if="loading"
        class="content-pane"
      >
        <div class="contest-loading">
          <div class="contest-loading__spinner" />
          <div class="contest-loading__text">
            正在同步竞赛详情...
          </div>
        </div>
      </main>

      <template v-else-if="contest && contestAccessible">
        <ContestAnnouncementRealtimeBridge
          :contest-id="contest.id"
          @updated="refreshAnnouncements"
        />

        <div
          class="workspace-tabbar top-tabs"
          role="tablist"
          aria-label="竞赛页面主切换"
        >
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

          <section
            v-else-if="activeWorkspaceTab === 'announcements'"
            id="contest-workspace-panel-announcements"
            class="workspace-panel"
            role="tabpanel"
            aria-labelledby="contest-workspace-tab-announcements"
          >
            <section class="contest-section">
              <div class="contest-section__head workspace-tab-heading">
                <div class="workspace-tab-heading__main">
                  <div class="workspace-overline">
                    Announcements
                  </div>
                  <h2 class="contest-section__title workspace-tab-heading__title">
                    公告
                  </h2>
                </div>
                <div class="contest-section__hint">
                  {{ announcements.length }} 条
                </div>
              </div>

              <ContestAnnouncementsPanel
                :announcements="announcements"
                :announcements-error="announcementsError"
              />
            </section>
          </section>

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
                @update:flag-input="flagInput = $event"
                @submit-flag="submitFlagAction"
              />
            </section>
          </section>

          <section
            v-else
            id="contest-workspace-panel-team"
            class="workspace-panel"
            role="tabpanel"
            aria-labelledby="contest-workspace-tab-team"
          >
            <section class="contest-section">
              <div class="contest-section__head workspace-tab-heading">
                <div class="workspace-tab-heading__main">
                  <div class="workspace-overline">
                    Team
                  </div>
                  <h2 class="contest-section__title workspace-tab-heading__title">
                    队伍
                  </h2>
                </div>
                <div class="contest-section__hint">
                  {{ memberCount }} 人
                </div>
              </div>

              <ContestTeamPanel
                v-if="!team"
                :team="null"
                @create-team="openCreateTeam"
                @join-team="openJoinTeam"
              />

              <div
                v-else
                class="team-board"
              >
                <div class="team-summary">
                  <div>
                    <div class="workspace-overline">
                      Current Team
                    </div>
                    <h3 class="team-summary__name">
                      {{ team.name }}
                    </h3>
                  </div>
                  <span
                    v-if="team.invite_code"
                    class="team-summary__invite"
                  >邀请码: {{ team.invite_code }}</span>
                </div>

                <ContestTeamPanel
                  :team="team"
                  :is-captain="isCaptain"
                  @kick-member="kickMember"
                />
              </div>
            </section>
          </section>
        </main>
      </template>

      <main
        v-else-if="contest"
        class="content-pane"
      >
        <div class="contest-not-found">
          <AppEmpty
            icon="Flag"
            title="当前竞赛暂未开放"
            description="该竞赛还处于筹备阶段，暂不对学生开放查看或报名。"
          >
            <template #action>
              <RouterLink
                class="ui-btn ui-btn--primary"
                to="/contests"
              >
                <Trophy class="h-4 w-4" />
                返回竞赛中心
              </RouterLink>
            </template>
          </AppEmpty>
        </div>
      </main>

      <main
        v-else
        class="content-pane"
      >
        <div class="contest-not-found">
          <AppEmpty
            icon="AlertTriangle"
            title="竞赛不存在或暂不可用"
            description="请返回竞赛中心重新选择竞赛，或稍后再试。"
          >
            <template #action>
              <RouterLink
                class="ui-btn ui-btn--primary"
                to="/contests"
              >
                <Trophy class="h-4 w-4" />
                返回竞赛中心
              </RouterLink>
            </template>
          </AppEmpty>
        </div>
      </main>
    </section>

    <CFocusedInputDialog
      :open="showCreateTeam"
      title="创建新队伍"
      description="为你的战队起一个响亮的代号。创建完成后，你可以生成邀请链接让其他队友加入。"
      width="35rem"
      aria-label="创建队伍"
      overlay-class="c-focused-input-shell--plain"
      :close-on-backdrop="false"
      @update:open="showCreateTeam = $event"
      @close="closeCreateTeam"
    >
      <template #icon>
        <UsersRound
          class="h-6 w-6"
          :stroke-width="2"
        />
      </template>

      <div class="contest-team-dialog-field">
        <label for="contest-create-team-name">队伍名称</label>
        <input
          id="contest-create-team-name"
          v-model="teamName"
          type="text"
          placeholder="例如：HackerG1"
          @keyup.enter="createTeamAction"
        >
      </div>

      <template #footer="{ close }">
        <button
          type="button"
          data-c-modal-action="ghost"
          @click="close"
        >
          取消
        </button>
        <button
          type="button"
          data-c-modal-action="primary"
          :disabled="creatingTeam"
          @click="createTeamAction"
        >
          {{ creatingTeam ? '创建中...' : '确认创建' }}
        </button>
      </template>
    </CFocusedInputDialog>

    <CFocusedInputDialog
      :open="showJoinTeam"
      title="加入现有队伍"
      description="输入队伍 ID 后立即加入当前战队。加入成功后，你会同步看到队伍成员与竞赛工作区。"
      width="34rem"
      aria-label="加入队伍"
      overlay-class="c-focused-input-shell--plain"
      :close-on-backdrop="false"
      @update:open="showJoinTeam = $event"
      @close="closeJoinTeam"
    >
      <template #icon>
        <UsersRound
          class="h-6 w-6"
          :stroke-width="2"
        />
      </template>

      <div class="contest-team-dialog-field">
        <label for="contest-join-team-id">队伍 ID</label>
        <input
          id="contest-join-team-id"
          v-model="teamIdInput"
          type="text"
          placeholder="输入队伍 ID"
          @keyup.enter="joinTeamAction"
        >
      </div>

      <template #footer="{ close }">
        <button
          type="button"
          data-c-modal-action="ghost"
          @click="close"
        >
          取消
        </button>
        <button
          type="button"
          data-c-modal-action="primary"
          :disabled="joiningTeam"
          @click="joinTeamAction"
        >
          {{ joiningTeam ? '加入中...' : '确认加入' }}
        </button>
      </template>
    </CFocusedInputDialog>
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

.workspace-panel {
  padding-top: 1.35rem;
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
  margin: 1.4rem 0;
  border-top: 1px solid color-mix(in srgb, var(--color-border-default) 86%, transparent);
}

.contest-divider--compact {
  margin-block: 1rem;
}




.contest-section__head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 0.8rem;
}

.contest-section__title:not(.workspace-tab-heading__title) {
  margin-top: 0.35rem;
  font-size: var(--font-size-1-10);
  font-weight: 700;
  color: var(--color-text-primary);
}

.contest-section__hint {
  font-size: var(--font-size-0-82);
  color: var(--color-text-secondary);
}

.team-board {
  margin-top: 1rem;
}

.team-summary {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.6rem 1rem;
}

.team-summary__name {
  margin-top: 0.35rem;
  font-size: var(--font-size-1-10);
  font-weight: 700;
  color: var(--color-text-primary);
}

.team-summary__invite {
  font-size: var(--font-size-0-78);
  color: var(--color-text-secondary);
}

.contest-not-found {
  display: grid;
  align-content: center;
}

.contest-team-dialog-field {
  display: grid;
  gap: 0.5rem;
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
