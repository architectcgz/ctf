<script setup lang="ts">
import { FileUp } from 'lucide-vue-next'

import type { AdminUserImportData } from '@/api/contracts'

defineProps<{
  importResult: AdminUserImportData | null
}>()

const emit = defineEmits<{
  returnOverview: []
  triggerImport: []
}>()
</script>

<template>
  <section class="user-panel user-panel--import">
    <section class="workspace-directory-section user-import-panel">
      <header class="workspace-tab-heading user-import-head">
        <div class="workspace-tab-heading__main">
          <div class="workspace-overline">
            User Import
          </div>
          <h2 class="workspace-page-title">
            导入用户
          </h2>
          <p class="workspace-page-copy">
            统一导入账号、角色与班级归属，导入完成后可回到工作台继续筛选和治理具体用户。
          </p>
        </div>

        <div class="header-actions user-panel-actions">
          <button
            id="user-return-overview"
            type="button"
            class="header-btn header-btn--ghost"
            @click="emit('returnOverview')"
          >
            返回工作台
          </button>
          <button
            type="button"
            class="header-btn header-btn--primary"
            @click="emit('triggerImport')"
          >
            <FileUp class="h-4 w-4" />
            批量导入
          </button>
        </div>
      </header>

      <div class="journal-note user-import-format">
        <div class="journal-note-label">
          CSV 格式
        </div>
        <div class="journal-note-helper">
          列顺序：`username,password,email,class_name,role,status,student_no,teacher_no,name`
        </div>
      </div>

      <section class="workspace-directory-section user-import-receipt-section">
        <header class="list-heading user-import-receipt-head">
          <div>
            <div class="journal-note-label">
              Import Receipt
            </div>
            <h2 class="list-heading__title">
              导入回执
            </h2>
          </div>
        </header>

        <div v-if="importResult" class="admin-receipt">
          <p>
            创建 {{ importResult.created }}，更新 {{ importResult.updated }}，失败
            {{ importResult.failed }}
          </p>
          <ul
            v-if="importResult.errors?.length"
            class="mt-3 space-y-2 text-[var(--color-danger)]"
          >
            <li
              v-for="item in importResult.errors.slice(0, 5)"
              :key="`${item.row}-${item.message}`"
            >
              第 {{ item.row }} 行：{{ item.message }}
            </li>
          </ul>
        </div>
        <div v-else class="admin-empty">
          还没有导入记录。
        </div>
      </section>
    </section>
  </section>
</template>

<style scoped>
.user-panel {
  display: grid;
  gap: var(--space-4);
}

.user-import-head {
  gap: var(--space-3);
}

.user-import-format {
  border: 1px dashed color-mix(in srgb, var(--journal-border) 72%, transparent);
  border-radius: 1.1rem;
  background: color-mix(in srgb, var(--journal-surface) 96%, transparent);
  padding: var(--space-4);
}

.admin-receipt {
  border-radius: 16px;
  border: 1px solid var(--journal-border);
  background: color-mix(in srgb, var(--journal-surface) 95%, transparent);
  padding: var(--space-4);
  font-size: var(--font-size-0-875);
  color: var(--journal-ink);
}

.admin-empty {
  border: 1px dashed color-mix(in srgb, var(--journal-border) 72%, transparent);
  border-radius: 16px;
  padding: var(--space-4);
  font-size: var(--font-size-0-875);
  color: var(--journal-muted);
}

@media (max-width: 767px) {
  .user-panel-actions {
    justify-content: flex-start;
  }
}
</style>
