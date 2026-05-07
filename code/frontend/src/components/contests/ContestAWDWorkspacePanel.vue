<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  ShieldAlert,
  Sword,
  Target,
  Activity,
  Wifi,
  Zap,
  BarChart3,
  History,
  Terminal,
  ExternalLink,
  RefreshCw,
} from 'lucide-vue-next'

import AppEmpty from '@/components/common/AppEmpty.vue'
import AWDDefenseOperationsPanel from '@/components/contests/awd/AWDDefenseOperationsPanel.vue'
import AWDDefenseServiceList from '@/components/contests/awd/AWDDefenseServiceList.vue'
import ScoreboardRealtimeBridge from '@/components/scoreboard/ScoreboardRealtimeBridge.vue'
import {
  getVSCodeSSHCommand,
  toDefenseServiceCards,
  useAwdDefenseServiceSelection,
  useContestAWDWorkspace,
} from '@/features/contest-awd-workspace'
import type {
  ContestAWDWorkspaceServiceData,
  ContestChallengeItem,
  ContestDetailData,
} from '@/api/contracts'
import { useToast } from '@/composables/useToast'
import { formatTime } from '@/utils/format'

const props = defineProps<{
  contest: ContestDetailData
  challenges: ContestChallengeItem[]
}>()

const toast = useToast()
const activeChallengeKey = ref('')
const flagInputs = ref<Record<string, string>>({})
const targetKeyword = ref('')
const showOnlyReachableTargets = ref(false)
const copiedSSHCommandKey = ref('')
const copiedSSHPasswordKey = ref('')

