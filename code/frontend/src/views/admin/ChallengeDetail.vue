<template>
  <section
    class="workspace-shell journal-shell journal-shell-admin journal-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
  >
    <header class="workspace-topbar">
      <div class="topbar-leading">
        <span class="workspace-overline">Challenge Workspace</span>
        <span class="class-chip">{{ workspaceLabel }}</span>
      </div>
      <div class="topbar-actions">
        <button
          v-if="challengeId"
          class="admin-btn admin-btn-primary"
          type="button"
          @click="openTopology"
        >
          拓扑编排
        </button>
        <button class="admin-btn admin-btn-ghost" type="button" @click="openChallengeList">
          返回题库
        </button>
      </div>
    </header>

    <nav class="top-tabs" role="tablist" aria-label="题目管理视图切换">
      <button
        v-for="(tab, index) in panelTabs"
        :id="tab.tabId"
        :key="tab.key"
        type="button"
        role="tab"
        class="top-tab"
        :class="{ active: activePanel === tab.key }"
        :aria-selected="activePanel === tab.key ? 'true' : 'false'"
        :aria-controls="tab.panelId"
        :tabindex="activePanel === tab.key ? 0 : -1"
        @click="switchPanel(tab.key)"
        @keydown="handleTabKeydown($event, index)"
      >
        {{ tab.label }}
      </button>
    </nav>

    <main class="content-pane">
      <div v-if="loading" class="flex items-center justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]"
        />
      </div>

      <template v-else-if="challenge">
        <section
          id="admin-challenge-panel-detail"
          class="tab-panel challenge-panel"
          role="tabpanel"
          aria-labelledby="admin-challenge-tab-detail"
          :aria-hidden="activePanel === 'detail' ? 'false' : 'true'"
          v-show="activePanel === 'detail'"
        >
          <div class="workspace-tab-heading">
            <div class="workspace-tab-heading__main">
              <div class="journal-note-label">Question Ops</div>
              <h1 class="workspace-page-title">题目管理</h1>
            </div>
            <p class="workspace-tab-copy">
              查看题目基础信息、提示与附件，并维护当前判题模式配置。
            </p>
          </div>

          <section class="challenge-summary">
            <div class="challenge-summary__head">
              <div class="journal-note-label">Challenge Profile</div>
              <h2 class="challenge-summary__title">{{ challenge.title }}</h2>
            </div>

            <dl class="challenge-meta-grid">
              <div class="challenge-meta-item">
                <dt>分类</dt>
                <dd>{{ challenge.category }}</dd>
              </div>
              <div class="challenge-meta-item">
                <dt>难度</dt>
                <dd>{{ challenge.difficulty }}</dd>
              </div>
              <div class="challenge-meta-item">
                <dt>分值</dt>
                <dd>{{ challenge.points }}</dd>
              </div>
              <div class="challenge-meta-item">
                <dt>状态</dt>
                <dd>{{ challenge.status }}</dd>
              </div>
              <div v-if="challenge.image_id" class="challenge-meta-item">
                <dt>镜像</dt>
                <dd class="challenge-meta-item__mono">ID #{{ challenge.image_id }}</dd>
              </div>
              <div class="challenge-meta-item">
                <dt>Flag 配置</dt>
                <dd class="challenge-meta-item__mono">{{ flagConfigSummary }}</dd>
              </div>
              <div v-if="challenge.attachment_url" class="challenge-meta-item challenge-meta-item--full">
                <dt>附件</dt>
                <dd>
                  <a
                    :href="challenge.attachment_url"
                    target="_blank"
                    rel="noreferrer"
                    class="challenge-link"
                  >
                    {{ challenge.attachment_url }}
                  </a>
                </dd>
              </div>
            </dl>
          </section>

          <ChallengeDescriptionPanel
            v-if="challenge.description"
            :content="challenge.description"
            label="描述"
            test-id="challenge-detail-description"
          />

          <section v-if="challenge.hints?.length" class="challenge-section">
            <div class="workspace-tab-heading">
              <div class="workspace-tab-heading__main">
                <div class="journal-note-label">Hints</div>
                <h2 class="workspace-tab-heading__title">提示管理</h2>
              </div>
            </div>

            <div class="hint-list">
              <article
                v-for="hint in challenge.hints"
                :key="hint.id || hint.level"
                class="hint-card"
              >
                <div class="hint-card__title">
                  Level {{ hint.level }}{{ hint.title ? ` · ${hint.title}` : '' }}
                </div>
                <div class="hint-card__content">{{ hint.content }}</div>
              </article>
            </div>
          </section>

          <section class="journal-panel challenge-section challenge-flag-panel p-5 md:p-6">
            <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
              <div>
                <div class="journal-note-label">Judge Mode</div>
                <h2 class="challenge-flag-panel__title">判题模式配置</h2>
                <p class="challenge-flag-panel__copy">
                  支持静态 Flag、动态前缀、正则判题和人工审核四种模式。保存后即时刷新当前题目配置。
                </p>
              </div>
              <div class="flag-summary-chip">
                {{ flagDraftSummary }}
              </div>
            </div>

            <div class="grid gap-4 md:grid-cols-2">
              <label class="flag-field">
                <span class="flag-field-label">判题模式</span>
                <select v-model="flagType" class="flag-field-input">
                  <option value="static">静态 Flag</option>
                  <option value="dynamic">动态前缀</option>
                  <option value="regex">正则匹配</option>
                  <option value="manual_review">人工审核</option>
                </select>
              </label>

              <label v-if="flagType === 'dynamic' || flagType === 'regex'" class="flag-field">
                <span class="flag-field-label">Flag 前缀</span>
                <input
                  v-model="flagPrefix"
                  type="text"
                  placeholder="例如：flag"
                  class="flag-field-input"
                />
              </label>

              <label v-if="flagType === 'static'" class="flag-field md:col-span-2">
                <span class="flag-field-label">静态 Flag</span>
                <input
                  v-model="flagValue"
                  type="text"
                  placeholder="例如：flag{demo}"
                  class="flag-field-input font-mono"
                />
              </label>

              <label v-if="flagType === 'regex'" class="flag-field md:col-span-2">
                <span class="flag-field-label">正则表达式</span>
                <input
                  v-model="flagRegex"
                  type="text"
                  placeholder="例如：^flag\\{demo-[0-9]+\\}$"
                  class="flag-field-input font-mono"
                />
              </label>
            </div>

            <div
              v-if="isSharedInstanceChallenge"
              class="challenge-flag-panel__warning"
            >
              共享实例只适用于无状态题。该模式不提供用户级答案隔离，静态/正则答案可能被转发；若需隔离答案，请使用 per_user 或 per_team。
            </div>

            <div
              v-if="flagType === 'manual_review'"
              class="challenge-flag-panel__warning"
            >
              学生提交的答案将进入教师审核队列。审核通过后才会计分并更新通过状态。
            </div>

            <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div class="text-sm text-[var(--journal-muted)]">当前配置：{{ flagConfigSummary }}</div>
              <button
                :disabled="saving"
                class="admin-btn admin-btn-primary"
                type="button"
                @click="saveFlagConfig"
              >
                {{ saving ? '保存中...' : '保存配置' }}
              </button>
            </div>
          </section>
        </section>

        <section
          id="admin-challenge-panel-writeup"
          class="tab-panel challenge-panel"
          role="tabpanel"
          aria-labelledby="admin-challenge-tab-writeup"
          :aria-hidden="activePanel === 'writeup' ? 'false' : 'true'"
          v-show="activePanel === 'writeup'"
        >
          <ChallengeWriteupManagePanel
            :challenge-id="challengeId"
            :challenge-title="challenge.title"
          />
        </section>
      </template>
    </main>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import type { AdminChallengeFlagPayload } from '@/api/admin'
