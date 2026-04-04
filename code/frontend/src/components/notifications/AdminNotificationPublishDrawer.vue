<script setup lang="ts">
import { watch } from 'vue'

import type { AdminNotificationPublishResult } from '@/api/contracts'
import { useAdminNotificationPublisher } from '@/composables/useAdminNotificationPublisher'
import { USER_ROLES, type UserRole } from '@/utils/constants'

const props = defineProps<{ open: boolean }>()

const emit = defineEmits<{
  close: []
  published: [receipt: AdminNotificationPublishResult]
}>()

const publisher = useAdminNotificationPublisher()

watch(
  () => props.open,
  (open) => {
    if (open) {
      publisher.reset()
    }
  }
)

function handleClose(): void {
  emit('close')
}

async function handleSubmit(): Promise<void> {
  const result = await publisher.submit()
  if (result) {
    emit('published', result)
    emit('close')
  }
}

function onAudienceChange(value: 'all' | 'role' | 'class' | 'user'): void {
  publisher.setAudienceTarget(value)
  if (value === 'class' && publisher.classOptions.value.length === 0) {
    void publisher.loadClasses()
  }
}

function toggleRole(role: UserRole, checked: boolean): void {
  if (checked) {
    if (!publisher.selectedRoles.value.includes(role)) {
      publisher.selectedRoles.value = [...publisher.selectedRoles.value, role]
    }
    return
  }
  publisher.selectedRoles.value = publisher.selectedRoles.value.filter((item) => item !== role)
}

function toggleClass(name: string, checked: boolean): void {
  if (checked) {
    if (!publisher.selectedClasses.value.includes(name)) {
      publisher.selectedClasses.value = [...publisher.selectedClasses.value, name]
    }
    return
  }
  publisher.selectedClasses.value = publisher.selectedClasses.value.filter((item) => item !== name)
}

function toggleUser(id: string, checked: boolean): void {
  if (checked) {
    if (!publisher.selectedUserIds.value.includes(id)) {
      publisher.selectedUserIds.value = [...publisher.selectedUserIds.value, id]
    }
    return
  }
  publisher.selectedUserIds.value = publisher.selectedUserIds.value.filter((item) => item !== id)
}

async function handleUserSearch(): Promise<void> {
  await publisher.searchUsers(publisher.userKeyword.value)
}
</script>