const {
  workspace,
  scoreboardRows,
  loading,
  error,
  hasTeam,
  submitResult,
  startingServiceKey,
  serviceActionPendingById,
  openingServiceKey,
  openingSSHKey,
  sshAccessByServiceId,
  openingTargetKey,
  submittingKey,
  shouldAutoRefresh,
  lastSyncedAt,
  refreshAll,
  restartService,
  openService,
  openDefenseSSH,
  openTarget,
  submitAttack,
} = useContestAWDWorkspace({
  contestId: computed(() => props.contest.id),
  contestStatus: computed(() => props.contest.status),
  formatAttackResultToast,
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

const runtimeChallenges = computed(() =>
  props.challenges.filter((item): item is ContestChallengeItem & { awd_service_id: string } =>
    Boolean(item.awd_service_id)
  )
)

const defenseServiceCards = computed(() =>
  toDefenseServiceCards({
    challenges: props.challenges,
    services: workspace.value?.services || [],
  })
)

const defenseServiceActionPendingById = computed(() => {
  const pendingById: Record<string, boolean> = {}
  for (const card of defenseServiceCards.value) {
    const service = servicesByServiceId.value.get(card.serviceId)
    pendingById[card.serviceId] = Boolean(
      startingServiceKey.value === card.serviceId ||
      serviceActionPendingById.value[card.serviceId] ||
      service?.instance_status === 'pending' ||
      service?.instance_status === 'creating' ||
      service?.operation_status === 'requested' ||
      service?.operation_status === 'provisioning' ||
      service?.operation_status === 'recovering'
    )
  }
  return pendingById
})

const { selectedServiceId, selectService } = useAwdDefenseServiceSelection(defenseServiceCards)

const challengeByChallengeId = computed(() => {
  const map = new Map<string, ContestChallengeItem>()
  for (const item of props.challenges) {
    map.set(getAWDChallengeId(item), item)
  }
  return map
})

const challengeByServiceId = computed(() => {
  const map = new Map<string, ContestChallengeItem>()
  for (const item of props.challenges) {
    if (item.awd_service_id) {
      map.set(item.awd_service_id, item)
    }
  }
  return map
})

const currentRound = computed(() => workspace.value?.current_round)
const myTeam = computed(() => workspace.value?.my_team ?? null)
const selectedDefenseServiceCard = computed(
  () => defenseServiceCards.value.find((card) => card.serviceId === selectedServiceId.value) || null
)
const selectedDefenseServiceTitle = computed(() => selectedDefenseServiceCard.value?.title || '')
const topScore = computed(() => scoreboardRows.value[0]?.score ?? 0)
const lastSyncedLabel = computed(() =>
  lastSyncedAt.value ? formatTime(lastSyncedAt.value) : '未同步'
)
const targetFilterKeyword = computed(() => targetKeyword.value.trim().toLowerCase())

const activeChallenge = computed(
  () =>
    runtimeChallenges.value.find(
      (item) => getChallengeRuntimeKey(item) === activeChallengeKey.value
    ) || null
)
const activeChallengeRuntimeKey = computed(() => getChallengeRuntimeKey(activeChallenge.value))

const activeTargets = computed(() => {
  if (!workspace.value || !activeChallenge.value) {
    return []
  }
  return workspace.value.targets.map((target) => ({
    ...target,
    active_service: target.services.find((service) =>
      isTargetServiceForChallenge(service, activeChallenge.value as ContestChallengeItem)
    ),
  }))
})

const filteredTargets = computed(() =>
  activeTargets.value.filter((target) => {
    const matchesKeyword =
      targetFilterKeyword.value.length === 0 ||
      target.team_name.toLowerCase().includes(targetFilterKeyword.value)
    const matchesReachable =
      !showOnlyReachableTargets.value || Boolean(target.active_service?.reachable)
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

  for (const challenge of runtimeChallenges.value) {
    const service = getWorkspaceService(challenge)
    if (!service) continue

    const issues: string[] = []
    let statusLabel = '正常'
    let tone: 'danger' | 'warning' = 'warning'

    if (service.service_status === 'compromised') {
      issues.push('已失陷')
      statusLabel = '严重'
      tone = 'danger'
    } else if (service.service_status === 'down' && service.instance_status !== 'running') {
      issues.push('已离线')
      statusLabel = '告警'
    }

    if ((service.attack_received ?? 0) > 0) {
      issues.push(`检测到 ${service.attack_received} 次攻击`)
    }

    if (issues.length === 0) continue

    items.push({
      challengeId: getAWDChallengeId(challenge),
      challengeTitle: challenge.title,
      statusLabel,
      tone,
      issues,
    })
  }
  return items
})

watch(
  () => runtimeChallenges.value.map((item) => getChallengeRuntimeKey(item)),
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

function eventDirectionLabel(direction: 'attack_in' | 'attack_out'): string {
  return direction === 'attack_out' ? '对外攻击' : '受到攻击'
}

function eventResultLabel(success: boolean): string {
  return success ? '成功' : '失败'
}

function formatServiceRef(serviceId?: string): string {
  return `服务 #${serviceId || '--'}`
}

function formatRoundStatusLabel(status?: string): string {
  switch (status) {
    case 'running':
      return '进行中'
    case 'frozen':
      return '已冻结'
    case 'finished':
    case 'completed':
    case 'ended':
      return '已结束'
    default:
      return '等待中'
  }
}

function getChallengeRuntimeKey(challenge: ContestChallengeItem | null | undefined): string {
  return challenge?.awd_service_id || ''
}

function getAWDChallengeId(challenge: ContestChallengeItem): string {
  return challenge.awd_challenge_id || challenge.challenge_id
}

function getChallengeTitleForEvent(event: {
  service_id?: string
  awd_challenge_id: string
}): string {
  if (event.service_id) {
    const matchedByService = challengeByServiceId.value.get(event.service_id)
    if (matchedByService) return matchedByService.title
  }
  return challengeByChallengeId.value.get(event.awd_challenge_id)?.title || event.awd_challenge_id
}

function getSubmitResultMessage(): string {
  if (!submitResult.value) return ''
  return formatAttackResultToast(submitResult.value)
}

function formatAttackResultToast(result: {
  service_id?: string
  awd_challenge_id: string
  is_success: boolean
  score_gained: number
}): string {
  const challengeTitle = getChallengeTitleForEvent(result)
  if (result.is_success) return `${challengeTitle}: 攻击成功，+${result.score_gained} 分`
  return `${challengeTitle}: 未获取到有效 Flag。`
}

function getWorkspaceService(
  challenge: ContestChallengeItem
): ContestAWDWorkspaceServiceData | undefined {
  if (!challenge.awd_service_id) return undefined
  return servicesByServiceId.value.get(challenge.awd_service_id)
}

function getSSHAccess(serviceId?: string) {
  if (!serviceId) return undefined
  return sshAccessByServiceId.value[serviceId]
}

function getSSHCommand(serviceId?: string): string {
  return getVSCodeSSHCommand(getSSHAccess(serviceId))
}

function openDefenseService(serviceId: string): void {
  const instanceId = servicesByServiceId.value.get(serviceId)?.instance_id
  if (!instanceId) return
  void openService(instanceId)
}

async function copyTextToClipboard(text: string, successMessage: string): Promise<boolean> {
  if (!text || typeof navigator === 'undefined' || !navigator.clipboard) {
    toast.error('复制失败，请手动选择文本')
    return false
  }

  try {
    await navigator.clipboard.writeText(text)
    toast.success(successMessage)
    return true
  } catch (err) {
    console.error(err)
    toast.error('复制失败，请手动选择文本')
    return false
  }
}

async function copySSHCommand(serviceId?: string): Promise<void> {
  if (!serviceId) return
  const copied = await copyTextToClipboard(getSSHCommand(serviceId), 'SSH 命令已复制')
  if (copied) {
    copiedSSHCommandKey.value = serviceId
  }
}

async function copySSHPassword(serviceId?: string): Promise<void> {
  if (!serviceId) return
  const password = getSSHAccess(serviceId)?.password || ''
  const copied = await copyTextToClipboard(password, 'SSH 密码已复制')
  if (copied) {
    copiedSSHPasswordKey.value = serviceId
  }
}

function isTargetServiceForChallenge(
  service: { service_id?: string; awd_challenge_id: string },
  challenge: ContestChallengeItem
): boolean {
  return Boolean(challenge.awd_service_id) && service.service_id === challenge.awd_service_id
}

function buildAttackStateKey(serviceKey: string, teamId: string): string {
  return `${serviceKey}:${teamId}`
}

async function handleSubmit(serviceKey: string, teamId: string): Promise<void> {
  const stateKey = buildAttackStateKey(serviceKey, teamId)
  const flag = flagInputs.value[stateKey] || ''
  const result = await submitAttack(serviceKey, Number(teamId), flag)
  if (result) {
    flagInputs.value[stateKey] = ''
  }
}
</script>

<template>
  <div class="awd-war-room">
    <ScoreboardRealtimeBridge
      v-if="contest.status === 'running' || contest.status === 'frozen'"
      :contest-id="contest.id"
      @updated="refreshAll"
    />

    <!-- HUD KPI Strip -->
    <header class="awd-hud-strip">
      <div class="hud-item">
        <div class="hud-label">当前回合</div>
        <div class="hud-value font-mono">
          {{ currentRound ? `#${String(currentRound.round_number).padStart(2, '0')}` : '--' }}
        </div>
        <div class="hud-helper">
          {{ formatRoundStatusLabel(currentRound?.status) }}
        </div>
      </div>
      <div class="hud-item">
        <div class="hud-label">我的战队</div>
        <div class="hud-value">
          {{ myTeam?.team_name || '未加入' }}
        </div>
        <div class="hud-helper">
          排名 #{{ scoreboardRows.find((r) => r.team_id === myTeam?.team_id)?.rank || '--' }}
        </div>
      </div>
      <div class="hud-item">
        <div class="hud-label">战队服务</div>
        <div class="hud-value font-mono">
          {{ workspace?.services.length || 0 }}
        </div>
        <div class="hud-helper">运行中服务</div>
      </div>
      <div class="hud-item">
        <div class="hud-label">最高分</div>
        <div class="hud-value hud-value--accent font-mono">
          {{ topScore }}
        </div>
        <div class="hud-helper">当前榜首</div>
      </div>
      <div class="hud-actions">
        <button class="hud-refresh-btn" :disabled="loading" @click="refreshAll">
          <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': loading }" />
          <span>{{ lastSyncedLabel }}</span>
        </button>
      </div>
    </header>

    <div v-if="loading && !workspace" class="war-room-loading">
      <div class="radar-scan" />
      <p>正在建立战场连接...</p>
    </div>

    <AppEmpty
      v-else-if="!hasTeam"
      icon="Users"
      title="先加入队伍"
      description="需要先加入队伍后才能进入 AWD 战场。"
      class="war-room-empty"
    />

    <div v-else class="war-room-grid">
      <!-- 1. Defense Monitor (Left) -->
      <aside class="war-room-col column-defense">
        <section class="ops-panel">
          <header class="ops-panel__header">
            <ShieldAlert class="ops-panel__icon ops-panel__icon--warning h-4 w-4" />
            <h3 class="ops-panel__title">我的防守</h3>
          </header>

          <div class="ops-panel__content custom-scrollbar">
            <!-- Alerts -->
            <div v-if="defenseAlerts.length > 0" class="defense-alerts">
              <div
                v-for="alert in defenseAlerts"
                :key="alert.challengeId"
                class="defense-alert"
                :class="alert.tone"
              >
                <div class="flex items-center justify-between">
                  <span class="alert-title">{{ alert.challengeTitle }}</span>
                  <span class="alert-badge">{{ alert.statusLabel }}</span>
                </div>
                <div class="alert-issues">
                  <span v-for="issue in alert.issues" :key="issue">{{ issue }}</span>
                </div>
              </div>
            </div>

            <AWDDefenseServiceList
              :services="defenseServiceCards"
              :selected-service-id="selectedServiceId"
              :opening-service-key="openingServiceKey"
              :opening-ssh-key="openingSSHKey"
              :service-action-pending-by-id="defenseServiceActionPendingById"
              @select-service="selectService"
              @open-service="openDefenseService"
              @request-ssh="openDefenseSSH"
              @restart-service="restartService"
            />
            <AWDDefenseOperationsPanel
              :service-card="selectedDefenseServiceCard"
              :service-title="selectedDefenseServiceTitle"
              :opening-service-key="openingServiceKey"
              :opening-ssh-key="openingSSHKey"
              :action-pending="
                selectedServiceId
                  ? Boolean(defenseServiceActionPendingById[selectedServiceId])
                  : false
              "
              :loading="loading"
              :access="getSSHAccess(selectedServiceId)"
              :copied-command="copiedSSHCommandKey === selectedServiceId"
              :copied-password="copiedSSHPasswordKey === selectedServiceId"
              @open-service="openDefenseService"
              @request-ssh="openDefenseSSH"
              @restart-service="restartService"
              @refresh="refreshAll"
              @copy-command="copySSHCommand"
              @copy-password="copySSHPassword"
            />
          </div>
        </section>
      </aside>

      <!-- 2. Attack Vector (Middle) -->
      <main class="war-room-col column-attack">
        <section class="ops-panel">
          <header class="ops-panel__header">
            <Sword class="ops-panel__icon ops-panel__icon--danger h-4 w-4" />
            <h3 class="ops-panel__title">攻击向量</h3>
          </header>

          <div class="ops-panel__toolbar">
            <div class="toolbar-field">
              <label>目标题目</label>
              <select
                id="awd-target-challenge"
                v-model="activeChallengeKey"
                class="war-room-select"
              >
                <option
                  v-for="challenge in runtimeChallenges"
                  :key="getChallengeRuntimeKey(challenge)"
                  :value="getChallengeRuntimeKey(challenge)"
                >
                  {{ challenge.title }}
                </option>
              </select>
            </div>
            <div class="toolbar-field">
              <label>队伍筛选</label>
              <input
                id="awd-target-search"
                v-model="targetKeyword"
                type="text"
                placeholder="按队伍名称筛选..."
                class="war-room-input"
              />
            </div>
          </div>

          <div class="ops-panel__content custom-scrollbar">
            <div v-if="runtimeChallenges.length === 0" class="panel-note">
              当前竞赛暂无可部署服务。
            </div>
            <div v-else-if="!activeChallenge" class="panel-note">请选择目标题目后开始攻击。</div>
            <div v-else-if="filteredTargets.length === 0" class="panel-note">
              当前题目下没有匹配的目标队伍。
            </div>
            <div v-else class="target-grid">
              <article v-for="target in filteredTargets" :key="target.team_id" class="target-card">
                <div class="target-info">
                  <div class="target-team font-black">
                    {{ target.team_name.toUpperCase() }}
                  </div>
                  <div class="target-ref">
                    {{ formatServiceRef(target.active_service?.service_id) }}
                  </div>
                  <div class="target-url font-mono">
                    {{ target.active_service?.reachable ? '代理链路已就绪' : '不可达' }}
                  </div>
                </div>
                <div class="target-action">
                  <button
                    :data-testid="`awd-open-target-${activeChallengeRuntimeKey}-${target.team_id}`"
                    :disabled="
                      !target.active_service?.reachable ||
                      openingTargetKey ===
                        buildAttackStateKey(activeChallengeRuntimeKey, target.team_id)
                    "
                    class="target-open-btn"
                    type="button"
                    @click="openTarget(activeChallengeRuntimeKey, target.team_id)"
                  >
                    <ExternalLink class="h-3 w-3" />
                    <span>{{
                      openingTargetKey ===
                      buildAttackStateKey(activeChallengeRuntimeKey, target.team_id)
                        ? '...'
                        : '打开'
                    }}</span>
                  </button>
                  <input
                    :value="
                      flagInputs[buildAttackStateKey(activeChallengeRuntimeKey, target.team_id)] ||
                      ''
                    "
                    placeholder="输入获取到的 Flag..."
                    class="flag-input"
                    @input="
                      flagInputs[buildAttackStateKey(activeChallengeRuntimeKey, target.team_id)] =
                        String(($event.target as HTMLInputElement).value)
                    "
                    @keyup.enter="handleSubmit(activeChallengeRuntimeKey, target.team_id)"
                  />
                  <button
                    :disabled="
                      !target.active_service?.reachable ||
                      submittingKey ===
                        buildAttackStateKey(activeChallengeRuntimeKey, target.team_id)
                    "
                    class="submit-btn"
                    @click="handleSubmit(activeChallengeRuntimeKey, target.team_id)"
                  >
                    {{
                      submittingKey ===
                      buildAttackStateKey(activeChallengeRuntimeKey, target.team_id)
                        ? '...'
                        : '提交'
                    }}
                  </button>
                </div>
              </article>
            </div>
          </div>

          <footer v-if="submitResult" class="ops-panel__footer">
            <div class="result-alert" :class="submitResult.is_success ? 'success' : 'danger'">
              <Terminal class="h-3.5 w-3.5" />
              <span>{{ getSubmitResultMessage() }}</span>
            </div>
          </footer>
        </section>
      </main>

      <!-- 3. Intelligence (Right) -->
      <aside class="war-room-col column-intel">
        <section class="ops-panel">
          <header class="ops-panel__header">
            <BarChart3 class="ops-panel__icon ops-panel__icon--accent h-4 w-4" />
            <h3 class="ops-panel__title">战场情报</h3>
          </header>
          <div class="ops-panel__content custom-scrollbar">
            <div
              v-for="item in scoreboardRows.slice(0, 10)"
              :key="item.team_id"
              class="intel-row"
              :class="{ 'is-me': item.team_id === myTeam?.team_id }"
            >
              <span class="intel-rank">#{{ item.rank }}</span>
              <span class="intel-name truncate">{{ item.team_name }}</span>
              <span class="intel-score font-mono">{{ item.score }}</span>
            </div>
          </div>
        </section>

        <section class="ops-panel">
          <header class="ops-panel__header">
            <History class="h-4 w-4 text-purple-500" />
            <h3 class="ops-panel__title">最近战报</h3>
          </header>
          <div class="ops-panel__content custom-scrollbar">
            <div
              v-for="event in workspace?.recent_events"
              :key="event.id"
              class="feedback-item"
              :class="event.direction"
            >
              <div class="flex items-center justify-between text-[10px] font-black">
                <span>{{ eventDirectionLabel(event.direction) }}</span>
                <span>{{ formatTime(event.created_at) }}</span>
              </div>
              <div class="mt-1 text-xs">
                <span>{{ event.peer_team_name }} / </span>
                <span data-testid="awd-feedback-challenge-title">{{
                  getChallengeTitleForEvent(event)
                }}</span>
              </div>
              <div class="feedback-ref">
                {{ formatServiceRef(event.service_id) }}
              </div>
              <div class="mt-1 flex items-center justify-between font-mono text-[10px]">
                <span
                  :class="event.is_success ? 'feedback-result--success' : 'feedback-result--muted'"
                  >{{ eventResultLabel(event.is_success) }}</span
                >
                <span class="feedback-score-gain">+{{ event.score_gained }}</span>
              </div>
            </div>
            <div v-if="workspace?.recent_events.length === 0" class="panel-note">
              暂无最近战报。
            </div>
          </div>
        </section>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.awd-war-room {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
  padding-top: var(--space-4);
}

/* HUD Strip */
.awd-hud-strip {
  display: grid;
  grid-template-columns: repeat(4, 1fr) auto;
  gap: var(--space-4);
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-default);
  border-radius: 1rem;
  padding: 1.25rem 1.5rem;
  box-shadow: var(--color-shadow-soft);
}

.hud-item {
  display: flex;
  flex-direction: column;
}

.hud-label {
  font-size: 10px;
  font-weight: 900;
  color: var(--color-text-muted);
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.hud-value {
  font-size: var(--font-size-24);
  font-weight: 900;
  color: var(--color-text-primary);
  margin: 0.25rem 0;
}

.hud-value--accent,
.ops-panel__icon--accent {
  color: var(--color-primary);
}

.ops-panel__icon--warning {
  color: var(--color-warning);
}

.ops-panel__icon--danger {
  color: var(--color-danger);
}

.hud-helper {
  font-size: 11px;
  font-weight: 800;
  color: var(--color-primary);
}

.hud-refresh-btn {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  padding: 0 1.25rem;
  border-left: 1px solid var(--color-border-subtle);
  color: var(--color-text-secondary);
  font-size: 11px;
  font-weight: 800;
  cursor: pointer;
  background: transparent;
  transition: all 0.2s ease;
}

.hud-refresh-btn:hover {
  color: var(--color-text-primary);
}

/* Layout Grid */
.war-room-grid {
  display: grid;
  grid-template-columns: minmax(20rem, 24rem) minmax(0, 1fr);
  grid-template-areas:
    'defense attack'
    'defense intel';
  gap: var(--space-5);
  flex: 1;
  min-height: 0;
  align-items: stretch;
}

.column-defense {
  grid-area: defense;
}

.column-attack {
  grid-area: attack;
}

.column-intel {
  grid-area: intel;
  display: grid;
  grid-template-columns: minmax(0, 0.9fr) minmax(0, 1.1fr);
  gap: var(--space-5);
}

.ops-panel {
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-default);
  border-radius: 1rem;
  display: flex;
  flex-direction: column;
  min-height: 0;
  box-shadow: var(--color-shadow-soft);
  overflow: hidden;
}

.column-intel .ops-panel {
  min-height: 18rem;
}

.ops-panel__header {
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--color-border-subtle);
  display: flex;
  align-items: center;
  gap: 0.75rem;
  background: var(--color-bg-elevated);
}

.ops-panel__title {
  font-size: 12px;
  font-weight: 900;
  letter-spacing: 0.15em;
  color: var(--color-text-primary);
  margin: 0;
}

.ops-panel__toolbar {
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--color-border-subtle);
  display: flex;
  gap: 1rem;
  background: var(--color-bg-surface);
}

.toolbar-field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  flex: 1;
}

