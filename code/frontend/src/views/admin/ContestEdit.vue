<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ChevronLeft } from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'

import { getContest, getContestAWDReadiness, updateContest } from '@/api/admin'
import type { AWDReadinessData, ContestDetailData } from '@/api/contracts'
import type { AdminContestUpdatePayload } from '@/api/admin'
import AdminContestFormPanel from '@/components/admin/contest/AdminContestFormPanel.vue'
import AWDReadinessOverrideDialog from '@/components/admin/contest/AWDReadinessOverrideDialog.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import {
  createFieldLocks,
  createContestStatusOptions,
  createDraftFromContest,
  normalizeEditableStatus,
  type AdminContestStatus,
  type ContestFormDraft,
} from '@/composables/useAdminContests'
import { ApiError } from '@/api/request'
import { useToast } from '@/composables/useToast'

interface AWDStartOverrideDialogState {
  open: boolean
  title: string
  readiness: AWDReadinessData | null
  confirmLoading: boolean
  pendingPayload: AdminContestUpdatePayload | null
}

const ERR_AWD_READINESS_BLOCKED = 14025

const route = useRoute()
const router = useRouter()
const toast = useToast()

const contestId = computed(() => String(route.params.id ?? ''))
const loading = ref(true)
const loadError = ref('')
const saving = ref(false)
const contest = ref<ContestDetailData | null>(null)
const editingBaseStatus = ref<AdminContestStatus | null>(null)
const formDraft = ref<ContestFormDraft | null>(null)
const awdStartOverrideDialogState = ref<AWDStartOverrideDialogState>(createDefaultAWDStartOverrideDialogState())

const fieldLocks = computed(() => createFieldLocks(editingBaseStatus.value))
const statusOptions = computed(() => createContestStatusOptions(editingBaseStatus.value))
const pageTitle = computed(() => (contest.value ? `编辑《${contest.value.title}》` : '编辑竞赛'))

function createDefaultAWDStartOverrideDialogState(): AWDStartOverrideDialogState {
  return {
    open: false,
    title: '',
    readiness: null,
    confirmLoading: false,
    pendingPayload: null,
  }
}

function toISOString(value: string): string {
  return new Date(value).toISOString()
}

function shouldGateAWDContestStart(mode: ContestDetailData['mode'] | null, targetStatus: AdminContestStatus): boolean {
  return mode === 'awd' && targetStatus === 'running'
}

function isAWDReadinessBlockedError(error: unknown): boolean {
  return error instanceof ApiError && error.code === ERR_AWD_READINESS_BLOCKED
}

function humanizeRequestError(error: unknown, fallback: string): string {
  if (error instanceof ApiError && error.message.trim()) {
    return error.message
  }
  if (error instanceof Error && error.message.trim()) {
    return error.message
  }
  return fallback
}

async function loadContestDetail(): Promise<void> {
  if (!contestId.value) {
    loadError.value = '缺少赛事 ID，无法进入编辑页。'
    loading.value = false
    return
  }

  loading.value = true
  loadError.value = ''
  try {
    const detail = await getContest(contestId.value)
    contest.value = detail
    editingBaseStatus.value = normalizeEditableStatus(detail.status)
    formDraft.value = createDraftFromContest(detail)
  } catch (error) {
    loadError.value = humanizeRequestError(error, '竞赛详情加载失败')
  } finally {
    loading.value = false
  }
}

function goBackToContestList() {
  void router.push({ name: 'ContestManage', query: { panel: 'list' } })
}

async function finalizeContestUpdateSuccess() {
  toast.success('竞赛已更新')
  awdStartOverrideDialogState.value = createDefaultAWDStartOverrideDialogState()
  goBackToContestList()
}

async function openAWDStartOverrideDialog(payload: AdminContestUpdatePayload) {
  const readiness = await getContestAWDReadiness(contestId.value)
  awdStartOverrideDialogState.value = {
    open: true,
    title: '启动赛事',
    readiness,
    confirmLoading: false,
    pendingPayload: payload,
  }
}

function handleAwdStartOverrideDialogOpenChange(value: boolean) {
  if (!value) {
    awdStartOverrideDialogState.value = createDefaultAWDStartOverrideDialogState()
  }
}

async function confirmAWDStartOverride(reason: string) {
  const payload = awdStartOverrideDialogState.value.pendingPayload
  const normalizedReason = reason.trim()
  if (!payload || !normalizedReason) {
    return
  }

  awdStartOverrideDialogState.value = {
    ...awdStartOverrideDialogState.value,
    confirmLoading: true,
  }

  try {
    await updateContest(
      contestId.value,
      {
        ...payload,
        force_override: true,
        override_reason: normalizedReason,
      },
      { suppressErrorToast: true }
    )
    await finalizeContestUpdateSuccess()
  } catch (error) {
    if (isAWDReadinessBlockedError(error)) {
      await openAWDStartOverrideDialog(payload)
      return
    }
    toast.error(humanizeRequestError(error, '竞赛更新失败'))
  } finally {
    if (awdStartOverrideDialogState.value.open) {
      awdStartOverrideDialogState.value = {
        ...awdStartOverrideDialogState.value,
        confirmLoading: false,
      }
    }
  }
}

