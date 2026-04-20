<script setup lang="ts">
import { computed, watch } from 'vue'
import { BellRing, CalendarRange, Clock3, Flag, Swords, Trophy, UsersRound } from 'lucide-vue-next'
import { RouterLink, useRoute } from 'vue-router'

import AppEmpty from '@/components/common/AppEmpty.vue'
import CFocusedInputDialog from '@/components/common/modal-templates/CFocusedInputDialog.vue'
import ContestAWDWorkspacePanel from '@/components/contests/ContestAWDWorkspacePanel.vue'
import ContestAnnouncementRealtimeBridge from '@/components/contests/ContestAnnouncementRealtimeBridge.vue'
import { useContestDetailPage } from '@/composables/useContestDetailPage'
import { useUrlSyncedTabs } from '@/composables/useUrlSyncedTabs'
import { useAuthStore } from '@/stores/auth'
import {
  getContestAccentColor,
  getModeLabel,
  getStatusLabel,
  isStudentVisibleContestStatus,
} from '@/utils/contest'
import { formatTime } from '@/utils/format'

type ContestWorkspaceTab = 'overview' | 'announcements' | 'challenges' | 'team'

const route = useRoute()
const authStore = useAuthStore()
const contestId = computed(() => String(route.params.id ?? ''))
const currentUserId = computed(() => authStore.user?.id)
const workspaceTabOrder: ContestWorkspaceTab[] = ['overview', 'announcements', 'challenges', 'team']
const {
  activeTab: activeWorkspaceTab,
  setTabButtonRef,
  selectTab: selectWorkspaceTab,
  handleTabKeydown: handleWorkspaceTabKeydown,
} = useUrlSyncedTabs<ContestWorkspaceTab>({
  orderedTabs: workspaceTabOrder,
  defaultTab: 'overview',
})

const {
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
} = useContestDetailPage({
  contestId,
  currentUserId,
})

const isAWDContest = computed(() => contest.value?.mode === 'awd')
const workspaceTabs = computed<Array<{ id: ContestWorkspaceTab; label: string }>>(() => [
  { id: 'overview', label: '概览' },
  { id: 'announcements', label: '公告' },
  { id: 'challenges', label: isAWDContest.value ? '战场' : '题目' },
  { id: 'team', label: '队伍' },
])

watch(
  () => contest.value?.mode,
  (mode) => {
    if (mode === 'awd' && !route.query.panel) {
      selectWorkspaceTab('challenges')
    }
  }
)

const solvedCount = computed(() => challenges.value.filter((item) => item.is_solved).length)
const totalPoints = computed(() =>
  challenges.value.reduce((sum, item) => sum + (item.points || 0), 0)
)
const memberCount = computed(() => team.value?.members.length ?? 0)

const selectedChallengeMeta = computed(() => {
  if (!selectedChallenge.value) return ''
  return `${selectedChallenge.value.category} · ${selectedChallenge.value.points} pts`
})

const contestAccentStyle = computed<Record<string, string> | undefined>(() => {
  if (!contest.value) return undefined
  return {
    '--contest-accent': getContestAccentColor(contest.value.status),
  }
})
const contestAccessible = computed(() =>
  contest.value ? isStudentVisibleContestStatus(contest.value.status) : false
)

