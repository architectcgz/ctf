import { computed, ref, type Ref } from 'vue'

import type { EnvironmentTemplateData } from '@/api/contracts'

interface UseTopologyTemplateSelectionOptions {
  templates: Ref<EnvironmentTemplateData[]>
}

export function useTopologyTemplateSelection(options: UseTopologyTemplateSelectionOptions) {
  const { templates } = options

  const templateKeyword = ref('')
  const selectedTemplateId = ref<string | null>(null)
  const templateName = ref('')
  const templateDescription = ref('')

  const canSaveTemplate = computed(() => templateName.value.trim().length > 0)
  const selectedTemplate = computed(
    () => templates.value.find((item) => item.id === selectedTemplateId.value) || null
  )
  const selectedTemplateSummary = computed(() => {
    if (!selectedTemplate.value) {
      return '尚未选中模板，可从下方模板库载入到当前草稿。'
    }
    return `${selectedTemplate.value.name} · 节点 ${selectedTemplate.value.nodes.length} · 网络 ${selectedTemplate.value.networks?.length || 0} · 使用 ${selectedTemplate.value.usage_count}`
  })

  function resetTemplateForm(template?: EnvironmentTemplateData | null) {
    selectedTemplateId.value = template?.id || null
    templateName.value = template?.name || ''
    templateDescription.value = template?.description || ''
  }

  function clearTemplateSelection() {
    resetTemplateForm(null)
  }

  function reconcileTemplateSelection() {
    if (
      selectedTemplateId.value &&
      !templates.value.some((item) => item.id === selectedTemplateId.value)
    ) {
      resetTemplateForm(null)
    }
  }

  return {
    templateKeyword,
    selectedTemplateId,
    templateName,
    templateDescription,
    canSaveTemplate,
    selectedTemplate,
    selectedTemplateSummary,
    resetTemplateForm,
    clearTemplateSelection,
    reconcileTemplateSelection,
  }
}
