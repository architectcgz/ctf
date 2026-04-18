<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import AppEmpty from '@/components/common/AppEmpty.vue'
import ScoreboardRealtimeBridge from '@/components/scoreboard/ScoreboardRealtimeBridge.vue'
import { useContestAWDWorkspace } from '@/composables/useContestAWDWorkspace'
import type {
  ContestAWDWorkspaceServiceData,
  ContestChallengeItem,
  ContestDetailData,
} from '@/api/contracts'
import { formatTime } from '@/utils/format'

const props = defineProps<{
  contest: ContestDetailData
  challenges: ContestChallengeItem[]
}>()

const activeChallengeKey = ref('')
const flagInputs = ref<Record<string, string>>({})
const targetKeyword = ref('')
const showOnlyReachableTargets = ref(false)

const {
  workspace,
  scoreboardRows,
  loading,
  error,
  hasTeam,
  submitResult,
  startingChallengeId,
  submittingKey,
  shouldAutoRefresh,
  lastSyncedAt,
  refreshAll,
  startService,
  submitAttack,
} = useContestAWDWorkspace({
  contestId: computed(() => props.contest.id),
  contestStatus: computed(() => props.contest.status),
})

const servicesByChallengeId = computed(() => {
  const map = new Map<string, ContestAWDWorkspaceServiceData>()
  for (const item of workspace.value?.services || []) {
    map.set(item.challenge_id, item)
  }
  return map
})

const servicesByServiceId = computed(() => {
  const map = new Map<string, ContestAWDWorkspaceServiceData>()
  for (const item of workspace.value?.services || []) {
    if (item.service_id) {
      map.set(item.service_id, item)
    }
  }
  return map
})

const currentRound = computed(() => workspace.value?.current_round)
const myTeam = computed(() => workspace.value?.my_team ?? null)
const topScore = computed(() => scoreboardRows.value[0]?.score ?? 0)
const lastSyncedLabel = computed(() =>
  lastSyncedAt.value ? formatTime(lastSyncedAt.value) : '未同步'
)
const targetFilterKeyword = computed(() => targetKeyword.value.trim().toLowerCase())

const activeChallenge = computed(
  () => props.challenges.find((item) => getChallengeRuntimeKey(item) === activeChallengeKey.value) || null
)
const activeChallengeRuntimeKey = computed(() => getChallengeRuntimeKey(activeChallenge.value))

const activeTargets = computed(() => {
  if (!workspace.value || !activeChallenge.value) {
    return []
  }
  return workspace.value.targets.map((target) => ({
    ...target,
    active_service: target.services.find(
      (service) => isTargetServiceForChallenge(service, activeChallenge.value as ContestChallengeItem)
    ),
  }))
})

const filteredTargets = computed(() =>
  activeTargets.value.filter((target) => {
    const matchesKeyword =
      targetFilterKeyword.value.length === 0 ||
      target.team_name.toLowerCase().includes(targetFilterKeyword.value)
    const matchesReachable =
      !showOnlyReachableTargets.value || Boolean(target.active_service?.access_url)
    return matchesKeyword && matchesReachable
  })
)

const defenseAlerts = computed(() => {
  const items: Array<{
    challengeId: string
    challengeTitle: string
    statusLabel: string
    tone: 'danger' | 'warning'
    issues: string[]
  }> = []

  for (const challenge of props.challenges) {
    const service = getWorkspaceService(challenge)
    if (!service) {
      continue
    }

    const issues: string[] = []
    let statusLabel = '留意'
    let tone: 'danger' | 'warning' = 'warning'

    if (service.service_status === 'compromised') {
      issues.push('本轮服务状态为失陷')
      statusLabel = '失陷'
      tone = 'danger'
    } else if (service.service_status === 'down') {
      issues.push('本轮服务状态为离线')
      statusLabel = '离线'
    }

    if ((service.attack_received ?? 0) > 0) {
      issues.push(`本轮收到 ${service.attack_received} 次攻击`)
    }

    if (issues.length === 0) {
      continue
    }

    items.push({
      challengeId: challenge.challenge_id,
      challengeTitle: challenge.title,
      statusLabel,
      tone,
      issues,
    })
  }

  return items
})

