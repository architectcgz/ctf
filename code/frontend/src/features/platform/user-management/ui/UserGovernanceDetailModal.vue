<script setup lang="ts">
import { UserRoundCheck } from 'lucide-vue-next'

import type { AdminUserListItem, UserStatus } from '@/api/contracts'
import AdminSurfaceModal from '@/shared/ui/common/modal-templates/AdminSurfaceModal.vue'

const props = defineProps<{
  user: AdminUserListItem | null
}>()

const emit = defineEmits<{
  close: []
  edit: [user: AdminUserListItem]
  delete: [user: AdminUserListItem]
}>()

const userStatusAccentMap: Record<UserStatus, string> = {
  active: 'var(--color-primary)',
  locked: 'var(--color-warning)',
  banned: 'var(--color-danger)',
  inactive: 'color-mix(in srgb, var(--journal-muted) 84%, var(--journal-ink))',
}

function getUserAccentColor(status: UserStatus): string {
  return userStatusAccentMap[status] ?? 'var(--color-primary)'
}

function getUserStatusStyle(status: UserStatus): Record<string, string> {
  const accent = getUserAccentColor(status)
  return {
    color: accent,
    borderColor: `color-mix(in srgb, ${accent} 18%, transparent)`,
    backgroundColor: `color-mix(in srgb, ${accent} 10%, var(--journal-surface))`,
  }
}

function getUserIdentity(user: AdminUserListItem): string {
  if (user.roles.includes('admin') || user.roles.includes('teacher')) {
    return user.teacher_no || '未设置'
  }
  if (user.roles.includes('student')) {
    return user.student_no || '未设置'
  }
  return '未设置'
}

function formatCreatedAt(value: string): string {
  return new Date(value).toLocaleString('zh-CN')
}
</script>

<template>
  <AdminSurfaceModal
    v-if="user"
    :open="!!user"
    :title="user.username"
    eyebrow="User Detail"
    width="40rem"
    @close="emit('close')"
    @update:open="($event) => { if (!$event) emit('close') }"
  >
    <div class="user-detail-drawer">
      <dl class="user-detail-list">
        <div class="user-detail-item">
          <dt>用户名</dt>
          <dd>{{ user.username }}</dd>
        </div>
        <div class="user-detail-item">
          <dt>姓名</dt>
          <dd>{{ user.name || user.username }}</dd>
        </div>
        <div class="user-detail-item">
          <dt>邮箱</dt>
          <dd>{{ user.email || '未填写邮箱' }}</dd>
        </div>
        <div class="user-detail-item">
          <dt>角色</dt>
          <dd>
            <div class="user-row__roles">
              <span
                v-for="role in user.roles"
                :key="`detail-${user.id}-${role}`"
                class="admin-role-chip"
              >
                <UserRoundCheck class="h-3.5 w-3.5" />
                {{ role }}
              </span>
            </div>
          </dd>
        </div>
        <div class="user-detail-item">
          <dt>状态</dt>
          <dd>
            <span class="admin-status-chip" :style="getUserStatusStyle(user.status)">
              {{ user.status }}
            </span>
          </dd>
        </div>
        <div class="user-detail-item">
          <dt>班级</dt>
          <dd>{{ user.class_name || '未分配班级' }}</dd>
        </div>
        <div class="user-detail-item">
          <dt>学号 / 工号</dt>
          <dd>{{ getUserIdentity(user) }}</dd>
        </div>
        <div class="user-detail-item">
          <dt>创建时间</dt>
          <dd>{{ formatCreatedAt(user.created_at) }}</dd>
        </div>
      </dl>
    </div>

    <template #footer>
      <div class="user-detail-actions">
        <button
          id="user-detail-close"
          type="button"
          class="ui-btn ui-btn--secondary user-action-btn"
          @click="emit('close')"
        >
          关闭
        </button>
        <button
          type="button"
          class="ui-btn ui-btn--secondary user-action-btn"
          @click="emit('edit', user)"
        >
          编辑
        </button>
        <button
          type="button"
          class="ui-btn ui-btn--danger user-action-btn"
          @click="emit('delete', user)"
        >
          删除
        </button>
      </div>
    </template>
  </AdminSurfaceModal>
</template>

<style scoped>
.user-detail-drawer {
  min-width: 0;
}

.user-detail-list {
  display: grid;
  gap: var(--space-4);
  margin: 0;
  padding: var(--space-5);
  overflow: auto;
}

.user-detail-item {
  display: grid;
  gap: var(--space-2);
  min-width: 0;
  padding-bottom: var(--space-4);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 54%, transparent);
}

.user-detail-item:last-child {
  padding-bottom: 0;
  border-bottom: 0;
}

.user-detail-item dt {
  color: var(--journal-muted);
  font-size: var(--font-size-11);
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.user-detail-item dd {
  min-width: 0;
  margin: 0;
  overflow-wrap: anywhere;
  color: var(--journal-ink);
  font-size: var(--font-size-14);
}

.user-row__roles {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.user-detail-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  padding: var(--space-5);
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 72%, transparent);
  background: color-mix(in srgb, var(--journal-surface-subtle) 72%, var(--journal-surface));
}

.user-detail-actions > .ui-btn {
  --ui-btn-height: 2.75rem;
  --ui-btn-radius: 1rem;
  --ui-btn-padding: var(--space-2-5) var(--space-4);
  --ui-btn-font-size: var(--font-size-0-875);
  --ui-btn-font-weight: 600;
  --ui-btn-focus-ring: color-mix(in srgb, var(--journal-accent) 18%, transparent);
}

.user-action-btn {
  --ui-btn-height: 2rem;
  --ui-btn-padding: var(--space-1-5) var(--space-3);
  --ui-btn-radius: 0.8rem;
  --ui-btn-font-size: var(--font-size-0-8125);
}

.user-detail-actions > .ui-btn.ui-btn--danger {
  --ui-btn-danger-border: color-mix(in srgb, var(--color-danger) 20%, transparent);
  --ui-btn-danger-background: color-mix(in srgb, var(--color-danger) 10%, var(--journal-surface));
  --ui-btn-danger-color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
}

.admin-status-chip,
.admin-role-chip {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1-5);
  border-radius: 999px;
  padding: var(--space-1-5) var(--space-3);
  font-size: var(--font-size-0-72);
  font-weight: 600;
}

.admin-status-chip {
  border: 1px solid color-mix(in srgb, var(--journal-accent) 14%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  color: var(--journal-accent);
}

.admin-role-chip {
  border: 1px solid color-mix(in srgb, var(--journal-accent) 16%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  color: var(--journal-accent);
}
</style>