function challengeClass(challengeId: string, solved: boolean): string[] {
  const active = selectedChallenge.value?.id === challengeId
  return [
    'contest-challenge',
    active ? 'contest-challenge--active' : '',
    solved ? 'contest-challenge--solved' : '',
  ]
}
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
          <section
            v-if="activeWorkspaceTab === 'overview'"
            id="contest-workspace-panel-overview"
            class="workspace-panel"
            role="tabpanel"
            aria-labelledby="contest-workspace-tab-overview"
          >
            <header class="contest-hero">
              <div class="contest-hero__main">
                <div class="workspace-overline">
                  Contest
                </div>
                <h1 class="contest-hero__title workspace-page-title">
                  {{ contest.title }}
                </h1>
                <p class="contest-hero__desc workspace-page-copy">
                  {{ contest.description || '当前竞赛暂未提供描述。' }}
                </p>

                <div class="contest-meta-strip">
                  <span class="contest-chip contest-chip--status">
                    {{ getStatusLabel(contest.status) }}
                  </span>
                  <span class="contest-chip contest-chip--neutral">
                    {{ getModeLabel(contest.mode) }}
                  </span>
                  <span class="contest-chip contest-chip--neutral">
                    {{ formatTime(contest.starts_at) }} ~ {{ formatTime(contest.ends_at) }}
                  </span>
                  <span
                    v-if="countdown"
                    class="contest-chip contest-chip--accent"
                  >
                    {{ countdown }}
                  </span>
                </div>
              </div>

              <aside class="contest-score-rail">
                <div class="contest-score-rail__label">
                  总分
                </div>
                <div class="contest-score-rail__value">
                  {{ totalPoints }} <small>pts</small>
                </div>
                <div class="contest-score-rail__note">
                  {{ challenges.length }} 题 · {{ solvedCount }} 已解 · {{ memberCount }} 人
                </div>
              </aside>
            </header>

            <div class="contest-divider" />

            <section class="contest-stat-grid metric-panel-grid">
              <article class="contest-stat metric-panel-card">
                <div class="contest-stat__label metric-panel-label">
                  队伍成员
                </div>
                <div class="contest-stat__value metric-panel-value">
                  {{ memberCount }}
                </div>
                <div class="contest-stat__hint metric-panel-helper">
                  当前队伍人数
                </div>
              </article>
              <article class="contest-stat metric-panel-card">
                <div class="contest-stat__label metric-panel-label">
                  题目数量
                </div>
                <div class="contest-stat__value metric-panel-value">
                  {{ challenges.length }}
                </div>
                <div class="contest-stat__hint metric-panel-helper">
                  本场竞赛题目总数
                </div>
              </article>
              <article class="contest-stat metric-panel-card">
                <div class="contest-stat__label metric-panel-label">
                  已解题目
                </div>
                <div class="contest-stat__value metric-panel-value">
                  {{ solvedCount }}
                </div>
                <div class="contest-stat__hint metric-panel-helper">
                  当前账号已完成数量
                </div>
              </article>
              <article class="contest-stat metric-panel-card">
                <div class="contest-stat__label metric-panel-label">
                  积分总览
                </div>
                <div class="contest-stat__value metric-panel-value">
                  {{ totalPoints }}
                </div>
                <div class="contest-stat__hint metric-panel-helper">
                  全部题目可获得积分
                </div>
              </article>
            </section>

            <div class="contest-divider" />

            <div class="contest-overview-grid">
              <section class="contest-section contest-section--flat">
                <div class="contest-section__head workspace-tab-heading">
                  <div class="workspace-tab-heading__main">
                    <div class="workspace-overline">
                      Rules
                    </div>
                    <h2 class="contest-section__title workspace-tab-heading__title">
                      竞赛规则
                    </h2>
                  </div>
                </div>
                <div class="contest-copy">
                  {{ contest.rules || '当前竞赛暂无额外规则说明。' }}
                </div>
              </section>

              <section class="contest-section contest-section--flat">
                <div class="contest-section__head workspace-tab-heading">
                  <div class="workspace-tab-heading__main">
                    <div class="workspace-overline">
                      Schedule
                    </div>
                    <h2 class="contest-section__title workspace-tab-heading__title">
                      赛程信息
                    </h2>
                  </div>
                </div>
                <div class="contest-copy-list">
                  <div class="contest-copy-row">
                    <span>开始时间</span>
                    <strong>{{ formatTime(contest.starts_at) }}</strong>
                  </div>
                  <div class="contest-copy-row">
                    <span>结束时间</span>
                    <strong>{{ formatTime(contest.ends_at) }}</strong>
                  </div>
                  <div class="contest-copy-row">
                    <span>参赛模式</span>
                    <strong>{{ getModeLabel(contest.mode) }}</strong>
                  </div>
                  <div class="contest-copy-row">
                    <span>冻结榜单</span>
                    <strong>{{ contest.scoreboard_frozen ? '是' : '否' }}</strong>
                  </div>
                </div>
              </section>
            </div>

            <div class="contest-divider" />

            <section class="contest-section contest-section--flat">
              <div class="contest-section__head workspace-tab-heading">
                <div class="workspace-tab-heading__main">
                  <div class="workspace-overline">
                    Announcements
                  </div>
                  <h2 class="contest-section__title workspace-tab-heading__title">
                    公告预览
                  </h2>
                </div>
                <div class="contest-section__hint">
                  {{ announcements.length }} 条
                </div>
              </div>

              <div
                v-if="announcementsError"
                class="contest-alert contest-alert--warning"
              >
                {{ announcementsError }}
              </div>

              <div
                v-else-if="announcements.length === 0"
                class="contest-inline-note"
              >
                当前竞赛暂无新的公告通知。
              </div>

              <div
                v-else
                class="announcement-list"
              >
                <article
                  v-for="announcement in announcements"
                  :key="announcement.id"
                  class="announcement-item"
                >
                  <div class="announcement-item__head">
                    <h3 class="announcement-item__title">
                      {{ announcement.title }}
                    </h3>
                    <time
                      class="announcement-item__time"
                      :datetime="announcement.created_at"
                    >
                      {{ formatTime(announcement.created_at) }}
                    </time>
                  </div>
                  <p
                    v-if="announcement.content"
                    class="announcement-item__content"
                  >
                    {{ announcement.content }}
                  </p>
                </article>
              </div>
            </section>
          </section>

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

              <div
                v-if="announcementsError"
                class="contest-alert contest-alert--warning"
              >
                {{ announcementsError }}
              </div>

              <div
                v-else-if="announcements.length === 0"
                class="contest-empty-state"
              >
                <AppEmpty
                  icon="Bell"
                  title="暂无公告"
                  description="当前竞赛暂无新的公告通知。"
                />
              </div>

              <div
                v-else
                class="announcement-list"
              >
                <article
                  v-for="announcement in announcements"
                  :key="announcement.id"
                  class="announcement-item"
                >
                  <div class="announcement-item__head">
                    <h3 class="announcement-item__title">
                      {{ announcement.title }}
                    </h3>
                    <time
                      class="announcement-item__time"
                      :datetime="announcement.created_at"
                    >
                      {{ formatTime(announcement.created_at) }}
                    </time>
                  </div>
                  <p
                    v-if="announcement.content"
                    class="announcement-item__content"
                  >
                    {{ announcement.content }}
                  </p>
                </article>
              </div>
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
                    {{ contest.mode === 'awd' ? 'Battle' : 'Challenges' }}
                  </div>
                  <h2 class="contest-section__title workspace-tab-heading__title">
                    {{ contest.mode === 'awd' ? '战场' : '题目' }}
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

              <div
                v-else-if="challenges.length === 0"
                class="contest-empty-state"
              >
                <AppEmpty
                  icon="Flag"
                  title="暂无题目"
                  description="当前竞赛尚未发布题目。"
                />
              </div>

              <div
                v-else
                class="contest-challenge-workspace"
              >
                <div class="contest-challenge-list">
                  <button
                    v-for="challenge in challenges"
                    :key="challenge.id"
                    type="button"
                    :class="challengeClass(challenge.id, challenge.is_solved)"
                    @click="selectChallenge(challenge)"
                  >
                    <div class="contest-challenge__head">
                      <h3 class="contest-challenge__title">
                        {{ challenge.title }}
                      </h3>
                      <span
                        v-if="challenge.is_solved"
                        class="contest-challenge__solved"
                      >✓</span>
                    </div>
                    <div class="contest-challenge__meta">
                      <span>{{ challenge.category }}</span>
                      <span>{{ challenge.points }} pts</span>
                      <span>{{ challenge.solved_count }} 人解出</span>
                    </div>
                  </button>
                </div>

                <article class="challenge-focus">
                  <template v-if="selectedChallenge">
                    <div class="challenge-focus__head">
                      <div>
                        <div class="workspace-overline">
                          Selected
                        </div>
                        <h3 class="challenge-focus__title">
                          {{ selectedChallenge.title }}
                        </h3>
                      </div>
                      <div class="challenge-focus__meta">
                        {{ selectedChallengeMeta }}
                      </div>
                    </div>

                    <div class="challenge-focus__stats">
                      <span class="contest-chip contest-chip--neutral">
                        解出人数 {{ selectedChallenge.solved_count }}
                      </span>
                      <span
                        v-if="selectedChallenge.is_solved"
                        class="contest-chip contest-chip--success"
                      >
                        已解出
                      </span>
                    </div>

                    <div class="contest-divider contest-divider--compact" />

                    <div class="challenge-focus__form">
                      <div>
                        <div class="workspace-overline">
                          Primary Action
                        </div>
                        <h4 class="challenge-focus__form-title">
                          提交 Flag
                        </h4>
                      </div>

                      <label
                        class="ui-field__label flag-submit__label"
                        for="contest-flag-input"
                      >
                        Flag
                      </label>
                      <div class="flag-submit">
                        <div class="ui-control-wrap flag-submit__control">
                          <input
                            id="contest-flag-input"
                            v-model="flagInput"
                            type="text"
                            placeholder="flag{...}"
                            class="ui-control"
                            @keyup.enter="submitFlagAction"
                          >
                        </div>
                        <button
                          type="button"
                          :disabled="submitting"
                          class="ui-btn ui-btn--primary"
                          @click="submitFlagAction"
                        >
                          {{ submitting ? '提交中...' : '提交' }}
                        </button>
                      </div>

                      <div
                        v-if="submitResult"
                        class="contest-alert"
                        :class="
                          submitResult.is_correct ? 'contest-alert--success' : 'contest-alert--danger'
                        "
                      >
                        {{
                          submitResult.is_correct
                            ? `正确！+${submitResult.points ?? 0} 分`
                            : submitResult.message
                        }}
                      </div>
                    </div>
                  </template>

                  <div
                    v-else
                    class="contest-inline-note"
                  >
                    从左侧选择题目后可在这里提交 Flag。
                  </div>
                </article>
              </div>
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

              <div
                v-if="!team"
                class="team-empty"
              >
                <div class="contest-inline-note">
                  当前账号尚未加入队伍。
                </div>
                <div class="team-actions">
                  <button
                    type="button"
                    class="ui-btn ui-btn--primary"
                    @click="openCreateTeam"
                  >
                    创建队伍
                  </button>
                  <button
                    type="button"
                    class="ui-btn ui-btn--ghost"
                    @click="openJoinTeam"
                  >
                    加入队伍
                  </button>
                </div>
              </div>

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

                <div class="team-member-list">
                  <div
                    v-for="member in team.members"
                    :key="member.user_id"
                    class="team-member"
                  >
                    <span class="team-member__name">{{ member.username }}</span>
                    <div class="team-member__actions">
                      <span
                        v-if="member.user_id === team.captain_user_id"
                        class="team-member__captain"
                      >
                        队长
                      </span>
                      <button
                        v-if="isCaptain && member.user_id !== team.captain_user_id"
                        type="button"
                        class="team-member__kick"
                        @click="kickMember(member.user_id)"
                      >
                        踢出
                      </button>
                    </div>
                  </div>
                </div>
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
  --page-top-tabs-border: color-mix(in srgb, var(--journal-border) 86%, transparent);
  --page-top-tab-min-height: 3rem;
  --page-top-tab-padding: 0.4rem 0.1rem 0.65rem;
  --page-top-tab-font-size: var(--font-size-0-90);
  --page-top-tab-font-weight: 600;
  --page-top-tab-color: var(--journal-muted);
  --page-top-tab-active-color: var(--journal-ink);
  --page-top-tab-active-border: color-mix(in srgb, var(--contest-accent) 72%, transparent);
  --journal-shell-hero-radial-strength: 10%;
  --journal-shell-hero-radial-size: 18rem;
  --journal-shell-hero-top-strength: 97%;
  --journal-shell-hero-end: color-mix(
    in srgb,
    var(--journal-surface-subtle) 95%,
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
  border: 3px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-top-color: var(--contest-accent);
  animation: contestDetailSpin 0.9s linear infinite;
}