.toolbar-field label {
  font-size: 9px;
  font-weight: 900;
  color: var(--color-text-muted);
  letter-spacing: 0.1em;
}

.war-room-select,
.war-room-input {
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border-default);
  border-radius: 0.5rem;
  color: var(--color-text-primary);
  font-size: 12px;
  font-weight: 700;
  padding: 0.5rem 0.75rem;
  outline: none;
  transition: all 0.2s ease;
}
.war-room-select:focus,
.war-room-input:focus {
  border-color: var(--color-primary);
}

.ops-panel__content {
  flex: 1;
  overflow-y: auto;
  padding: 1.25rem;
}

/* Defense Components */
.defense-alert {
  background: var(--color-warning-soft);
  border: 1px solid color-mix(in srgb, var(--color-warning) 20%, transparent);
  border-left: 3px solid var(--color-warning);
  border-radius: 0.5rem;
  padding: 0.75rem 1rem;
  margin-bottom: 0.75rem;
}

.defense-alert.danger {
  background: var(--color-danger-soft);
  border-color: color-mix(in srgb, var(--color-danger) 20%, transparent);
  border-left-color: var(--color-danger);
}

.alert-title {
  font-size: 12px;
  font-weight: 800;
  color: var(--color-text-primary);
}
.alert-badge {
  font-size: 10px;
  font-weight: 900;
}
.alert-issues {
  font-size: 10px;
  font-weight: 700;
  color: var(--color-text-secondary);
  margin-top: 0.35rem;
  display: flex;
  gap: 0.5rem;
}