import { configureChallengeFlag, getChallengeDetail } from '@/api/admin'
import type { AdminChallengeListItem, FlagType } from '@/api/contracts'
import ChallengeDescriptionPanel from '@/components/admin/challenge/ChallengeDescriptionPanel.vue'
import ChallengeWriteupManagePanel from '@/components/admin/writeup/ChallengeWriteupManagePanel.vue'
import { useToast } from '@/composables/useToast'

type ChallengePanelKey = 'detail' | 'writeup'

const panelTabs = [
  {
    key: 'detail' as const,
    label: '题目管理',
    tabId: 'admin-challenge-tab-detail',
    panelId: 'admin-challenge-panel-detail',
  },
  {
    key: 'writeup' as const,
    label: '题解管理',
    tabId: 'admin-challenge-tab-writeup',
    panelId: 'admin-challenge-panel-writeup',
  },
]

const route = useRoute()
const router = useRouter()
const toast = useToast()

const loading = ref(true)
const saving = ref(false)
const challenge = ref<AdminChallengeListItem | null>(null)
const flagType = ref<FlagType>('static')
const flagValue = ref('')
const flagRegex = ref('')
const flagPrefix = ref('')

const challengeId = computed(() => String(route.params.id || ''))
const activePanel = computed<ChallengePanelKey>(() => resolvePanel(route.query.panel))
const workspaceLabel = computed(() => challenge.value?.title || '题目管理')
const flagConfigSummary = computed(() => summarizeFlagConfig(challenge.value?.flag_config))
const isSharedInstanceChallenge = computed(() => challenge.value?.instance_sharing === 'shared')
const flagDraftSummary = computed(() =>
  summarizeFlagConfig({
    configured: true,
    flag_type: flagType.value,
    flag_regex: flagRegex.value.trim() || undefined,
    flag_prefix: flagPrefix.value.trim() || undefined,
  })
)

