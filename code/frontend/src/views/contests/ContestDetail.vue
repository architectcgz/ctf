<script setup lang="ts">
import { computed } from 'vue'
import {
  BellRing,
  CalendarRange,
  Clock3,
  Flag,
  Swords,
  Trophy,
  UsersRound,
} from 'lucide-vue-next'
import { RouterLink, useRoute } from 'vue-router'

import AppEmpty from '@/components/common/AppEmpty.vue'
import { useContestDetailPage } from '@/composables/useContestDetailPage'
import { useAuthStore } from '@/stores/auth'
import { getContestAccentColor, getModeLabel, getStatusLabel } from '@/utils/contest'
import { formatTime } from '@/utils/format'

const route = useRoute()
const authStore = useAuthStore()
const contestId = computed(() => String(route.params.id ?? ''))
const currentUserId = computed(() => authStore.user?.id)
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
} = useContestDetailPage({
  contestId,
  currentUserId,
})

const solvedCount = computed(() => challenges.value.filter((item) => item.is_solved).length)
const totalPoints = computed(() => challenges.value.reduce((sum, item) => sum + (item.points || 0), 0))
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
  return ['contest-challenge', active ? 'contest-challenge--active' : '', solved ? 'contest-challenge--solved' : '']
}
</script>

<template>
  <div
    class="contest-detail-view space-y-6"
    :style="contestAccentStyle"
  >
    <div
      v-if="loading"
      class="contest-loading"
    >
      <div class="contest-loading__spinner" />
      <div class="contest-loading__text">
        正在同步竞赛详情...
      </div>
    </div>

    <div
      v-else-if="contest"
      class="contest-shell-card rounded-[30px] border px-6 py-6 md:px-8"
    >
      <section class="contest-hero">
        <div class="contest-hero__kicker">
          Contest Mission Control
        </div>
        <h1 class="contest-hero__title">
          {{ contest.title }}
        </h1>
        <p class="contest-hero__desc">
          {{ contest.description || '当前竞赛暂未提供描述，进入题目区后可直接开始解题与提交。' }}
        </p>

        <div class="contest-hero__chips">
          <span
            class="contest-chip contest-chip--status"
          >
            {{ getStatusLabel(contest.status) }}
          </span>
          <span class="contest-chip contest-chip--neutral">
            {{ getModeLabel(contest.mode) }}
          </span>
        </div>

        <div class="contest-hero__meta">
          <div class="contest-hero__meta-item">
            <CalendarRange class="h-4 w-4" />
            <span>{{ formatTime(contest.starts_at) }} ~ {{ formatTime(contest.ends_at) }}</span>
          </div>
          <div
            v-if="countdown"
            class="contest-hero__meta-item contest-hero__meta-item--strong"
          >
            <Clock3 class="h-4 w-4" />
            <span>{{ countdown }}</span>
          </div>
        </div>
      </section>

      <section class="contest-kpis">
        <article class="contest-kpi">
          <div class="contest-kpi__label">
            队伍成员
          </div>
          <div class="contest-kpi__value">
            {{ memberCount }}
          </div>
          <div class="contest-kpi__hint">
            当前已加入队伍的人数
          </div>
        </article>
        <article class="contest-kpi">
          <div class="contest-kpi__label">
            题目数量
          </div>
          <div class="contest-kpi__value">
            {{ challenges.length }}
          </div>
          <div class="contest-kpi__hint">
            本场竞赛可解题目总数
          </div>
        </article>
        <article class="contest-kpi">
          <div class="contest-kpi__label">
            已解题目
          </div>
          <div class="contest-kpi__value">
            {{ solvedCount }}
          </div>
          <div class="contest-kpi__hint">
            当前账号已完成题目数量
          </div>
        </article>
        <article class="contest-kpi">
          <div class="contest-kpi__label">
            题目总分
          </div>
          <div class="contest-kpi__value">
            {{ totalPoints }}
          </div>
          <div class="contest-kpi__hint">
            全部题目可获得积分
          </div>
        </article>
      </section>

      <section class="contest-panel">
        <header class="contest-panel__header">
          <div class="contest-panel__title-wrap">
            <UsersRound class="h-4 w-4" />
            <h2 class="contest-panel__title">
              队伍
            </h2>
          </div>
          <span class="contest-panel__meta">{{ memberCount }} 人</span>
        </header>

        <div
          v-if="!team"
          class="team-actions"
        >
          <button
            type="button"
            class="contest-btn contest-btn--primary"
            @click="openCreateTeam"
          >
            创建队伍
          </button>
          <button
            type="button"
            class="contest-btn contest-btn--ghost"
            @click="openJoinTeam"
          >
            加入队伍
          </button>
        </div>

        <div
          v-else
          class="space-y-3"
        >
          <div class="team-summary">
            <h3 class="team-summary__name">
              {{ team.name }}
            </h3>
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

      <section class="contest-panel">
        <header class="contest-panel__header">
          <div class="contest-panel__title-wrap">
            <BellRing class="h-4 w-4" />
            <h2 class="contest-panel__title">
              公告
            </h2>
          </div>
          <span class="contest-panel__meta">{{ announcements.length }} 条</span>
        </header>

        <div
          v-if="announcementsError"
          class="contest-alert contest-alert--warning"
        >
          {{ announcementsError }}
        </div>

        <AppEmpty
          v-else-if="announcements.length === 0"
          icon="Bell"
          title="暂无公告"
          description="当前竞赛暂无新的公告通知。"
        />

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

      <section class="contest-panel">
        <header class="contest-panel__header">
          <div class="contest-panel__title-wrap">
            <Swords class="h-4 w-4" />
            <h2 class="contest-panel__title">
              题目
            </h2>
          </div>
          <span class="contest-panel__meta">{{ solvedCount }} / {{ challenges.length }} 已解</span>
        </header>

        <AppEmpty
          v-if="challenges.length === 0"
          icon="Flag"
          title="暂无题目"
          description="当前竞赛尚未发布题目。"
        />

        <div
          v-else
          class="contest-challenge-grid"
        >
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
              >
                ✓
              </span>
            </div>
            <div class="contest-challenge__meta">
              <span>{{ challenge.category }}</span>
              <span>{{ challenge.points }} pts</span>
            </div>
          </button>
        </div>
      </section>

      <section
        v-if="selectedChallenge"
        class="contest-panel contest-panel--flag"
      >
        <header class="contest-panel__header">
          <div class="contest-panel__title-wrap">
            <Flag class="h-4 w-4" />
            <h2 class="contest-panel__title">
              提交 Flag - {{ selectedChallenge.title }}
            </h2>
          </div>
          <span class="contest-panel__meta">{{ selectedChallengeMeta }}</span>
        </header>

        <div class="flag-submit">
          <input
            v-model="flagInput"
            placeholder="flag{...}"
            class="flag-submit__input"
            @keyup.enter="submitFlagAction"
          >
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
          {{ submitResult.is_correct ? `正确！+${submitResult.points ?? 0} 分` : submitResult.message }}
        </div>
      </section>
    </div>

    <div
      v-else
      class="contest-not-found"
    >
      <AppEmpty
        icon="AlertTriangle"
        title="竞赛不存在或暂不可用"
        description="请返回竞赛中心重新选择竞赛，或稍后再试。"
      >
        <template #action>
          <RouterLink
            class="contest-btn contest-btn--primary"
            to="/contests"
          >
            <Trophy class="h-4 w-4" />
            返回竞赛中心
          </RouterLink>
        </template>
      </AppEmpty>
    </div>

    <div
      v-if="showCreateTeam"
      class="contest-modal-overlay"
      @click.self="closeCreateTeam"
    >
      <div class="contest-modal">
        <h3 class="contest-modal__title">
          创建队伍
        </h3>
        <input
          v-model="teamName"
          placeholder="队伍名称"
          class="contest-modal__input"
          @keyup.enter="createTeamAction"
        >
        <div class="contest-modal__actions">
          <button
            type="button"
            class="contest-btn contest-btn--ghost"
            @click="closeCreateTeam"
          >
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

    <div
      v-if="showJoinTeam"
      class="contest-modal-overlay"
      @click.self="closeJoinTeam"
    >
      <div class="contest-modal">
        <h3 class="contest-modal__title">
          加入队伍
        </h3>
        <input
          v-model="teamIdInput"
          placeholder="队伍 ID"
          class="contest-modal__input"
          @keyup.enter="joinTeamAction"
        >
        <div class="contest-modal__actions">
          <button
            type="button"
            class="contest-btn contest-btn--ghost"
            @click="closeJoinTeam"
          >
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
.contest-detail-view {
  --contest-accent: var(--color-primary);
  --journal-ink: #0f172a;
  --journal-muted: #64748b;
  --journal-border: rgba(226, 232, 240, 0.8);
  --journal-surface: #ffffff;
  --journal-surface-subtle: rgba(248, 250, 252, 0.92);
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.contest-shell-card {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--contest-accent) 8%, transparent), transparent 20rem),
    linear-gradient(180deg, rgba(248, 250, 252, 0.98), rgba(241, 245, 249, 0.95));
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.05);
}

