import type { ComputedRef, Ref } from 'vue'

import {
  deleteChallengeTopology,
  exportChallengePackage,
  saveChallengeTopology,
} from '@/api/admin/authoring'
import type { ChallengeTopologyData } from '@/api/contracts'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useToast } from '@/composables/useToast'

import {
  createDraftFromTopology,
  serializeTopologyDraft,
  type TopologyEditorDraft,
} from './topologyDraft'

interface UseTopologyPersistenceActionsOptions {
  challengeId: string
  isTemplateLibraryMode: ComputedRef<boolean>
  draft: Ref<TopologyEditorDraft>
  topology: Ref<ChallengeTopologyData | null>
  saving: Ref<boolean>
  exporting: Ref<boolean>
  applyTopologyDraft: (draft: TopologyEditorDraft) => void
  applyEmptyTemplateDraft: () => void
  loadTemplates: () => Promise<void>
  reloadAll: () => Promise<void>
}

export function useTopologyPersistenceActions(options: UseTopologyPersistenceActionsOptions) {
  const {
    challengeId,
    isTemplateLibraryMode,
    draft,
    topology,
    saving,
    exporting,
    applyTopologyDraft,
    applyEmptyTemplateDraft,
    loadTemplates,
    reloadAll,
  } = options
  const toast = useToast()

  async function handleSaveTopology() {
    saving.value = true
    try {
      const saved = await saveChallengeTopology(challengeId, serializeTopologyDraft(draft.value))
      topology.value = saved
      applyTopologyDraft(createDraftFromTopology(saved))
      toast.success('题目拓扑已保存')
      await loadTemplates()
    } finally {
      saving.value = false
    }
  }

  async function handleExportPackage() {
    if (isTemplateLibraryMode.value) {
      return
    }
    exporting.value = true
    try {
      const exported = await exportChallengePackage(challengeId)
      toast.success('题目包已导出')
      await reloadAll()
      if (typeof window !== 'undefined' && exported.download_url) {
        window.open(exported.download_url, '_blank', 'noopener')
      }
    } finally {
      exporting.value = false
    }
  }

  async function handleDeleteTopology() {
    if (!topology.value) {
      toast.warning('当前题目还没有已保存的拓扑')
      return
    }
    const confirmed = await confirmDestructiveAction({
      title: '删除题目拓扑',
      message: '确认删除当前题目已保存的拓扑吗？删除后需要重新保存才能恢复。',
      confirmButtonText: '确认删除',
    })
    if (!confirmed) {
      return
    }
    saving.value = true
    try {
      await deleteChallengeTopology(challengeId)
      topology.value = null
      applyEmptyTemplateDraft()
      toast.success('题目拓扑已删除')
    } catch (error) {
      const message =
        error instanceof Error && error.message.trim() ? error.message : '删除题目拓扑失败'
      toast.error(message)
    } finally {
      saving.value = false
    }
  }

  return {
    handleSaveTopology,
    handleExportPackage,
    handleDeleteTopology,
  }
}
