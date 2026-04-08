<script setup lang="ts">
import { computed, ref } from 'vue'
import { BellRing, CalendarRange, Clock3, Flag, Swords, Trophy, UsersRound } from 'lucide-vue-next'
import { RouterLink, useRoute } from 'vue-router'

import AppEmpty from '@/components/common/AppEmpty.vue'
import ContestAnnouncementRealtimeBridge from '@/components/contests/ContestAnnouncementRealtimeBridge.vue'
import { useContestDetailPage } from '@/composables/useContestDetailPage'
import { useAuthStore } from '@/stores/auth'
import { getContestAccentColor, getModeLabel, getStatusLabel } from '@/utils/contest'
import { formatTime } from '@/utils/format'

type ContestWorkspaceTab = 'overview' | 'announcements' | 'challenges' | 'team'

const workspaceTabs: Array<{ id: ContestWorkspaceTab; label: string }> = [
  { id: 'overview', label: '概览' },
  { id: 'announcements', label: '公告' },
  { id: 'challenges', label: '题目' },
  { id: 'team', label: '队伍' },
]

const route = useRoute()
const authStore = useAuthStore()
const contestId = computed(() => String(route.params.id ?? ''))
const currentUserId = computed(() => authStore.user?.id)
const contestWorkspaceTabSet = new Set<ContestWorkspaceTab>(workspaceTabs.map((tab) => tab.id))

function resolveWorkspaceTabFromLocation(): ContestWorkspaceTab {
  if (typeof window === 'undefined') return 'overview'
  if (!window.location.pathname || window.location.pathname === '/') return 'overview'
  const panel = new URLSearchParams(window.location.search).get('panel')
  if (panel && contestWorkspaceTabSet.has(panel as ContestWorkspaceTab)) {
    return panel as ContestWorkspaceTab
  }
  return 'overview'
}

function syncWorkspacePanelToLocation(tab: ContestWorkspaceTab): void {
  if (typeof window === 'undefined') return
  const url = new URL(window.location.href)
  url.searchParams.set('panel', tab)
  window.history.replaceState(window.history.state, '', `${url.pathname}${url.search}${url.hash}`)
}

const activeWorkspaceTab = ref<ContestWorkspaceTab>(resolveWorkspaceTabFromLocation())

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

function challengeClass(challengeId: string, solved: boolean): string[] {
  const active = selectedChallenge.value?.id === challengeId
  return [
    'contest-challenge',
    active ? 'contest-challenge--active' : '',
    solved ? 'contest-challenge--solved' : '',
  ]
}

function focusWorkspaceTab(id: string): void {
  requestAnimationFrame(() => {
    document.getElementById(id)?.focus()
  })
}

function selectWorkspaceTab(tab: ContestWorkspaceTab): void {
  if (activeWorkspaceTab.value === tab) return
  activeWorkspaceTab.value = tab
  syncWorkspacePanelToLocation(tab)
}

function handleWorkspaceTabKeydown(event: KeyboardEvent, currentTab: ContestWorkspaceTab): void {
  const currentIndex = workspaceTabs.findIndex((item) => item.id === currentTab)
  if (currentIndex < 0) return

  if (event.key === 'ArrowRight') {
    const nextTab = workspaceTabs[(currentIndex + 1) % workspaceTabs.length]
    selectWorkspaceTab(nextTab.id)
    focusWorkspaceTab(`contest-workspace-tab-${nextTab.id}`)
  } else if (event.key === 'ArrowLeft') {
    const nextTab = workspaceTabs[(currentIndex - 1 + workspaceTabs.length) % workspaceTabs.length]
    selectWorkspaceTab(nextTab.id)
    focusWorkspaceTab(`contest-workspace-tab-${nextTab.id}`)
  } else if (event.key === 'Home') {
    selectWorkspaceTab(workspaceTabs[0].id)
    focusWorkspaceTab(`contest-workspace-tab-${workspaceTabs[0].id}`)
  } else if (event.key === 'End') {
    const lastTab = workspaceTabs[workspaceTabs.length - 1]
    selectWorkspaceTab(lastTab.id)
    focusWorkspaceTab(`contest-workspace-tab-${lastTab.id}`)
  } else {
    return
  }

  event.preventDefault()
}
</script>