.contest-loading__text {
  font-size: var(--font-size-0-88);
  color: var(--journal-muted);
}

.workspace-panel {
  padding-top: 1.35rem;
}

.contest-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 15rem;
  gap: 1.25rem;
}

.contest-hero__title {
  margin-top: 0.85rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.contest-hero__desc {
  margin-top: 0.8rem;
  max-width: 60ch;
  color: var(--journal-muted);
}

.contest-meta-strip {
  margin-top: 1rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.55rem;
}

.contest-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 86%, transparent);
  padding: 0.36rem 0.78rem;
  font-size: var(--font-size-0-76);
  font-weight: 700;
  color: var(--journal-muted);
}

.contest-chip--status,
.contest-chip--accent,
.contest-chip--success {
  color: color-mix(in srgb, var(--contest-accent) 88%, var(--journal-ink));
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
  color: color-mix(in srgb, var(--color-success) 88%, var(--journal-ink));
}

.contest-score-rail {
  align-self: start;
  border-inline-start: 1px solid color-mix(in srgb, var(--journal-border) 86%, transparent);
  padding-inline-start: 1.1rem;
}

.contest-score-rail__label {
  font-size: var(--font-size-0-74);
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.contest-score-rail__value {
  margin-top: 0.7rem;
  font-size: var(--font-size-2-10);
  font-weight: 700;
  line-height: 1;
  color: var(--journal-ink);
}

.contest-score-rail__value small {
  font-size: var(--font-size-0-85);
  color: var(--journal-muted);
}

.contest-score-rail__note {
  margin-top: 0.7rem;
  font-size: var(--font-size-0-84);
  line-height: 1.7;
  color: var(--journal-muted);
}

.contest-divider {
  margin: 1.4rem 0;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 86%, transparent);
}

