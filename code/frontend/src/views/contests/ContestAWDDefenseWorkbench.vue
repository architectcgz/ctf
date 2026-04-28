<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import {
  ArrowLeft,
  CheckCircle2,
  ChevronUp,
  FileText,
  Folder,
  RefreshCw,
  Save,
  Server,
  Terminal,
} from 'lucide-vue-next'

import AppEmpty from '@/components/common/AppEmpty.vue'
import {
  getContestChallenges,
  getContestDetail,
} from '@/api/contest'
import type {
  AWDDefenseDirectoryEntryData,
  ContestChallengeItem,
  ContestDetailData,
} from '@/api/contracts'
import { useContestAWDWorkspace } from '@/composables/useContestAWDWorkspace'
import { formatTime } from '@/utils/format'

const route = useRoute()
const contestId = computed(() => String(route.params.id ?? ''))
const serviceId = computed(() => String(route.params.serviceId ?? ''))

const contest = ref<ContestDetailData | null>(null)
const challenges = ref<ContestChallengeItem[]>([])
const pageLoading = ref(false)
const pageError = ref('')
const initializedServiceKey = ref('')

const {
  workspace,
  loading,
  error,
  hasTeam,
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
  sshAccessByServiceId,
  openingSSHKey,
  lastSyncedAt,
  refreshAll,
  openDefenseSSH,
  openDefenseDirectory,
  openDefenseFile,
  openDefenseWorkbench,
  saveDefenseFile,
  runDefenseCommand,
} = useContestAWDWorkspace({
  contestId,
  contestStatus: computed(() => contest.value?.status),
})

const activeService = computed(() =>
  workspace.value?.services.find((item) => item.service_id === serviceId.value)
)

const activeChallenge = computed(() =>
  challenges.value.find(
    (item) =>
      item.awd_service_id === serviceId.value ||
      item.challenge_id === activeService.value?.challenge_id
  )
)

const serviceStatusLabel = computed(() => {
  switch (activeService.value?.service_status) {
    case 'up':
      return '正常'
    case 'down':
      return '离线'
    case 'compromised':
      return '失陷'
    default:
      return '待命'
  }
})

const lastSyncedLabel = computed(() =>
  lastSyncedAt.value ? formatTime(lastSyncedAt.value) : '未同步'
)

const sshAccess = computed(() =>
  serviceId.value ? sshAccessByServiceId.value[serviceId.value] : undefined
)

async function loadPageContext(): Promise<void> {
  if (!contestId.value || pageLoading.value) {
    return
  }

  pageLoading.value = true
  pageError.value = ''
  try {
    const [nextContest, nextChallenges] = await Promise.all([
      getContestDetail(contestId.value),
      getContestChallenges(contestId.value),
    ])
    contest.value = nextContest
    challenges.value = nextChallenges
  } catch (err) {
    console.error(err)
    pageError.value = err instanceof Error ? err.message : '加载防守工作台失败'
  } finally {
    pageLoading.value = false
  }
}

function getDefenseParentPath(): string {
  const current = defenseDirectoryPath.value || '.'
  if (current === '.') {
    return '.'
  }
  const parts = current.split('/').filter(Boolean)
  parts.pop()
  return parts.length > 0 ? parts.join('/') : '.'
}

function openDefenseEntry(entry: AWDDefenseDirectoryEntryData): void {
  if (entry.type === 'dir') {
    void openDefenseDirectory(entry.path)
    return
  }
  if (entry.type === 'file') {
    void openDefenseFile(entry.path)
  }
}

function openDefenseParentDirectory(): void {
  void openDefenseDirectory(getDefenseParentPath())
}

function refreshWorkbench(): void {
  initializedServiceKey.value = ''
  void refreshAll()
}

function handleOpenDefenseSSH(): void {
  void openDefenseSSH(serviceId.value)
}

function handleOpenDefenseFile(): void {
  void openDefenseFile(defenseFilePath.value)
}

function handleSaveDefenseFile(): void {
  void saveDefenseFile()
}

function handleRunDefenseCommand(): void {
  void runDefenseCommand()
}

watch(
  contestId,
  () => {
    initializedServiceKey.value = ''
    void loadPageContext()
  },
  { immediate: true }
)

watch(
  () => [serviceId.value, workspace.value?.services.length, loading.value] as const,
  ([nextServiceId]) => {
    if (!nextServiceId || loading.value || initializedServiceKey.value === nextServiceId) {
      return
    }
    initializedServiceKey.value = nextServiceId
    void openDefenseWorkbench(nextServiceId)
  },
  { immediate: true }
)
</script>

