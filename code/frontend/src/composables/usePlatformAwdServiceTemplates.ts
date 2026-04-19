import { computed, ref, watch } from 'vue'
import { useDebounceFn } from '@vueuse/core'

import {
  createAdminAwdServiceTemplate,
  deleteAdminAwdServiceTemplate,
  listAdminAwdServiceTemplates,
  updateAdminAwdServiceTemplate,
  type AdminAwdServiceTemplateCreatePayload,
  type AdminAwdServiceTemplateUpdatePayload,
} from '@/api/admin'
import { ApiError } from '@/api/request'
import type {
  AdminAwdServiceTemplateData,
  AWDDeploymentMode,
  AWDServiceTemplateStatus,
  AWDServiceType,
  ChallengeDifficulty,
} from '@/api/contracts'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'

type AwdServiceTypeFilter = AWDServiceType | ''
type AwdServiceStatusFilter = AWDServiceTemplateStatus | ''

export interface PlatformAwdServiceTemplateFormDraft {
  name: string
  slug: string
  category: string
  difficulty: ChallengeDifficulty
  description: string
  service_type: AWDServiceType
  deployment_mode: AWDDeploymentMode
  status: AWDServiceTemplateStatus
}

function createEmptyDraft(): PlatformAwdServiceTemplateFormDraft {
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

export function usePlatformAwdServiceTemplates() {
  const toast = useToast()
  const keyword = ref('')
  const serviceTypeFilter = ref<AwdServiceTypeFilter>('')
  const statusFilter = ref<AwdServiceStatusFilter>('')
  const dialogOpen = ref(false)
  const saving = ref(false)
  const editingTemplateId = ref<string | null>(null)
  const formDraft = ref<PlatformAwdServiceTemplateFormDraft>(createEmptyDraft())

  const pagination = usePagination<AdminAwdServiceTemplateData>(({ page, page_size }) =>
    listAdminAwdServiceTemplates({
      page,
      page_size,
      keyword: keyword.value.trim() || undefined,
      service_type: serviceTypeFilter.value || undefined,
      status: statusFilter.value || undefined,
    })
  )

  const dialogMode = computed<'create' | 'edit'>(() => (editingTemplateId.value ? 'edit' : 'create'))

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
    editingTemplateId.value = null
    formDraft.value = createEmptyDraft()
    dialogOpen.value = true
  }

  function openEditDialog(template: AdminAwdServiceTemplateData) {
    editingTemplateId.value = template.id
    formDraft.value = {
      name: template.name,
      slug: template.slug,
      category: template.category,
      difficulty: template.difficulty,
      description: template.description,
      service_type: template.service_type,
      deployment_mode: template.deployment_mode,
      status: template.status,
    }
    dialogOpen.value = true
  }

  function closeDialog() {
    dialogOpen.value = false
  }

  async function saveTemplate(draft: PlatformAwdServiceTemplateFormDraft) {
    saving.value = true

    try {
      if (editingTemplateId.value) {
        const payload: AdminAwdServiceTemplateUpdatePayload = {
          name: draft.name.trim(),
          slug: draft.slug.trim(),
          category: draft.category.trim(),
          difficulty: draft.difficulty,
          description: draft.description.trim(),
          service_type: draft.service_type,
          deployment_mode: draft.deployment_mode,
          status: draft.status,
        }
        await updateAdminAwdServiceTemplate(editingTemplateId.value, payload)
        toast.success('AWD 服务模板已更新')
      } else {
        const payload: AdminAwdServiceTemplateCreatePayload = {
          name: draft.name.trim(),
          slug: draft.slug.trim(),
          category: draft.category.trim(),
          difficulty: draft.difficulty,
          description: draft.description.trim() || undefined,
          service_type: draft.service_type,
          deployment_mode: draft.deployment_mode,
        }
        await createAdminAwdServiceTemplate(payload)
        toast.success('AWD 服务模板已创建')
      }

      dialogOpen.value = false
      await pagination.refresh()
    } catch (error) {
      toast.error(
        humanizeRequestError(error, editingTemplateId.value ? '更新 AWD 服务模板失败' : '创建 AWD 服务模板失败')
      )
    } finally {
      saving.value = false
    }
  }

  async function removeTemplate(template: AdminAwdServiceTemplateData) {
    const confirmed = await confirmDestructiveAction({
      message: `确定要删除模板 ${template.name} 吗？`,
    })
    if (!confirmed) {
      return
    }

    try {
      await deleteAdminAwdServiceTemplate(template.id)
      toast.success(`已删除模板 ${template.name}`)
      await pagination.refresh()
    } catch (error) {
      toast.error(humanizeRequestError(error, '删除 AWD 服务模板失败'))
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
    formDraft,
    openCreateDialog,
    openEditDialog,
    closeDialog,
    saveTemplate,
    removeTemplate,
  }
}
