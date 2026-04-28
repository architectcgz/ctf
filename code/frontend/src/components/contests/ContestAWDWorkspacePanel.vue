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
  Folder,
  FileText,
} from 'lucide-vue-next'

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
  startingServiceKey,
  openingServiceKey,
  openingSSHKey,
  sshAccessByServiceId,
  activeDefenseServiceId,
  defenseDirectory,
  defenseDirectoryPath,
  defenseFile,
  defenseDraft,
  defenseFilePath,
  loadingDefenseDirectory,
  loadingDefenseFile,
  savingDefenseFile,
  runningDefenseCommand,
  defenseCommand,
  defenseCommandResult,
  openingTargetKey,
  submittingKey,
  shouldAutoRefresh,
  lastSyncedAt,
  refreshAll,
  startService,
  openService,
  openDefenseSSH,
  openDefenseDirectory,
  openDefenseFile,
  openDefenseWorkbench,
  saveDefenseFile,
  runDefenseCommand,
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

const challengeByChallengeId = computed(() => {
  const map = new Map<string, ContestChallengeItem>()
  for (const item of props.challenges) {
    map.set(item.challenge_id, item)
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
    } else if (service.service_status === 'down') {
      issues.push('已离线')
      statusLabel = '告警'
    }

    if ((service.attack_received ?? 0) > 0) {
      issues.push(`检测到 ${service.attack_received} 次攻击`)
    }

    if (issues.length === 0) continue

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

function getServiceStatusLabel(status?: string): string {
  switch (status) {
    case 'up':
      return '正常'
    case 'down':
      return '离线'
    case 'compromised':
      return '失陷'
    default:
      return '待命'
  }
}

function getServiceStatusClass(status?: string): string {
  return `status-badge status-badge--${status || 'pending'}`
}

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

function getServiceStartKey(challenge: ContestChallengeItem): string {
  return challenge.awd_service_id || ''
}

function getChallengeTitleForEvent(event: { service_id?: string; challenge_id: string }): string {
  if (event.service_id) {
    const matchedByService = challengeByServiceId.value.get(event.service_id)
    if (matchedByService) return matchedByService.title
  }
  return challengeByChallengeId.value.get(event.challenge_id)?.title || event.challenge_id
}

function getSubmitResultMessage(): string {
  if (!submitResult.value) return ''
  return formatAttackResultToast(submitResult.value)
}

function formatAttackResultToast(result: {
  service_id?: string
  challenge_id: string
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

function getActiveDefenseTitle(): string {
  const challenge = runtimeChallenges.value.find(
    (item) => item.awd_service_id === activeDefenseServiceId.value
  )
  return challenge?.title || '防守工作台'
}

function getDefenseParentPath(): string {
  if (!defenseDirectoryPath.value || defenseDirectoryPath.value === '.') {
    return '.'
  }
  const parts = defenseDirectoryPath.value.split('/').filter(Boolean)
  parts.pop()
  return parts.length > 0 ? parts.join('/') : '.'
}

function openDefenseEntry(entry: { path: string; type: 'file' | 'dir' | 'other' }): void {
  if (entry.type === 'dir') {
    void openDefenseDirectory(entry.path)
    return
  }
  if (entry.type === 'file') {
    void openDefenseFile(entry.path)
  }
}

function isTargetServiceForChallenge(
  service: { service_id?: string; challenge_id: string },
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
            <h3 class="ops-panel__title">防守监控</h3>
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

            <!-- Services -->
            <div class="asset-list mt-4">
              <div class="asset-header">战队服务</div>
              <div v-if="runtimeChallenges.length === 0" class="panel-note">
                当前竞赛暂无可部署服务。
              </div>
              <div v-for="challenge in runtimeChallenges" :key="challenge.id" class="asset-item">
                <div class="asset-main">
                  <div class="flex items-center justify-between">
                    <div class="asset-title-stack">
                      <span class="asset-title">{{ challenge.title }}</span>
                      <span class="asset-ref">{{
                        formatServiceRef(challenge.awd_service_id)
                      }}</span>
                    </div>
                    <span
                      :class="getServiceStatusClass(getWorkspaceService(challenge)?.service_status)"
                    >
                      {{ getServiceStatusLabel(getWorkspaceService(challenge)?.service_status) }}
                    </span>
                  </div>
                  <div class="asset-meta font-mono text-[10px]">
                    {{
                      getWorkspaceService(challenge)?.instance_id
                        ? '已通过平台代理就绪'
                        : '等待分配实例'
                    }}
                  </div>
                  <div v-if="getSSHAccess(challenge.awd_service_id)" class="asset-ssh">
                    <div class="asset-ssh__label">SSH DEFENSE</div>
                    <code class="asset-ssh__command">{{
                      getSSHAccess(challenge.awd_service_id)?.command
                    }}</code>
                    <div class="asset-ssh__password font-mono">
                      PASS {{ getSSHAccess(challenge.awd_service_id)?.password }}
                    </div>
                  </div>
                </div>
                <div class="asset-actions">
                  <button
                    v-if="getWorkspaceService(challenge)?.instance_id"
                    :disabled="openingServiceKey === getWorkspaceService(challenge)?.instance_id"
                    class="asset-btn"
                    @click="
                      getWorkspaceService(challenge)?.instance_id &&
                      openService(getWorkspaceService(challenge)?.instance_id || '')
                    "
                  >
                    <ExternalLink class="h-3 w-3" />
                  </button>
                  <button
                    v-if="getWorkspaceService(challenge)?.instance_id"
                    :disabled="openingSSHKey === getServiceStartKey(challenge)"
                    class="asset-btn"
                    @click="challenge.awd_service_id && openDefenseSSH(challenge.awd_service_id)"
                  >
                    {{ openingSSHKey === getServiceStartKey(challenge) ? '...' : 'SSH' }}
                  </button>
                  <button
                    v-if="getWorkspaceService(challenge)?.instance_id"
                    :disabled="
                      loadingDefenseFile && activeDefenseServiceId === challenge.awd_service_id
                    "
                    class="asset-btn asset-btn--defense"
                    @click="
                      challenge.awd_service_id && openDefenseWorkbench(challenge.awd_service_id)
                    "
                  >
                    {{
                      loadingDefenseFile && activeDefenseServiceId === challenge.awd_service_id
                        ? '...'
                        : '防守'
                    }}
                  </button>
                  <button
                    :disabled="startingServiceKey === getServiceStartKey(challenge)"
                    class="asset-btn asset-btn--primary"
                    @click="challenge.awd_service_id && startService(challenge.awd_service_id)"
                  >
                    {{ startingServiceKey === getServiceStartKey(challenge) ? '...' : '重启' }}
                  </button>
                </div>
              </div>
            </div>

            <div v-if="activeDefenseServiceId" class="defense-workbench">
              <div class="defense-workbench__head">
                <div>
                  <div class="asset-header">防守工作台</div>
                  <div class="defense-workbench__title">{{ getActiveDefenseTitle() }}</div>
                </div>
                <button
                  class="asset-btn asset-btn--primary"
                  :disabled="savingDefenseFile || !defenseFile"
                  @click="saveDefenseFile"
                >
                  {{ savingDefenseFile ? '保存中' : '保存' }}
                </button>
              </div>
              <div class="defense-file-row">
                <input v-model="defenseFilePath" class="war-room-input" placeholder="app.py" />
                <button
                  class="asset-btn"
                  :disabled="loadingDefenseFile"
                  @click="openDefenseFile(defenseFilePath)"
                >
                  读取
                </button>
              </div>
              <div class="defense-browser">
                <div class="defense-file-list">
                  <div class="defense-file-list__bar">
                    <span>{{ defenseDirectoryPath }}</span>
                    <button
                      class="asset-btn"
                      :disabled="loadingDefenseDirectory || defenseDirectoryPath === '.'"
                      @click="openDefenseDirectory(getDefenseParentPath())"
                    >
                      上级
                    </button>
                  </div>
                  <button
                    v-for="entry in defenseDirectory?.entries || []"
                    :key="entry.path"
                    class="defense-file-entry"
                    :class="{ 'defense-file-entry--active': defenseFile?.path === entry.path }"
                    :disabled="
                      loadingDefenseDirectory || loadingDefenseFile || entry.type === 'other'
                    "
                    @click="openDefenseEntry(entry)"
                  >
                    <Folder v-if="entry.type === 'dir'" class="h-3.5 w-3.5" />
                    <FileText v-else class="h-3.5 w-3.5" />
                    <span>{{ entry.name }}</span>
                  </button>
                  <div
                    v-if="!loadingDefenseDirectory && (defenseDirectory?.entries.length || 0) === 0"
                    class="defense-file-empty"
                  >
                    当前目录为空
                  </div>
                </div>
                <div class="defense-editor-pane">
                  <textarea
                    v-model="defenseDraft"
                    class="defense-editor"
                    spellcheck="false"
                    :disabled="loadingDefenseFile || !defenseFile"
                  />
                  <div class="defense-file-meta">
                    {{
                      defenseFile ? `${defenseFile.path} · ${defenseFile.size} bytes` : '未载入文件'
                    }}
                  </div>
                </div>
              </div>
              <div class="defense-command">
                <input
                  v-model="defenseCommand"
                  class="war-room-input"
                  placeholder="ls"
                  @keyup.enter="runDefenseCommand()"
                />
                <button
                  class="asset-btn asset-btn--primary"
                  :disabled="runningDefenseCommand"
                  @click="runDefenseCommand()"
                >
                  {{ runningDefenseCommand ? '执行中' : '执行' }}
                </button>
              </div>
              <pre v-if="defenseCommandResult" class="defense-output">{{
                defenseCommandResult.output || '(无输出)'
              }}</pre>
            </div>
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
        <section class="ops-panel h-1/2 mb-4">
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

        <section class="ops-panel h-1/2">
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
  grid-template-columns: 22rem 1fr 22rem;
  gap: var(--space-5);
  flex: 1;
  min-height: 0;
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

.asset-header {
  font-size: 10px;
  font-weight: 900;
  color: var(--color-text-muted);
  letter-spacing: 0.1em;
  margin-bottom: 0.75rem;
}

.asset-item {
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border-default);
  padding: 1rem;
  border-radius: 0.75rem;
  margin-bottom: 0.75rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-3);
}

.asset-main {
  min-width: 0;
  flex: 1;
}

.asset-actions {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-shrink: 0;
}

.asset-title-stack {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}
.asset-title {
  font-size: 13px;
  font-weight: 800;
  color: var(--color-text-primary);
}
.asset-ref,
.target-ref,
.feedback-ref {
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
}
.asset-meta {
  color: var(--color-text-muted);
  margin-top: 0.35rem;
}

.asset-ssh {
  margin-top: var(--space-3);
  border: 1px solid color-mix(in srgb, var(--color-primary) 24%, transparent);
  border-radius: 0.625rem;
  background: color-mix(in srgb, var(--color-primary) 8%, var(--color-bg-surface));
  padding: var(--space-2);
}

.asset-ssh__label {
  font-size: 9px;
  font-weight: 900;
  letter-spacing: 0.1em;
  color: var(--color-primary);
}

.asset-ssh__command,
.asset-ssh__password {
  display: block;
  margin-top: var(--space-1);
  color: var(--color-text-primary);
  font-size: 11px;
  line-height: 1.45;
  overflow-wrap: anywhere;
}

.asset-ssh__command {
  font-family: var(--font-family-mono);
}

.status-badge {
  font-size: 10px;
  font-weight: 900;
  padding: 0.2rem 0.6rem;
  border-radius: 99px;
}

.status-badge--up {
  background: var(--color-success-soft);
  color: var(--color-success);
}
.status-badge--down {
  background: var(--color-danger-soft);
  color: var(--color-danger);
}
.status-badge--compromised {
  background: var(--color-warning-soft);
  color: var(--color-warning);
}

.asset-btn {
  width: 2.25rem;
  height: 2.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.5rem;
  background: var(--color-bg-surface);
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border-default);
  cursor: pointer;
  transition: all 0.2s ease;
}

.asset-btn:hover {
  color: var(--color-primary);
  border-color: var(--color-primary);
}

.asset-btn--primary {
  width: auto;
  padding: 0 1rem;
  font-size: 11px;
  font-weight: 900;
  background: var(--color-primary-soft);
  color: var(--color-primary);
  border-color: color-mix(in srgb, var(--color-primary) 20%, transparent);
}

.asset-btn--primary:hover {
  background: var(--color-primary);
  color: var(--color-bg-base);
}

.asset-btn--defense {
  width: auto;
  padding: 0 var(--space-3);
  color: var(--color-success);
}

.defense-workbench {
  margin-top: var(--space-5);
  padding: var(--space-3);
  border: 1px solid color-mix(in srgb, var(--color-success) 24%, transparent);
  border-radius: 0.875rem;
  background: color-mix(in srgb, var(--color-success) 7%, var(--color-bg-elevated));
}

.defense-workbench__head,
.defense-file-row,
.defense-command {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.defense-workbench__head {
  justify-content: space-between;
  margin-bottom: var(--space-3);
}

.defense-workbench__title {
  color: var(--color-text-primary);
  font-size: var(--font-size-14);
  font-weight: 900;
}

.defense-file-row,
.defense-command {
  margin-top: var(--space-2);
}

.defense-file-row .war-room-input,
.defense-command .war-room-input {
  min-width: 0;
  flex: 1;
}

.defense-browser {
  display: grid;
  grid-template-columns: minmax(9rem, 0.4fr) minmax(0, 1fr);
  gap: var(--space-3);
  margin-top: var(--space-3);
}

.defense-file-list {
  min-height: 18rem;
  max-height: 24rem;
  overflow: auto;
  border: 1px solid var(--color-border-default);
  border-radius: 0.75rem;
  background: var(--color-bg-surface);
}

.defense-file-list__bar {
  position: sticky;
  top: 0;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
  padding: var(--space-2);
  border-bottom: 1px solid var(--color-border-subtle);
  background: var(--color-bg-surface);
  color: var(--color-text-muted);
  font-size: var(--font-size-12);
  font-weight: 800;
}

.defense-file-list__bar span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.defense-file-entry {
  display: flex;
  width: 100%;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
  font-weight: 800;
  text-align: left;
}

.defense-file-entry:hover:not(:disabled),
.defense-file-entry--active {
  background: color-mix(in srgb, var(--color-success) 14%, transparent);
  color: var(--color-text-primary);
}

.defense-file-empty {
  padding: var(--space-3);
  color: var(--color-text-muted);
  font-size: var(--font-size-12);
  font-weight: 800;
}

.defense-editor-pane {
  min-width: 0;
}

.defense-editor {
  width: 100%;
  min-height: 18rem;
  padding: var(--space-3);
  border: 1px solid var(--color-border-default);
  border-radius: 0.75rem;
  background: var(--color-bg-surface);
  color: var(--color-text-primary);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-12);
  line-height: 1.55;
  resize: vertical;
  outline: none;
}

.defense-editor:focus {
  border-color: var(--color-success);
}

.defense-file-meta {
  margin-top: var(--space-2);
  color: var(--color-text-muted);
  font-size: var(--font-size-12);
  font-weight: 800;
}

.defense-output {
  max-height: 12rem;
  margin-top: var(--space-3);
  padding: var(--space-3);
  overflow: auto;
  white-space: pre-wrap;
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.75rem;
  background: color-mix(in srgb, black 28%, var(--color-bg-surface));
  color: var(--color-text-primary);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-12);
  line-height: 1.5;
}

@media (max-width: 900px) {
  .defense-browser {
    grid-template-columns: 1fr;
  }
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
  }
  .column-defense,
  .column-intel {
    grid-row: auto;
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