<template>
  <section class="workspace-shell journal-shell journal-shell-user awd-defense-page">
    <header class="defense-topbar">
      <div class="defense-topbar__main">
        <RouterLink
          class="defense-back-link"
          :to="{ name: 'ContestDetail', params: { id: contestId }, query: { panel: 'challenges' } }"
        >
          <ArrowLeft class="defense-icon" />
          <span>返回战场</span>
        </RouterLink>
        <div class="defense-title-block">
          <div class="workspace-overline">AWD Defense</div>
          <h1 class="defense-title">
            {{ activeChallenge?.title || '防守工作台' }}
          </h1>
          <p class="defense-summary">
            {{ contest?.title || '正在同步竞赛信息' }}
          </p>
        </div>
      </div>

      <div class="defense-status-rail">
        <span class="defense-status-pill" :class="`defense-status-pill--${activeService?.service_status || 'pending'}`">
          <CheckCircle2 class="defense-icon" />
          {{ serviceStatusLabel }}
        </span>
        <button class="defense-action-btn" type="button" :disabled="loading" @click="refreshWorkbench">
          <RefreshCw class="defense-icon" :class="{ 'defense-icon--spin': loading }" />
          <span>{{ lastSyncedLabel }}</span>
        </button>
      </div>
    </header>

    <main class="defense-content">
      <div v-if="pageLoading || (loading && !workspace)" class="defense-state">
        正在加载防守工作台...
      </div>

      <AppEmpty
        v-else-if="pageError || error"
        icon="AlertTriangle"
        title="加载失败"
        :description="pageError || error"
        class="defense-empty"
      />

      <AppEmpty
        v-else-if="!hasTeam"
        icon="Users"
        title="先加入队伍"
        description="需要加入队伍后才能进入防守工作台。"
        class="defense-empty"
      />

      <AppEmpty
        v-else-if="!activeService"
        icon="Server"
        title="服务不可用"
        description="当前服务不在你的 AWD 战场服务列表中。"
        class="defense-empty"
      />

      <div v-else class="defense-workspace-grid">
        <aside class="defense-pane defense-pane--side">
          <div class="defense-pane__head">
            <Folder class="defense-icon" />
            <span>文件目录</span>
          </div>

          <div class="defense-directory-bar">
            <span :title="defenseDirectoryPath">{{ defenseDirectoryPath }}</span>
            <button
              class="defense-icon-btn"
              type="button"
              :disabled="loadingDefenseDirectory || defenseDirectoryPath === '.'"
              aria-label="返回上级目录"
              @click="openDefenseParentDirectory"
            >
              <ChevronUp class="defense-icon" />
            </button>
          </div>

          <div class="defense-file-list">
            <button
              v-for="entry in defenseDirectory?.entries || []"
              :key="entry.path"
              class="defense-file-entry"
              :class="{ 'defense-file-entry--active': defenseFile?.path === entry.path }"
              :disabled="loadingDefenseDirectory || loadingDefenseFile || entry.type === 'other'"
              type="button"
              @click="openDefenseEntry(entry)"
            >
              <Folder v-if="entry.type === 'dir'" class="defense-icon" />
              <FileText v-else class="defense-icon" />
              <span :title="entry.path">{{ entry.name }}</span>
            </button>
            <div
              v-if="!loadingDefenseDirectory && (defenseDirectory?.entries.length || 0) === 0"
              class="defense-muted"
            >
              当前目录为空
            </div>
          </div>

          <div class="defense-ssh-panel">
            <div class="defense-pane__head defense-pane__head--compact">
              <Server class="defense-icon" />
              <span>SSH</span>
            </div>
            <button
              class="defense-action-btn defense-action-btn--full"
              type="button"
              :disabled="openingSSHKey === serviceId"
              @click="handleOpenDefenseSSH"
            >
              {{ openingSSHKey === serviceId ? '生成中' : '生成连接' }}
            </button>
            <div v-if="sshAccess" class="defense-ssh-info">
              <code>{{ sshAccess.command }}</code>
              <span>PASS {{ sshAccess.password }}</span>
            </div>
          </div>
        </aside>

        <section class="defense-pane defense-pane--editor">
          <div class="defense-editor-toolbar">
            <div class="defense-file-path">
              <FileText class="defense-icon" />
              <input
                v-model="defenseFilePath"
                class="defense-input"
                placeholder="app.py"
                @keyup.enter="handleOpenDefenseFile"
              />
            </div>
            <div class="defense-editor-actions">
              <button
                class="defense-action-btn"
                type="button"
                :disabled="loadingDefenseFile"
                @click="handleOpenDefenseFile"
              >
                读取
              </button>
              <button
                class="defense-action-btn defense-action-btn--primary"
                type="button"
                :disabled="savingDefenseFile || !defenseFile"
                @click="handleSaveDefenseFile"
              >
                <Save class="defense-icon" />
                <span>{{ savingDefenseFile ? '保存中' : '保存' }}</span>
              </button>
            </div>
          </div>

          <textarea
            v-model="defenseDraft"
            class="defense-code-editor"
            spellcheck="false"
            :disabled="loadingDefenseFile || !defenseFile"
          />

          <div class="defense-editor-meta">
            {{ defenseFile ? `${defenseFile.path} · ${defenseFile.size} bytes` : '未载入文件' }}
          </div>
        </section>

        <aside class="defense-pane defense-pane--terminal">
          <div class="defense-pane__head">
            <Terminal class="defense-icon" />
            <span>命令执行</span>
          </div>
          <div class="defense-command-row">
            <input
              v-model="defenseCommand"
              class="defense-input"
              placeholder="ls"
              @keyup.enter="handleRunDefenseCommand"
            />
            <button
              class="defense-action-btn defense-action-btn--primary"
              type="button"
              :disabled="runningDefenseCommand"
              @click="handleRunDefenseCommand"
            >
              {{ runningDefenseCommand ? '执行中' : '执行' }}
            </button>
          </div>
          <pre class="defense-output">{{ defenseCommandResult?.output || '(暂无输出)' }}</pre>
        </aside>
      </div>
    </main>
  </section>
