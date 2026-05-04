<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  ShieldAlert,
  Sword,
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
  buildOpenSSHConfig,
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
const copiedSSHConfigKey = ref('')

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
const selectedDefenseServiceTitle = computed(
  () =>
    defenseServiceCards.value.find((card) => card.serviceId === selectedServiceId.value)?.title ||
    ''
)
const selectedDefenseServiceCard = computed(
  () => defenseServiceCards.value.find((card) => card.serviceId === selectedServiceId.value) || null
)
const selectedWorkspaceService = computed(() =>
  selectedServiceId.value ? servicesByServiceId.value.get(selectedServiceId.value) || null : null
)

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

function getOpenSSHConfig(serviceId?: string): string {
  const access = getSSHAccess(serviceId)
  return buildOpenSSHConfig(access?.ssh_profile)
}

function getSSHCommand(serviceId?: string): string {
  return getVSCodeSSHCommand(getSSHAccess(serviceId))
}

function openDefenseService(serviceId: string): void {
  const instanceId = servicesByServiceId.value.get(serviceId)?.instance_id
  if (!instanceId) return
  void openService(instanceId)
}

function requestDefenseSSH(serviceId: string): void {
  selectService(serviceId)
  void openDefenseSSH(serviceId)
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
  const copied = await copyTextToClipboard(getSSHCommand(serviceId), 'VS Code SSH 命令已复制')
  if (copied) {
    copiedSSHCommandKey.value = serviceId
  }
}

async function copySSHConfig(serviceId?: string): Promise<void> {
  const config = getOpenSSHConfig(serviceId)
  if (!serviceId) return
  const copied = await copyTextToClipboard(config, 'OpenSSH 配置已复制')
  if (copied) {
    copiedSSHConfigKey.value = serviceId
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

    <header class="awd-hud-strip">
      <div class="hud-heading">
        <div class="hud-label">AWD Workspace</div>
        <h3>战场态势</h3>
      </div>
      <div class="hud-metrics" aria-label="战场概览">
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
              @request-ssh="requestDefenseSSH"
              @restart-service="restartService"
            />

            <AWDDefenseOperationsPanel
              :service-card="selectedDefenseServiceCard"
              :service="selectedWorkspaceService"
              :service-title="selectedDefenseServiceTitle"
              :opening-service-key="openingServiceKey"
              :opening-ssh-key="openingSSHKey"
              :action-pending="Boolean(defenseServiceActionPendingById[selectedServiceId])"
              :loading="loading"
              :access="getSSHAccess(selectedServiceId)"
              :copied-command="copiedSSHCommandKey === selectedServiceId"
              :copied-config="copiedSSHConfigKey === selectedServiceId"
              @open-service="openDefenseService"
              @request-ssh="requestDefenseSSH"
              @restart-service="restartService"
              @refresh="refreshAll"
              @copy-command="copySSHCommand"
              @copy-config="copySSHConfig"
            />
          </div>
        </section>
      </aside>

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
                        ? '打开中'
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
                        ? '提交中'
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

      <aside class="war-room-col column-intel">
        <section class="ops-panel ops-panel--compact">
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

        <section class="ops-panel ops-panel--compact">
          <header class="ops-panel__header">
            <History class="ops-panel__icon ops-panel__icon--history h-4 w-4" />
            <h3 class="ops-panel__title">最近战报</h3>
          </header>
          <div class="ops-panel__content custom-scrollbar">
            <div
              v-for="event in workspace?.recent_events"
              :key="event.id"
              class="feedback-item"
              :class="event.direction"
            >
              <div class="feedback-topline">
                <span>{{ eventDirectionLabel(event.direction) }}</span>
                <span>{{ formatTime(event.created_at) }}</span>
              </div>
              <div class="feedback-title">
                <span>{{ event.peer_team_name }} / </span>
                <span data-testid="awd-feedback-challenge-title">{{
                  getChallengeTitleForEvent(event)
                }}</span>
              </div>
              <div class="feedback-ref">
                {{ formatServiceRef(event.service_id) }}
              </div>
              <div class="feedback-result-row">
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

.awd-hud-strip {
  display: grid;
  grid-template-columns: minmax(10rem, 0.8fr) minmax(0, 2.4fr) auto;
  align-items: stretch;
  gap: var(--space-5);
  background:
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--color-primary) 7%, transparent),
      transparent 46%
    ),
    var(--color-bg-surface);
  border: 1px solid var(--color-border-default);
  border-radius: var(--ui-control-radius-lg);
  padding: var(--space-5) var(--space-6);
  box-shadow: var(--color-shadow-soft);
}