.contest-divider--compact {
  margin-block: 1rem;
}

.contest-stat-grid {
  --metric-panel-grid-gap: 0.85rem;
  --metric-panel-columns: repeat(4, minmax(0, 1fr));
}

.contest-stat {
  min-width: 0;
}

.contest-overview-grid {
  display: grid;
  gap: 1.25rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
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
  color: var(--journal-ink);
}

.contest-section__hint {
  font-size: var(--font-size-0-82);
  color: var(--journal-muted);
}

.contest-section--flat + .contest-section--flat {
  border-top: 0;
}

.contest-copy,
.contest-copy-list {
  margin-top: 1rem;
}

.contest-copy {
  white-space: pre-wrap;
  font-size: var(--font-size-0-92);
  line-height: 1.8;
  color: var(--journal-ink);
}

.contest-copy-list {
  display: grid;
  gap: 0.8rem;
}

.contest-copy-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem 1rem;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 82%, transparent);
  padding-bottom: 0.8rem;
  font-size: var(--font-size-0-88);
  color: var(--journal-muted);
}

.contest-copy-row strong {
  color: var(--journal-ink);
}

.contest-empty-state {
  margin-top: 1rem;
}

.announcement-list {
  margin-top: 1rem;
  display: grid;
  gap: 0.9rem;
}

