<template>
  <section
    class="workspace-shell journal-shell journal-shell-admin journal-hero flex min-h-full flex-1 flex-col"
  >
    <AdminChallengeTopbarPanel
      :workspace-label="workspaceLabel"
      :has-challenge-id="Boolean(challengeId)"
      @open-topology="openTopology"
      @open-challenge-list="openChallengeList"
    />

    <nav
      class="top-tabs"
      role="tablist"
      aria-label="题目管理视图切换"
    >
      <button
        v-for="(tab, index) in panelTabs"
        :id="tab.tabId"
        :key="tab.key"
        :ref="(element) => setTabButtonRef(tab.key, element as HTMLButtonElement | null)"
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
      <div
        v-if="loading"
        class="flex items-center justify-center py-12"
      >
        <div
          class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]"
        />
      </div>

      <template v-else-if="challenge">
        <!--
          <div class="workspace-overline">Challenge Profile</div>
          class="challenge-overview-summary progress-strip metric-panel-grid metric-panel-default-surface"
          <h2 class="list-heading__title">基础信息</h2>
          <div class="journal-note-label">Hints</div>
          <h2 class="list-heading__title">提示管理</h2>
        -->
        <AdminChallengeProfilePanel
          v-show="activePanel === 'detail'"
          id="admin-challenge-panel-detail"
          class="tab-panel challenge-panel"
          role="tabpanel"
          aria-labelledby="admin-challenge-tab-detail"
          :aria-hidden="activePanel === 'detail' ? 'false' : 'true'"
          :challenge="challenge"
          :downloading-attachment="downloadingAttachment"
          :flag-config-summary="flagConfigSummary"
          :flag-draft-summary="flagDraftSummary"
          :flag-type="flagType"
          :flag-value="flagValue"
          :flag-regex="flagRegex"
          :flag-prefix="flagPrefix"
          :saving="saving"
          :is-shared-instance-challenge="isSharedInstanceChallenge"
          @download-attachment="downloadAttachment"
          @save-flag-config="saveFlagConfig"
          @update:flag-type="flagType = $event"
          @update:flag-value="flagValue = $event"
          @update:flag-regex="flagRegex = $event"
          @update:flag-prefix="flagPrefix = $event"
        />

        <section
          v-show="activePanel === 'writeup'"
          id="admin-challenge-panel-writeup"
          class="tab-panel challenge-panel"
          role="tabpanel"
          aria-labelledby="admin-challenge-tab-writeup"
          :aria-hidden="activePanel === 'writeup' ? 'false' : 'true'"
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
import { computed, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import type { AdminChallengeFlagPayload } from '@/api/admin'
import { configureChallengeFlag, getChallengeDetail } from '@/api/admin'
import { downloadAttachment as downloadChallengeAttachment } from '@/api/challenge'
import type { AdminChallengeListItem, FlagType } from '@/api/contracts'
import AdminChallengeProfilePanel from '@/components/platform/challenge/AdminChallengeProfilePanel.vue'
import AdminChallengeTopbarPanel from '@/components/platform/challenge/AdminChallengeTopbarPanel.vue'
import ChallengeWriteupManagePanel from '@/components/platform/writeup/ChallengeWriteupManagePanel.vue'
import { useRouteQueryTabs } from '@/composables/useRouteQueryTabs'
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
const downloadingAttachment = ref(false)
const challenge = ref<AdminChallengeListItem | null>(null)
const flagType = ref<FlagType>('static')
const flagValue = ref('')
const flagRegex = ref('')
const flagPrefix = ref('')
let redirectTimer: ReturnType<typeof setTimeout> | null = null

const challengeId = computed(() => String(route.params.id || ''))
const panelTabOrder = panelTabs.map((tab) => tab.key) as ChallengePanelKey[]
const {
  activeTab: activePanel,
  setTabButtonRef,
  selectTab: switchPanel,
  handleTabKeydown,
} = useRouteQueryTabs<ChallengePanelKey>({
  route,
  router,
  orderedTabs: panelTabOrder,
  defaultTab: 'detail',
  routeName: 'PlatformChallengeDetail',
  routeParams: route.params,
})
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

function openTopology(): void {
  if (!challengeId.value) return
  void router.push(`/platform/challenges/${challengeId.value}/topology`)
}

function openChallengeList(): void {
  void router.push('/platform/challenges')
}

function clearRedirectTimer(): void {
  if (redirectTimer === null) {
    return
  }
  clearTimeout(redirectTimer)
  redirectTimer = null
}

async function downloadAttachment(): Promise<void> {
  const attachmentURL = challenge.value?.attachment_url?.trim()
  if (!attachmentURL) return

  try {
    const parsed = new URL(attachmentURL, window.location.origin)
    if (parsed.origin !== window.location.origin) {
      window.open(attachmentURL, '_blank', 'noopener')
      return
    }
  } catch {
    // fallback to axios download for relative urls
  }

  downloadingAttachment.value = true
  try {
    const { blob, filename } = await downloadChallengeAttachment(attachmentURL)
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    link.remove()
    URL.revokeObjectURL(url)
  } catch {
    toast.error('下载附件失败')
  } finally {
    downloadingAttachment.value = false
  }
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
    clearRedirectTimer()
    redirectTimer = setTimeout(() => {
      redirectTimer = null
      void router.push('/platform/challenges')
    }, 1500)
  } finally {
    loading.value = false
  }
}

async function saveFlagConfig() {
  if (isSharedInstanceChallenge.value && flagType.value === 'dynamic') {
    toast.error(
      '共享实例只适用于无状态题，不支持动态 Flag；若需隔离答案，请使用 per_user 或 per_team'
    )
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

onUnmounted(() => {
  clearRedirectTimer()
})
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

.tab-panel {
  display: grid;
  gap: var(--space-5);
}

.challenge-panel {
  padding-top: var(--space-6);
}
</style>
