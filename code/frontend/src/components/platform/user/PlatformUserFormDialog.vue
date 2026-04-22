<script setup lang="ts">
import { reactive, watch } from 'vue'

import AdminSurfaceModal from '@/components/common/modal-templates/AdminSurfaceModal.vue'
import type { PlatformUserFormDraft } from '@/composables/usePlatformUsers'
import { USER_ROLES } from '@/utils/constants'

const props = defineProps<{
  open: boolean
  mode: 'create' | 'edit'
  draft: PlatformUserFormDraft
  saving: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  save: [value: PlatformUserFormDraft]
}>()

const localDraft = reactive<PlatformUserFormDraft>({
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

const fieldErrors = reactive<Partial<Record<keyof PlatformUserFormDraft, string>>>({})

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
  if (props.saving) {
    return
  }

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
  <AdminSurfaceModal
    :open="open"
    :title="mode === 'create' ? '创建用户' : '编辑用户'"
    :subtitle="
      mode === 'create'
        ? '填写账号、角色和身份字段，保存后即可进入用户目录继续治理。'
        : '更新用户基础资料、角色和状态，保持用户目录与班级上下文同步。'
    "
    eyebrow="User Governance"
    width="40rem"
    @close="closeDialog"
    @update:open="emit('update:open', $event)"
  >
    <form
      class="admin-user-form"
      @submit.prevent="handleSubmit"
    >
      <div class="admin-user-form__grid">
        <label
          class="ui-field admin-user-form__field"
          for="user-username"
        >
          <span class="ui-field__label">用户名</span>
          <span
            class="ui-control-wrap"
            :class="{
              'is-disabled': mode === 'edit',
              'is-error': !!fieldErrors.username,
            }"
          >
            <input
              id="user-username"
              v-model="localDraft.username"
              type="text"
              :disabled="mode === 'edit'"
              class="ui-control"
              placeholder="例如：alice"
            >
          </span>
          <p
            v-if="fieldErrors.username"
            class="ui-field__error"
          >{{ fieldErrors.username }}</p>
        </label>

        <label
          class="ui-field admin-user-form__field"
          for="user-name"
        >
          <span class="ui-field__label">姓名</span>
          <span
            class="ui-control-wrap"
            :class="{ 'is-error': !!fieldErrors.name }"
          >
            <input
              id="user-name"
              v-model="localDraft.name"
              type="text"
              class="ui-control"
              placeholder="例如：Alice Zhang"
            >
          </span>
          <p
            v-if="fieldErrors.name"
            class="ui-field__error"
          >{{ fieldErrors.name }}</p>
        </label>

        <label
          class="ui-field admin-user-form__field admin-user-form__field--wide"
          for="user-password"
        >
          <span class="ui-field__label">
            {{ mode === 'create' ? '初始密码' : '重置密码（可选）' }}
          </span>
          <span
            class="ui-control-wrap"
            :class="{ 'is-error': !!fieldErrors.password }"
          >
            <input
              id="user-password"
              v-model="localDraft.password"
              type="password"
              class="ui-control"
              :placeholder="mode === 'create' ? '至少 8 位' : '留空则保持不变'"
            >
          </span>
          <p
            v-if="fieldErrors.password"
            class="ui-field__error"
          >{{ fieldErrors.password }}</p>
        </label>
      </div>

      <div class="admin-user-form__grid">
        <label
          class="ui-field admin-user-form__field"
          for="user-email"
        >
          <span class="ui-field__label">邮箱</span>
          <span
            class="ui-control-wrap"
            :class="{ 'is-error': !!fieldErrors.email }"
          >
            <input
              id="user-email"
              v-model="localDraft.email"
              type="email"
              class="ui-control"
              placeholder="user@example.com"
            >
          </span>
          <p
            v-if="fieldErrors.email"
            class="ui-field__error"
          >{{ fieldErrors.email }}</p>
        </label>

        <label
          class="ui-field admin-user-form__field"
          :for="localDraft.role === 'teacher' ? 'user-teacher-no' : 'user-student-no'"
        >
          <span class="ui-field__label">
            {{ localDraft.role === 'teacher' ? '教师工号' : '学生学号' }}
          </span>
          <span class="ui-control-wrap">
            <input
              :id="localDraft.role === 'teacher' ? 'user-teacher-no' : 'user-student-no'"
              :value="localDraft.role === 'teacher' ? localDraft.teacher_no : localDraft.student_no"
              type="text"
              class="ui-control"
              :placeholder="localDraft.role === 'teacher' ? '例如：T2024001' : '例如：20240001'"
              @input="
                localDraft.role === 'teacher'
                  ? (localDraft.teacher_no = ($event.target as HTMLInputElement).value)
                  : (localDraft.student_no = ($event.target as HTMLInputElement).value)
              "
            >
          </span>
        </label>
      </div>

      <div class="admin-user-form__grid admin-user-form__grid--single">
        <label
          class="ui-field admin-user-form__field"
          for="user-class-name"
        >
          <span class="ui-field__label">班级</span>
          <span class="ui-control-wrap">
            <input
              id="user-class-name"
              v-model="localDraft.class_name"
              type="text"
              class="ui-control"
              placeholder="例如：Class A"
            >
          </span>
        </label>
      </div>

      <div class="admin-user-form__grid">
        <label
          class="ui-field admin-user-form__field"
          for="user-role"
        >
          <span class="ui-field__label">角色</span>
          <span class="ui-control-wrap">
            <select
              id="user-role"
              v-model="localDraft.role"
              class="ui-control"
            >
              <option
                v-for="role in USER_ROLES"
                :key="role"
                :value="role"
              >
                {{ role }}
              </option>
            </select>
          </span>
          <p class="ui-field__hint">
            `student` 仅保留学号，`teacher` 仅保留工号，其他角色会忽略这两个字段。
          </p>
        </label>

        <label
          class="ui-field admin-user-form__field"
          for="user-status"
        >
          <span class="ui-field__label">状态</span>
          <span class="ui-control-wrap">
            <select
              id="user-status"
              v-model="localDraft.status"
              class="ui-control"
            >
              <option value="active">active</option>
              <option value="inactive">inactive</option>
              <option value="locked">locked</option>
              <option value="banned">banned</option>
            </select>
          </span>
        </label>
      </div>
    </form>

    <template #footer>
      <div class="admin-user-form__footer">
        <button
          type="button"
          class="ui-btn ui-btn--secondary"
          @click="closeDialog"
        >
          取消
        </button>
        <button
          type="button"
          class="ui-btn ui-btn--primary"
          :disabled="saving"
          @click="handleSubmit"
        >
          {{ saving ? '保存中...' : '保存用户' }}
        </button>
      </div>
    </template>
  </AdminSurfaceModal>
</template>

<style scoped>
.admin-user-form {
  display: grid;
  gap: var(--space-5);
}

.admin-user-form__grid {
  display: grid;
  gap: var(--space-4);
}

.admin-user-form__field {
  --ui-field-gap: var(--space-2);
}

.admin-user-form__field--wide {
  grid-column: 1 / -1;
}

.admin-user-form__footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-2);
}

@media (min-width: 640px) {
  .admin-user-form__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .admin-user-form__grid--single {
    grid-template-columns: minmax(0, 1fr);
  }
}

@media (max-width: 639px) {
  .admin-user-form__footer {
    flex-direction: column-reverse;
  }

  .admin-user-form__footer > .ui-btn {
    width: 100%;
  }
}
</style>
