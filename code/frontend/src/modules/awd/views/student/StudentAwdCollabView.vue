<script setup lang="ts">
import StudentAwdWorkspaceLayout from '@/modules/awd/layouts/StudentAwdWorkspaceLayout.vue'

import { useStudentAwdWorkspacePage } from './useStudentAwdWorkspacePage'

const { pageModel, layoutProps } = useStudentAwdWorkspacePage('collab')
</script>

<template>
  <StudentAwdWorkspaceLayout v-bind="layoutProps">
    <div class="awd-student-page awd-collab-grid">
      <section class="awd-collab-panel">
        <div class="workspace-overline">Priorities</div>
        <h2>当前协作优先级</h2>
        <ol class="awd-collab-list">
          <li
            v-for="item in pageModel.collab.priorities"
            :key="item"
          >
            {{ item }}
          </li>
        </ol>
      </section>

      <section class="awd-collab-panel">
        <div class="workspace-overline">Notes</div>
        <h2>队伍同步提醒</h2>
        <ul class="awd-collab-list">
          <li
            v-for="item in pageModel.collab.notes"
            :key="item"
          >
            {{ item }}
          </li>
        </ul>
      </section>
    </div>
  </StudentAwdWorkspaceLayout>
</template>

<style scoped>
.awd-collab-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
}

.awd-collab-panel {
  display: grid;
  gap: 0.7rem;
  padding: 1rem 1.1rem;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 76%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--color-bg-surface) 94%, var(--color-bg-base));
}

.awd-collab-panel h2 {
  margin: 0;
  font-size: var(--font-size-18);
}

.awd-collab-list {
  margin: 0;
  padding-left: 1.1rem;
  display: grid;
  gap: 0.5rem;
  color: var(--color-text-secondary);
  font-size: var(--font-size-13);
}

@media (max-width: 900px) {
  .awd-collab-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
