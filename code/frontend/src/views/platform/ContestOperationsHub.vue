<script setup lang="ts">
import ContestOperationsHubHeroPanel from '@/components/platform/contest/ContestOperationsHubHeroPanel.vue'
import ContestOperationsHubWorkspacePanel from '@/components/platform/contest/ContestOperationsHubWorkspacePanel.vue'
import { useContestOperationsHubPage } from '@/features/platform-contests'

const {
  loading,
  loadError,
  operableContests,
  runningContestCount,
  frozenContestCount,
  preferredContest,
  loadContests,
  handleEnterOperations,
  handleBackToContestDirectory,
} = useContestOperationsHubPage()
</script>

<template>
  <section
    class="journal-shell journal-shell-admin journal-notes-card journal-hero workspace-shell flex min-h-full flex-1 flex-col"
  >
    <div class="workspace-grid">
      <main class="content-pane contest-ops-content">
        <ContestOperationsHubHeroPanel
          :operable-contest-count="operableContests.length"
          :running-contest-count="runningContestCount"
          :frozen-contest-count="frozenContestCount"
          :preferred-contest-title="preferredContest ? preferredContest.title : '暂无'"
          @back="void handleBackToContestDirectory()"
        />

        <ContestOperationsHubWorkspacePanel
          :loading="loading"
          :load-error="loadError"
          :operable-contests="operableContests"
          @retry="void loadContests()"
          @back="void handleBackToContestDirectory()"
          @enter-operations="void handleEnterOperations($event)"
        />
      </main>
    </div>
  </section>
</template>

<style scoped>
.contest-ops-content {
  display: flex;
  flex-direction: column;
  gap: var(--workspace-directory-page-block-gap, var(--space-5));
}
</style>