<template>
  <Transition name="publish-overlay">
    <div
      v-if="open"
      class="publish-overlay fixed inset-0 z-40 flex justify-end bg-slate-950/45"
      @click.self="handleClose"
    >
      <aside
        class="publish-panel flex h-full w-full max-w-xl flex-col border-l bg-[var(--color-bg-elevated)]"
        role="dialog"
        aria-modal="true"
        aria-label="发布通知"
      >
          <header class="publish-header border-b px-6 py-4">
            <div class="text-sm font-semibold uppercase tracking-[0.22em] text-[var(--color-text-muted)]">
              Admin Actions
            </div>
            <div class="mt-2 flex items-center justify-between">
              <h3 class="text-2xl font-semibold text-[var(--color-text)]">发布通知</h3>
              <button type="button" class="publish-close-btn" @click="handleClose">关闭</button>
            </div>
            <p class="mt-2 text-sm text-[var(--color-text-muted)]">
              配置通知内容与受众范围，发布后会立即写入目标用户通知中心。
            </p>
          </header>

          <form class="publish-form flex-1 space-y-5 overflow-y-auto px-6 py-5" @submit.prevent="handleSubmit">
            <label class="publish-field">
              <span class="publish-label">通知类型</span>
              <select v-model="publisher.form.type" class="publish-input">
                <option value="system">系统</option>
                <option value="contest">竞赛</option>
                <option value="challenge">训练</option>
                <option value="team">团队</option>
              </select>
            </label>

            <label class="publish-field">
              <span class="publish-label">标题</span>
              <input
                v-model="publisher.form.title"
                type="text"
                class="publish-input"
                placeholder="例如：平台将于今晚进行维护"
              />
              <span v-if="publisher.errors.title" class="publish-error">{{ publisher.errors.title }}</span>
            </label>

            <label class="publish-field">
              <span class="publish-label">内容</span>
              <textarea
                v-model="publisher.form.content"
                rows="5"
                class="publish-input"
                placeholder="输入通知详情，支持纯文本。"
              />
              <span v-if="publisher.errors.content" class="publish-error">{{ publisher.errors.content }}</span>
            </label>

            <label class="publish-field">
              <span class="publish-label">跳转链接（可选）</span>
              <input
                v-model="publisher.form.link"
                type="text"
                class="publish-input"
                placeholder="例如：https://ctf.example.com/contests/12"
              />
            </label>

            <fieldset class="publish-field">
              <legend class="publish-label">发送范围</legend>
              <div class="publish-audience-grid mt-2">
                <label
                  v-for="item in [
                    { value: 'all', label: '所有用户' },
                    { value: 'role', label: '按角色' },
                    { value: 'class', label: '按班级' },
                    { value: 'user', label: '指定用户' },
                  ]"
                  :key="item.value"
                  class="publish-target-option"
                >
                  <input
                    type="radio"
                    name="audience-target"
                    :value="item.value"
                    :checked="publisher.audienceTarget.value === item.value"
                    @change="onAudienceChange(item.value as 'all' | 'role' | 'class' | 'user')"
                  />
                  <span>{{ item.label }}</span>
                </label>
              </div>

              <div v-if="publisher.audienceTarget.value === 'role'" class="publish-subsection mt-3">
                <label v-for="role in USER_ROLES" :key="role" class="publish-check-item">
                  <input
                    type="checkbox"
                    :checked="publisher.selectedRoles.value.includes(role)"
                    @change="toggleRole(role, ($event.target as HTMLInputElement).checked)"
                  />
                  <span>{{ role }}</span>
                </label>
              </div>

              <div v-else-if="publisher.audienceTarget.value === 'class'" class="publish-subsection mt-3">
                <div class="mb-2 flex items-center justify-between">
                  <span class="text-xs text-[var(--color-text-muted)]">班级候选来自 /academy/classes</span>
                  <button
                    type="button"
                    class="publish-inline-btn"
                    :disabled="publisher.loadingClasses.value"
                    @click="publisher.loadClasses"
                  >
                    {{ publisher.loadingClasses.value ? '加载中...' : '刷新班级' }}
                  </button>
                </div>
                <div class="publish-options-list">
                  <label v-for="item in publisher.classOptions.value" :key="item.name" class="publish-check-item">
                    <input
                      type="checkbox"
                      :checked="publisher.selectedClasses.value.includes(item.name)"
                      @change="toggleClass(item.name, ($event.target as HTMLInputElement).checked)"
                    />
                    <span>{{ item.name }}</span>
                  </label>
                  <div
                    v-if="publisher.classOptions.value.length === 0 && !publisher.loadingClasses.value"
                    class="publish-empty"
                  >
                    暂无班级数据，请点击“刷新班级”。
                  </div>
                </div>
              </div>

              <div v-else-if="publisher.audienceTarget.value === 'user'" class="publish-subsection mt-3">
                <div class="mb-2 flex gap-2">
                  <input
                    v-model="publisher.userKeyword.value"
                    type="text"
                    class="publish-input"
                    placeholder="输入用户名/学号搜索"
                    @keyup.enter.prevent="handleUserSearch"
                  />
                  <button
                    type="button"
                    class="publish-inline-btn"
                    :disabled="publisher.loadingUsers.value"
                    @click="handleUserSearch"
                  >
                    {{ publisher.loadingUsers.value ? '搜索中...' : '搜索' }}
                  </button>
                </div>
                <div class="publish-options-list">
                  <label v-for="user in publisher.userOptions.value" :key="user.id" class="publish-check-item">
                    <input
                      type="checkbox"
                      :checked="publisher.selectedUserIds.value.includes(user.id)"
                      @change="toggleUser(user.id, ($event.target as HTMLInputElement).checked)"
                    />
                    <span>{{ user.name || user.username }}（{{ user.username }}）</span>
                  </label>
                  <div
                    v-if="publisher.userOptions.value.length === 0 && !publisher.loadingUsers.value"
                    class="publish-empty"
                  >
                    输入关键词并搜索用户。
                  </div>
                </div>
              </div>

              <span v-if="publisher.errors.audience" class="publish-error mt-2">{{ publisher.errors.audience }}</span>
            </fieldset>
          </form>

          <footer class="publish-footer border-t px-6 py-4">
            <button type="button" class="publish-btn" @click="handleClose">取消</button>
            <button type="button" class="publish-btn publish-btn-primary" :disabled="publisher.submitting.value" @click="handleSubmit">
              {{ publisher.submitting.value ? '发布中...' : '确认发布' }}
            </button>
          </footer>
      </aside>
    </div>
  </Transition>