.announcement-item {
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 82%, transparent);
  padding-bottom: 0.9rem;
}

.announcement-item__head {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.45rem 1rem;
}

.announcement-item__title {
  font-size: var(--font-size-0-96);
  font-weight: 700;
  color: var(--journal-ink);
}

.announcement-item__time {
  font-size: var(--font-size-0-76);
  color: var(--journal-muted);
}

.announcement-item__content {
  margin-top: 0.55rem;
  white-space: pre-wrap;
  font-size: var(--font-size-0-88);
  line-height: 1.75;
  color: var(--journal-muted);
}

.contest-challenge-workspace {
  margin-top: 1rem;
  display: grid;
  gap: 1.25rem;
  grid-template-columns: minmax(0, 18rem) minmax(0, 1fr);
}

.contest-challenge-list {
  display: grid;
  gap: 0.45rem;
}

.contest-challenge {
  border: 0;
  border-inline-start: 2px solid
    color-mix(in srgb, var(--contest-accent) 24%, var(--journal-border));
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 82%, transparent);
  background: transparent;
  padding: 0.75rem 0.35rem 0.75rem 0.85rem;
  text-align: left;
  transition:
    border-color 150ms ease,
    background 150ms ease,
    color 150ms ease;
}

.contest-challenge:hover,
.contest-challenge:focus-visible {
  border-inline-start-color: color-mix(in srgb, var(--contest-accent) 72%, var(--journal-border));
  background: color-mix(in srgb, var(--contest-accent) 5%, transparent);
  outline: none;
}

.contest-challenge--active {
  border-inline-start-color: color-mix(in srgb, var(--contest-accent) 86%, var(--journal-border));
  background: color-mix(in srgb, var(--contest-accent) 7%, transparent);
}

.contest-challenge--solved {
  border-inline-start-color: color-mix(in srgb, var(--color-success) 68%, var(--journal-border));
}

.contest-challenge__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.55rem;
}

.contest-challenge__title {
  font-size: var(--font-size-0-90);
  font-weight: 700;
  color: var(--journal-ink);
}

.contest-challenge__solved {
  color: var(--color-success);
  font-weight: 700;
}

