<script setup lang="ts">
import AwdContextHero from '@/modules/awd/components/AwdContextHero.vue'
import AwdPageNav from '@/modules/awd/components/AwdPageNav.vue'
import AwdStatusShell from '@/modules/awd/components/AwdStatusShell.vue'
import type { AwdHeroMetric, AwdPageDefinition } from '@/modules/awd/types'

withDefaults(
  defineProps<{
    eyebrow?: string
    shellClass?: string
    contestTitle: string
    pageTitle: string
    pageDescription?: string
    pages: Array<AwdPageDefinition<string>>
    currentPage: string
    heroMetrics?: AwdHeroMetric[]
    resolvePath?: (pageKey: string) => string
    loading?: boolean
    error?: string
    empty?: boolean
    emptyTitle?: string
    emptyDescription?: string
  }>(),
  {
    eyebrow: 'AWD Workspace',
    shellClass: '',
    pageDescription: '',
    heroMetrics: () => [],
    resolvePath: undefined,
    loading: false,
    error: '',
    empty: false,
    emptyTitle: '暂无数据',
    emptyDescription: '',
  }
)
</script>

<template>
  <section
    class="workspace-shell awd-workspace-layout"
    :class="shellClass"
  >
    <div class="content-pane awd-workspace-layout__content">
      <AwdContextHero
        :eyebrow="eyebrow"
        :contest-title="contestTitle"
        :page-title="pageTitle"
        :page-description="pageDescription"
        :metrics="heroMetrics"
      />

      <div class="awd-workspace-layout__body">
        <aside class="awd-workspace-layout__rail">
          <AwdPageNav
            :items="pages"
            :current-page="currentPage"
            :resolve-path="resolvePath"
          />
        </aside>

        <main class="awd-workspace-layout__main">
          <AwdStatusShell
            :loading="loading"
            :error="error"
            :empty="empty"
            :empty-title="emptyTitle"
            :empty-description="emptyDescription"
          >
            <slot />
          </AwdStatusShell>
        </main>
      </div>
    </div>
  </section>
</template>

<style scoped>
.awd-workspace-layout {
  min-height: 100%;
}

.awd-workspace-layout__content {
  display: grid;
  gap: 1.5rem;
}

.awd-workspace-layout__body {
  display: grid;
  grid-template-columns: 260px minmax(0, 1fr);
  gap: 1.25rem;
  align-items: start;
}

.awd-workspace-layout__rail {
  position: sticky;
  top: 1.5rem;
}

.awd-workspace-layout__main {
  min-width: 0;
  min-height: 360px;
  padding: 1.25rem;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 76%, transparent);
  border-radius: 1.25rem;
  background: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  box-shadow: 0 18px 42px color-mix(in srgb, var(--color-shadow-soft) 58%, transparent);
}

@media (max-width: 1024px) {
  .awd-workspace-layout__body {
    grid-template-columns: minmax(0, 1fr);
  }

  .awd-workspace-layout__rail {
    position: static;
  }
}
</style>