</template>

<style scoped>
.awd-defense-page {
  min-height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--color-bg-surface);
}

.defense-topbar {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-5);
  align-items: end;
  padding: var(--space-6) var(--space-6) var(--space-5);
  border-bottom: 1px solid var(--color-border-subtle);
}

.defense-topbar__main {
  display: grid;
  gap: var(--space-4);
  min-width: 0;
}

.defense-back-link,
.defense-action-btn,
.defense-icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  border: 1px solid var(--color-border-default);
  background: var(--color-bg-elevated);
  color: var(--color-text-secondary);
  border-radius: var(--radius-lg);
  min-height: var(--control-height-sm);
  padding: 0 var(--space-3);
  font-size: var(--font-size-13);
  font-weight: 700;
  transition:
    border-color var(--transition-fast),
    color var(--transition-fast),
    background var(--transition-fast);
}

.defense-back-link:hover,
.defense-action-btn:hover,
.defense-icon-btn:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.defense-action-btn:disabled,
.defense-icon-btn:disabled {
  cursor: not-allowed;
  opacity: 0.58;
}

.defense-title-block {
  display: grid;
  gap: var(--space-2);
  min-width: 0;
}

.defense-title {
  margin: 0;
  color: var(--color-text-primary);
  font-size: var(--workspace-page-title-font-size);
  line-height: var(--workspace-page-title-line-height);
  letter-spacing: var(--workspace-page-title-letter-spacing);
}

.defense-summary {
  margin: 0;
  max-width: 58rem;
  color: var(--color-text-secondary);
  font-size: var(--font-size-15);
  line-height: 1.7;
}

.defense-status-rail {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex-wrap: wrap;
  justify-content: flex-end;
}

.defense-status-pill {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  min-height: var(--control-height-sm);
  padding: 0 var(--space-3);
  border-radius: var(--radius-lg);
  border: 1px solid color-mix(in srgb, var(--color-border-default) 84%, transparent);
  font-size: var(--font-size-13);
  font-weight: 800;
  color: var(--color-text-secondary);
  background: var(--color-bg-elevated);
}

.defense-status-pill--up {
  color: var(--color-success);
  border-color: color-mix(in srgb, var(--color-success) 24%, transparent);
  background: color-mix(in srgb, var(--color-success) 8%, var(--color-bg-elevated));
}

.defense-status-pill--down,
.defense-status-pill--compromised {
  color: var(--color-danger);
  border-color: color-mix(in srgb, var(--color-danger) 24%, transparent);
  background: color-mix(in srgb, var(--color-danger) 8%, var(--color-bg-elevated));
}

.defense-content {
  flex: 1;
  min-height: 0;
  padding: var(--space-5) var(--space-6) var(--space-6);
}

.defense-workspace-grid {
  min-height: 42rem;
  display: grid;
  grid-template-columns: minmax(15rem, 18rem) minmax(0, 1fr) minmax(18rem, 22rem);
  grid-template-rows: minmax(0, 1fr);
  gap: var(--space-4);
}

