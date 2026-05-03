import type { ComputedRef, Ref } from 'vue'

import {
  getChallengeDetail,
  getChallengeTopology,
  getEnvironmentTemplates,
  getImages,
} from '@/api/admin/authoring'
import type {
  AdminChallengeListItem,
  AdminImageListItem,
  ChallengeTopologyData,
  EnvironmentTemplateData,
} from '@/api/contracts'

import { createDraftFromTopology, type TopologyEditorDraft } from './topologyDraft'

interface UseTopologyDataLoaderOptions {
  challengeId: string
  loading: Ref<boolean>
  isTemplateLibraryMode: ComputedRef<boolean>
  selectedTemplateId: Ref<string | null>
  templateKeyword: Ref<string>
  templates: Ref<EnvironmentTemplateData[]>
  challenge: Ref<AdminChallengeListItem | null>
  topology: Ref<ChallengeTopologyData | null>
  images: Ref<AdminImageListItem[]>
  applyTopologyDraft: (next: TopologyEditorDraft) => void
  applyEmptyTemplateDraft: () => void
  resetTemplateForm: (template?: EnvironmentTemplateData | null) => void
  reconcileTemplateSelection: () => void
}

export function useTopologyDataLoader(options: UseTopologyDataLoaderOptions) {
  const {
    challengeId,
    loading,
    isTemplateLibraryMode,
    selectedTemplateId,
    templateKeyword,
    templates,
    challenge,
    topology,
    images,
    applyTopologyDraft,
    applyEmptyTemplateDraft,
    resetTemplateForm,
    reconcileTemplateSelection,
  } = options

  async function loadTemplates() {
    templates.value = await getEnvironmentTemplates(templateKeyword.value.trim() || undefined)
    reconcileTemplateSelection()
  }

  async function loadPageData() {
    loading.value = true
    try {
      if (isTemplateLibraryMode.value) {
        const imageResult = await getImages({ page: 1, page_size: 200 })
        challenge.value = null
        topology.value = null
        images.value = imageResult.list
        if (!selectedTemplateId.value) {
          applyEmptyTemplateDraft()
          resetTemplateForm(null)
        }
        return
      }

      const [challengeDetail, imageResult, currentTopology] = await Promise.all([
        getChallengeDetail(challengeId),
        getImages({ page: 1, page_size: 200 }),
        getChallengeTopology(challengeId),
      ])

      challenge.value = challengeDetail
      images.value = imageResult.list
      topology.value = currentTopology
      applyTopologyDraft(createDraftFromTopology(currentTopology))
      resetTemplateForm(
        currentTopology?.template_id
          ? templates.value.find((item) => item.id === currentTopology.template_id) || null
          : null
      )
    } finally {
      loading.value = false
    }
  }

  async function reloadAll() {
    await loadTemplates()
    await loadPageData()
  }

  return {
    loadTemplates,
    loadPageData,
    reloadAll,
  }
}