.contest-challenge__meta {
  margin-top: 0.35rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem 0.7rem;
  font-size: var(--font-size-0-76);
  color: var(--journal-muted);
}

.challenge-focus {
  min-width: 0;
}

.challenge-focus__head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 0.75rem;
}

.challenge-focus__title {
  margin-top: 0.35rem;
  font-size: var(--font-size-1-20);
  font-weight: 700;
  color: var(--journal-ink);
}

.challenge-focus__meta {
  font-size: var(--font-size-0-82);
  color: var(--journal-muted);
}

.challenge-focus__stats {
  margin-top: 1rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.55rem;
}

.challenge-focus__form-title {
  margin-top: 0.35rem;
  font-size: var(--font-size-1-00);
  font-weight: 700;
  color: var(--journal-ink);
}

.flag-submit {
  margin-top: 0.55rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
}

.flag-submit__label {
  display: inline-flex;
  margin-top: 1rem;
}

.flag-submit__control {
  flex: 1 1 18rem;
  min-width: 0;
}

.contest-alert {
  margin-top: 0.8rem;
  border-inline-start: 2px solid transparent;
  padding: 0.6rem 0.75rem;
  font-size: var(--font-size-0-84);
  line-height: 1.6;
}

.contest-alert--success {
  border-inline-start-color: color-mix(in srgb, var(--color-success) 60%, transparent);
  background: color-mix(in srgb, var(--color-success) 8%, transparent);
  color: color-mix(in srgb, var(--color-success) 86%, var(--journal-ink));
}

.contest-alert--danger {
  border-inline-start-color: color-mix(in srgb, var(--color-danger) 60%, transparent);
  background: color-mix(in srgb, var(--color-danger) 8%, transparent);
  color: color-mix(in srgb, var(--color-danger) 86%, var(--journal-ink));
}

.contest-alert--warning {
  border-inline-start-color: color-mix(in srgb, var(--color-warning) 60%, transparent);
  background: color-mix(in srgb, var(--color-warning) 8%, transparent);
  color: color-mix(in srgb, var(--color-warning) 88%, var(--journal-ink));
}

.contest-inline-note {
  border-inline-start: 2px solid color-mix(in srgb, var(--journal-border) 84%, transparent);
  padding-inline-start: 0.85rem;
  font-size: var(--font-size-0-88);
  line-height: 1.7;
  color: var(--journal-muted);
}

.team-empty {
  margin-top: 1rem;
  display: grid;
  gap: 0.9rem;
}

.team-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
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
  color: var(--journal-ink);
}

.team-summary__invite {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-0-78);
  color: var(--journal-muted);
}

.team-member-list {
  margin-top: 1rem;
  display: grid;
  gap: 0.55rem;
}

.team-member {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.45rem 1rem;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 82%, transparent);
  padding-bottom: 0.75rem;
}

.team-member__name {
  font-size: var(--font-size-0-90);
  color: var(--journal-ink);
}

.team-member__actions {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.team-member__captain {
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--contest-accent) 28%, transparent);
  background: color-mix(in srgb, var(--contest-accent) 10%, transparent);
  padding: 0.2rem 0.55rem;
  font-size: var(--font-size-0-72);
  font-weight: 700;
  color: color-mix(in srgb, var(--contest-accent) 84%, var(--journal-ink));
}

.team-member__kick {
  border: 0;
  background: transparent;
  padding: 0;
  font-size: var(--font-size-0-78);
  color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
}

.team-member__kick:hover,
.team-member__kick:focus-visible {
  text-decoration: underline;
  outline: none;
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

@media (max-width: 1100px) {
  .contest-stat-grid,
  .contest-overview-grid,
  .contest-challenge-workspace {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .contest-challenge-workspace {
    grid-template-columns: minmax(0, 1fr);
  }
}

@media (max-width: 860px) {
  .journal-shell {
    padding: 1rem;
  }

  .contest-hero {
    grid-template-columns: 1fr;
  }

  .contest-score-rail {
    border-inline-start: 0;
    border-top: 1px solid color-mix(in srgb, var(--journal-border) 86%, transparent);
    padding-inline-start: 0;
    padding-top: 1rem;
  }

  .contest-stat-grid,
  .contest-overview-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}

@media (max-width: 640px) {
  .flag-submit {
    flex-direction: column;
  }
}
</style>