function resolvePanel(rawPanel: unknown): ChallengePanelKey {
  const panel = Array.isArray(rawPanel) ? rawPanel[0] : rawPanel
  return panel === 'writeup' ? 'writeup' : 'detail'
}

function buildPanelQuery(panel: ChallengePanelKey) {
  const nextQuery = { ...route.query }
  if (panel === 'detail') {
    delete nextQuery.panel
  } else {
    nextQuery.panel = panel
  }
  return nextQuery
}

function switchPanel(panel: ChallengePanelKey): void {
  if (!challengeId.value || panel === activePanel.value) return

  void router.replace({
    name: 'AdminChallengeDetail',
    params: { id: challengeId.value },
    query: buildPanelQuery(panel),
  })
}

function handleTabKeydown(event: KeyboardEvent, index: number): void {
  let targetIndex = index

  switch (event.key) {
    case 'ArrowRight':
      targetIndex = (index + 1) % panelTabs.length
      break
    case 'ArrowLeft':
      targetIndex = (index - 1 + panelTabs.length) % panelTabs.length
      break
    case 'Home':
      targetIndex = 0
      break
    case 'End':
      targetIndex = panelTabs.length - 1
      break
    default:
      return
  }

  event.preventDefault()
  switchPanel(panelTabs[targetIndex].key)
}

function openTopology(): void {
  if (!challengeId.value) return
  void router.push(`/platform/challenges/${challengeId.value}/topology`)
}

function openChallengeList(): void {
  void router.push('/platform/challenges')
}

function summarizeFlagConfig(config?: AdminChallengeListItem['flag_config']): string {
  if (!config?.configured) return '未配置'

  switch (config.flag_type) {
    case 'static':
      return '静态 Flag'
    case 'dynamic':
      return `动态 Flag / 前缀 ${config.flag_prefix || 'flag'}`
    case 'regex':
      return `正则匹配 / ${config.flag_regex || '未填写'}`
    case 'manual_review':
      return '人工审核'
    default:
      return '未配置'
  }
}

function hydrateFlagForm(item: AdminChallengeListItem | null): void {
  const config = item?.flag_config
  flagType.value = config?.flag_type ?? 'static'
  flagValue.value = ''
  flagRegex.value = config?.flag_regex ?? ''
  flagPrefix.value = config?.flag_prefix ?? ''
}

