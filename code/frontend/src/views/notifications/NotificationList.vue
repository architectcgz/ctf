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
            <div class="notification-title-line">
              <h1 class="notification-title workspace-page-title">
                通知中心
              </h1>
              <div
                class="notification-head-stats"
                aria-label="消息概况"
              >
                <div
                  v-for="stat in headStats"
                  :key="stat.key"
                  class="notification-head-stat"
                >
                  <span class="notification-head-stat__label">{{ stat.label }}</span>
                  <strong class="notification-head-stat__value">{{ stat.value }}</strong>
                </div>
              </div>
            </div>
            <p class="notification-subtitle">
              系统、竞赛和训练相关通知会在这里按时间顺序汇总。
            </p>
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
              <button
                type="button"
                class="ui-btn ui-btn--secondary"
                @click="markCurrentPageRead"
              >
                本页已读
              </button>
              <button
                type="button"
                class="ui-btn ui-btn--secondary"
                @click="handleRefresh"
              >
                <RefreshCw class="h-4 w-4" />
                刷新
              </button>
            </div>
          </div>
        </header>

        <p
          v-if="probeMessage"
          class="notification-probe-note"
        >
          {{ probeMessage }}
        </p>

        <section
          v-if="!loading && !hasLoadError"
          class="notification-filter-section"
          aria-label="消息分类"
        >
          <NotificationCategoryFilter
            :total="total"
            :selected-category="selectedCategory"
            :selected-category-label="selectedCategoryLabel"
            :category-options="categoryOptions"
            @select-category="selectCategory"
          />
        </section>

        <div
          v-if="loading"
          class="notification-loading"
        >
          <div class="notification-loading-spinner" />
        </div>

        <AppEmpty
          v-else-if="hasLoadError"
          class="notification-empty-state"
          icon="AlertTriangle"
          title="通知加载失败"
          :description="loadErrorMessage"
        >
          <template #action>
            <button
              type="button"
              class="ui-btn ui-btn--secondary"
              @click="handleRefresh"
            >
              重新加载
            </button>
          </template>
        </AppEmpty>

        <AppEmpty
          v-else-if="list.length === 0"
          class="notification-empty-state"
          icon="Inbox"
          title="暂无通知"
          description="新的系统、竞赛、团队和训练消息会在这里汇总展示。"
        />

        <template v-else>
          <section
            class="notification-directory"
            aria-label="通知目录"
          >
            <div class="notification-directory-top">
              <h2 class="notification-directory-title">
                {{ selectedCategoryLabel }}消息
              </h2>
            </div>

            <div class="notification-directory-head">
              <span>类型</span>
              <span>标题与内容</span>
              <span>时间</span>
              <span>状态</span>
            </div>

            <button
              v-for="item in list"
              :key="item.id"
              type="button"
              class="notification-row"
              :class="{ 'notification-row-unread': item.unread }"
              @click="openNotificationDetail(item)"
            >
              <div class="notification-row-type">
                <span class="notification-chip">{{ typeLabel(item.type) }}</span>
              </div>
              <div class="notification-row-main">
                <div
                  class="notification-row-title"
                  :title="item.title"
                >
                  {{ item.title }}
                </div>
                <div
                  class="notification-row-copy"
                  :title="item.content"
                >
                  {{ item.content }}
                </div>
              </div>
              <div class="notification-row-time">
                {{ formatDate(item.created_at) }}
              </div>
              <div class="notification-row-state">
                <span
                  class="notification-state-chip"
                  :class="{ 'notification-state-chip-unread': item.unread }"
                >
                  {{ item.unread ? '未读' : '已读' }}
                </span>
              </div>
            </button>
          </section>

          <div
            v-if="total > 0"
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
        </template>
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

.notification-title-line {
  display: flex;
  flex-wrap: wrap;
  align-items: end;
  gap: var(--space-3) var(--space-4);
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
  gap: 12px;
}

.notification-head-stat {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  min-height: 44px;
  padding: 0 14px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 82%, transparent);
  border-radius: 14px;
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
}

.notification-head-stat__label {
  font-size: var(--font-size-13);
  font-weight: 600;
  color: var(--journal-muted);
}

.notification-head-stat__value {
  font-family: var(--font-family-mono);
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
  padding: 80px 0;
}

.notification-loading-spinner {
  width: 32px;
  height: 32px;
  border: 4px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-top-color: var(--journal-accent);
  border-radius: 999px;
  animation: notificationSpin 900ms linear infinite;
}

:deep(.notification-empty-state) {
  margin-top: 24px;
  border-top-style: solid;
  border-bottom-style: solid;
  border-top-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-bottom-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.notification-filter-section {
  margin-top: var(--space-5);
}

.notification-directory {
  margin-top: var(--space-4);
}

.notification-directory-head {
  display: grid;
  grid-template-columns: 140px minmax(0, 1fr) 180px 120px;
  gap: 16px;
  padding: 0 0 12px;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  font-size: var(--font-size-11);
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.notification-row {
  display: grid;
  grid-template-columns: 140px minmax(0, 1fr) 180px 120px;
  gap: 16px;
  align-items: center;
  width: 100%;
  padding: 18px 0;
  border: 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: transparent;
  text-align: left;
  cursor: pointer;
}

.notification-row-unread {
  box-shadow: inset 2px 0 0 color-mix(in srgb, var(--journal-accent) 56%, transparent);
}

.notification-row:hover,
.notification-row:focus-visible {
  background: color-mix(in srgb, var(--journal-accent) 5%, transparent);
  outline: none;
}

.notification-chip,
.notification-state-chip {
  display: inline-flex;
  align-items: center;
  min-height: 26px;
  padding: 0 9px;
  border-radius: 8px;
  font-size: var(--font-size-12);
  font-weight: 600;
}

.notification-chip {
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent);
}

.notification-row-main {
  min-width: 0;
}

.notification-row-title {
  font-size: var(--font-size-15);
  font-weight: 700;
  color: var(--journal-ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notification-row-copy {
  margin-top: 6px;
  display: -webkit-box;
  font-size: var(--font-size-13);
  line-height: 1.6;
  color: var(--journal-muted);
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.notification-row-time {
  font-size: var(--font-size-13);
  line-height: 1.6;
  color: var(--journal-muted);
}

.notification-state-chip {
  background: color-mix(in srgb, var(--journal-muted) 10%, transparent);
  color: var(--journal-muted);
}

.notification-state-chip-unread {
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent);
}

.notification-pagination {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

@keyframes notificationSpin {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
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