.hud-heading {
  display: flex;
  min-width: 0;
  flex-direction: column;
  justify-content: center;
  gap: var(--space-2);
}

.hud-heading h3 {
  margin: 0;
  color: var(--color-text-primary);
  font-size: var(--font-size-22);
  font-weight: 900;
  line-height: 1.1;
}

.hud-metrics {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-3);
}

.hud-item {
  display: flex;
  flex-direction: column;
  min-width: 0;
  justify-content: center;
  border-left: 1px solid color-mix(in srgb, var(--color-border-default) 76%, transparent);
  padding-left: var(--space-4);
}

.hud-label {
  font-size: var(--font-size-11);
  font-weight: 900;
  color: var(--color-text-muted);
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.hud-value {
  font-size: var(--font-size-24);
  font-weight: 900;
  color: var(--color-text-primary);
  margin: var(--space-1) 0;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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

.ops-panel__icon--history {
  color: var(--color-text-secondary);
}

.hud-helper {
  font-size: var(--font-size-11);
  font-weight: 800;
  color: var(--color-text-secondary);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.hud-refresh-btn {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-width: 7.5rem;
  padding: 0 var(--space-4);
  border: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  border-radius: var(--ui-control-radius-sm);
  color: var(--color-text-secondary);
  font-size: var(--font-size-11);
  font-weight: 800;
  cursor: pointer;
  background: color-mix(in srgb, var(--color-bg-elevated) 72%, transparent);
  transition: all 0.2s ease;
}

.hud-refresh-btn:hover:not(:disabled) {
  color: var(--color-text-primary);
  border-color: color-mix(in srgb, var(--color-primary) 36%, transparent);
}

.hud-refresh-btn:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.war-room-grid {
  display: grid;
  grid-template-columns: minmax(18rem, 0.9fr) minmax(28rem, 1.7fr) minmax(18rem, 0.9fr);
  gap: var(--space-4);
  flex: 1;
  min-height: 0;
}

.war-room-col {
  min-width: 0;
}

.column-intel {
  display: grid;
  gap: var(--space-4);
  align-content: start;
}

.ops-panel {
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-default);
  border-radius: var(--ui-control-radius-lg);
  display: flex;
  flex-direction: column;
  min-height: 0;
  box-shadow: var(--color-shadow-soft);
  overflow: hidden;
}

.ops-panel--compact {
  min-height: 0;
}

.ops-panel__header {
  padding: var(--space-4) var(--space-5);
  border-bottom: 1px solid var(--color-border-subtle);
  display: flex;
  align-items: center;
  gap: var(--space-3);
  background: color-mix(in srgb, var(--color-bg-elevated) 78%, var(--color-bg-surface));
}

.ops-panel__title {
  font-size: var(--font-size-12);
  font-weight: 900;
  letter-spacing: 0.15em;
  color: var(--color-text-primary);
  margin: 0;
  text-transform: uppercase;
}

.ops-panel__toolbar {
  padding: var(--space-4) var(--space-5);
  border-bottom: 1px solid var(--color-border-subtle);
  display: grid;
  grid-template-columns: minmax(12rem, 0.8fr) minmax(14rem, 1fr);
  gap: var(--space-4);
  background: var(--color-bg-surface);
}

.toolbar-field {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  min-width: 0;
}

.toolbar-field label {
  font-size: var(--font-size-11);
  font-weight: 900;
  color: var(--color-text-muted);
  letter-spacing: 0.1em;
}

.war-room-select,
.war-room-input {
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border-default);
  border-radius: var(--ui-control-radius-sm);
  color: var(--color-text-primary);
  font-size: var(--font-size-12);
  font-weight: 700;
  min-height: var(--ui-control-height-sm);
  padding: 0 var(--space-3);
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
  padding: var(--space-5);
}

.defense-alerts {
  display: grid;
  gap: var(--space-2);
  margin-bottom: var(--space-4);
}

.defense-alert {
  background: var(--color-warning-soft);
  border: 1px solid color-mix(in srgb, var(--color-warning) 20%, transparent);
  border-left: var(--space-0-5) solid var(--color-warning);
  border-radius: var(--ui-control-radius-sm);
  padding: var(--space-3) var(--space-4);
}

.defense-alert.danger {
  background: var(--color-danger-soft);
  border-color: color-mix(in srgb, var(--color-danger) 20%, transparent);
  border-left-color: var(--color-danger);
}

.alert-title {
  font-size: var(--font-size-12);
  font-weight: 800;
  color: var(--color-text-primary);
}
.alert-badge {
  font-size: var(--font-size-11);
  font-weight: 900;
}
.alert-issues {
  font-size: var(--font-size-11);
  font-weight: 700;
  color: var(--color-text-secondary);
  margin-top: var(--space-2);
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.target-ref,
.feedback-ref {
  font-size: var(--font-size-11);
  font-weight: 800;
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
}

.target-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(19rem, 1fr));
  gap: var(--space-3);
}

