<script setup lang="ts">
import { reactive, watch } from 'vue'

import type { AdminUserFormDraft } from '@/composables/useAdminUsers'
import { USER_ROLES } from '@/utils/constants'

const props = defineProps<{
  open: boolean
  mode: 'create' | 'edit'
  draft: AdminUserFormDraft
  saving: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  save: [value: AdminUserFormDraft]
}>()

const localDraft = reactive<AdminUserFormDraft>({
  username: '',
  name: '',
  password: '',
  email: '',
  student_no: '',
  teacher_no: '',
  class_name: '',
  role: 'student',
  status: 'active',
})

const fieldErrors = reactive<Partial<Record<keyof AdminUserFormDraft, string>>>({})

watch(
  () => props.draft,
  (draft) => {
    Object.assign(localDraft, draft)
    resetErrors()
  },
  { immediate: true, deep: true }
)

function resetErrors() {
  fieldErrors.username = ''
  fieldErrors.name = ''
  fieldErrors.password = ''
  fieldErrors.email = ''
}

function closeDialog() {
  emit('update:open', false)
}

function validate(): boolean {
  resetErrors()

  if (!localDraft.username.trim()) {
    fieldErrors.username = '请填写用户名'
  }

  if (localDraft.name.trim().length > 64) {
    fieldErrors.name = '姓名不能超过 64 个字符'
  }

  if (props.mode === 'create' && !localDraft.password.trim()) {
    fieldErrors.password = '创建用户时必须设置初始密码'
  }

  if (localDraft.email.trim() && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(localDraft.email.trim())) {
    fieldErrors.email = '邮箱格式不正确'
  }

  return !fieldErrors.username && !fieldErrors.name && !fieldErrors.password && !fieldErrors.email
}

function handleSubmit() {
  if (!validate()) {
    return
  }

  emit('save', {
    username: localDraft.username,
    name: localDraft.name,
    password: localDraft.password,
    email: localDraft.email,
    student_no: localDraft.student_no,
    teacher_no: localDraft.teacher_no,
    class_name: localDraft.class_name,
    role: localDraft.role,
    status: localDraft.status,
  })
}
</script>

<template>
  <ElDialog
    :model-value="open"
    :title="mode === 'create' ? '创建用户' : '编辑用户'"
    width="640px"
    @close="closeDialog"
    @update:model-value="emit('update:open', $event)"
  >
    <form class="space-y-5" @submit.prevent="handleSubmit">
      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <label class="text-sm font-medium text-[var(--color-text-primary)]" for="user-username">用户名</label>
          <input
            id="user-username"
            v-model="localDraft.username"
            type="text"
            :disabled="mode === 'edit'"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary disabled:cursor-not-allowed disabled:opacity-60"
            placeholder="例如：alice"
          />
          <p v-if="fieldErrors.username" class="text-xs text-[var(--color-danger)]">
            {{ fieldErrors.username }}
          </p>
        </div>

        <div class="space-y-2">
          <label class="text-sm font-medium text-[var(--color-text-primary)]" for="user-name">姓名</label>
          <input
            id="user-name"
            v-model="localDraft.name"
            type="text"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
            placeholder="例如：Alice Zhang"
          />
          <p v-if="fieldErrors.name" class="text-xs text-[var(--color-danger)]">
            {{ fieldErrors.name }}
          </p>
        </div>

        <div class="space-y-2 sm:col-span-2">
          <label class="text-sm font-medium text-[var(--color-text-primary)]" for="user-password">
            {{ mode === 'create' ? '初始密码' : '重置密码（可选）' }}
          </label>
          <input
            id="user-password"
            v-model="localDraft.password"
            type="password"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
            :placeholder="mode === 'create' ? '至少 8 位' : '留空则保持不变'"
          />
          <p v-if="fieldErrors.password" class="text-xs text-[var(--color-danger)]">
            {{ fieldErrors.password }}
          </p>
        </div>
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <label class="text-sm font-medium text-[var(--color-text-primary)]" for="user-email">邮箱</label>
          <input
            id="user-email"
            v-model="localDraft.email"
            type="email"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
            placeholder="user@example.com"
          />
          <p v-if="fieldErrors.email" class="text-xs text-[var(--color-danger)]">{{ fieldErrors.email }}</p>
        </div>

        <div class="space-y-2">
          <label
            class="text-sm font-medium text-[var(--color-text-primary)]"
            :for="localDraft.role === 'teacher' ? 'user-teacher-no' : 'user-student-no'"
          >
            {{ localDraft.role === 'teacher' ? '教师工号' : '学生学号' }}
          </label>
          <input
            :id="localDraft.role === 'teacher' ? 'user-teacher-no' : 'user-student-no'"
            :value="localDraft.role === 'teacher' ? localDraft.teacher_no : localDraft.student_no"
            type="text"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
            :placeholder="localDraft.role === 'teacher' ? '例如：T2024001' : '例如：20240001'"
            @input="
              localDraft.role === 'teacher'
                ? (localDraft.teacher_no = ($event.target as HTMLInputElement).value)
                : (localDraft.student_no = ($event.target as HTMLInputElement).value)
            "
          />
        </div>
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <label class="text-sm font-medium text-[var(--color-text-primary)]" for="user-class-name">班级</label>
          <input
            id="user-class-name"
            v-model="localDraft.class_name"
            type="text"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
            placeholder="例如：Class A"
          />
        </div>
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <label class="text-sm font-medium text-[var(--color-text-primary)]" for="user-role">角色</label>
          <select
            id="user-role"
            v-model="localDraft.role"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          >
            <option v-for="role in USER_ROLES" :key="role" :value="role">
              {{ role }}
            </option>
          </select>
          <p class="text-xs text-[var(--color-text-muted)]">
            `student` 仅保留学号，`teacher` 仅保留工号，其他角色会忽略这两个字段。
          </p>
        </div>

        <div class="space-y-2">
          <label class="text-sm font-medium text-[var(--color-text-primary)]" for="user-status">状态</label>
          <select
            id="user-status"
            v-model="localDraft.status"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          >
            <option value="active">active</option>
            <option value="inactive">inactive</option>
            <option value="locked">locked</option>
            <option value="banned">banned</option>
          </select>
        </div>
      </div>
    </form>

    <template #footer>
      <div class="flex items-center justify-end gap-2">
        <button
          type="button"
          class="rounded-xl border border-border px-4 py-2 text-sm text-[var(--color-text-primary)] transition hover:border-primary"
          @click="closeDialog"
        >
          取消
        </button>
        <button
          type="button"
          class="rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
          :disabled="saving"
          @click="handleSubmit"
        >
          {{ saving ? '保存中...' : '保存用户' }}
        </button>
      </div>
    </template>
  </ElDialog>
</template>