async function loadChallenge(id: string): Promise<void> {
  if (!id) {
    challenge.value = null
    loading.value = false
    return
  }

  try {
    challenge.value = await getChallengeDetail(id)
    hydrateFlagForm(challenge.value)
  } catch {
    challenge.value = null
    toast.error('加载失败')
    setTimeout(() => {
      void router.push('/platform/challenges')
    }, 1500)
  } finally {
    loading.value = false
  }
}

async function saveFlagConfig() {
  if (isSharedInstanceChallenge.value && flagType.value === 'dynamic') {
    toast.error('共享实例只适用于无状态题，不支持动态 Flag；若需隔离答案，请使用 per_user 或 per_team')
    return
  }

  const payload: AdminChallengeFlagPayload = {
    flag_type: flagType.value,
  }

  if (flagType.value === 'static') {
    if (!flagValue.value.trim()) {
      toast.error('请填写静态 Flag')
      return
    }
    payload.flag = flagValue.value.trim()
  }

  if (flagType.value === 'dynamic') {
    if (!flagPrefix.value.trim()) {
      toast.error('请填写动态 Flag 前缀')
      return
    }
    payload.flag_prefix = flagPrefix.value.trim()
  }

  if (flagType.value === 'regex') {
    if (!flagRegex.value.trim()) {
      toast.error('请填写正则表达式')
      return
    }
    payload.flag_regex = flagRegex.value.trim()
    if (flagPrefix.value.trim()) {
      payload.flag_prefix = flagPrefix.value.trim()
    }
  }

  saving.value = true
  try {
    await configureChallengeFlag(challengeId.value, payload)
    toast.success('Flag 配置已保存')
    loading.value = true
    await loadChallenge(challengeId.value)
  } catch {
    toast.error('保存 Flag 配置失败')
  } finally {
    saving.value = false
  }
}

watch(
  challengeId,
  (id) => {
    loading.value = true
    void loadChallenge(id)
  },
  { immediate: true }
)
</script>

<style scoped>
.journal-shell {
  --journal-topbar-padding-bottom: var(--space-3);
  --page-top-tabs-gap: var(--space-7);
  --page-top-tabs-margin: var(--space-2-5) calc(var(--space-6) * -1) 0;
  --page-top-tabs-padding: 0 var(--space-6);
  --page-top-tabs-border: color-mix(in srgb, var(--journal-ink) 10%, transparent);
  --page-top-tab-min-height: 52px;
  --page-top-tab-padding: var(--space-2-5) 0 var(--space-3-5);
  --page-top-tab-font-size: var(--font-size-15);
  --page-top-tab-active-color: color-mix(in srgb, var(--journal-accent) 74%, var(--journal-ink));
  --page-top-tab-active-border: color-mix(in srgb, var(--journal-accent) 86%, var(--journal-ink));
  --journal-shell-hero-radial-strength: 7%;
  --journal-shell-hero-radial-size: 22rem;
  --journal-shell-hero-end: var(--journal-surface);
  --journal-shell-hero-shadow: 0 22px 50px var(--color-shadow-soft);
}

