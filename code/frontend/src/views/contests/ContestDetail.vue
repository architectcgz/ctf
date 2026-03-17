<template>
  <div class="space-y-6">
    <div
      v-if="loading"
      class="flex justify-center py-12"
    >
      <div
        class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--color-border-default)] border-t-[var(--color-primary)]"
      />
    </div>

    <div
      v-else-if="contest"
      class="space-y-6"
    >
      <!-- 竞赛信息 -->
      <div
        class="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6"
      >
        <h1 class="text-3xl font-bold text-[var(--color-text-primary)]">
          {{ contest.title }}
        </h1>
        <p class="mt-3 text-[var(--color-text-secondary)]">
          {{ contest.description }}
        </p>

        <div class="mt-4 flex items-center gap-4 text-sm">
          <span
            class="rounded px-2 py-0.5 text-xs font-medium"
            :class="getStatusBadgeClass(contest.status)"
          >
            {{ getStatusLabel(contest.status) }}
          </span>
          <span class="text-[var(--color-text-secondary)]">{{ getModeLabel(contest.mode) }}</span>
        </div>

        <div class="mt-4 font-mono text-sm text-[var(--color-text-secondary)]">
          {{ formatTime(contest.starts_at) }} ~ {{ formatTime(contest.ends_at) }}
        </div>

        <!-- 倒计时 -->
        <div
          v-if="countdown"
          class="mt-4 text-lg font-semibold text-[var(--color-primary)]"
        >
          {{ countdown }}
        </div>
      </div>

      <!-- 队伍管理 -->
      <div
        class="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6"
      >
        <h2 class="text-xl font-bold text-[var(--color-text-primary)]">
          队伍
        </h2>

        <div
          v-if="!team"
          class="mt-4 space-y-3"
        >
          <button
            class="rounded bg-[var(--color-primary)] px-4 py-2 text-sm font-medium text-white hover:opacity-90"
            @click="openCreateTeam"
          >
            创建队伍
          </button>
          <button
            class="ml-2 rounded border border-[var(--color-border-default)] px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] hover:bg-[var(--color-bg-hover)]"
            @click="openJoinTeam"
          >
            加入队伍
          </button>
        </div>

        <div
          v-else
          class="mt-4"
        >
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-[var(--color-text-primary)]">
              {{ team.name }}
            </h3>
            <span
              v-if="team.invite_code"
              class="font-mono text-sm text-[var(--color-text-secondary)]"
            >邀请码: {{ team.invite_code }}</span>
          </div>
          <div class="mt-3 space-y-2">
            <div
              v-for="member in team.members"
              :key="member.user_id"
              class="flex items-center justify-between rounded border border-[var(--color-border-default)] p-2"
            >
              <span class="text-[var(--color-text-primary)]">{{ member.username }}</span>
              <div class="flex items-center gap-2">
                <span
                  v-if="member.user_id === team.captain_user_id"
                  class="text-xs text-[var(--color-text-muted)]"
                >队长</span>
                <button
                  v-if="isCaptain && member.user_id !== team.captain_user_id"
                  class="text-xs text-red-500 hover:underline"
                  @click="kickMember(member.user_id)"
                >
                  踢出
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 竞赛公告 -->
      <div
        class="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6"
      >
        <div class="flex items-center justify-between gap-3">
          <h2 class="text-xl font-bold text-[var(--color-text-primary)]">
            公告
          </h2>
          <span class="text-sm text-[var(--color-text-muted)]">{{ announcements.length }} 条</span>
        </div>

        <div
          v-if="announcementsError"
          class="mt-4 rounded border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-sm text-amber-600"
        >
          {{ announcementsError }}
        </div>
        <div
          v-else-if="announcements.length === 0"
          class="mt-4 text-center text-[var(--color-text-muted)]"
        >
          暂无公告
        </div>
        <div
          v-else
          class="mt-4 space-y-3"
        >
          <article
            v-for="announcement in announcements"
            :key="announcement.id"
            class="rounded border border-[var(--color-border-default)] p-4"
          >
            <div class="flex items-start justify-between gap-4">
              <h3 class="text-base font-semibold text-[var(--color-text-primary)]">
                {{ announcement.title }}
              </h3>
              <time
                class="shrink-0 text-xs text-[var(--color-text-muted)]"
                :datetime="announcement.created_at"
              >
                {{ formatTime(announcement.created_at) }}
              </time>
            </div>
            <p
              v-if="announcement.content"
              class="mt-2 whitespace-pre-wrap text-sm leading-6 text-[var(--color-text-secondary)]"
            >
              {{ announcement.content }}
            </p>
          </article>
        </div>
      </div>

      <!-- 题目列表 -->
      <div
        class="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6"
      >
        <h2 class="text-xl font-bold text-[var(--color-text-primary)]">
          题目
        </h2>
        <div
          v-if="challenges.length === 0"
          class="mt-4 text-center text-[var(--color-text-muted)]"
        >
          暂无题目
        </div>
        <div
          v-else
          class="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-3"
        >
          <div
            v-for="chal in challenges"
            :key="chal.id"
            class="cursor-pointer rounded border border-[var(--color-border-default)] p-4 hover:border-[var(--color-primary)]"
            @click="selectChallenge(chal)"
          >
            <div class="flex items-center justify-between">
              <h3 class="font-semibold text-[var(--color-text-primary)]">
                {{ chal.title }}
              </h3>
              <span
                v-if="chal.is_solved"
                class="text-green-500"
              >✓</span>
            </div>
            <div class="mt-2 flex items-center gap-2 text-xs text-[var(--color-text-secondary)]">
              <span>{{ chal.category }}</span>
              <span>{{ chal.points }} pts</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Flag 提交 -->
      <div
        v-if="selectedChallenge"
        class="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6"
      >
        <h2 class="text-xl font-bold text-[var(--color-text-primary)]">
          提交 Flag - {{ selectedChallenge.title }}
        </h2>
        <div class="mt-4 flex gap-2">
          <input
            v-model="flagInput"
            placeholder="flag{...}"
            class="flex-1 rounded border border-[var(--color-border-default)] bg-[var(--color-bg-default)] px-3 py-2 text-[var(--color-text-primary)]"
            @keyup.enter="submitFlagAction"
          >
          <button
            :disabled="submitting"
            class="rounded bg-[var(--color-primary)] px-4 py-2 text-sm font-medium text-white hover:opacity-90 disabled:opacity-50"
            @click="submitFlagAction"
          >
            {{ submitting ? '提交中...' : '提交' }}
          </button>
        </div>
        <div
          v-if="submitResult"
          class="mt-3 rounded p-3"
          :class="
            submitResult.is_correct
              ? 'bg-green-500/10 text-green-500'
              : 'bg-red-500/10 text-red-500'
          "
        >
          {{
            submitResult.is_correct ? `正确！+${submitResult.points ?? 0} 分` : submitResult.message
          }}
        </div>
      </div>
    </div>

    <!-- 创建队伍弹窗 -->
    <div
      v-if="showCreateTeam"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      @click.self="closeCreateTeam"
    >
      <div
        class="w-full max-w-md rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6"
      >
        <h3 class="text-lg font-bold text-[var(--color-text-primary)]">
          创建队伍
        </h3>
        <input
          v-model="teamName"
          placeholder="队伍名称"
          class="mt-4 w-full rounded border border-[var(--color-border-default)] bg-[var(--color-bg-default)] px-3 py-2 text-[var(--color-text-primary)]"
          @keyup.enter="createTeamAction"
        >
        <div class="mt-4 flex justify-end gap-2">
          <button
            class="rounded border border-[var(--color-border-default)] px-4 py-2 text-sm text-[var(--color-text-primary)]"
            @click="closeCreateTeam"
          >
            取消
          </button>
          <button
            :disabled="creatingTeam"
            class="rounded bg-[var(--color-primary)] px-4 py-2 text-sm text-white disabled:opacity-50"
            @click="createTeamAction"
          >
            {{ creatingTeam ? '创建中...' : '创建' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 加入队伍弹窗 -->
    <div
      v-if="showJoinTeam"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      @click.self="closeJoinTeam"
    >
      <div
        class="w-full max-w-md rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6"
      >
        <h3 class="text-lg font-bold text-[var(--color-text-primary)]">
          加入队伍
        </h3>
        <input
          v-model="teamIdInput"
          placeholder="队伍 ID"
          class="mt-4 w-full rounded border border-[var(--color-border-default)] bg-[var(--color-bg-default)] px-3 py-2 text-[var(--color-text-primary)]"
          @keyup.enter="joinTeamAction"
        >
        <div class="mt-4 flex justify-end gap-2">
          <button
            class="rounded border border-[var(--color-border-default)] px-4 py-2 text-sm text-[var(--color-text-primary)]"
            @click="closeJoinTeam"
          >
            取消
          </button>
          <button
            :disabled="joiningTeam"
            class="rounded bg-[var(--color-primary)] px-4 py-2 text-sm text-white disabled:opacity-50"
            @click="joinTeamAction"
          >
            {{ joiningTeam ? '加入中...' : '加入' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useContestDetailPage } from '@/composables/useContestDetailPage'
import { useAuthStore } from '@/stores/auth'
import { formatTime } from '@/utils/format'
import { getStatusLabel, getModeLabel, getStatusBadgeClass } from '@/utils/contest'

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
</script>
