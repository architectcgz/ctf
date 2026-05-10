<script setup lang="ts">
import ChallengeImportHeroPanel from '@/components/platform/challenge/ChallengeImportHeroPanel.vue'
import ChallengeImportQueuePanel from '@/components/platform/challenge/ChallengeImportQueuePanel.vue'
import ChallengePackageImportEntry from '@/components/platform/challenge/ChallengePackageImportEntry.vue'
import ChallengeImportUploadResultsPanel from '@/components/platform/challenge/ChallengeImportUploadResultsPanel.vue'
import { useChallengeImportManagePage } from '@/features/challenge-package-import'

const {
  uploading,
  queueLoading,
  selectedFileName,
  queue,
  uploadResults,
  queueCount,
  handleSelectPackage,
  openPackageFormatGuide,
  backToChallenges,
  inspectImportTask,
  formatDateTime,
} = useChallengeImportManagePage()
</script>

<template>
  <div class="workspace-shell challenge-import-shell journal-shell journal-shell-admin journal-notes-card">
    <main class="content-pane challenge-import-content">
        <section class="challenge-import-panel">
          <ChallengeImportHeroPanel
            @back="void backToChallenges()"
            @open-guide="void openPackageFormatGuide()"
          />

          <section
            id="challenge-import-workspace"
            class="workspace-directory-section challenge-import-directory challenge-workspace-section"
          >
            <header class="list-heading challenge-section-heading">
              <div>
                <div class="workspace-overline">
                  Challenge Package
                </div>
                <h2 class="list-heading__title">
                  导入题目包
                </h2>
                <p class="challenge-section-copy">
                  上传压缩包后先进入预览，再确认是否写入题库。格式规范与示例已收敛到当前导入页。
                </p>
              </div>
              <div class="challenge-directory-meta">
                共 {{ queueCount }} 个待处理任务
              </div>
            </header>

            <ChallengePackageImportEntry
              :hide-header="true"
              :uploading="uploading"
              :selected-file-name="selectedFileName"
              @select="handleSelectPackage"
            />
          </section>

          <ChallengeImportUploadResultsPanel
            :upload-results="uploadResults"
            :format-date-time="formatDateTime"
          />

          <ChallengeImportQueuePanel
            :queue-loading="queueLoading"
            :queue-count="queueCount"
            :queue="queue"
            :format-date-time="formatDateTime"
            @inspect="inspectImportTask"
          />
        </section>
    </main>
  </div>
</template>

<style scoped>
.challenge-import-shell {
  --workspace-brand: var(--journal-accent);
  --workspace-brand-ink: color-mix(in srgb, var(--journal-accent) 74%, var(--journal-ink));
  --challenge-page-bg: color-mix(in srgb, var(--journal-surface-subtle) 90%, var(--color-bg-base));
  --challenge-page-surface: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-surface));
  --challenge-page-surface-subtle: color-mix(
    in srgb,
    var(--color-bg-elevated) 82%,
    var(--color-bg-surface)
  );
  --challenge-page-surface-elevated: color-mix(
    in srgb,
    var(--color-bg-elevated) 90%,
    var(--color-bg-surface)
  );
  --challenge-page-line: color-mix(in srgb, var(--journal-border) 84%, transparent);
  --challenge-page-line-strong: color-mix(in srgb, var(--journal-border) 92%, transparent);
  --challenge-page-text: color-mix(in srgb, var(--journal-ink) 94%, transparent);
  --challenge-page-muted: color-mix(in srgb, var(--journal-muted) 92%, transparent);
  --challenge-page-faint: color-mix(in srgb, var(--color-text-muted) 90%, transparent);
  --challenge-page-accent: color-mix(in srgb, var(--workspace-brand) 88%, var(--challenge-page-text));
  border: none;
  background: var(--challenge-page-bg);
}

.challenge-import-content {
  display: grid;
  gap: 2rem;
  background: transparent;
}

.challenge-import-panel {
  display: grid;
  gap: 2rem;
  min-width: 0;
}

.challenge-workspace-section {
  scroll-margin-top: 6rem;
}

.challenge-import-directory {
  display: grid;
  gap: 1.5rem;
}

.challenge-section-heading {
  align-items: flex-start;
  gap: 1rem;
}

.challenge-section-copy {
  margin: 0.5rem 0 0;
  max-width: 44rem;
  font-size: 0.92rem;
  line-height: 1.6;
  color: var(--challenge-page-muted);
}

.challenge-directory-state,
.challenge-directory-meta {
  color: var(--challenge-page-muted);
}

</style>