.defense-pane {
  min-width: 0;
  min-height: 0;
  border: 1px solid var(--color-border-default);
  border-radius: var(--radius-xl);
  background: var(--color-bg-elevated);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.defense-pane__head,
.defense-editor-toolbar,
.defense-directory-bar,
.defense-command-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.defense-pane__head {
  min-height: var(--control-height-md);
  padding: 0 var(--space-4);
  border-bottom: 1px solid var(--color-border-subtle);
  color: var(--color-text-primary);
  font-size: var(--font-size-13);
  font-weight: 800;
}

.defense-pane__head--compact {
  min-height: auto;
  padding: 0;
  border-bottom: 0;
}

.defense-directory-bar {
  justify-content: space-between;
  padding: var(--space-3) var(--space-3) 0;
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
}

.defense-directory-bar span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.defense-icon-btn {
  width: var(--control-height-sm);
  padding: 0;
  flex-shrink: 0;
}

.defense-file-list {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: var(--space-3);
}

.defense-file-entry {
  width: 100%;
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  align-items: center;
  gap: var(--space-2);
  min-height: var(--control-height-sm);
  padding: 0 var(--space-2);
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--color-text-secondary);
  font-size: var(--font-size-13);
  text-align: left;
}

.defense-file-entry span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.defense-file-entry:hover,
.defense-file-entry--active {
  color: var(--color-primary);
  border-color: color-mix(in srgb, var(--color-primary) 24%, transparent);
  background: color-mix(in srgb, var(--color-primary) 7%, transparent);
}

.defense-ssh-panel {
  display: grid;
  gap: var(--space-3);
  padding: var(--space-4);
  border-top: 1px solid var(--color-border-subtle);
}

.defense-ssh-info {
  display: grid;
  gap: var(--space-2);
  color: var(--color-text-secondary);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-12);
  overflow-wrap: anywhere;
}

.defense-action-btn--full {
  width: 100%;
}

.defense-action-btn--primary {
  color: var(--color-primary);
  border-color: color-mix(in srgb, var(--color-primary) 28%, transparent);
  background: color-mix(in srgb, var(--color-primary) 8%, var(--color-bg-elevated));
}

.defense-editor-toolbar {
  justify-content: space-between;
  padding: var(--space-3);
  border-bottom: 1px solid var(--color-border-subtle);
}

.defense-file-path {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  min-width: 0;
  flex: 1;
}

.defense-editor-actions {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-shrink: 0;
}

.defense-input {
  width: 100%;
  min-height: var(--control-height-sm);
  border: 1px solid var(--color-border-default);
  border-radius: var(--radius-md);
  background: var(--color-bg-surface);
  color: var(--color-text-primary);
  padding: 0 var(--space-3);
  font-size: var(--font-size-13);
  outline: none;
}

.defense-input:focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 12%, transparent);
}

.defense-code-editor {
  flex: 1;
  min-height: 0;
  width: 100%;
  resize: none;
  border: 0;
  outline: 0;
  padding: var(--space-4);
  background: var(--color-bg-base);
  color: var(--color-text-primary);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-13);
  line-height: 1.65;
}

.defense-editor-meta {
  padding: var(--space-2) var(--space-4);
  border-top: 1px solid var(--color-border-subtle);
  color: var(--color-text-muted);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-12);
}

.defense-pane--terminal {
  min-height: 0;
}

.defense-command-row {
  padding: var(--space-3);
  border-bottom: 1px solid var(--color-border-subtle);
}

.defense-output {
  flex: 1;
  min-height: 0;
  margin: 0;
  padding: var(--space-4);
  overflow: auto;
  background: var(--color-bg-base);
  color: var(--color-text-primary);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-12);
  line-height: 1.6;
  white-space: pre-wrap;
}

.defense-state,
.defense-empty,
.defense-muted {
  color: var(--color-text-secondary);
  font-size: var(--font-size-14);
}

.defense-state,
.defense-empty {
  padding: var(--space-8);
}

.defense-muted {
  padding: var(--space-4);
}

.defense-icon {
  width: 1em;
  height: 1em;
  flex-shrink: 0;
}

.defense-icon--spin {
  animation: defense-spin 900ms linear infinite;
}

@keyframes defense-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 1180px) {
  .defense-workspace-grid {
    grid-template-columns: minmax(14rem, 17rem) minmax(0, 1fr);
  }

  .defense-pane--terminal {
    grid-column: 1 / -1;
    min-height: 18rem;
  }
}

@media (max-width: 860px) {
  .defense-topbar {
    grid-template-columns: 1fr;
    align-items: start;
  }

  .defense-status-rail {
    justify-content: flex-start;
  }

  .defense-workspace-grid {
    grid-template-columns: 1fr;
  }

  .defense-pane--side,
  .defense-pane--terminal {
    min-height: 18rem;
  }
}
</style>