watch(
  () => props.challenges.map((item) => getChallengeRuntimeKey(item)),
  (challengeKeys) => {
    if (challengeKeys.length === 0) {
      activeChallengeKey.value = ''
      return
    }
    if (!challengeKeys.includes(activeChallengeKey.value)) {
      activeChallengeKey.value = challengeKeys[0]
    }
  },
  { immediate: true }
)

function getServiceStatusLabel(status?: string): string {
  switch (status) {
    case 'up':
      return '在线'
    case 'down':
      return '离线'
    case 'compromised':
      return '失陷'
    default:
      return '待启动'
  }
}

function getServiceStatusClass(status?: string): string {
  switch (status) {
    case 'up':
      return 'contest-chip contest-chip--success'
    case 'down':
      return 'contest-chip contest-chip--neutral'
    case 'compromised':
      return 'contest-chip contest-chip--status'
    default:
      return 'contest-chip contest-chip--neutral'
  }
}

function eventDirectionLabel(direction: 'attack_in' | 'attack_out'): string {
  return direction === 'attack_out' ? '我方攻击' : '我方被打'
}

function eventResultLabel(success: boolean): string {
  return success ? '命中' : '未命中'
}

function formatServiceRef(serviceId?: string): string {
  return `Service #${serviceId || '--'}`
}

function getChallengeRuntimeKey(challenge: ContestChallengeItem | null | undefined): string {
  if (!challenge) {
    return ''
  }
  return challenge.awd_service_id || challenge.challenge_id
}

function getWorkspaceService(challenge: ContestChallengeItem): ContestAWDWorkspaceServiceData | undefined {
  if (challenge.awd_service_id) {
    return (
      servicesByServiceId.value.get(challenge.awd_service_id) ||
      servicesByChallengeId.value.get(challenge.challenge_id)
    )
  }
  return servicesByChallengeId.value.get(challenge.challenge_id)
}

function isTargetServiceForChallenge(
  service: { service_id?: string; challenge_id: string },
  challenge: ContestChallengeItem
): boolean {
  if (challenge.awd_service_id) {
    return (
      service.service_id === challenge.awd_service_id ||
      service.challenge_id === challenge.challenge_id
    )
  }
  return service.challenge_id === challenge.challenge_id
}

function buildAttackStateKey(serviceKey: string, teamId: string): string {
  return `${serviceKey}:${teamId}`
}

function buildAttackSubmissionKey(challengeId: string, teamId: string): string {
  return `${challengeId}:${teamId}`
}

function getDefenseAlertClass(tone: 'danger' | 'warning'): string {
  return tone === 'danger'
    ? 'awd-defense-row awd-defense-row--danger'
    : 'awd-defense-row awd-defense-row--warning'
}

async function handleSubmit(challengeId: string, serviceKey: string, teamId: string): Promise<void> {
  const stateKey = buildAttackStateKey(serviceKey, teamId)
  const flag = flagInputs.value[stateKey] || ''
  const result = await submitAttack(challengeId, Number(teamId), flag)
  if (result) {
    flagInputs.value[stateKey] = ''
  }
}
</script>