.contest-loading {
  display: grid;
  justify-items: center;
  gap: 0.65rem;
  padding: 3rem 0;
}

.contest-loading__spinner {
  width: 2rem;
  height: 2rem;
  border-radius: 999px;
  border: 3px solid var(--color-border-default);
  border-top-color: var(--contest-accent);
  animation: contestDetailSpin 0.9s linear infinite;
}

.contest-loading__text {
  font-size: 0.86rem;
  color: var(--color-text-secondary);
}

.contest-hero {
  padding: 0.2rem 0 1rem;
}

.contest-hero__kicker {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--contest-accent) 24%, transparent);
  background: color-mix(in srgb, var(--contest-accent) 8%, transparent);
  padding: 0.2rem 0.75rem;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--contest-accent) 84%, var(--journal-ink));
}

.contest-hero__title {
  margin-top: 0.78rem;
  font-size: clamp(1.4rem, 3.5vw, 2rem);
  font-weight: 700;
  line-height: 1.2;
  color: var(--journal-ink);
}

.contest-hero__desc {
  margin-top: 0.56rem;
  font-size: 0.9rem;
  line-height: 1.7;
  color: var(--journal-muted);
}

.contest-hero__chips {
  margin-top: 0.78rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
}

.contest-chip {
  border-radius: 999px;
  padding: 0.2rem 0.62rem;
  font-size: 0.74rem;
  font-weight: 700;
}

