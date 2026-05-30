import { computed, ref, watch } from 'vue'
import { useDebounceFn } from '@vueuse/core'

import {
  createAdminAwdChallenge,
  deleteAdminAwdChallenge,
  listAdminAwdChallenges,
  updateAdminAwdChallenge,
  type AdminAwdChallengeCreatePayload,
  type AdminAwdChallengeUpdatePayload,
} from '@/api/admin/awd-authoring'
import { ApiError } from '@/api/request'
import type {
  AdminAwdChallengeData,
  AWDDeploymentMode,
  AWDChallengeStatus,
  AWDServiceType,
  ChallengeCategory,
  ChallengeDifficulty,
} from '@/api/contracts'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'
import { useAwdChallengeImportFlow } from './useAwdChallengeImportFlow'

type AwdServiceTypeFilter = AWDServiceType | ''
type AwdServiceStatusFilter = AWDChallengeStatus | ''

export interface PlatformAwdChallengeFormDraft {
  name: string
  slug: string
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
  description: string
  service_type: AWDServiceType
  deployment_mode: AWDDeploymentMode
  status: AWDChallengeStatus
}

export interface PlatformAwdChallengeImportUploadResult {
  id: string
  status: 'success' | 'error'
  fileName: string
  message: string
  createdAt: string
  code?: number
  requestId?: string
}

function createEmptyDraft(): PlatformAwdChallengeFormDraft {
  return {
    name: '',
    slug: '',
    category: 'web',
    difficulty: 'medium',
    description: '',
    service_type: 'web_http',
    deployment_mode: 'single_container',
    status: 'draft',
  }
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

export function usePlatformAwdChallenges() {
  const toast = useToast()
  const keyword = ref('')
  const serviceTypeFilter = ref<AwdServiceTypeFilter>('')
  const statusFilter = ref<AwdServiceStatusFilter>('')
  const dialogOpen = ref(false)
  const saving = ref(false)
  const editingChallengeId = ref<string | null>(null)
  const formDraft = ref<PlatformAwdChallengeFormDraft>(createEmptyDraft())

  const pagination = usePagination<AdminAwdChallengeData>(({ page, page_size }) =>
    listAdminAwdChallenges({
      page,
      page_size,
      keyword: keyword.value.trim() || undefined,
      service_type: serviceTypeFilter.value || undefined,
      status: statusFilter.value || undefined,
    })
  )

  const dialogMode = computed<'create' | 'edit'>(() => (editingChallengeId.value ? 'edit' : 'create'))

  type DebouncedRefresh = ReturnType<typeof useDebounceFn> & {
    cancel?: () => void
  }
  const scheduleKeywordRefresh = useDebounceFn(() => {
    void pagination.changePage(1)
  }, 250) as DebouncedRefresh

  watch(keyword, () => {
    scheduleKeywordRefresh()
  })

  watch([serviceTypeFilter, statusFilter], async () => {
    scheduleKeywordRefresh.cancel?.()
    await pagination.changePage(1)
  })

  function openCreateDialog() {
    editingChallengeId.value = null
    formDraft.value = createEmptyDraft()
    dialogOpen.value = true
  }

  function openEditDialog(challenge: AdminAwdChallengeData) {
    editingChallengeId.value = challenge.id
    formDraft.value = {
      name: challenge.name,
      slug: challenge.slug,
      category: challenge.category,
      difficulty: challenge.difficulty,
      description: challenge.description,
      service_type: challenge.service_type,
      deployment_mode: challenge.deployment_mode,
      status: challenge.status,
    }
    dialogOpen.value = true
  }

  function closeDialog() {
    dialogOpen.value = false
  }

  const {
    uploading,
    queueLoading,
    selectedImportFileName,
    importQueue,
    uploadResults,
    refreshImportQueue,
    selectImportPackages,
    commitImportPreview,
  } = useAwdChallengeImportFlow({
    refreshChallenges: pagination.refresh,
    humanizeRequestError,
    notifySuccess: (message) => {
      toast.success(message)
    },
    notifyError: (message) => {
      toast.error(message)
    },
  })

  async function saveChallenge(draft: PlatformAwdChallengeFormDraft) {
    saving.value = true

    try {
      if (editingChallengeId.value) {
        const payload: AdminAwdChallengeUpdatePayload = {
          name: draft.name.trim(),
          slug: draft.slug.trim(),
          category: draft.category,
          difficulty: draft.difficulty,
          description: draft.description.trim(),
          service_type: draft.service_type,
          deployment_mode: draft.deployment_mode,
          status: draft.status,
        }
        await updateAdminAwdChallenge(editingChallengeId.value, payload)
        toast.success('AWD 题目已更新')
      } else {
        const payload: AdminAwdChallengeCreatePayload = {
          name: draft.name.trim(),
          slug: draft.slug.trim(),
          category: draft.category,
          difficulty: draft.difficulty,
          description: draft.description.trim() || undefined,
          service_type: draft.service_type,
          deployment_mode: draft.deployment_mode,
        }
        await createAdminAwdChallenge(payload)
        toast.success('AWD 题目已创建')
      }

      dialogOpen.value = false
      await pagination.refresh()
    } catch (error) {
      toast.error(
        humanizeRequestError(error, editingChallengeId.value ? '更新 AWD 题目失败' : '创建 AWD 题目失败')
      )
    } finally {
      saving.value = false
    }
  }

  async function removeChallenge(challenge: AdminAwdChallengeData) {
    const confirmed = await confirmDestructiveAction({
      message: `确定要删除 AWD 题目 ${challenge.name} 吗？`,
    })
    if (!confirmed) {
      return
    }

    try {
      await deleteAdminAwdChallenge(challenge.id)
      toast.success(`已删除 AWD 题目 ${challenge.name}`)
      await pagination.refresh()
    } catch (error) {
      toast.error(humanizeRequestError(error, '删除 AWD 题目失败'))
    }
  }

  return {
    ...pagination,
    keyword,
    serviceTypeFilter,
    statusFilter,
    dialogOpen,
    dialogMode,
    saving,
    uploading,
    queueLoading,
    selectedImportFileName,
    importQueue,
    uploadResults,
    formDraft,
    openCreateDialog,
    openEditDialog,
    closeDialog,
    refreshImportQueue,
    selectImportPackages,
    commitImportPreview,
    saveChallenge,
    removeChallenge,
  }
}
