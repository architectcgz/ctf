<script setup lang="ts">
import type { EnvironmentTemplateData } from '@/api/contracts'
import TopologyPackageContextPanel from './TopologyPackageContextPanel.vue'
import TopologyStatusNotes from './TopologyStatusNotes.vue'
import TopologyTemplateSidePanel from './TopologyTemplateSidePanel.vue'

type TopologyStatusCard = {
  eyebrow: string
  title: string
  subtitle: string
}

type PackageSourceSummary = {
  title: string
  subtitle: string
  canExport: boolean
}

type PackageBaselineSummary = {
  entryNodeKey: string
  networkCount: number
  nodeCount: number
}

type PackageFile = {
  path: string
  size: number
}

type PackageRevision = {
  id: string
  revision_no: number
  source_type: 'imported' | 'exported'
  package_slug?: string
  topology_source_path?: string
  created_at: string
}

defineProps<{
  statusCard: TopologyStatusCard
  secondaryCard: TopologyStatusCard
  packageSourceSummary: PackageSourceSummary
  packageBaselineSummary: PackageBaselineSummary | null
  packageFiles: PackageFile[]
  packageRevisionHistory: PackageRevision[]
  exporting: boolean
  selectedTemplateSummary: string
  selectedTemplateId: string | null
  templates: EnvironmentTemplateData[]
  templateKeyword: string
  templateName: string
  templateDescription: string
  templateBusy: boolean
}>()

const emit = defineEmits<{
  exportPackage: []
  'update:templateKeyword': [value: string]
  'update:templateName': [value: string]
  'update:templateDescription': [value: string]
  loadTemplate: [template: EnvironmentTemplateData]
  clearTemplateSelection: []
  searchTemplates: []
  resetTemplateForm: [template: EnvironmentTemplateData]
  applyTemplate: [template: EnvironmentTemplateData]
  deleteTemplate: [templateId: string]
  resetTemplateEditor: []
  createTemplate: []
  updateTemplate: []
}>()
</script>

<template>
  <aside class="context-rail topology-context-rail">
    <div class="topology-context-stack">
      <TopologyStatusNotes
        mode="challenge"
        :status-card="statusCard"
        :secondary-card="secondaryCard"
      />

      <TopologyPackageContextPanel
        :package-source-summary="packageSourceSummary"
        :package-baseline-summary="packageBaselineSummary"
        :package-files="packageFiles"
        :package-revision-history="packageRevisionHistory"
        :exporting="exporting"
        @export-package="emit('exportPackage')"
      />

      <TopologyTemplateSidePanel
        :template-keyword="templateKeyword"
        :template-name="templateName"
        :template-description="templateDescription"
        :is-template-library-mode="false"
        :selected-template-summary="selectedTemplateSummary"
        :selected-template-id="selectedTemplateId"
        :templates="templates"
        :template-busy="templateBusy"
        @update:template-keyword="emit('update:templateKeyword', $event)"
        @update:template-name="emit('update:templateName', $event)"
        @update:template-description="emit('update:templateDescription', $event)"
        @load-template="emit('loadTemplate', $event)"
        @clear-template-selection="emit('clearTemplateSelection')"
        @search-templates="emit('searchTemplates')"
        @reset-template-form="emit('resetTemplateForm', $event)"
        @apply-template="emit('applyTemplate', $event)"
        @delete-template="emit('deleteTemplate', $event)"
        @reset-template-editor="emit('resetTemplateEditor')"
        @create-template="emit('createTemplate')"
        @update-template="emit('updateTemplate')"
      />
    </div>
  </aside>
</template>
