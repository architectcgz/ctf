<script setup lang="ts">
import {
  ContestOperationsHubHeroPanel,
  ContestOperationsHubWorkspacePanel,
  useContestOperationsHubPage,
} from '@/features/platform/contests'

const {
  changeContestPage,
  loading,
  loadError,
  operableContests,
  page,
  total,
  totalPages,
  runningContestCount,
  frozenContestCount,
  preferredContest,
  loadContests,
  buildContestOperationsRoute,
  backToContestDirectoryRoute,
} = useContestOperationsHubPage()
</script>

<template>
  <section
    class="journal-shell journal-shell-admin journal-notes-card journal-hero workspace-shell flex min-h-full flex-1 flex-col"
  >
    <main class="content-pane contest-ops-content">
        <ContestOperationsHubHeroPanel
          :operable-contest-count="total"
          :running-contest-count="runningContestCount"
          :frozen-contest-count="frozenContestCount"
          :preferred-contest-title="preferredContest ? preferredContest.title : '暂无'"
          :back-route="backToContestDirectoryRoute"
        />

        <ContestOperationsHubWorkspacePanel
          :loading="loading"
          :load-error="loadError"
          :operable-contests="operableContests"
          :page="page"
          :total="total"
          :total-pages="totalPages"
          :back-route="backToContestDirectoryRoute"
          :build-operations-route="buildContestOperationsRoute"
          @retry="void loadContests()"
          @change-page="void changeContestPage($event)"
        />
    </main>
  </section>
</template>

<style scoped>
.contest-ops-content {
  display: flex;
  flex-direction: column;
  gap: var(--workspace-directory-page-block-gap, var(--space-5));
}
</style>
