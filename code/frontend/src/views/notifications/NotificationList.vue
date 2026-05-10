<script setup lang="ts">
import { RefreshCw } from 'lucide-vue-next'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AdminNotificationPublishDrawer from '@/components/notifications/AdminNotificationPublishDrawer.vue'
import NotificationCategoryFilter from '@/components/notifications/NotificationCategoryFilter.vue'
import PagePaginationControls from '@/components/common/PagePaginationControls.vue'
import { useNotificationListPage } from '@/features/notifications'
import { formatDate } from '@/utils/format'

const {
  publishDrawerOpen,
  probeMessage,
  list,
  total,
  page,
  pageSize,
  loading,
  changePage,
  totalPages,
  hasLoadError,
  loadErrorMessage,
  headStats,
  categoryOptions,
  selectedCategory,
  selectedCategoryLabel,
  canPublishNotification,
  typeLabel,
  selectCategory,
  openNotificationDetail,
  markCurrentPageRead,
  openPublishDrawer,
  closePublishDrawer,
  handlePublishSuccess,
  handleRefresh,
} = useNotificationListPage()
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-user journal-hero flex min-h-full flex-1 flex-col space-y-6"
  >
    <main class="content-pane">
      <div class="notification-page">
        <header class="notification-topbar">
          <div class="notification-heading">
            <div class="workspace-overline">Notifications</div>
            <h1 class="notification-title workspace-page-title">通知中心</h1>
            <p class="notification-subtitle">系统、竞赛和训练相关通知会在这里按时间顺序汇总。</p>
          </div>

          <div class="notification-topbar-meta">
            <div class="notification-actions">
              <button
                v-if="canPublishNotification"
                type="button"
                class="ui-btn ui-btn--primary"
                @click="openPublishDrawer"
              >
                发布通知
              </button>
              <button type="button" class="ui-btn ui-btn--secondary" @click="markCurrentPageRead">
                本页已读
              </button>
              <button type="button" class="ui-btn ui-btn--secondary" @click="handleRefresh">
                <RefreshCw class="h-4 w-4" />
                刷新
              </button>
            </div>
          </div>
        </header>

        <p v-if="probeMessage" class="notification-probe-note">
          {{ probeMessage }}
        </p>

        <section
          class="student-directory-section workspace-directory-section"
          aria-label="通知目录"
        >
          <section
            class="student-directory-shell notification-directory-shell workspace-directory-list"
          >
            <header class="student-directory-shell__head student-directory-list-heading list-heading">
              <div class="student-directory-shell__heading student-directory-list-heading__body">
                <div
                  class="journal-note-label student-directory-shell__eyebrow student-directory-list-heading__eyebrow"
                >
                  Notification Directory
                </div>
                <h2 class="student-directory-shell__title student-directory-list-heading__title">
                  {{ selectedCategoryLabel }}消息
                </h2>
              </div>
              <div class="student-directory-shell__meta">共 {{ total }} 条</div>
            </header>

            <div
              v-if="loading"
              class="notification-loading student-directory-state workspace-directory-loading"
            >
              <div class="student-directory-spinner" />
            </div>

            <AppEmpty
              v-else-if="hasLoadError"
              class="notification-empty-state student-directory-state workspace-directory-empty"
              icon="AlertTriangle"
              title="通知加载失败"
              :description="loadErrorMessage"
            >
              <template #action>
                <button type="button" class="ui-btn ui-btn--secondary" @click="handleRefresh">
                  重新加载
                </button>
              </template>
            </AppEmpty>

            <section
              v-else
              class="student-directory-filters notification-filter-section"
              aria-label="消息分类"
            >
              <NotificationCategoryFilter
                :total="total"
                :selected-category="selectedCategory"
                :selected-category-label="selectedCategoryLabel"
                :category-options="categoryOptions"
                @select-category="selectCategory"
              />
              <div class="notification-head-stats" aria-label="消息概况">
                <div v-for="stat in headStats" :key="stat.key" class="notification-head-stat">
                  <span class="notification-head-stat__label">{{ stat.label }}</span>
                  <strong class="notification-head-stat__value">{{ stat.value }}</strong>
                </div>
              </div>
            </section>

            <AppEmpty
              v-if="!loading && !hasLoadError && list.length === 0"
              class="notification-empty-state student-directory-state workspace-directory-empty"
              icon="Inbox"
              title="暂无通知"
              description="新的系统、竞赛、团队和训练消息会在这里汇总展示。"
            />

            <section
              v-else-if="!loading && !hasLoadError"
              class="notification-directory"
              aria-label="通知目录"
            >
              <div class="workspace-directory-grid-head notification-directory-head">
                <span>类型</span>
                <span>标题与内容</span>
                <span>时间</span>
                <span>状态</span>
              </div>

              <button
                v-for="item in list"
                :key="item.id"
                type="button"
                class="workspace-directory-grid-row notification-row"
                :class="{ 'notification-row-unread': item.unread }"
                @click="openNotificationDetail(item)"
              >
                <div class="notification-row-type">
                  <span class="workspace-directory-status-pill notification-chip">{{
                    typeLabel(item.type)
                  }}</span>
                </div>
                <div class="workspace-directory-cell notification-row-main">
                  <div
                    class="notification-row-title"
                    :class="'workspace-directory-row-title'"
                    :title="item.title"
                  >
                    {{ item.title }}
                  </div>
                  <div
                    class="notification-row-copy"
                    :class="[
                      'workspace-directory-row-subtitle',
                      'workspace-directory-row-subtitle--clamp',
                    ]"
                    :title="item.content"
                  >
                    {{ item.content }}
                  </div>
                </div>
                <div class="workspace-directory-compact-text notification-row-time">
                  {{ formatDate(item.created_at) }}
                </div>
                <div class="notification-row-state">
                  <span
                    class="workspace-directory-status-pill notification-state-chip"
                    :class="{
                      'workspace-directory-status-pill--primary notification-state-chip-unread':
                        item.unread,
                      'workspace-directory-status-pill--muted': !item.unread,
                    }"
                  >
                    {{ item.unread ? '未读' : '已读' }}
                  </span>
                </div>
              </button>
            </section>

            <div
              v-if="list.length > 0 && total > 0"
              class="notification-pagination workspace-directory-pagination"
            >
              <PagePaginationControls
                :page="page"
                :total-pages="totalPages"
                :total="total"
                :total-label="`共 ${total} 条`"
                @change-page="changePage"
              />
            </div>
          </section>
        </section>
      </div>
    </main>

    <AdminNotificationPublishDrawer
      :open="publishDrawerOpen"
      @close="closePublishDrawer"
      @published="handlePublishSuccess"
    />
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-shell-surface-subtle: color-mix(
    in srgb,
    var(--color-bg-elevated) 74%,
    var(--color-bg-base)
  );
  --journal-shell-accent: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
  --journal-shell-hero-end: color-mix(
    in srgb,
    var(--journal-surface-subtle) 94%,
    var(--color-bg-base)
  );
}

