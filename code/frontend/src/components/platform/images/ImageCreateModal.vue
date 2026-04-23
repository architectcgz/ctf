<script setup lang="ts">
import type { AdminImagePayload } from '@/api/admin'
import AdminSurfaceModal from '@/components/common/modal-templates/AdminSurfaceModal.vue'

const props = defineProps<{
  open: boolean
  creating: boolean
  form: AdminImagePayload
}>()

const emit = defineEmits<{
  close: []
  'update:open': [value: boolean]
  'update:name': [value: string]
  'update:tag': [value: string]
  'update:description': [value: string]
  submit: []
}>()
</script>

<template>
  <AdminSurfaceModal
    class="image-create-modal"
    :open="open"
    :frosted="true"
    title="创建镜像"
    subtitle="填写镜像名称、标签和说明，提交后会进入镜像目录并参与构建状态跟踪。"
    eyebrow="Image Registry"
    width="31.25rem"
    @close="emit('close')"
    @update:open="emit('update:open', $event)"
  >
    <form
      class="image-create-form"
      @submit.prevent="emit('submit')"
    >
      <label class="ui-field image-create-field">
        <span class="ui-field__label">
          镜像名称
          <span
            class="ui-field__required"
            aria-hidden="true"
          >*</span>
        </span>
        <span class="ui-control-wrap">
          <input
            :value="form.name"
            type="text"
            class="ui-control"
            placeholder="例如：ubuntu"
            @input="emit('update:name', ($event.target as HTMLInputElement).value)"
          >
        </span>
      </label>

      <label class="ui-field image-create-field">
        <span class="ui-field__label">
          标签
          <span
            class="ui-field__required"
            aria-hidden="true"
          >*</span>
        </span>
        <span class="ui-control-wrap">
          <input
            :value="form.tag"
            type="text"
            class="ui-control"
            placeholder="例如：22.04"
            @input="emit('update:tag', ($event.target as HTMLInputElement).value)"
          >
        </span>
      </label>

      <label class="ui-field image-create-field">
        <span class="ui-field__label">描述</span>
        <span class="ui-control-wrap image-create-field__textarea">
          <textarea
            :value="form.description"
            class="ui-control"
            rows="3"
            placeholder="镜像说明（可选）"
            @input="emit('update:description', ($event.target as HTMLTextAreaElement).value)"
          />
        </span>
      </label>
    </form>
    <template #footer>
      <div class="image-create-dialog__footer">
        <button
          type="button"
          class="ui-btn ui-btn--secondary"
          @click="emit('close')"
        >
          取消
        </button>
        <button
          type="button"
          :disabled="creating"
          class="ui-btn ui-btn--primary"
          @click="emit('submit')"
        >
          {{ creating ? '创建中...' : '创建' }}
        </button>
      </div>
    </template>
  </AdminSurfaceModal>
</template>

<style scoped>
.image-create-form {
  display: grid;
  gap: var(--space-4);
}

.image-create-field {
  --ui-field-gap: var(--space-2);
}

.image-create-field__textarea {
  align-items: stretch;
}

.image-create-field__textarea .ui-control {
  min-height: 6.25rem;
  resize: vertical;
}

.image-create-dialog__footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-2);
}

@media (max-width: 720px) {
  .image-create-dialog__footer {
    flex-direction: column-reverse;
  }

  .image-create-dialog__footer > .ui-btn {
    width: 100%;
  }
}
</style>