.contest-chip--status {
  border: 1px solid color-mix(in srgb, var(--contest-accent) 34%, transparent);
  background: color-mix(in srgb, var(--contest-accent) 14%, transparent);
  color: color-mix(in srgb, var(--contest-accent) 84%, var(--color-text-primary));
}

.contest-chip--neutral {
  border: 1px solid color-mix(in srgb, var(--journal-border) 90%, transparent);
  color: var(--journal-muted);
}

.contest-hero__meta {
  margin-top: 0.78rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem 0.95rem;
}

.contest-hero__meta-item {
  display: inline-flex;
  align-items: center;
  gap: 0.38rem;
  font-size: 0.8rem;
  color: var(--journal-muted);
}

.contest-hero__meta-item--strong {
  color: color-mix(in srgb, var(--contest-accent) 88%, var(--color-text-primary));
  font-weight: 600;
}

.contest-kpis {
  display: grid;
  gap: 0.65rem;
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px dashed rgba(148, 163, 184, 0.58);
}

.contest-kpi {
  border-radius: 18px;
  border: 1px solid rgba(148, 163, 184, 0.14);
  background: rgba(255, 255, 255, 0.56);
  padding: 0.9rem 1rem;
}

.contest-kpi__label {
  font-size: 0.68rem;
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.contest-kpi__value {
  margin-top: 0.36rem;
  font-size: 1.5rem;
  font-weight: 700;
  line-height: 1.1;
  color: var(--journal-ink);
}

.contest-kpi__hint {
  margin-top: 0.32rem;
  font-size: 0.78rem;
  line-height: 1.5;
  color: var(--journal-muted);
}

.contest-panel {
  border-top: 1px dashed rgba(148, 163, 184, 0.58);
  padding-top: 0.9rem;
  margin-top: 1.5rem;
}

.contest-panel--flag {
  border-top-color: rgba(148, 163, 184, 0.58);
}

.contest-panel__header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.55rem;
  margin-bottom: 0.7rem;
}