</template>

<style scoped>
.publish-panel {
  border-color: color-mix(in srgb, var(--color-border, #d8e1ec) 78%, transparent);
}

.publish-header,
.publish-footer {
  border-color: color-mix(in srgb, var(--color-border, #d8e1ec) 78%, transparent);
}

.publish-field {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.publish-label {
  font-size: 0.78rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  color: var(--color-text-muted, #64748b);
}

.publish-input {
  width: 100%;
  border: 1px solid color-mix(in srgb, var(--color-border, #d8e1ec) 82%, transparent);
  border-radius: 0.75rem;
  background: var(--color-bg-elevated, #fff);
  padding: 0.58rem 0.75rem;
  color: var(--color-text, #0f172a);
  font-size: 0.9rem;
}

.publish-input:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--color-primary, #3b82f6) 45%, transparent);
  outline-offset: 1px;
}

.publish-audience-grid {
  display: grid;
  gap: 0.5rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.publish-target-option,
.publish-check-item {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.88rem;
  color: var(--color-text, #0f172a);
}

.publish-subsection {
  border: 1px dashed color-mix(in srgb, var(--color-border, #d8e1ec) 80%, transparent);
  border-radius: 0.75rem;
  padding: 0.75rem;
}

.publish-options-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-height: 13rem;
  overflow-y: auto;
}

.publish-empty {
  font-size: 0.8rem;
  color: var(--color-text-muted, #64748b);
}

.publish-inline-btn,
.publish-btn,
.publish-close-btn {
  border: 1px solid color-mix(in srgb, var(--color-border, #d8e1ec) 80%, transparent);
  border-radius: 0.75rem;
  padding: 0.45rem 0.75rem;
  font-size: 0.82rem;
  color: var(--color-text, #0f172a);
  background: var(--color-bg-soft, var(--color-bg-elevated, var(--color-bg-surface)));
  cursor: pointer;
}

.publish-inline-btn:disabled,
.publish-btn:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.publish-btn {
  min-width: 6rem;
}

.publish-btn-primary {
  border-color: color-mix(in srgb, var(--color-primary, #3b82f6) 45%, transparent);
  background: color-mix(in srgb, var(--color-primary, #3b82f6) 14%, transparent);
  color: var(--color-primary, #3b82f6);
}

.publish-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.6rem;
}

.publish-error {
  font-size: 0.78rem;
  color: var(--color-danger, #dc2626);
}

.publish-overlay-enter-active,
.publish-overlay-leave-active {
  transition: opacity 0.2s ease;
}

.publish-overlay-enter-from,
.publish-overlay-leave-to {
  opacity: 0;
}

@media (max-width: 640px) {
  .publish-audience-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
