<template>
  <div class="space-y-6">
    <div v-if="loading" class="flex justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--color-border-default)] border-t-[var(--color-primary)]"></div>
    </div>

    <div v-else-if="contest" class="space-y-6">
      <!-- 竞赛信息 -->
      <div class="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6">
        <h1 class="text-3xl font-bold text-[var(--color-text-primary)]">{{ contest.title }}</h1>
        <p class="mt-3 text-[var(--color-text-secondary)]">{{ contest.description }}</p>

        <div class="mt-4 flex items-center gap-4 text-sm">
          <span class="rounded px-2 py-0.5 text-xs font-medium" :class="getStatusBadgeClass(contest.status)">
            {{ getStatusLabel(contest.status) }}
          </span>
          <span class="text-[var(--color-text-secondary)]">{{ getModeLabel(contest.mode) }}</span>
        </div>

        <div class="mt-4 font-mono text-sm text-[var(--color-text-secondary)]">
          {{ formatTime(contest.starts_at) }} ~ {{ formatTime(contest.ends_at) }}
        </div>

        <!-- 倒计时 -->
        <div v-if="countdown" class="mt-4 text-lg font-semibold text-[var(--color-primary)]">
          {{ countdown }}
        </div>
      </div>

      <!-- 队伍管理 -->
      <div class="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6">
        <h2 class="text-xl font-bold text-[var(--color-text-primary)]">队伍</h2>

        <div v-if="!team" class="mt-4 space-y-3">
          <button @click="showCreateTeam = true" class="rounded bg-[var(--color-primary)] px-4 py-2 text-sm font-medium text-white hover:opacity-90">
            创建队伍
          </button>
          <button @click="showJoinTeam = true" class="ml-2 rounded border border-[var(--color-border-default)] px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] hover:bg-[var(--color-bg-hover)]">
            加入队伍
          </button>
        </div>

        <div v-else class="mt-4">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-[var(--color-text-primary)]">{{ team.name }}</h3>
            <span v-if="team.invite_code" class="font-mono text-sm text-[var(--color-text-secondary)]">邀请码: {{ team.invite_code }}</span>
          </div>
          <div class="mt-3 space-y-2">
            <div v-for="member in team.members" :key="member.user_id" class="flex items-center justify-between rounded border border-[var(--color-border-default)] p-2">
              <span class="text-[var(--color-text-primary)]">{{ member.username }}</span>
              <div class="flex items-center gap-2">
                <span v-if="member.user_id === team.captain_user_id" class="text-xs text-[var(--color-text-muted)]">队长</span>
                <button v-if="isCaptain && member.user_id !== team.captain_user_id" @click="kickMember(member.user_id)" class="text-xs text-red-500 hover:underline">
                  踢出
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 题目列表 -->
      <div class="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6">
        <h2 class="text-xl font-bold text-[var(--color-text-primary)]">题目</h2>
        <div v-if="challenges.length === 0" class="mt-4 text-center text-[var(--color-text-muted)]">暂无题目</div>
        <div v-else class="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <div v-for="chal in challenges" :key="chal.id" @click="selectChallenge(chal)" class="cursor-pointer rounded border border-[var(--color-border-default)] p-4 hover:border-[var(--color-primary)]">
            <div class="flex items-center justify-between">
              <h3 class="font-semibold text-[var(--color-text-primary)]">{{ chal.title }}</h3>
              <span v-if="chal.is_solved" class="text-green-500">✓</span>
            </div>
            <div class="mt-2 flex items-center gap-2 text-xs text-[var(--color-text-secondary)]">
              <span>{{ chal.category }}</span>
              <span>{{ chal.points }} pts</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Flag 提交 -->
      <div v-if="selectedChallenge" class="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6">
        <h2 class="text-xl font-bold text-[var(--color-text-primary)]">提交 Flag - {{ selectedChallenge.title }}</h2>
        <div class="mt-4 flex gap-2">
          <input v-model="flagInput" @keyup.enter="submitFlag" placeholder="flag{...}" class="flex-1 rounded border border-[var(--color-border-default)] bg-[var(--color-bg-default)] px-3 py-2 text-[var(--color-text-primary)]" />
          <button @click="submitFlag" :disabled="submitting" class="rounded bg-[var(--color-primary)] px-4 py-2 text-sm font-medium text-white hover:opacity-90 disabled:opacity-50">
            {{ submitting ? '提交中...' : '提交' }}
          </button>
        </div>
        <div v-if="submitResult" class="mt-3 rounded p-3" :class="submitResult.correct ? 'bg-green-500/10 text-green-500' : 'bg-red-500/10 text-red-500'">
          {{ submitResult.correct ? `正确！+${submitResult.points_earned} 分` : '错误' }}
        </div>
      </div>
    </div>

    <!-- 创建队伍弹窗 -->
    <div v-if="showCreateTeam" @click.self="closeCreateTeam" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div class="w-full max-w-md rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6">
        <h3 class="text-lg font-bold text-[var(--color-text-primary)]">创建队伍</h3>
        <input v-model="teamName" @keyup.enter="createTeamAction" placeholder="队伍名称" class="mt-4 w-full rounded border border-[var(--color-border-default)] bg-[var(--color-bg-default)] px-3 py-2 text-[var(--color-text-primary)]" />
        <div class="mt-4 flex justify-end gap-2">
          <button @click="closeCreateTeam" class="rounded border border-[var(--color-border-default)] px-4 py-2 text-sm text-[var(--color-text-primary)]">取消</button>
          <button @click="createTeamAction" :disabled="creatingTeam" class="rounded bg-[var(--color-primary)] px-4 py-2 text-sm text-white disabled:opacity-50">
            {{ creatingTeam ? '创建中...' : '创建' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 加入队伍弹窗 -->
    <div v-if="showJoinTeam" @click.self="closeJoinTeam" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div class="w-full max-w-md rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6">
        <h3 class="text-lg font-bold text-[var(--color-text-primary)]">加入队伍</h3>
        <input v-model="teamIdInput" @keyup.enter="joinTeamAction" placeholder="队伍 ID" class="mt-4 w-full rounded border border-[var(--color-border-default)] bg-[var(--color-bg-default)] px-3 py-2 text-[var(--color-text-primary)]" />
        <div class="mt-4 flex justify-end gap-2">
          <button @click="closeJoinTeam" class="rounded border border-[var(--color-border-default)] px-4 py-2 text-sm text-[var(--color-text-primary)]">取消</button>
          <button @click="joinTeamAction" :disabled="joiningTeam" class="rounded bg-[var(--color-primary)] px-4 py-2 text-sm text-white disabled:opacity-50">
            {{ joiningTeam ? '加入中...' : '加入' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { getContestDetail, getContestChallenges, getMyTeam, createTeam, joinTeam, kickTeamMember, submitContestFlag } from '@/api/contest'
import type { ContestDetailData, ContestChallengeItem, TeamData, SubmitFlagData } from '@/api/contracts'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'
import { formatTime, formatDuration } from '@/utils/format'
import { getStatusLabel, getModeLabel, getStatusBadgeClass } from '@/utils/contest'

const route = useRoute()
const authStore = useAuthStore()
const toast = useToast()
const contest = ref<ContestDetailData | null>(null)
const team = ref<TeamData | null>(null)
const challenges = ref<ContestChallengeItem[]>([])
const loading = ref(false)
const countdown = ref('')
const selectedChallenge = ref<ContestChallengeItem | null>(null)
const flagInput = ref('')
const submitting = ref(false)
const submitResult = ref<SubmitFlagData | null>(null)
const showCreateTeam = ref(false)
const showJoinTeam = ref(false)
const teamName = ref('')
const teamIdInput = ref('')
const creatingTeam = ref(false)
const joiningTeam = ref(false)

let timer: number | null = null

const isCaptain = computed(() => team.value && authStore.user && team.value.captain_user_id === authStore.user.id)

onMounted(async () => {
  loading.value = true
  try {
    const contestId = route.params.id as string
    contest.value = await getContestDetail(contestId)

    const [teamData, challengesData] = await Promise.all([
      getMyTeam(contestId).catch(() => null),
      getContestChallenges(contestId).catch(() => [])
    ])

    team.value = teamData
    challenges.value = challengesData

    startCountdown()
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

function startCountdown() {
  if (timer) {
    clearInterval(timer)
    timer = null
  }

  if (!contest.value) return

  timer = window.setInterval(() => {
    if (!contest.value) {
      if (timer) clearInterval(timer)
      timer = null
      return
    }

    const now = Date.now()
    const start = new Date(contest.value.starts_at).getTime()
    const end = new Date(contest.value.ends_at).getTime()

    if (now < start) {
      countdown.value = `距离开始: ${formatDuration(start - now)}`
    } else if (now < end) {
      countdown.value = `距离结束: ${formatDuration(end - now)}`
    } else {
      countdown.value = ''
      if (timer) {
        clearInterval(timer)
        timer = null
      }
    }
  }, 1000)
}

function selectChallenge(chal: ContestChallengeItem) {
  selectedChallenge.value = chal
  flagInput.value = ''
  submitResult.value = null
}

async function submitFlag() {
  const flag = flagInput.value.trim()
  if (!flag) {
    toast.warning('请输入 Flag')
    return
  }
  if (flag.length < 5 || flag.length > 200) {
    toast.warning('Flag 长度应在 5-200 字符之间')
    return
  }
  if (!selectedChallenge.value || !contest.value) return

  submitting.value = true
  submitResult.value = null

  try {
    const result = await submitContestFlag(contest.value.id, selectedChallenge.value.id, flag)
    submitResult.value = result

    if (result.correct) {
      const idx = challenges.value.findIndex(c => c.id === selectedChallenge.value!.id)
      if (idx !== -1) challenges.value[idx].is_solved = true
      flagInput.value = ''
    }
  } catch (err) {
    console.error(err)
    toast.error(err instanceof Error ? err.message : '提交失败，请稍后重试')
  } finally {
    submitting.value = false
  }
}

async function createTeamAction() {
  const name = teamName.value.trim()
  if (!name) {
    toast.warning('请输入队伍名称')
    return
  }
  if (name.length < 2 || name.length > 50) {
    toast.warning('队伍名称长度应在 2-50 字符之间')
    return
  }
  if (!contest.value || creatingTeam.value) return

  creatingTeam.value = true
  try {
    team.value = await createTeam(contest.value.id, { name })
    showCreateTeam.value = false
    teamName.value = ''
    toast.success('创建队伍成功')
  } catch (err) {
    console.error(err)
    toast.error(err instanceof Error ? err.message : '创建队伍失败')
  } finally {
    creatingTeam.value = false
  }
}

function closeCreateTeam() {
  showCreateTeam.value = false
  teamName.value = ''
}

async function joinTeamAction() {
  const teamId = teamIdInput.value.trim()
  if (!teamId) {
    toast.warning('请输入队伍 ID')
    return
  }
  if (!contest.value || joiningTeam.value) return

  joiningTeam.value = true
  try {
    await joinTeam(contest.value.id, teamId)
    team.value = await getMyTeam(contest.value.id)
    showJoinTeam.value = false
    teamIdInput.value = ''
    toast.success('加入队伍成功')
  } catch (err) {
    console.error(err)
    toast.error(err instanceof Error ? err.message : '加入队伍失败')
  } finally {
    joiningTeam.value = false
  }
}

function closeJoinTeam() {
  showJoinTeam.value = false
  teamIdInput.value = ''
}

async function kickMember(userId: string) {
  if (!contest.value || !team.value || !confirm('确定踢出该成员？')) return

  try {
    await kickTeamMember(contest.value.id, team.value.id, userId)
    team.value = await getMyTeam(contest.value.id)
    toast.success('已踢出成员')
  } catch (err) {
    console.error(err)
    toast.error(err instanceof Error ? err.message : '踢出成员失败')
  }
}
</script>