<template>
  <div class="contest-page-shell" :style="contestAccentStyle">
    <section
      class="journal-shell journal-shell-user journal-hero contest-detail-view min-h-full rounded-[30px] border"
    >
      <div v-if="loading" class="contest-loading">
        <div class="contest-loading__spinner" />
        <div class="contest-loading__text">正在同步竞赛详情...</div>
      </div>

      <template v-else-if="contest">
        <ContestAnnouncementRealtimeBridge :contest-id="contest.id" @updated="refreshAnnouncements" />

        <div class="workspace-tabbar" role="tablist" aria-label="竞赛页面主切换">
          <button
            v-for="tab in workspaceTabs"
            :id="`contest-workspace-tab-${tab.id}`"
            :key="tab.id"
            type="button"
            role="tab"
            class="workspace-tab"
            :class="{ 'workspace-tab--active': activeWorkspaceTab === tab.id }"
            :aria-selected="activeWorkspaceTab === tab.id"
            :aria-controls="`contest-workspace-panel-${tab.id}`"
            :tabindex="activeWorkspaceTab === tab.id ? 0 : -1"
            @click="selectWorkspaceTab(tab.id)"
            @keydown="handleWorkspaceTabKeydown($event, tab.id)"
          >
            {{ tab.label }}
          </button>
        </div>

        <section
          v-if="activeWorkspaceTab === 'overview'"
          id="contest-workspace-panel-overview"
          class="workspace-panel"
          role="tabpanel"
          aria-labelledby="contest-workspace-tab-overview"
        >
          <header class="contest-hero">
            <div class="contest-hero__main">
              <div class="contest-overline">Contest</div>
              <h1 class="contest-hero__title">{{ contest.title }}</h1>
              <p class="contest-hero__desc">
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
                <span v-if="countdown" class="contest-chip contest-chip--accent">
                  {{ countdown }}
                </span>
              </div>
            </div>

            <aside class="contest-score-rail">
              <div class="contest-score-rail__label">总分</div>
              <div class="contest-score-rail__value">{{ totalPoints }} <small>pts</small></div>
              <div class="contest-score-rail__note">
                {{ challenges.length }} 题 · {{ solvedCount }} 已解 · {{ memberCount }} 人
              </div>
            </aside>
          </header>

          <div class="contest-divider" />

          <section class="contest-stat-grid">
            <article class="contest-stat">
              <div class="contest-stat__label">队伍成员</div>
              <div class="contest-stat__value">{{ memberCount }}</div>
              <div class="contest-stat__hint">当前队伍人数</div>
            </article>
            <article class="contest-stat">
              <div class="contest-stat__label">题目数量</div>
              <div class="contest-stat__value">{{ challenges.length }}</div>
              <div class="contest-stat__hint">本场竞赛题目总数</div>
            </article>
            <article class="contest-stat">
              <div class="contest-stat__label">已解题目</div>
              <div class="contest-stat__value">{{ solvedCount }}</div>
              <div class="contest-stat__hint">当前账号已完成数量</div>
            </article>
            <article class="contest-stat">
              <div class="contest-stat__label">积分总览</div>
              <div class="contest-stat__value">{{ totalPoints }}</div>
              <div class="contest-stat__hint">全部题目可获得积分</div>
            </article>
          </section>

          <div class="contest-divider" />

          <div class="contest-overview-grid">
            <section class="contest-section contest-section--flat">
              <div class="contest-section__head workspace-tab-heading">
                <div class="workspace-tab-heading__main">
                  <div class="contest-overline">Rules</div>
                  <h2 class="contest-section__title workspace-tab-heading__title">竞赛规则</h2>
                </div>
              </div>
              <div class="contest-copy">{{ contest.rules || '当前竞赛暂无额外规则说明。' }}</div>
            </section>

            <section class="contest-section contest-section--flat">
              <div class="contest-section__head workspace-tab-heading">
                <div class="workspace-tab-heading__main">
                  <div class="contest-overline">Schedule</div>
                  <h2 class="contest-section__title workspace-tab-heading__title">赛程信息</h2>
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
                <div class="contest-overline">Announcements</div>
                <h2 class="contest-section__title workspace-tab-heading__title">公告预览</h2>
              </div>
              <div class="contest-section__hint">{{ announcements.length }} 条</div>
            </div>

            <div v-if="announcementsError" class="contest-alert contest-alert--warning">
              {{ announcementsError }}
            </div>

            <div v-else-if="announcements.length === 0" class="contest-inline-note">
              当前竞赛暂无新的公告通知。
            </div>

            <div v-else class="announcement-list">
              <article
                v-for="announcement in announcements"
                :key="announcement.id"
                class="announcement-item"
              >
                <div class="announcement-item__head">
                  <h3 class="announcement-item__title">{{ announcement.title }}</h3>
                  <time class="announcement-item__time" :datetime="announcement.created_at">
                    {{ formatTime(announcement.created_at) }}
                  </time>
                </div>
                <p v-if="announcement.content" class="announcement-item__content">
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
                <div class="contest-overline">Announcements</div>
                <h2 class="contest-section__title workspace-tab-heading__title">公告</h2>
              </div>
              <div class="contest-section__hint">{{ announcements.length }} 条</div>
            </div>

            <div v-if="announcementsError" class="contest-alert contest-alert--warning">
              {{ announcementsError }}
            </div>

            <div v-else-if="announcements.length === 0" class="contest-empty-state">
              <AppEmpty icon="Bell" title="暂无公告" description="当前竞赛暂无新的公告通知。" />
            </div>

            <div v-else class="announcement-list">
              <article
                v-for="announcement in announcements"
                :key="announcement.id"
                class="announcement-item"
              >
                <div class="announcement-item__head">
                  <h3 class="announcement-item__title">{{ announcement.title }}</h3>
                  <time class="announcement-item__time" :datetime="announcement.created_at">
                    {{ formatTime(announcement.created_at) }}
                  </time>
                </div>
                <p v-if="announcement.content" class="announcement-item__content">
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
                <div class="contest-overline">Challenges</div>
                <h2 class="contest-section__title workspace-tab-heading__title">题目</h2>
              </div>
              <div class="contest-section__hint">{{ solvedCount }} / {{ challenges.length }} 已解</div>
            </div>

            <div v-if="challenges.length === 0" class="contest-empty-state">
              <AppEmpty icon="Flag" title="暂无题目" description="当前竞赛尚未发布题目。" />
            </div>

            <div v-else class="contest-challenge-workspace">
              <div class="contest-challenge-list">
                <button
                  v-for="challenge in challenges"
                  :key="challenge.id"
                  type="button"
                  :class="challengeClass(challenge.id, challenge.is_solved)"
                  @click="selectChallenge(challenge)"
                >
                  <div class="contest-challenge__head">
                    <h3 class="contest-challenge__title">{{ challenge.title }}</h3>
                    <span v-if="challenge.is_solved" class="contest-challenge__solved">✓</span>
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
                      <div class="contest-overline">Selected</div>
                      <h3 class="challenge-focus__title">{{ selectedChallenge.title }}</h3>
                    </div>
                    <div class="challenge-focus__meta">{{ selectedChallengeMeta }}</div>
                  </div>

                  <div class="challenge-focus__stats">
                    <span class="contest-chip contest-chip--neutral">
                      解出人数 {{ selectedChallenge.solved_count }}
                    </span>
                    <span v-if="selectedChallenge.is_solved" class="contest-chip contest-chip--success">
                      已解出
                    </span>
                  </div>

                  <div class="contest-divider contest-divider--compact" />

                  <div class="challenge-focus__form">
                    <div>
                      <div class="contest-overline">Primary Action</div>
                      <h4 class="challenge-focus__form-title">提交 Flag</h4>
                    </div>

                    <label class="flag-submit__label" for="contest-flag-input">Flag</label>
                    <div class="flag-submit">
                      <input
                        id="contest-flag-input"
                        v-model="flagInput"
                        type="text"
                        placeholder="flag{...}"
                        class="flag-submit__input"
                        @keyup.enter="submitFlagAction"
                      />
                      <button
                        type="button"
                        :disabled="submitting"
                        class="contest-btn contest-btn--primary"
                        @click="submitFlagAction"
                      >
                        {{ submitting ? '提交中...' : '提交' }}
                      </button>
                    </div>

                    <div
                      v-if="submitResult"
                      class="contest-alert"
                      :class="submitResult.is_correct ? 'contest-alert--success' : 'contest-alert--danger'"
                    >
                      {{
                        submitResult.is_correct
                          ? `正确！+${submitResult.points ?? 0} 分`
                          : submitResult.message
                      }}
                    </div>
                  </div>
                </template>

                <div v-else class="contest-inline-note">从左侧选择题目后可在这里提交 Flag。</div>
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
                <div class="contest-overline">Team</div>
                <h2 class="contest-section__title workspace-tab-heading__title">队伍</h2>
              </div>
              <div class="contest-section__hint">{{ memberCount }} 人</div>
            </div>

            <div v-if="!team" class="team-empty">
              <div class="contest-inline-note">当前账号尚未加入队伍。</div>
              <div class="team-actions">
                <button type="button" class="contest-btn contest-btn--primary" @click="openCreateTeam">
                  创建队伍
                </button>
                <button type="button" class="contest-btn contest-btn--ghost" @click="openJoinTeam">
                  加入队伍
                </button>
              </div>
            </div>

            <div v-else class="team-board">
              <div class="team-summary">
                <div>
                  <div class="contest-overline">Current Team</div>
                  <h3 class="team-summary__name">{{ team.name }}</h3>
                </div>
                <span v-if="team.invite_code" class="team-summary__invite">邀请码: {{ team.invite_code }}</span>
              </div>

              <div class="team-member-list">
                <div v-for="member in team.members" :key="member.user_id" class="team-member">
                  <span class="team-member__name">{{ member.username }}</span>
                  <div class="team-member__actions">
                    <span v-if="member.user_id === team.captain_user_id" class="team-member__captain">
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
      </template>

      <div v-else class="contest-not-found">
        <AppEmpty
          icon="AlertTriangle"
          title="竞赛不存在或暂不可用"
          description="请返回竞赛中心重新选择竞赛，或稍后再试。"
        >
          <template #action>
            <RouterLink class="contest-btn contest-btn--primary" to="/contests">
              <Trophy class="h-4 w-4" />
              返回竞赛中心
            </RouterLink>
          </template>
        </AppEmpty>
      </div>
    </section>

    <div v-if="showCreateTeam" class="contest-modal-overlay" @click.self="closeCreateTeam">
      <div class="contest-modal">
        <h3 class="contest-modal__title">创建队伍</h3>
        <input
          v-model="teamName"
          type="text"
          placeholder="队伍名称"
          class="contest-modal__input"
          @keyup.enter="createTeamAction"
        />
        <div class="contest-modal__actions">
          <button type="button" class="contest-btn contest-btn--ghost" @click="closeCreateTeam">
            取消
          </button>
          <button
            type="button"
            :disabled="creatingTeam"
            class="contest-btn contest-btn--primary"
            @click="createTeamAction"
          >
            {{ creatingTeam ? '创建中...' : '创建' }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="showJoinTeam" class="contest-modal-overlay" @click.self="closeJoinTeam">
      <div class="contest-modal">
        <h3 class="contest-modal__title">加入队伍</h3>
        <input
          v-model="teamIdInput"
          type="text"
          placeholder="队伍 ID"
          class="contest-modal__input"
          @keyup.enter="joinTeamAction"
        />
        <div class="contest-modal__actions">
          <button type="button" class="contest-btn contest-btn--ghost" @click="closeJoinTeam">
            取消
          </button>
          <button
            type="button"
            :disabled="joiningTeam"
            class="contest-btn contest-btn--primary"
            @click="joinTeamAction"
          >
            {{ joiningTeam ? '加入中...' : '加入' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.contest-page-shell {
  --contest-accent: var(--color-primary);
  --journal-shell-accent: var(--contest-accent);
  --journal-shell-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-shell-surface: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --journal-shell-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 76%, var(--color-bg-base));
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

.journal-shell {
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
  font-size: 0.88rem;
  color: var(--journal-muted);
}

.workspace-tabbar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 86%, transparent);
  padding-bottom: 0.9rem;
}

.workspace-tab {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.5rem;
  border: 0;
  border-bottom: 2px solid transparent;
  background: transparent;
  padding: 0.4rem 0.1rem 0.65rem;
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--journal-muted);
  transition:
    color 150ms ease,
    border-color 150ms ease;
}

