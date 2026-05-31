<template>
  <NotificationListWorkspace
    :publish-drawer-open="publishDrawerOpen"
    :probe-message="probeMessage"
    :list="list"
    :total="total"
    :page="page"
    :loading="loading"
    :total-pages="totalPages"
    :has-load-error="hasLoadError"
    :load-error-message="loadErrorMessage"
    :head-stats="headStats"
    :category-options="categoryOptions"
    :selected-category="selectedCategory"
    :selected-category-label="selectedCategoryLabel"
    :can-publish-notification="canPublishNotification"
    :type-label="typeLabel"
    :select-category="selectCategory"
    :notification-detail-route="notificationDetailRoute"
    :mark-current-page-read="markCurrentPageRead"
    :open-publish-drawer="openPublishDrawer"
    :close-publish-drawer="closePublishDrawer"
    :handle-publish-success="handlePublishSuccess"
    :handle-refresh="handleRefresh"
    :change-page="changePage"
  />
</template>

<script setup lang="ts">
import { useNotificationListPage } from '@/features/notifications'
import { NotificationListWorkspace } from '@/widgets/notification-list-workspace'

const {
  publishDrawerOpen,
  probeMessage,
  list,
  total,
  page,
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
  notificationDetailRoute,
  markCurrentPageRead,
  openPublishDrawer,
  closePublishDrawer,
  handlePublishSuccess,
  handleRefresh,
} = useNotificationListPage()
</script>
