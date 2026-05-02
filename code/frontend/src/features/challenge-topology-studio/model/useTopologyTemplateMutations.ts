import { ref, type Ref } from 'vue'

import {
  createEnvironmentTemplate,
  deleteEnvironmentTemplate,
  updateEnvironmentTemplate,
} from '@/api/admin/authoring'
import type { EnvironmentTemplateData } from '@/api/contracts'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useToast } from '@/composables/useToast'

import {
  serializeEnvironmentTemplateDraft,
  type TopologyEditorDraft,
} from './topologyDraft'

interface UseTopologyTemplateMutationsOptions {
  canSaveTemplate: Readonly<Ref<boolean>>
  templateName: Ref<string>
  templateDescription: Ref<string>
  selectedTemplateId: Readonly<Ref<string | null>>
  templates: Readonly<Ref<EnvironmentTemplateData[]>>
  draft: Readonly<Ref<TopologyEditorDraft>>
  isTemplateLibraryMode: Readonly<Ref<boolean>>
  applyEmptyTemplateDraft: () => void
  resetTemplateForm: (template?: EnvironmentTemplateData | null) => void
  loadTemplates: () => Promise<void>
}

export function useTopologyTemplateMutations(options: UseTopologyTemplateMutationsOptions) {
  const {
    canSaveTemplate,
    templateName,
    templateDescription,
    selectedTemplateId,
    templates,
    draft,
    isTemplateLibraryMode,
    applyEmptyTemplateDraft,
    resetTemplateForm,
    loadTemplates,
  } = options
  const toast = useToast()
  const templateBusy = ref(false)

  async function handleCreateTemplate() {
    if (!canSaveTemplate.value) {
      toast.warning('请先填写模板名称')
      return
    }

    templateBusy.value = true
    try {
      const created = await createEnvironmentTemplate(
        serializeEnvironmentTemplateDraft(
          templateName.value,
          templateDescription.value,
          draft.value
        )
      )
      resetTemplateForm(created)
      toast.success('模板已创建')
      await loadTemplates()
    } finally {
      templateBusy.value = false
    }
  }

  async function handleUpdateTemplate() {
    if (!selectedTemplateId.value) {
      toast.warning('请先选择模板')
      return
    }
    if (!canSaveTemplate.value) {
      toast.warning('请先填写模板名称')
      return
    }

    templateBusy.value = true
    try {
      const updated = await updateEnvironmentTemplate(
        selectedTemplateId.value,
        serializeEnvironmentTemplateDraft(
          templateName.value,
          templateDescription.value,
          draft.value
        )
      )
      resetTemplateForm(updated)
      toast.success('模板已更新')
      await loadTemplates()
    } finally {
      templateBusy.value = false
    }
  }

  async function handleDeleteTemplate(templateId: string) {
    const template = templates.value.find((item) => item.id === templateId)
    const confirmed = await confirmDestructiveAction({
      title: '删除环境模板',
      message: `确认删除模板“${template?.name || templateId}”吗？该操作不可撤销。`,
      confirmButtonText: '删除模板',
      cancelButtonText: '取消',
    })
    if (!confirmed) {
      return
    }
    templateBusy.value = true
    try {
      await deleteEnvironmentTemplate(templateId)
      if (selectedTemplateId.value === templateId) {
        if (isTemplateLibraryMode.value) {
          applyEmptyTemplateDraft()
        }
        resetTemplateForm(null)
      }
      toast.success('模板已删除')
      await loadTemplates()
    } catch (error) {
      const message =
        error instanceof Error && error.message.trim() ? error.message : '删除模板失败'
      toast.error(message)
    } finally {
      templateBusy.value = false
    }
  }

  return {
    templateBusy,
    handleCreateTemplate,
    handleUpdateTemplate,
    handleDeleteTemplate,
  }
}