.contest-panel__title-wrap {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  color: color-mix(in srgb, var(--contest-accent) 82%, var(--journal-ink));
}

.contest-panel__title {
  font-size: 1rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.contest-panel__meta {
  font-size: 0.78rem;
  color: var(--journal-muted);
}

.contest-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  border-radius: 10px;
  border: 1px solid transparent;
  padding: 0.5rem 0.82rem;
  font-size: 0.84rem;
  font-weight: 600;
  transition: all 180ms ease;
}

.contest-btn:disabled {
  opacity: 0.58;
  cursor: not-allowed;
}

.contest-btn--primary {
  border-color: color-mix(in srgb, var(--contest-accent) 45%, transparent);
  background: color-mix(in srgb, var(--contest-accent) 90%, #0b4f60);
  color: #f8feff;
}

.contest-btn--primary:hover:not(:disabled) {
  transform: translateY(-1px);
  filter: brightness(1.03);
}

.contest-btn--ghost {
  border-color: color-mix(in srgb, var(--journal-border) 86%, transparent);
  color: var(--journal-ink);
  background: rgba(255, 255, 255, 0.6);
}

.contest-btn--ghost:hover {
  border-color: color-mix(in srgb, var(--contest-accent) 32%, transparent);
}

.team-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.55rem;
}

.team-summary {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.5rem 0.8rem;
}

.team-summary__name {
  font-size: 1rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.team-summary__invite {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
  font-size: 0.76rem;
  color: var(--color-text-secondary);
}

.team-member-list {
  display: grid;
  gap: 0.45rem;
}

.team-member {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.4rem 0.75rem;
  border-bottom: 1px solid var(--color-border-default);
  padding: 0.58rem 0.2rem 0.58rem 0.15rem;
}

.team-member__name {
  font-size: 0.88rem;
  color: var(--color-text-primary);
}

.team-member__actions {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
}

.team-member__captain {
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--contest-accent) 38%, transparent);
  background: color-mix(in srgb, var(--contest-accent) 12%, transparent);
  padding: 0.12rem 0.42rem;
  font-size: 0.7rem;
  color: color-mix(in srgb, var(--contest-accent) 84%, var(--color-text-primary));
}

.team-member__kick {
  border: 0;
  background: transparent;
  font-size: 0.75rem;
  color: color-mix(in srgb, var(--color-danger) 88%, var(--color-text-primary));
}

.team-member__kick:hover {
  text-decoration: underline;
}

.announcement-list {
  display: grid;
  gap: 0.55rem;
}

.announcement-item {
  border-bottom: 1px solid var(--color-border-default);
  padding: 0.62rem 0.2rem 0.62rem 0.1rem;
}

.announcement-item__head {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.4rem 0.65rem;
}