async function handleSave(draft: ContestFormDraft): Promise<void> {
  if (!contest.value) {
    return
  }

  saving.value = true
  try {
    const payload: AdminContestUpdatePayload = {
      title: draft.title.trim(),
      description: draft.description.trim(),
      mode: draft.mode,
      starts_at: toISOString(draft.starts_at),
      ends_at: toISOString(draft.ends_at),
      status: draft.status,
    }

    if (shouldGateAWDContestStart(contest.value.mode, draft.status)) {
      try {
        await updateContest(contestId.value, payload, { suppressErrorToast: true })
        await finalizeContestUpdateSuccess()
      } catch (error) {
        if (isAWDReadinessBlockedError(error)) {
          await openAWDStartOverrideDialog(payload)
          return
        }
        toast.error(humanizeRequestError(error, '竞赛更新失败'))
      }
      return
    }

    await updateContest(contestId.value, payload)
    await finalizeContestUpdateSuccess()
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  void loadContestDetail()
})
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-admin journal-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
  >
    <header class="workspace-topbar">
      <div class="topbar-leading">
        <span class="workspace-overline">Contest Workspace</span>
        <span class="class-chip">竞赛编辑</span>
      </div>
      <div class="topbar-actions">
        <button class="admin-btn admin-btn-ghost" type="button" @click="goBackToContestList">
          <ChevronLeft class="h-4 w-4" />
          返回竞赛目录
        </button>
      </div>
    </header>

    <main class="content-pane">
      <div v-if="loading" class="flex justify-center py-12">
        <AppLoading>正在同步竞赛详情...</AppLoading>
      </div>

      <AppEmpty
        v-else-if="loadError"
        title="竞赛详情加载失败"
        :description="loadError"
        icon="AlertTriangle"
      >
        <template #action>
          <button type="button" class="admin-btn admin-btn-ghost" @click="goBackToContestList">
            返回竞赛目录
          </button>
        </template>
      </AppEmpty>

      <template v-else-if="formDraft">
        <section class="workspace-directory-section contest-edit-section">
          <header class="contest-edit-header">
            <div class="workspace-tab-heading__main">
              <div class="workspace-overline">Contest Editor</div>
              <h1 class="workspace-page-title">编辑竞赛</h1>
              <p class="workspace-page-copy">
                {{ pageTitle }}。保存成功后会回到赛事目录，便于继续查看列表或进入后续编排。
              </p>
            </div>
          </header>

          <AdminContestFormPanel
            :mode="'edit'"
            :draft="formDraft"
            :saving="saving"
            :status-options="statusOptions"
            :field-locks="fieldLocks"
            :show-cancel="false"
            :note="'保存后将返回赛事目录列表；AWD 赛事切换到进行中时仍会执行就绪检查。'"
            @save="handleSave"
          />
        </section>
      </template>
    </main>

    <AWDReadinessOverrideDialog
      :open="awdStartOverrideDialogState.open"
      :title="awdStartOverrideDialogState.title"
      :readiness="awdStartOverrideDialogState.readiness"
      :confirm-loading="awdStartOverrideDialogState.confirmLoading"
      @update:open="handleAwdStartOverrideDialogOpenChange"
      @confirm="confirmAWDStartOverride"
    />
  </section>
</template>

<style scoped>
.content-pane {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  gap: var(--space-6);
}

.contest-edit-section {
  display: grid;
  gap: var(--space-5);
  padding: var(--space-5) var(--space-5-5);
}

.contest-edit-header {
  display: grid;
  gap: var(--space-4);
}

.admin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: 2.75rem;
  border-radius: 1rem;
  padding: var(--space-2-5) var(--space-4);
  font-size: var(--font-size-0-875);
  font-weight: 600;
  transition: all 150ms ease;
}

.admin-btn-ghost {
  border: 1px solid color-mix(in srgb, var(--journal-border) 76%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  color: var(--journal-ink);
}

.admin-btn-ghost:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  color: var(--journal-accent);
}

@media (max-width: 767px) {
  .content-pane {
    gap: var(--space-5);
    padding: var(--space-5) var(--space-4) var(--space-6);
  }

  .contest-edit-section {
    padding: var(--space-4-5) var(--space-4);
  }
}
</style>