.workspace-tab:hover,
.workspace-tab:focus-visible {
  color: var(--journal-ink);
  outline: none;
}

.workspace-tab--active {
  border-bottom-color: color-mix(in srgb, var(--contest-accent) 72%, transparent);
  color: var(--journal-ink);
}

.workspace-panel {
  padding-top: 1.35rem;
}

.contest-overline {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--contest-accent) 88%, var(--journal-ink));
}

.contest-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 15rem;
  gap: 1.25rem;
}

.contest-hero__title {
  margin-top: 0.85rem;
  font-size: clamp(2rem, 3vw, 2.9rem);
  font-weight: 700;
  line-height: 1.08;
  color: var(--journal-ink);
}

.contest-hero__desc {
  margin-top: 0.8rem;
  max-width: 60ch;
  font-size: 0.95rem;
  line-height: 1.8;
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
  font-size: 0.76rem;
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
  font-size: 0.74rem;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.contest-score-rail__value {
  margin-top: 0.7rem;
  font-size: 2.1rem;
  font-weight: 700;
  line-height: 1;
  color: var(--journal-ink);
}

.contest-score-rail__value small {
  font-size: 0.85rem;
  color: var(--journal-muted);
}

.contest-score-rail__note {
  margin-top: 0.7rem;
  font-size: 0.84rem;
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
  display: grid;
  gap: 0.85rem;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.contest-stat {
  min-width: 0;
}

.contest-stat__label {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.contest-stat__value {
  margin-top: 0.45rem;
  font-size: 1.65rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.contest-stat__hint {
  margin-top: 0.35rem;
  font-size: 0.82rem;
  line-height: 1.6;
  color: var(--journal-muted);
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
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.contest-section__hint {
  font-size: 0.82rem;
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
  font-size: 0.92rem;
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
  font-size: 0.88rem;
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
  font-size: 0.96rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.announcement-item__time {
  font-size: 0.76rem;
  color: var(--journal-muted);
}

.announcement-item__content {
  margin-top: 0.55rem;
  white-space: pre-wrap;
  font-size: 0.88rem;
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
  border-inline-start: 2px solid color-mix(in srgb, var(--contest-accent) 24%, var(--journal-border));
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
  font-size: 0.9rem;
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
  font-size: 0.76rem;
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
  font-size: 1.2rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.challenge-focus__meta {
  font-size: 0.82rem;
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
  font-size: 1rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.flag-submit__label {
  display: inline-flex;
  margin-top: 1rem;
  font-size: 0.84rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.flag-submit {
  margin-top: 0.55rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
}

.flag-submit__input {
  flex: 1 1 18rem;
  min-width: 0;
  min-height: 2.85rem;
  border-radius: 16px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 82%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 95%, var(--color-bg-base));
  padding: 0.7rem 0.95rem;
  color: var(--journal-ink);
  outline: none;
  transition:
    border-color 150ms ease,
    box-shadow 150ms ease;
}

.flag-submit__input:focus {
  border-color: color-mix(in srgb, var(--contest-accent) 46%, transparent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--contest-accent) 10%, transparent);
}

.contest-alert {
  margin-top: 0.8rem;
  border-inline-start: 2px solid transparent;
  padding: 0.6rem 0.75rem;
  font-size: 0.84rem;
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
  font-size: 0.88rem;
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
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.team-summary__invite {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', monospace;
  font-size: 0.78rem;
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
  font-size: 0.9rem;
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
  font-size: 0.72rem;
  font-weight: 700;
  color: color-mix(in srgb, var(--contest-accent) 84%, var(--journal-ink));
}

.team-member__kick {
  border: 0;
  background: transparent;
  padding: 0;
  font-size: 0.78rem;
  color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
}

.team-member__kick:hover,
.team-member__kick:focus-visible {
  text-decoration: underline;
  outline: none;
}

.contest-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  min-height: 2.7rem;
  border-radius: 999px;
  border: 1px solid transparent;
  padding: 0.65rem 1rem;
  font-size: 0.88rem;
  font-weight: 600;
  transition:
    border-color 150ms ease,
    background 150ms ease,
    color 150ms ease;
}

.contest-btn:disabled {
  opacity: 0.58;
  cursor: not-allowed;
}

.contest-btn--primary {
  border-color: color-mix(in srgb, var(--contest-accent) 24%, transparent);
  background: color-mix(in srgb, var(--contest-accent) 12%, var(--journal-surface));
  color: color-mix(in srgb, var(--contest-accent) 88%, var(--journal-ink));
}

.contest-btn--ghost {
  border-color: color-mix(in srgb, var(--journal-border) 84%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base));
  color: var(--journal-ink);
}

.contest-not-found {
  display: grid;
  align-content: center;
}

.contest-modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: grid;
  place-items: center;
  background: color-mix(in srgb, black 45%, transparent);
  padding: 1rem;
}

.contest-modal {
  width: min(28rem, 100%);
  border: 1px solid var(--journal-border);
  border-radius: 24px;
  background: color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base));
  padding: 1rem;
}

.contest-modal__title {
  font-size: 1rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.contest-modal__input {
  margin-top: 0.8rem;
  width: 100%;
  min-height: 2.8rem;
  border-radius: 16px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 82%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  padding: 0.65rem 0.9rem;
  color: var(--journal-ink);
  outline: none;
}

.contest-modal__input:focus {
  border-color: color-mix(in srgb, var(--contest-accent) 44%, transparent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--contest-accent) 10%, transparent);
}

.contest-modal__actions {
  margin-top: 0.8rem;
  display: flex;
  justify-content: flex-end;
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
  .workspace-tabbar {
    overflow-x: auto;
    flex-wrap: nowrap;
  }

  .workspace-tab {
    flex: 0 0 auto;
    min-width: max-content;
  }

  .flag-submit {
    flex-direction: column;
  }
}
</style>