.workspace-overline {
  font-size: var(--font-size-0-70);
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.content-pane {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
}

.topbar-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.tab-panel {
  display: grid;
  gap: var(--space-5);
}

.challenge-panel {
  padding-top: var(--space-6);
}

.challenge-summary {
  display: grid;
  gap: var(--space-4);
}

.challenge-summary__head {
  display: grid;
  gap: var(--space-1);
}

.challenge-summary__title {
  margin: 0;
  font-size: var(--font-size-1-20);
  font-weight: 700;
  color: var(--journal-ink);
}

.challenge-meta-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.challenge-meta-item {
  display: grid;
  gap: var(--space-1-5);
  padding: var(--space-4) 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.challenge-meta-item:nth-child(odd) {
  padding-right: var(--space-4);
}

.challenge-meta-item:nth-child(even) {
  padding-left: var(--space-4);
}

.challenge-meta-item dt {
  font-size: var(--font-size-0-74);
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.challenge-meta-item dd {
  margin: 0;
  font-size: var(--font-size-0-92);
  font-weight: 600;
  color: var(--journal-ink);
  word-break: break-word;
}

.challenge-meta-item__mono {
  font-family:
    'IBM Plex Mono',
    'JetBrains Mono',
    'SFMono-Regular',
    'Consolas',
    monospace;
}

.challenge-meta-item--full {
  grid-column: 1 / -1;
  padding-inline: 0;
}

.challenge-link {
  color: var(--journal-accent);
  text-decoration: underline;
  text-decoration-thickness: 1px;
  text-underline-offset: 0.15em;
}

.challenge-section {
  display: grid;
  gap: var(--space-4);
}

.hint-list {
  display: grid;
  gap: var(--space-3);
}

.hint-card {
  display: grid;
  gap: var(--space-2);
  border-left: 2px solid color-mix(in srgb, var(--journal-accent) 26%, transparent);
  padding: var(--space-3) 0 var(--space-3) var(--space-4);
  background: color-mix(in srgb, var(--journal-surface-subtle) 85%, transparent);
}

.hint-card__title {
  font-size: var(--font-size-0-90);
  font-weight: 700;
  color: var(--journal-ink);
}

.hint-card__content {
  font-size: var(--font-size-0-88);
  line-height: 1.7;
  color: var(--journal-ink);
}

.challenge-flag-panel {
  display: grid;
  gap: var(--space-5);
}

.challenge-flag-panel__title {
  margin: var(--space-2) 0 0;
  font-size: var(--font-size-1-08);
  font-weight: 700;
  color: var(--journal-ink);
}

.challenge-flag-panel__copy {
  margin: var(--space-2) 0 0;
  font-size: var(--font-size-0-88);
  line-height: 1.7;
  color: var(--journal-muted);
}

.challenge-flag-panel__warning {
  border-radius: 1rem;
  border: 1px solid color-mix(in srgb, var(--color-warning) 30%, transparent);
  background: color-mix(in srgb, var(--color-warning) 10%, transparent);
  padding: var(--space-4);
  font-size: var(--font-size-0-88);
  line-height: 1.7;
  color: var(--journal-ink);
}

.admin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.45rem;
  border-radius: 0.75rem;
  border: 1px solid transparent;
  padding: var(--space-2) var(--space-4);
  font-size: var(--font-size-0-875);
  font-weight: 600;
  transition:
    border-color 150ms ease,
    background 150ms ease,
    color 150ms ease,
    box-shadow 150ms ease;
}

.admin-btn:focus-visible {
  outline: none;
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 18%, transparent);
}

.admin-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.admin-btn-primary {
  border-color: color-mix(in srgb, var(--journal-accent) 18%, transparent);
  background: var(--journal-accent);
  color: #fff;
}

.admin-btn-ghost {
  border-color: var(--journal-border);
  background: color-mix(in srgb, var(--journal-surface) 96%, transparent);
  color: var(--journal-ink);
}

.flag-summary-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid rgba(37, 99, 235, 0.18);
  background: rgba(37, 99, 235, 0.08);
  padding: var(--space-2) var(--space-3-5);
  font-size: var(--font-size-0-80);
  font-weight: 600;
  color: var(--journal-accent);
}

.flag-field {
  display: flex;
  flex-direction: column;
  gap: var(--space-2-5);
}

.flag-field-label {
  font-size: var(--font-size-0-82);
  font-weight: 600;
  color: var(--journal-ink);
}

.flag-field-input {
  min-height: 2.9rem;
  border: 1px solid var(--journal-border);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--journal-surface) 96%, transparent);
  padding: var(--space-3) var(--space-4);
  font-size: var(--font-size-0-92);
  color: var(--journal-ink);
  outline: none;
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease;
}

.flag-field-input:focus {
  border-color: rgba(37, 99, 235, 0.42);
  box-shadow: 0 0 0 4px rgba(37, 99, 235, 0.12);
}

@media (max-width: 900px) {
  .challenge-meta-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .challenge-meta-item,
  .challenge-meta-item:nth-child(odd),
  .challenge-meta-item:nth-child(even) {
    padding-inline: 0;
  }
}
</style>
