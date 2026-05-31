<script setup lang="ts">
import { GitBranch } from 'lucide-vue-next'

import SectionCard from '@/shared/ui/common/SectionCard.vue'

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

const props = defineProps<{
  packageSourceSummary: PackageSourceSummary
  packageBaselineSummary: PackageBaselineSummary | null
  packageFiles: PackageFile[]
  packageRevisionHistory: PackageRevision[]
  exporting: boolean
}>()

const emit = defineEmits<{
  exportPackage: []
}>()

function formatFileSize(size: number): string {
  if (size < 1024) {
    return `${size} B`
  }
  if (size < 1024 * 1024) {
    return `${(size / 1024).toFixed(1)} KB`
  }
  return `${(size / (1024 * 1024)).toFixed(1)} MB`
}
</script>

<template>
  <div class="topology-context-stack">
    <SectionCard title="题包来源" subtitle="当前拓扑是否来自题包导入，以及是否已经偏离基线。">
      <div class="space-y-3">
        <div class="rounded-2xl border border-border bg-elevated px-4 py-4">
          <div class="text-xs uppercase tracking-[0.18em] text-text-muted">
            当前状态
          </div>
          <div class="mt-2 text-base font-semibold text-text-primary">
            {{ props.packageSourceSummary.title }}
          </div>
          <p class="mt-2 text-sm leading-6 text-text-secondary">
            {{ props.packageSourceSummary.subtitle }}
          </p>
        </div>

        <div v-if="props.packageBaselineSummary" class="grid gap-3 sm:grid-cols-2">
          <div class="rounded-2xl border border-border bg-surface px-4 py-3">
            <div class="text-xs uppercase tracking-[0.18em] text-text-muted">
              基线入口
            </div>
            <div class="mt-2 text-sm font-semibold text-text-primary">
              {{ props.packageBaselineSummary.entryNodeKey }}
            </div>
          </div>
          <div class="rounded-2xl border border-border bg-surface px-4 py-3">
            <div class="text-xs uppercase tracking-[0.18em] text-text-muted">
              基线规模
            </div>
            <div class="mt-2 text-sm font-semibold text-text-primary">
              {{ props.packageBaselineSummary.nodeCount }} 节点 /
              {{ props.packageBaselineSummary.networkCount }} 网络
            </div>
          </div>
        </div>

        <button
          type="button"
          class="ui-btn ui-btn--ghost topology-action-btn"
          :disabled="props.exporting || !props.packageSourceSummary.canExport"
          @click="emit('exportPackage')"
        >
          <GitBranch class="h-4 w-4" />
          {{ props.exporting ? '正在生成导出包...' : '导出并下载完整题目包' }}
        </button>
      </div>
    </SectionCard>

    <SectionCard title="题包文件" subtitle="当前基线修订中保留下来的源码、Dockerfile 和拓扑描述文件。">
      <div v-if="props.packageFiles.length" class="space-y-2">
        <div
          v-for="file in props.packageFiles.slice(0, 10)"
          :key="file.path"
          class="rounded-2xl border border-border bg-elevated px-4 py-3"
        >
          <div class="text-sm font-medium text-text-primary">
            {{ file.path }}
          </div>
          <div class="mt-1 text-xs text-text-muted">
            {{ formatFileSize(file.size) }}
          </div>
        </div>
        <div v-if="props.packageFiles.length > 10" class="text-xs text-text-muted">
          其余 {{ props.packageFiles.length - 10 }} 个文件已省略，导出包会完整保留。
        </div>
      </div>
      <div
        v-else
        class="rounded-2xl border border-dashed border-border px-4 py-6 text-sm text-text-muted"
      >
        当前题目还没有可展示的题包文件清单。
      </div>
    </SectionCard>

    <SectionCard
      title="修订历史"
      subtitle="导入和导出都会生成题包修订，导出后会把当前拓扑设为新的干净基线。"
    >
      <div v-if="props.packageRevisionHistory.length" class="space-y-2">
        <div
          v-for="revision in props.packageRevisionHistory.slice(0, 6)"
          :key="revision.id"
          class="rounded-2xl border border-border bg-elevated px-4 py-3"
        >
          <div class="flex items-center justify-between gap-3">
            <div class="text-sm font-medium text-text-primary">
              r{{ revision.revision_no }} ·
              {{ revision.source_type === 'exported' ? '导出' : '导入' }}
            </div>
            <div class="text-xs text-text-muted">
              {{ revision.created_at }}
            </div>
          </div>
          <div class="mt-2 text-xs leading-6 text-text-secondary">
            {{ revision.package_slug || '未记录 slug' }}
            <span v-if="revision.topology_source_path">
              · {{ revision.topology_source_path }}
            </span>
          </div>
        </div>
      </div>
      <div
        v-else
        class="rounded-2xl border border-dashed border-border px-4 py-6 text-sm text-text-muted"
      >
        当前题目还没有题包修订历史。
      </div>
    </SectionCard>
  </div>
</template>