.target-ref,
.feedback-ref {
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
}

/* Attack Components */
.target-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(20rem, 1fr));
  gap: 1.25rem;
}

.target-card {
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border-default);
  padding: 1.25rem;
  border-radius: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.target-team {
  font-size: 14px;
  letter-spacing: 0.05em;
  color: var(--color-primary);
}
.target-url {
  font-size: 11px;
  color: var(--color-text-muted);
}
.target-ref {
  margin-top: 0.2rem;
}

.target-action {
  display: flex;
  gap: 0.5rem;
}

.target-open-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-1);
  min-width: 4.5rem;
  border: 1px solid var(--color-border-default);
  border-radius: 0.5rem;
  background: var(--color-bg-surface);
  color: var(--color-text-secondary);
  font-size: 12px;
  font-weight: 900;
  cursor: pointer;
  transition: all 0.2s ease;
}

.target-open-btn:hover:not(:disabled) {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.target-open-btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.flag-input {
  flex: 1;
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-default);
  padding: 0.5rem 0.75rem;
  border-radius: 0.5rem;
  color: var(--color-text-primary);
  font-family: var(--font-family-mono);
  font-size: 13px;
  outline: none;
  transition: border-color 0.2s ease;
}
.flag-input:focus {
  border-color: var(--color-primary);
}

