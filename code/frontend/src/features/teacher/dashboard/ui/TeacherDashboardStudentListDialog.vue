<script setup lang="ts">
import { computed } from 'vue'

import type { AppRouteTarget } from '@/shared/lib/navigation/routeTarget'
import AppRouteLink from '@/shared/ui/navigation/AppRouteLink.vue'
import AdminSurfaceModal from '@/shared/ui/common/modal-templates/AdminSurfaceModal.vue'

interface StudentTarget {
  id: string
  title: string
  className: string
  detail: string
  chips: string[]
  route: AppRouteTarget | null
}

const props = defineProps<{
  modelValue: boolean
  title: string
  students: StudentTarget[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit('update:modelValue', value),
})

const subtitle = computed(() =>
  props.students.length > 0
    ? `共 ${props.students.length} 名学生，先确认名单，再进入个人复盘。`
    : '当前没有可直接进入个人复盘的学生样本。'
)

function closeDialog(): void {
  dialogVisible.value = false
}
</script>

<template>
  <AdminSurfaceModal
    :open="dialogVisible"
    :title="title"
    :subtitle="subtitle"
    eyebrow="Student List"
    width="52rem"
    @close="closeDialog"
    @update:open="dialogVisible = $event"
  >
    <div class="student-list-dialog">
      <div v-if="students.length > 0" class="student-list-dialog__list">
        <article
          v-for="student in students"
          :key="student.id"
          class="student-list-dialog__item"
        >
          <div class="student-list-dialog__main">
            <div class="student-list-dialog__title-row">
              <h3 class="student-list-dialog__title">{{ student.title }}</h3>
              <div class="student-list-dialog__chips">
                <span
                  v-for="chip in student.chips"
                  :key="chip"
                  class="workspace-directory-status-pill workspace-directory-status-pill--muted"
                >
                  {{ chip }}
                </span>
              </div>
            </div>
            <p class="student-list-dialog__detail">{{ student.detail }}</p>
          </div>

          <AppRouteLink
            v-if="student.route"
            :to="student.route"
            class="ui-btn ui-btn--ghost ui-btn--sm student-list-dialog__action"
          >
            查看复盘
          </AppRouteLink>
        </article>
      </div>
      <div v-else class="student-list-dialog__empty">
        暂无可展示的学生名单
      </div>
    </div>

    <template #footer>
      <div class="student-list-dialog__footer">
        <button
          type="button"
          class="ui-btn ui-btn--secondary"
          @click="closeDialog"
        >
          关闭
        </button>
      </div>
    </template>
  </AdminSurfaceModal>
</template>

<style scoped>
.student-list-dialog {
  display: grid;
  gap: var(--space-4);
}

.student-list-dialog__list {
  display: grid;
  gap: var(--space-3);
  max-height: min(60vh, 34rem);
  overflow: auto;
  padding-right: var(--space-1);
}

.student-list-dialog__item {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-4);
  align-items: start;
  padding: var(--space-4);
  border: 1px solid color-mix(in srgb, var(--color-border-default) 72%, transparent);
  border-radius: 16px;
  background: color-mix(in srgb, var(--color-bg-surface) 92%, transparent);
}

.student-list-dialog__main {
  min-width: 0;
}

.student-list-dialog__title-row {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: var(--space-3);
}

.student-list-dialog__title {
  flex: 1 1 14rem;
  min-width: 0;
  margin: 0;
  font-size: var(--font-size-16);
  font-weight: 800;
  color: var(--color-text-primary);
}

.student-list-dialog__chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.student-list-dialog__detail {
  margin: var(--space-2) 0 0;
  font-size: var(--font-size-14);
  line-height: 1.7;
  color: var(--color-text-secondary);
}

.student-list-dialog__action {
  align-self: center;
}

.student-list-dialog__empty {
  padding: var(--space-5);
  border: 1px dashed color-mix(in srgb, var(--color-border-default) 70%, transparent);
  border-radius: 16px;
  text-align: center;
  font-size: var(--font-size-14);
  color: var(--color-text-secondary);
}

.student-list-dialog__footer {
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 760px) {
  .student-list-dialog__item {
    grid-template-columns: 1fr;
  }

  .student-list-dialog__action {
    justify-self: start;
  }
}
</style>
