<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import ChallengePackageImportReview from '@/components/admin/challenge/ChallengePackageImportReview.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useChallengePackageImport } from '@/composables/useChallengePackageImport'

const route = useRoute()
const router = useRouter()

const importId = computed(() => {
  const raw = route.params.importId
  if (typeof raw === 'string') {
    return raw
  }
  if (Array.isArray(raw)) {
    return raw[0] ?? ''
  }
  return ''
})

const { preview, uploading, committing, hasPreview, loadPreview, resetPreview, commitPreview } =
  useChallengePackageImport({
    onCommitted: async () => {
      await router.push({ name: 'ChallengeManage', query: { panel: 'manage' } })
    },
  })

async function syncPreviewByRoute(): Promise<void> {
  resetPreview()
  const id = importId.value.trim()
  if (!id) {
    return
  }
  await loadPreview(id)
}

async function backToImportPanel(): Promise<void> {
  await router.push({ name: 'ChallengeManage', query: { panel: 'import' } })
}

async function backToQueuePanel(): Promise<void> {
  await router.push({ name: 'ChallengeManage', query: { panel: 'queue' } })
}

async function handleCommitPreview(): Promise<void> {
  await commitPreview()
}

onMounted(() => {
  void syncPreviewByRoute()
})

watch(importId, () => {
  void syncPreviewByRoute()
})
</script>

<template>
  <section
    class="journal-shell journal-shell-admin journal-notes-card journal-hero flex min-h-full flex-1 flex-col rounded-[24px] border px-6 py-6 md:px-8"
  >
    <header class="workspace-topbar">
      <div class="topbar-leading">
        <span class="workspace-overline">Challenge Workspace</span>
        <span class="class-chip">导入预览</span>
      </div>
      <button class="nav-back" type="button" @click="void backToImportPanel()">
        返回导入题目包
      </button>
    </header>

    <PageHeader
      class="import-preview-page-header"
      eyebrow="Import Preview"
      title="导入预览"
      description="请仔细检查题目，确认题头、运行信息和提示内容后再正式导入题库。"
    />

    <div v-if="uploading" class="import-preview-loading flex items-center justify-center py-12">
      <div
        class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]"
      />
    </div>

    <AppEmpty
      v-else-if="!hasPreview || !preview"
      class="import-preview-empty"
      icon="Flag"
      title="未找到导入预览"
      description="该导入任务可能已失效或已被清理，请返回导入页重新上传题目包。"
    >
      <template #action>
        <div class="import-preview-empty__actions">
          <button class="nav-back" type="button" @click="void backToImportPanel()">
            返回导入页
          </button>
          <button class="nav-back" type="button" @click="void backToQueuePanel()">
            查看待确认导入
          </button>
        </div>
      </template>
    </AppEmpty>

    <ChallengePackageImportReview
      v-else
      :preview="preview"
      :committing="committing"
      @confirm="void handleCommitPreview()"
      @reset="void backToImportPanel()"
    />
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-topbar-padding-bottom: var(--space-3);
  --journal-shell-hero-radial-strength: 7%;
  --journal-shell-hero-radial-size: 22rem;
  --journal-shell-hero-end: var(--journal-surface);
  --journal-shell-hero-shadow: 0 22px 50px var(--color-shadow-soft);
}

.workspace-overline {
  font-size: var(--font-size-0-70);
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.nav-back {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.35rem;
  border-radius: 0.75rem;
  border: 1px solid var(--journal-border);
  padding: var(--space-2) var(--space-3-5);
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  color: var(--journal-ink);
  font-size: var(--font-size-0-88);
  font-weight: 600;
}

.import-preview-empty {
  margin-top: var(--space-4);
}

.import-preview-empty__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}
</style>