.submit-btn {
  background: var(--color-danger);
  color: var(--color-bg-base);
  border: none;
  padding: 0 1.25rem;
  border-radius: 0.5rem;
  font-weight: 900;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}
.submit-btn:hover:not(:disabled) {
  background: color-mix(in srgb, var(--color-danger) 80%, var(--color-bg-base));
}
.submit-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.feedback-result--success,
.feedback-score-gain {
  color: var(--color-success);
}

.feedback-result--muted {
  color: var(--color-text-muted);
}

/* Intel Components */
.intel-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.65rem 0;
  border-bottom: 1px solid var(--color-border-subtle);
  font-size: 13px;
}

.intel-row.is-me {
  color: var(--color-primary);
}
.intel-rank {
  font-weight: 900;
  color: var(--color-text-muted);
  width: 2rem;
}
.intel-name {
  flex: 1;
  font-weight: 700;
  color: var(--color-text-primary);
}
.intel-score {
  font-weight: 900;
  color: var(--color-text-primary);
}
.is-me .intel-name,
.is-me .intel-score {
  color: var(--color-primary);
}

.feedback-item {
  background: var(--color-bg-elevated);
  padding: 0.75rem 1rem;
  border-radius: 0.5rem;
  border-left: 2px solid var(--color-border-default);
  margin-bottom: 0.75rem;
}

