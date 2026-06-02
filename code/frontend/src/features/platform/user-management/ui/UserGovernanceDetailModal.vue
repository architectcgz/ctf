<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Copy, LogOut, UserRoundCheck } from 'lucide-vue-next'

import type { AdminUserListItem, UserStatus } from '@/api/contracts'
import AdminSurfaceModal from '@/shared/ui/common/modal-templates/AdminSurfaceModal.vue'
import { getUserName } from '@/entities/user'
import { usePlatformUserSessions } from '../model/usePlatformUserSessions'

const props = defineProps<{
  user: AdminUserListItem | null
}>()

const emit = defineEmits<{
  close: []
  edit: [user: AdminUserListItem]
  delete: [user: AdminUserListItem]
}>()

const userId = computed(() => props.user?.id)

const {
  sessions,
  loading: sessionsLoading,
  revokingSessionId,
  revokingAll,
  copiedSessionId,
  revokeOne: handleRevokeSession,
  revokeAll: handleRevokeAll,
  copySessionIdToClipboard: copySessionId,
  reset: resetSessions,
} = usePlatformUserSessions(() => userId.value)

const showRevokeAllConfirm = ref(false)

// 切换用户或关窗时清除确认态，防止跨用户泄漏
watch(
  () => props.user?.id,
  (newId, oldId) => {
    if (newId !== oldId) {
      showRevokeAllConfirm.value = false
    }
  }
)
watch(
  () => props.user,
  (newUser) => {
    if (!newUser) {
      showRevokeAllConfirm.value = false
      resetSessions()
    }
  }
)

function handleRevokeAllWithConfirm() {
  showRevokeAllConfirm.value = true
}

async function handleRevokeAllConfirmed() {
  await handleRevokeAll()
  showRevokeAllConfirm.value = false
}

function formatExpiresAt(value: string): string {
  return new Date(value).toLocaleString('zh-CN')
}

function truncateSessionId(id: string): string {
  return id.length > 12 ? `${id.slice(0, 6)}…${id.slice(-4)}` : id
}

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
          <dd>{{ getUserName(user, '未设置姓名') }}</dd>
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

      <div class="user-detail-section">
        <div class="user-detail-section__header">
          <h3 class="user-detail-section__title">活跃会话</h3>
          <button
            v-if="sessions.length > 0"
            type="button"
            class="user-detail-section__action-btn user-detail-section__action-btn--danger"
            :disabled="revokingAll"
            @click="handleRevokeAllWithConfirm"
          >
            {{ revokingAll ? '撤销中…' : '撤销全部' }}
          </button>
        </div>

        <div v-if="sessionsLoading" class="user-session-state">
          <span class="user-session-state__text">加载中…</span>
        </div>
        <div v-else-if="sessions.length === 0" class="user-session-state">
          <span class="user-session-state__text">无活跃会话</span>
        </div>
        <ul v-else class="user-session-list">
          <li
            v-for="s in sessions"
            :key="s.id"
            class="user-session-item"
          >
            <div class="user-session-item__meta">
              <div class="user-session-item__id-row">
                <span class="user-session-item__id" :title="s.id">
                  {{ truncateSessionId(s.id) }}
                </span>
                <button
                  type="button"
                  class="user-session-item__copy-btn"
                  :title="copiedSessionId === s.id ? '已复制' : '复制会话 ID'"
                  @click.stop="copySessionId(s.id)"
                >
                  <Copy class="h-3 w-3" />
                </button>
              </div>
              <span class="user-session-item__expiry">
                过期：{{ formatExpiresAt(s.expires_at) }}
              </span>
            </div>
            <button
              type="button"
              class="user-session-item__revoke-btn"
              :disabled="revokingSessionId === s.id"
              @click="handleRevokeSession(s.id)"
            >
              <LogOut class="h-3.5 w-3.5" />
              {{ revokingSessionId === s.id ? '…' : '撤销' }}
            </button>
          </li>
        </ul>

        <!-- Revoke all confirmation -->
        <div v-if="showRevokeAllConfirm" class="user-session-confirm">
          <p class="user-session-confirm__text">
            确认撤销该用户的所有活跃会话？用户将被强制登出。
          </p>
          <div class="user-session-confirm__actions">
            <button
              type="button"
              class="ui-btn ui-btn--secondary user-session-confirm__btn"
              @click="showRevokeAllConfirm = false"
            >
              取消
            </button>
            <button
              type="button"
              class="ui-btn ui-btn--danger user-session-confirm__btn"
              :disabled="revokingAll"
              @click="handleRevokeAllConfirmed"
            >
              {{ revokingAll ? '撤销中…' : '确认撤销全部会话' }}
            </button>
          </div>
        </div>
      </div>
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