.notification-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.notification-subtitle {
  max-width: 720px;
}

.notification-probe-note {
  margin-top: 12px;
  font-size: var(--font-size-13);
  font-weight: 600;
  line-height: 1.7;
  color: color-mix(in srgb, var(--journal-accent) 84%, var(--journal-muted));
}

.notification-topbar-meta {
  display: grid;
  justify-items: end;
  gap: 12px;
}

.notification-head-stats {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-start;
  gap: var(--space-3);
}

.notification-head-stat {
  display: inline-flex;
  align-items: center;
  gap: var(--space-3);
  min-height: var(--ui-control-height-md);
  padding: 0 var(--space-4);
  border: 1px solid color-mix(in srgb, var(--journal-border) 82%, transparent);
  border-radius: var(--ui-control-radius-md);
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
}

.notification-head-stat__label {
  font-size: var(--font-size-13);
  font-weight: 600;
  color: var(--journal-muted);
}

.notification-head-stat__value {
  font-size: var(--font-size-16);
  font-weight: 700;
  color: var(--journal-ink);
}

.notification-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
}

.notification-loading {
  display: flex;
  align-items: center;
  justify-content: center;
}

:deep(.notification-empty-state) {
  margin-top: 0;
}

.notification-filter-section {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.notification-directory-shell {
  --workspace-directory-grid-columns: 8.75rem minmax(0, 1fr) 11.25rem 7.5rem;
}

.notification-row {
  cursor: pointer;
}

.notification-chip {
  border-color: color-mix(in srgb, var(--journal-accent) 22%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent);
}

.notification-row-main {
  min-width: 0;
}

.notification-row-title {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notification-row-copy {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  overflow: hidden;
}

.notification-pagination {
  margin-top: var(--workspace-directory-gap-pagination);
}

@media (max-width: 1180px) {
  .notification-topbar-meta {
    width: 100%;
    justify-items: start;
  }

  .notification-head-stats,
  .notification-actions {
    justify-content: flex-start;
  }

  .notification-directory-head {
    display: none;
  }

  .notification-row {
    grid-template-columns: 1fr;
  }
}
</style>