<template>
  <section class="awd-workspace-shell">
    <ScoreboardRealtimeBridge
      v-if="contest.status === 'running' || contest.status === 'frozen'"
      :contest-id="contest.id"
      @updated="refreshAll"
    />

    <section class="awd-workspace-summary metric-panel-grid">
      <article class="metric-panel-card">
        <div class="metric-panel-label">当前轮次</div>
        <div class="metric-panel-value">
          {{ currentRound ? `Round ${currentRound.round_number}` : '等待开赛' }}
        </div>
        <div class="metric-panel-helper">
          {{
            currentRound
              ? `${currentRound.attack_score}/${currentRound.defense_score} 分值`
              : '轮次尚未开始'
          }}
        </div>
      </article>
      <article class="metric-panel-card">
        <div class="metric-panel-label">我的队伍</div>
        <div class="metric-panel-value">{{ myTeam?.team_name || '未加入' }}</div>
        <div class="metric-panel-helper">{{ hasTeam ? '已进入战场视角' : '请先加入队伍' }}</div>
      </article>
      <article class="metric-panel-card">
        <div class="metric-panel-label">服务目录</div>
        <div class="metric-panel-value">{{ workspace?.services.length || 0 }}</div>
        <div class="metric-panel-helper">当前已同步到工作台的本队服务数</div>
      </article>
      <article class="metric-panel-card">
        <div class="metric-panel-label">榜首分数</div>
        <div class="metric-panel-value">{{ topScore }}</div>
        <div class="metric-panel-helper">实时排行榜 Top 1 官方总分</div>
      </article>
    </section>

    <div v-if="loading && !workspace" class="contest-loading">
      <div class="contest-loading__spinner" />
      <div class="contest-loading__text">正在同步 AWD 战场...</div>
    </div>

    <div v-else-if="error" class="contest-alert contest-alert--warning">
      {{ error }}
    </div>

    <AppEmpty
      v-else-if="!hasTeam"
      icon="Users"
      title="先加入队伍"
      description="先在队伍页创建或加入队伍，再进入 AWD 战场获取目标和提交 stolen flag。"
    />

    <div v-else class="awd-workspace-layout">
      <div class="awd-workspace-main">
        <section class="contest-section contest-section--flat">
          <div class="contest-section__head workspace-tab-heading">
            <div class="workspace-tab-heading__main">
              <div class="workspace-overline">Defense</div>
              <h2 class="contest-section__title workspace-tab-heading__title">防守告警</h2>
            </div>
            <div class="contest-section__hint">
              {{ defenseAlerts.length > 0 ? `${defenseAlerts.length} 项待处理` : '当前轮次无异常' }}
            </div>
          </div>

          <div v-if="defenseAlerts.length === 0" class="contest-inline-note">
            当前轮次未发现本队服务异常，战场会继续保持同步。
          </div>

          <div v-else class="awd-defense-list">
            <article
              v-for="alert in defenseAlerts"
              :key="alert.challengeId"
              :class="getDefenseAlertClass(alert.tone)"
            >
              <div class="awd-defense-row__head">
                <div class="awd-defense-row__title">
                  <h3>{{ alert.challengeTitle }}</h3>
                  <span
                    class="contest-chip"
                    :class="
                      alert.tone === 'danger' ? 'contest-chip--status' : 'contest-chip--neutral'
                    "
                  >
                    {{ alert.statusLabel }}
                  </span>
                </div>
              </div>
              <div class="awd-defense-row__issues">
                <span v-for="issue in alert.issues" :key="issue">{{ issue }}</span>
              </div>
            </article>
          </div>
        </section>

        <section class="contest-section contest-section--flat">
          <div class="contest-section__head workspace-tab-heading">
            <div class="workspace-tab-heading__main">
              <div class="workspace-overline">My Services</div>
              <h2 class="contest-section__title workspace-tab-heading__title">我的服务</h2>
            </div>
            <div class="contest-section__hint">{{ challenges.length }} 题</div>
          </div>

          <div v-if="challenges.length === 0" class="contest-inline-note">
            当前赛事还没有发布可用服务。
          </div>

          <div v-else class="awd-service-list">
            <article v-for="challenge in challenges" :key="challenge.id" class="awd-service-row">
              <div class="awd-service-row__main">
                <div class="awd-service-row__head">
                  <h3>{{ challenge.title }}</h3>
                  <span
                    :class="getServiceStatusClass(getWorkspaceService(challenge)?.service_status)"
                  >
                    {{
                      getServiceStatusLabel(getWorkspaceService(challenge)?.service_status)
                    }}
                  </span>
                </div>
                <div class="awd-service-row__meta">
                  <span>{{ challenge.category }}</span>
                  <span>{{ challenge.points }} pts</span>
                  <span v-if="getWorkspaceService(challenge)?.service_id">
                    {{ formatServiceRef(getWorkspaceService(challenge)?.service_id) }}
                  </span>
                  <span v-if="getWorkspaceService(challenge)?.access_url">
                    {{ getWorkspaceService(challenge)?.access_url }}
                  </span>
                  <span v-else>尚未生成访问地址</span>
                </div>
              </div>

              <div class="awd-service-row__actions" role="group" aria-label="服务操作">
                <a
                  v-if="getWorkspaceService(challenge)?.access_url"
                  class="ui-btn ui-btn--ghost"
                  :href="getWorkspaceService(challenge)?.access_url"
                  target="_blank"
                  rel="noreferrer"
                >
                  打开服务
                </a>
                <button
                  type="button"
                  class="ui-btn ui-btn--primary"
                  :disabled="startingChallengeId === challenge.challenge_id"
                  @click="startService(challenge.challenge_id)"
                >
                  {{
                    startingChallengeId === challenge.challenge_id
                      ? '启动中...'
                      : getWorkspaceService(challenge)?.access_url
                        ? '刷新服务'
                        : '启动服务'
                  }}
                </button>
              </div>
            </article>
          </div>
        </section>

        <section class="contest-section contest-section--flat">
          <div class="contest-section__head workspace-tab-heading">
            <div class="workspace-tab-heading__main">
              <div class="workspace-overline">Targets</div>
              <h2 class="contest-section__title workspace-tab-heading__title">目标目录</h2>
            </div>
            <div class="contest-section__hint">
              {{ filteredTargets.length }}/{{ activeTargets.length }} 支队伍
            </div>
          </div>

          <div class="awd-target-toolbar">
            <div class="awd-target-toolbar__field">
              <label class="ui-field__label" for="awd-target-challenge">攻击题目</label>
              <div class="ui-control-wrap awd-target-control">
                <select id="awd-target-challenge" v-model="activeChallengeKey" class="ui-control">
                  <option v-if="challenges.length === 0" value="" disabled>
                    当前没有可选攻击题目
                  </option>
                  <option
                    v-for="challenge in challenges"
                    :key="getChallengeRuntimeKey(challenge)"
                    :value="getChallengeRuntimeKey(challenge)"
                  >
                    {{ challenge.title }}
                  </option>
                </select>
              </div>
            </div>

            <div class="awd-target-toolbar__field awd-target-toolbar__field--wide">
              <label class="ui-field__label" for="awd-target-search">筛选队伍</label>
              <div class="ui-control-wrap awd-target-control awd-target-search">
                <input
                  id="awd-target-search"
                  v-model="targetKeyword"
                  type="text"
                  class="ui-control"
                  placeholder="输入队伍名"
                />
              </div>
            </div>

            <label class="awd-target-toggle" for="awd-target-reachable-only">
              <input
                id="awd-target-reachable-only"
                v-model="showOnlyReachableTargets"
                type="checkbox"
              />
              <span>仅看可用地址</span>
            </label>
          </div>

          <div v-if="!activeChallenge" class="contest-inline-note">当前没有可选攻击题目。</div>

          <div v-else-if="filteredTargets.length === 0" class="contest-inline-note">
            没有匹配的目标队伍。
          </div>

          <div v-else class="awd-target-list">
            <article v-for="target in filteredTargets" :key="target.team_id" class="awd-target-row">
                <div class="awd-target-row__main">
                  <div class="awd-target-row__head">
                    <div class="contest-chip contest-chip--neutral">{{ target.team_name }}</div>
                    <div
                      v-if="target.active_service?.service_id"
                      class="awd-runtime-ref-chip"
                    >
                      {{ formatServiceRef(target.active_service.service_id) }}
                    </div>
                    <div class="awd-target-row__url">
                      {{ target.active_service?.access_url || '当前没有可用目标地址' }}
                    </div>
                </div>
              </div>

              <div class="awd-target-row__form">
                <div class="ui-control-wrap flag-submit__control">
                  <input
                    :value="flagInputs[buildAttackStateKey(activeChallengeRuntimeKey, target.team_id)] || ''"
                    type="text"
                    class="ui-control"
                    placeholder="flag{...}"
                    @input="
                      flagInputs[buildAttackStateKey(activeChallengeRuntimeKey, target.team_id)] =
                        String(($event.target as HTMLInputElement).value)
                    "
                  />
                </div>
                <button
                  type="button"
                  class="ui-btn ui-btn--primary"
                  :disabled="
                    !target.active_service?.access_url ||
                    submittingKey ===
                      buildAttackSubmissionKey(activeChallenge.challenge_id, target.team_id)
                  "
                  @click="
                    handleSubmit(
                      activeChallenge.challenge_id,
                      activeChallengeRuntimeKey,
                      target.team_id
                    )
                  "
                >
                  {{
                    submittingKey ===
                    buildAttackSubmissionKey(activeChallenge.challenge_id, target.team_id)
                      ? '提交中...'
                      : '提交 stolen flag'
                  }}
                </button>
              </div>
            </article>
          </div>

          <div
            v-if="submitResult"
            class="contest-alert"
            :class="submitResult.is_success ? 'contest-alert--success' : 'contest-alert--danger'"
          >
            {{
              submitResult.is_success
                ? `攻击成功，+${submitResult.score_gained} 分`
                : '攻击未命中有效 flag'
            }}
          </div>
        </section>
      </div>

      <aside class="awd-workspace-rail">
        <section class="contest-section contest-section--flat">
          <div class="contest-section__head workspace-tab-heading">
            <div class="workspace-tab-heading__main">
              <div class="workspace-overline">Status</div>
              <h2 class="contest-section__title workspace-tab-heading__title">战场状态</h2>
            </div>
            <button
              type="button"
              class="ui-btn ui-btn--ghost ui-btn--sm"
              :disabled="loading"
              @click="refreshAll"
            >
              {{ loading ? '刷新中...' : '立即刷新' }}
            </button>
          </div>

          <div class="awd-rail-list">
            <div class="awd-rail-row">
              <span>当前轮次</span>
              <strong>{{
                currentRound ? `Round ${currentRound.round_number}` : '等待开赛'
              }}</strong>
            </div>
            <div class="awd-rail-row">
              <span>轮次状态</span>
              <strong>{{ currentRound?.status || 'pending' }}</strong>
            </div>
            <div class="awd-rail-row">
              <span>同步策略</span>
              <strong>{{ shouldAutoRefresh ? '每 15 秒自动刷新' : '手动刷新' }}</strong>
            </div>
            <div class="awd-rail-row">
              <span>最近同步</span>
              <strong>{{ lastSyncedLabel }}</strong>
            </div>
            <div class="awd-rail-row">
              <span>轮次更新时间</span>
              <strong>{{ currentRound ? formatTime(currentRound.updated_at) : '未记录' }}</strong>
            </div>
          </div>
        </section>

        <section class="contest-section contest-section--flat">
          <div class="contest-section__head workspace-tab-heading">
            <div class="workspace-tab-heading__main">
              <div class="workspace-overline">Scoreboard</div>
              <h2 class="contest-section__title workspace-tab-heading__title">实时榜单</h2>
            </div>
          </div>

          <div v-if="scoreboardRows.length === 0" class="contest-inline-note">
            当前还没有榜单数据。
          </div>

          <div v-else class="awd-scoreboard-list">
            <div v-for="item in scoreboardRows" :key="item.team_id" class="awd-scoreboard-row">
              <span class="awd-scoreboard-rank">{{ item.rank }}</span>
              <span class="awd-scoreboard-team">{{ item.team_name }}</span>
              <strong class="awd-scoreboard-score">{{ item.score }}</strong>
            </div>
          </div>
        </section>

        <section class="contest-section contest-section--flat">
          <div class="contest-section__head workspace-tab-heading">
            <div class="workspace-tab-heading__main">
              <div class="workspace-overline">Feedback</div>
              <h2 class="contest-section__title workspace-tab-heading__title">最近反馈</h2>
            </div>
          </div>

          <div v-if="workspace?.recent_events.length === 0" class="contest-inline-note">
            当前轮次还没有与你队伍相关的攻防反馈。
          </div>

          <div v-else class="awd-event-list">
            <article
              v-for="event in workspace?.recent_events"
              :key="event.id"
              class="awd-event-row"
            >
              <div class="awd-event-row__head">
                <span class="contest-chip contest-chip--neutral">{{
                  eventDirectionLabel(event.direction)
                }}</span>
                <span>{{ event.peer_team_name }}</span>
              </div>
              <div class="awd-event-row__meta">
                <span>{{ eventResultLabel(event.is_success) }}</span>
                <span>{{ event.score_gained }} 分</span>
                <span v-if="event.service_id">{{ formatServiceRef(event.service_id) }}</span>
                <span>{{ formatTime(event.created_at) }}</span>
              </div>
            </article>
          </div>
        </section>
      </aside>
    </div>
  </section>