.announcement-item__title {
  font-size: 0.92rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.announcement-item__time {
  font-size: 0.73rem;
  color: var(--color-text-muted);
}

.announcement-item__content {
  margin-top: 0.45rem;
  white-space: pre-wrap;
  font-size: 0.82rem;
  line-height: 1.65;
  color: var(--color-text-secondary);
}

.contest-challenge-grid {
  display: grid;
  gap: 0.52rem;
  grid-template-columns: repeat(auto-fill, minmax(13.5rem, 1fr));
}

.contest-challenge {
  border: 0;
  border-left: 2px solid color-mix(in srgb, var(--contest-accent) 24%, var(--color-border-default));
  border-bottom: 1px solid var(--color-border-default);
  background: transparent;
  padding: 0.62rem 0.4rem 0.62rem 0.72rem;
  text-align: left;
  transition: border-color 180ms ease, transform 180ms ease;
}

.contest-challenge:hover,
.contest-challenge:focus-visible {
  border-left-color: color-mix(in srgb, var(--contest-accent) 70%, var(--color-border-default));
  background: color-mix(in srgb, var(--contest-accent) 4%, transparent);
  outline: none;
}

.contest-challenge--active {
  border-left-color: color-mix(in srgb, var(--contest-accent) 88%, var(--color-border-default));
  background: color-mix(in srgb, var(--contest-accent) 6%, transparent);
}

.contest-challenge--solved {
  border-left-color: color-mix(in srgb, var(--color-success) 66%, var(--color-border-default));
}

.contest-challenge__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.55rem;
}

.contest-challenge__title {
  font-size: 0.87rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.contest-challenge__solved {
  color: var(--color-success);
  font-weight: 700;
}

.contest-challenge__meta {
  margin-top: 0.35rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem 0.65rem;
  font-size: 0.74rem;
  color: var(--color-text-secondary);
}

.flag-submit {
  display: flex;
  flex-wrap: wrap;
  gap: 0.52rem;
}

.flag-submit__input {
  flex: 1 1 15rem;
  min-width: 0;
  border-radius: 6px;
  border: 1px solid var(--color-border-default);
  background: transparent;
  padding: 0.58rem 0.68rem;
  color: var(--color-text-primary);
  outline: none;
}

.flag-submit__input:focus {
  border-color: color-mix(in srgb, var(--contest-accent) 54%, var(--color-border-default));
}

.contest-alert {
  margin-top: 0.72rem;
  border-left: 2px solid transparent;
  padding: 0.52rem 0.62rem;
  font-size: 0.82rem;
  line-height: 1.55;
}

.contest-alert--success {
  border-left-color: color-mix(in srgb, var(--color-success) 60%, transparent);
  background: color-mix(in srgb, var(--color-success) 6%, transparent);
  color: color-mix(in srgb, var(--color-success) 80%, var(--color-text-primary));
}

.contest-alert--danger {
  border-left-color: color-mix(in srgb, var(--color-danger) 62%, transparent);
  background: color-mix(in srgb, var(--color-danger) 6%, transparent);
  color: color-mix(in srgb, var(--color-danger) 82%, var(--color-text-primary));
}

.contest-alert--warning {
  border-left-color: color-mix(in srgb, var(--color-warning) 60%, transparent);
  background: color-mix(in srgb, var(--color-warning) 6%, transparent);
  color: color-mix(in srgb, var(--color-warning) 86%, var(--color-text-primary));
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
  border-radius: 10px;
  border: 1px solid var(--color-border-default);
  background: var(--color-bg-surface);
  padding: 0.9rem;
}

.contest-modal__title {
  font-size: 1rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.contest-modal__input {
  margin-top: 0.72rem;
  width: 100%;
  border-radius: 6px;
  border: 1px solid var(--color-border-default);
  background: transparent;
  padding: 0.55rem 0.62rem;
  color: var(--color-text-primary);
  outline: none;
}

.contest-modal__input:focus {
  border-color: color-mix(in srgb, var(--contest-accent) 52%, var(--color-border-default));
}

.contest-modal__actions {
  margin-top: 0.75rem;
  display: flex;
  justify-content: flex-end;
  gap: 0.48rem;
}

.contest-not-found :deep(.contest-btn) {
  text-decoration: none;
}

@media (min-width: 900px) {
  .contest-kpis {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

:global([data-theme='light']) .contest-hero {
  border-bottom-color: color-mix(in srgb, var(--contest-accent) 18%, var(--color-border-default));
}

@keyframes contestDetailSpin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