.feedback-item.attack_out {
  border-left-color: var(--color-danger);
}
.feedback-item.attack_in {
  border-left-color: var(--color-warning);
}
.feedback-ref {
  margin-top: 0.35rem;
}

.panel-note {
  font-size: 12px;
  font-weight: 700;
  color: var(--color-text-muted);
  text-align: center;
  padding: 3rem 0;
}

.ops-panel__footer {
  padding: 1rem 1.25rem;
  border-top: 1px solid var(--color-border-subtle);
  background: var(--color-bg-surface);
}

.result-alert {
  padding: 0.65rem 1rem;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 12px;
  font-weight: 900;
}

.result-alert.success {
  background: var(--color-success-soft);
  color: var(--color-success);
  border: 1px solid color-mix(in srgb, var(--color-success) 20%, transparent);
}
.result-alert.danger {
  background: var(--color-danger-soft);
  color: var(--color-danger);
  border: 1px solid color-mix(in srgb, var(--color-danger) 20%, transparent);
}

.war-room-loading {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2rem;
  font-size: 14px;
  font-weight: 900;
  color: var(--color-primary);
  padding: 4rem 0;
}

.radar-scan {
  width: 5rem;
  height: 5rem;
  border: 2px solid var(--color-primary);
  border-radius: 50%;
  position: relative;
}

.radar-scan::after {
  content: '';
  position: absolute;
  inset: 0;
  background: conic-gradient(from 0deg, var(--color-primary), transparent);
  border-radius: 50%;
  animation: radar 2s linear infinite;
  opacity: 0.3;
}

@keyframes radar {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 1280px) {
  .war-room-grid {
    grid-template-columns: 1fr;
    grid-template-areas:
      'defense'
      'attack'
      'intel';
  }
  .column-intel {
    grid-template-columns: 1fr;
  }
}

.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: var(--color-border-default);
  border-radius: 10px;
}
</style>
