import type { Ref } from 'vue'

import { saveChallengeTopology } from '@/api/admin/authoring'
import type { EnvironmentTemplateData } from '@/api/contracts'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useToast } from '@/composables/useToast'

import { createDraftFromTemplate, type TopologyEditorDraft } from './topologyDraft'

interface UseTopologyTemplateApplyOptions {
  challengeId: string
  templateBusy: Ref<boolean>
  applyTopologyDraft: (draft: TopologyEditorDraft) => void
  resetTemplateForm: (template?: EnvironmentTemplateData | null) => void
  onAfterApplied: () => Promise<void>
}

export function useTopologyTemplateApply(options: UseTopologyTemplateApplyOptions) {
  const {
    challengeId,
    templateBusy,
    applyTopologyDraft,
    resetTemplateForm,
    onAfterApplied,
  } = options
  const toast = useToast()

  function loadTemplateIntoDraft(template: EnvironmentTemplateData) {
    applyTopologyDraft(createDraftFromTemplate(template))
    resetTemplateForm(template)
    toast.success('模板已载入编辑器草稿')
  }

  async function handleApplyTemplate(template: EnvironmentTemplateData) {
    const confirmed = await confirmDestructiveAction({
      title: '应用模板',
      message: `确认将模板“${template.name}”应用到当前题目吗？已保存拓扑会被模板覆盖。`,
      confirmButtonText: '确认覆盖',
    })
    if (!confirmed) {
      return
    }
    templateBusy.value = true
    try {
      await saveChallengeTopology(challengeId, { template_id: Number(template.id) })
      toast.success('模板已应用到题目')
      await onAfterApplied()
    } finally {
      templateBusy.value = false
    }
  }

  return {
    loadTemplateIntoDraft,
    handleApplyTemplate,
  }
}