.target-card {
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border-default);
  padding: var(--space-4);
  border-radius: var(--ui-control-radius-sm);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  min-width: 0;
}

.target-team {
  font-size: var(--font-size-14);
  letter-spacing: 0.05em;
  color: var(--color-primary);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.target-url {
  font-size: var(--font-size-11);
  color: var(--color-text-muted);
}
.target-ref {
  margin-top: var(--space-1);
}

.target-action {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  gap: var(--space-2);
  align-items: stretch;
}

.target-open-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-1);
  min-width: 4.5rem;
  border: 1px solid var(--color-border-default);
  border-radius: var(--ui-control-radius-sm);
  background: var(--color-bg-surface);
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
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
  min-width: 0;
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-default);
  padding: 0 var(--space-3);
  border-radius: var(--ui-control-radius-sm);
  color: var(--color-text-primary);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-13);
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
  padding: 0 var(--space-4);
  border-radius: var(--ui-control-radius-sm);
  font-weight: 900;
  font-size: var(--font-size-12);
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

.intel-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3) 0;
  border-bottom: 1px solid var(--color-border-subtle);
  font-size: var(--font-size-13);
}

.intel-row.is-me {
  color: var(--color-primary);
}
.intel-rank {
  font-weight: 900;
  color: var(--color-text-muted);
  width: var(--space-8);
}
.intel-name {
  flex: 1;
  min-width: 0;
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
  padding: var(--space-3) var(--space-4);
  border-radius: var(--ui-control-radius-sm);
  border-left: var(--space-0-5) solid var(--color-border-default);
  margin-bottom: var(--space-3);
}

.feedback-item.attack_out {
  border-left-color: var(--color-danger);
}
.feedback-item.attack_in {
  border-left-color: var(--color-warning);
}
.feedback-ref {
  margin-top: var(--space-2);
}

.feedback-topline,
.feedback-result-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-11);
  font-weight: 900;
}

.feedback-title {
  margin-top: var(--space-1);
  color: var(--color-text-primary);
  font-size: var(--font-size-12);
  font-weight: 700;
}

.feedback-result-row {
  margin-top: var(--space-1);
}

.panel-note {
  font-size: var(--font-size-12);
  font-weight: 700;
  color: var(--color-text-muted);
  text-align: center;
  padding: var(--space-12) 0;
}

.ops-panel__footer {
  padding: var(--space-4) var(--space-5);
  border-top: 1px solid var(--color-border-subtle);
  background: var(--color-bg-surface);
}

.result-alert {
  padding: var(--space-3) var(--space-4);
  border-radius: var(--ui-control-radius-sm);
  display: flex;
  align-items: center;
  gap: var(--space-3);
  font-size: var(--font-size-12);
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
  gap: var(--space-8);
  font-size: var(--font-size-14);
  font-weight: 900;
  color: var(--color-primary);
  padding: var(--space-12) 0;
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

@media (max-width: 72rem) {
  .awd-hud-strip {
    grid-template-columns: 1fr;
  }

  .hud-metrics {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .hud-refresh-btn {
    min-height: var(--ui-control-height-sm);
  }

  .war-room-grid {
    grid-template-columns: 1fr;
  }
  .column-defense,
  .column-intel {
    grid-row: auto;
  }
}

@media (max-width: 48rem) {
  .awd-hud-strip {
    padding: var(--space-4);
  }

  .hud-metrics,
  .ops-panel__toolbar,
  .target-action {
    grid-template-columns: 1fr;
  }

  .hud-item {
    border-left: 0;
    border-top: 1px solid color-mix(in srgb, var(--color-border-default) 76%, transparent);
    padding: var(--space-3) 0 0;
  }

  .target-open-btn,
  .submit-btn {
    min-height: var(--ui-control-height-sm);
  }
}

.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: var(--color-border-default);
  border-radius: 999px;
}
</style>
