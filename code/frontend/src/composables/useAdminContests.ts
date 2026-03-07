import { computed, ref, watch } from 'vue'

import {
  createContest,
  getContests,
  updateContest,
  type AdminContestCreatePayload,
  type AdminContestUpdatePayload,
} from '@/api/admin'
import type { ContestDetailData, ContestStatus } from '@/api/contracts'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'

type AdminContestStatus = Extract<ContestStatus, 'draft' | 'registering' | 'running' | 'frozen' | 'ended'>
type StatusFilter = 'all' | AdminContestStatus

export interface ContestFormDraft {
  title: string
  description: string
  mode: 'jeopardy' | 'awd'
  starts_at: string
  ends_at: string
  status: AdminContestStatus
}

interface ContestFieldLocks {
  mode: boolean
  starts_at: boolean
  ends_at: boolean
}

const STATUS_TRANSITIONS: Record<AdminContestStatus, AdminContestStatus[]> = {
  draft: ['draft', 'registering'],
  registering: ['registering', 'draft', 'running'],
  running: ['running', 'frozen', 'ended'],
  frozen: ['frozen', 'ended'],
  ended: ['ended'],
}

function toDateTimeLocal(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return ''
  }
  const offset = date.getTimezoneOffset()
  const localDate = new Date(date.getTime() - offset * 60_000)
  return localDate.toISOString().slice(0, 16)
}

function toISOString(value: string): string {
  return new Date(value).toISOString()
}

function createEmptyDraft(): ContestFormDraft {
  const startAt = new Date()
  startAt.setMinutes(0, 0, 0)
  startAt.setHours(startAt.getHours() + 1)

  const endAt = new Date(startAt)
  endAt.setHours(endAt.getHours() + 2)

  return {
    title: '',
    description: '',
    mode: 'jeopardy',
    starts_at: toDateTimeLocal(startAt.toISOString()),
    ends_at: toDateTimeLocal(endAt.toISOString()),
    status: 'draft',
  }
}

function createDraftFromContest(contest: ContestDetailData): ContestFormDraft {
  return {
    title: contest.title,
    description: contest.description || '',
    mode: contest.mode === 'awd' ? 'awd' : 'jeopardy',
    starts_at: toDateTimeLocal(contest.starts_at),
    ends_at: toDateTimeLocal(contest.ends_at),
    status: normalizeEditableStatus(contest.status),
  }
}

function normalizeEditableStatus(status: ContestStatus): AdminContestStatus {
  if (status === 'draft' || status === 'registering' || status === 'running' || status === 'frozen' || status === 'ended') {
    return status
  }
  return 'draft'
}

function createFieldLocks(status: AdminContestStatus | null): ContestFieldLocks {
  if (!status) {
    return {
      mode: false,
      starts_at: false,
      ends_at: false,
    }
  }

  return {
    mode: status !== 'draft',
    starts_at: status === 'registering' || status === 'running' || status === 'ended',
    ends_at: status === 'running' || status === 'ended',
  }
}

export function useAdminContests() {
  const toast = useToast()
  const statusFilter = ref<StatusFilter>('all')
  const dialogOpen = ref(false)
  const saving = ref(false)
  const editingContestId = ref<string | null>(null)
  const editingBaseStatus = ref<AdminContestStatus | null>(null)
  const formDraft = ref<ContestFormDraft>(createEmptyDraft())

  const pagination = usePagination<ContestDetailData>(({ page, page_size }) =>
    getContests({
      page,
      page_size,
      status: statusFilter.value === 'all' ? undefined : statusFilter.value,
    })
  )

  const mode = computed<'create' | 'edit'>(() => (editingContestId.value ? 'edit' : 'create'))
  const fieldLocks = computed(() => createFieldLocks(editingBaseStatus.value))
  const statusOptions = computed(() => {
    if (!editingBaseStatus.value) {
      return []
    }
    return STATUS_TRANSITIONS[editingBaseStatus.value].map((status) => ({ label: status, value: status }))
  })

  watch(statusFilter, async () => {
    await pagination.changePage(1)
  })

  function openCreateDialog() {
    editingContestId.value = null
    editingBaseStatus.value = null
    formDraft.value = createEmptyDraft()
    dialogOpen.value = true
  }

  function openEditDialog(contest: ContestDetailData) {
    editingContestId.value = contest.id
    editingBaseStatus.value = normalizeEditableStatus(contest.status)
    formDraft.value = createDraftFromContest(contest)
    dialogOpen.value = true
  }

  function closeDialog() {
    dialogOpen.value = false
  }

  async function saveContest(draft: ContestFormDraft) {
    const title = draft.title.trim()
    const description = draft.description.trim()

    if (!title) {
      toast.error('请填写竞赛标题')
      return
    }

    saving.value = true
    try {
      if (editingContestId.value) {
        const payload: AdminContestUpdatePayload = {
          title,
          description,
          mode: draft.mode,
          starts_at: toISOString(draft.starts_at),
          ends_at: toISOString(draft.ends_at),
          status: draft.status,
        }
        await updateContest(editingContestId.value, payload)
        toast.success('竞赛已更新')
      } else {
        const payload: AdminContestCreatePayload = {
          title,
          description,
          mode: draft.mode,
          starts_at: toISOString(draft.starts_at),
          ends_at: toISOString(draft.ends_at),
        }
        await createContest(payload)
        toast.success('竞赛已创建')
      }

      dialogOpen.value = false
      await pagination.refresh()
    } finally {
      saving.value = false
    }
  }

  return {
    ...pagination,
    statusFilter,
    dialogOpen,
    mode,
    saving,
    formDraft,
    fieldLocks,
    statusOptions,
    openCreateDialog,
    openEditDialog,
    closeDialog,
    saveContest,
  }
}