</template>

<style scoped>
.awd-workspace-shell {
  display: grid;
  gap: 1rem;
}

.contest-section {
  border-radius: 24px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 82%, transparent);
  background: color-mix(in srgb, var(--journal-surface-subtle) 92%, transparent);
  padding: 1rem;
}

.contest-section__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 0.95rem;
}

.contest-section__title {
  margin: 0.35rem 0 0;
  color: var(--journal-ink);
  font-size: var(--font-size-1-12);
}

.contest-section__hint {
  color: var(--journal-muted);
  font-size: var(--font-size-0-82);
}

.contest-inline-note {
  border-radius: 16px;
  border: 1px dashed color-mix(in srgb, var(--journal-border) 78%, transparent);
  padding: 0.9rem 1rem;
  color: var(--journal-muted);
  background: color-mix(in srgb, var(--journal-surface) 88%, transparent);
}

.contest-alert {
  margin-top: 0.95rem;
  border-radius: 16px;
  padding: 0.8rem 0.95rem;
  font-size: var(--font-size-0-84);
}

.contest-alert--success {
  background: color-mix(in srgb, #34d399 12%, transparent);
  color: color-mix(in srgb, #34d399 88%, var(--journal-ink));
}

.contest-alert--danger {
  background: color-mix(in srgb, #f87171 12%, transparent);
  color: color-mix(in srgb, #f87171 88%, var(--journal-ink));
}

.contest-alert--warning {
  background: color-mix(in srgb, #fbbf24 12%, transparent);
  color: color-mix(in srgb, #fbbf24 88%, var(--journal-ink));
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

.contest-chip--status {
  color: color-mix(in srgb, #f59e0b 88%, var(--journal-ink));
  border-color: color-mix(in srgb, #f59e0b 28%, transparent);
  background: color-mix(in srgb, #f59e0b 11%, transparent);
}

.contest-chip--success {
  color: color-mix(in srgb, #34d399 88%, var(--journal-ink));
  border-color: color-mix(in srgb, #34d399 26%, transparent);
  background: color-mix(in srgb, #34d399 10%, transparent);
}

.contest-chip--neutral {
  background: color-mix(in srgb, var(--journal-surface) 86%, transparent);
}

.awd-runtime-ref-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 28%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  padding: 0.32rem 0.72rem;
  color: var(--journal-accent-strong);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-0-74);
  font-weight: 700;
  white-space: nowrap;
}

.flag-submit__control {
  flex: 1 1 12rem;
  min-width: 0;
}

.awd-workspace-summary {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.awd-workspace-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.45fr) minmax(18rem, 0.8fr);
  gap: 1rem;
}

.awd-workspace-main,
.awd-workspace-rail {
  display: grid;
  gap: 1rem;
}

.awd-service-list,
.awd-defense-list,
.awd-target-list,
.awd-event-list,
.awd-scoreboard-list {
  display: grid;
  gap: 0.85rem;
}

.awd-defense-row,
.awd-service-row,
.awd-target-row,
.awd-event-row,
.awd-scoreboard-row {
  display: grid;
  gap: 0.75rem;
  border: 1px solid color-mix(in srgb, var(--journal-border) 84%, transparent);
  border-radius: 18px;
  background: color-mix(in srgb, var(--journal-surface-subtle) 92%, transparent);
  padding: 1rem;
}

.awd-defense-row {
  gap: 0.85rem;
}

.awd-defense-row--danger {
  border-color: color-mix(in srgb, #f87171 28%, transparent);
  background: color-mix(in srgb, #f87171 10%, var(--journal-surface-subtle));
}

.awd-defense-row--warning {
  border-color: color-mix(in srgb, #f59e0b 24%, transparent);
  background: color-mix(in srgb, #f59e0b 8%, var(--journal-surface-subtle));
}

.awd-service-row {
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
}

.awd-defense-row__head,
.awd-service-row__head,
.awd-target-row__head,
.awd-event-row__head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.65rem;
}

.awd-defense-row__title {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.65rem;
}

.awd-defense-row__title h3,
.awd-service-row__head h3 {
  margin: 0;
  font-size: 1rem;
  color: var(--journal-ink);
}

.awd-defense-row__issues,
.awd-service-row__meta,
.awd-event-row__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  color: var(--journal-muted);
  font-size: var(--font-size-0-82);
}

.awd-service-row__actions,
.awd-target-row__form {
  display: flex;
  flex-wrap: wrap;
  gap: 0.65rem;
  align-items: center;
}

.awd-target-toolbar {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
  margin-bottom: 0.95rem;
}

.awd-target-toolbar__field {
  display: grid;
  gap: 0.45rem;
}

.awd-target-toolbar__field--wide {
  grid-column: span 2;
}

.awd-target-toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.55rem;
  color: var(--journal-muted);
  font-size: var(--font-size-0-80);
  font-weight: 700;
}

.awd-target-control {
  width: 100%;
}

.awd-target-row__url {
  font-family: var(--font-family-mono, monospace);
  font-size: var(--font-size-0-80);
  color: var(--journal-muted);
  word-break: break-all;
}

.awd-rail-list {
  display: grid;
  gap: 0.75rem;
}

.awd-rail-row,
.awd-scoreboard-row {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 0.75rem;
}

.awd-scoreboard-rank {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.9rem;
  height: 1.9rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 86%, transparent);
  color: var(--journal-ink);
  font-weight: 700;
}

.awd-scoreboard-team {
  min-width: 0;
  color: var(--journal-ink);
}

.awd-scoreboard-score {
  font-family: var(--font-family-mono, monospace);
  color: var(--journal-ink);
}

@media (max-width: 1080px) {
  .awd-workspace-summary {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .awd-workspace-layout {
    grid-template-columns: minmax(0, 1fr);
  }
}

@media (max-width: 720px) {
  .awd-workspace-summary {
    grid-template-columns: minmax(0, 1fr);
  }

  .awd-target-toolbar {
    grid-template-columns: minmax(0, 1fr);
  }

  .awd-target-toolbar__field--wide {
    grid-column: auto;
  }

  .awd-service-row {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