/* Session section */
.user-detail-section {
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 72%, transparent);
  padding: var(--space-5);
}

.user-detail-section__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  margin-bottom: var(--space-4);
}

.user-detail-section__title {
  margin: 0;
  color: var(--journal-muted);
  font-size: var(--font-size-11);
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.user-detail-section__action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-1-5);
  border: 1px solid var(--admin-control-border);
  border-radius: 0.8rem;
  padding: var(--space-1-5) var(--space-3);
  background: transparent;
  color: var(--journal-muted);
  font-size: var(--font-size-0-75);
  font-weight: 600;
  cursor: pointer;
  transition: color 0.15s, border-color 0.15s, background-color 0.15s;
}

.user-detail-section__action-btn:hover:not(:disabled) {
  border-color: var(--journal-accent);
  color: var(--journal-accent);
}

.user-detail-section__action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.user-detail-section__action-btn--danger:hover:not(:disabled) {
  border-color: var(--color-danger);
  color: var(--color-danger);
  background: color-mix(in srgb, var(--color-danger) 8%, transparent);
}

.user-session-state {
  padding: var(--space-4) 0;
}

.user-session-state__text {
  color: var(--journal-muted);
  font-size: var(--font-size-0-875);
}

.user-session-list {
  display: grid;
  gap: var(--space-2);
  margin: 0;
  padding: 0;
  list-style: none;
}

.user-session-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  border: 1px solid color-mix(in srgb, var(--journal-border) 54%, transparent);
  border-radius: 0.8rem;
  padding: var(--space-2-5) var(--space-3);
  min-width: 0;
}

.user-session-item__meta {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  min-width: 0;
}

.user-session-item__id-row {
  display: flex;
  align-items: center;
  gap: var(--space-1-5);
  min-width: 0;
}

.user-session-item__id {
  font-family: monospace;
  font-size: var(--font-size-0-8125);
  color: var(--journal-ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-session-item__copy-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 1.5rem;
  height: 1.5rem;
  border: none;
  border-radius: 0.25rem;
  background: transparent;
  color: var(--journal-muted);
  cursor: pointer;
  transition: color 0.15s, background-color 0.15s;
}

.user-session-item__copy-btn:hover {
  color: var(--journal-accent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
}

.user-session-item__expiry {
  font-size: var(--font-size-0-75);
  color: var(--journal-muted);
}

.user-session-item__revoke-btn {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1-5);
  flex-shrink: 0;
  border: 1px solid color-mix(in srgb, var(--color-danger) 28%, transparent);
  border-radius: 0.6rem;
  padding: var(--space-1) var(--space-2-5);
  background: transparent;
  color: var(--color-danger);
  font-size: var(--font-size-0-75);
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.15s;
}

.user-session-item__revoke-btn:hover:not(:disabled) {
  background: color-mix(in srgb, var(--color-danger) 12%, transparent);
}

.user-session-item__revoke-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.user-session-confirm {
  margin-top: var(--space-4);
  border: 1px solid color-mix(in srgb, var(--color-danger) 24%, transparent);
  border-radius: 0.8rem;
  padding: var(--space-4);
  background: color-mix(in srgb, var(--color-danger) 6%, var(--journal-surface));
}

.user-session-confirm__text {
  margin: 0 0 var(--space-3);
  color: var(--journal-ink);
  font-size: var(--font-size-0-875);
}

.user-session-confirm__actions {
  display: flex;
  gap: var(--space-3);
  justify-content: flex-end;
}

.user-session-confirm__btn {
  --ui-btn-height: 2rem;
  --ui-btn-padding: var(--space-1-5) var(--space-3);
  --ui-btn-radius: 0.8rem;
  --ui-btn-font-size: var(--font-size-0-8125);
}

.user-session-confirm__btn.ui-btn--danger {
  --ui-btn-danger-border: color-mix(in srgb, var(--color-danger) 28%, transparent);
  --ui-btn-danger-background: color-mix(in srgb, var(--color-danger) 12%, var(--journal-surface));
  --ui-btn-danger-color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
}
</style>
