<script setup lang="ts">
import { resolveDashboardPanelComponent } from '@/components/dashboard/student/dashboardPanelRegistry'
import { useStudentDashboardPage } from '@/features/student-dashboard'

const {
  loading,
  error,
  progress,
  panelTabs,
  activePanel,
  setTabButtonRef,
  switchPanel,
  handleTabKeydown,
  loadDashboard,
  resolveDashboardPanelBindings,
} = useStudentDashboardPage()
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-user journal-hero flex min-h-full flex-1 flex-col"
  >
    <nav
      class="top-tabs"
      role="tablist"
      aria-label="学生仪表盘视图切换"
    >
      <button
        v-for="(tab, index) in panelTabs"
        :id="tab.tabId"
        :key="tab.tabId"
        :ref="(element) => setTabButtonRef(tab.key, element as HTMLButtonElement | null)"
        type="button"
        role="tab"
        class="top-tab"
        :class="{ active: activePanel === tab.key }"
        :aria-selected="activePanel === tab.key ? 'true' : 'false'"
        :aria-controls="tab.panelId"
        :tabindex="activePanel === tab.key ? 0 : -1"
        @click="switchPanel(tab.key)"
        @keydown="handleTabKeydown($event, index)"
      >
        {{ tab.label }}
      </button>
    </nav>

    <main class="content-pane">
        <div
          v-if="error"
          class="workspace-alert"
          role="alert"
          aria-live="polite"
        >
          {{ error }}
          <button
            type="button"
            class="workspace-alert-action"
            @click="loadDashboard"
          >
            重试
          </button>
        </div>

        <div
          v-if="loading"
          class="dashboard-loading-grid"
        >
          <div
            v-for="index in 4"
            :key="index"
            class="dashboard-loading-item"
          />
        </div>

        <template v-else-if="progress">
          <component
            :is="resolveDashboardPanelComponent(tab.key)"
            v-for="tab in panelTabs"
            v-show="activePanel === tab.key"
            :id="tab.panelId"
            :key="tab.panelId"
            class="tab-panel"
            :class="{ active: activePanel === tab.key }"
            role="tabpanel"
            :aria-labelledby="tab.tabId"
            :aria-hidden="activePanel === tab.key ? 'false' : 'true'"
            v-bind="resolveDashboardPanelBindings(tab.key)"
          />
        </template>
    </main>
  </section>
</template>

<style scoped>
.workspace-shell {
  --workspace-line-soft: var(--journal-border);
  --workspace-faint: var(--journal-muted);
  --workspace-brand: var(--journal-accent);
  --workspace-brand-ink: var(--journal-accent-strong);
  --workspace-brand-soft: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  --workspace-page: var(--color-bg-base);
  --workspace-shell-bg: var(--journal-surface);
  --workspace-danger: var(--color-danger);
  --workspace-shadow-shell: var(--journal-shell-hero-shadow, 0 22px 50px var(--color-shadow-soft));
  --workspace-font-sans: var(--font-family-sans);
  --journal-track: color-mix(in srgb, var(--color-bg-surface) 84%, var(--color-bg-base));
  --journal-shell-hero-end: color-mix(
    in srgb,
    var(--journal-surface-subtle) 94%,
    var(--color-bg-base)
  );
  flex: 1 1 auto;
}

.content-pane {
  min-height: 0;
}

.tab-panel {
  min-height: 0;
}

.workspace-alert {
  margin-bottom: 18px;
  padding: 16px 18px;
  border: 1px solid color-mix(in srgb, var(--workspace-danger) 24%, var(--workspace-line-soft));
  border-radius: 18px;
  background: color-mix(in srgb, var(--workspace-danger) 6%, transparent);
  font-size: var(--font-size-13);
  line-height: 1.7;
  color: var(--journal-ink);
}

.workspace-alert-action {
  margin-left: 10px;
  border: 0;
  background: transparent;
  font-weight: 600;
  text-decoration: underline;
  color: inherit;
  cursor: pointer;
}

.dashboard-loading-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.dashboard-loading-item {
  height: 7.5rem;
  border-radius: 18px;
  background: var(--journal-track);
  animation: dashboardPulse 1.1s ease-in-out infinite;
}

@keyframes dashboardPulse {
  0%,
  100% {
    opacity: 0.6;
  }

  50% {
    opacity: 1;
  }
}

@media (max-width: 860px) {
  .top-tabs {
    gap: 18px;
    padding: 0 18px;
  }

  .content-pane {
    padding: 18px;
  }
}

@media (max-width: 640px) {
  .dashboard-loading-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
