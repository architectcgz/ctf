<script setup lang="ts">
import { reactive, watch } from 'vue'

import SlideOverDrawer from '@/components/common/modal-templates/SlideOverDrawer.vue'
import type { PlatformAwdChallengeFormDraft } from '@/composables/usePlatformAwdChallenges'
import type { ChallengeCategory } from '@/api/contracts'

const categoryOptions: Array<{ value: ChallengeCategory; label: string }> = [
  { value: 'web', label: 'Web' },
  { value: 'pwn', label: 'Pwn' },
  { value: 'reverse', label: 'Reverse' },
  { value: 'crypto', label: 'Crypto' },
  { value: 'misc', label: 'Misc' },
  { value: 'forensics', label: 'Forensics' },
]

const props = defineProps<{
  open: boolean
  mode: 'create' | 'edit'
  draft: PlatformAwdChallengeFormDraft
  saving: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  save: [value: PlatformAwdChallengeFormDraft]
}>()

const localDraft = reactive<PlatformAwdChallengeFormDraft>({
  name: '',
  slug: '',
  category: 'web',
  difficulty: 'medium',
  description: '',
  service_type: 'web_http',
  deployment_mode: 'single_container',
  status: 'draft',
})

const fieldErrors = reactive<Partial<Record<keyof PlatformAwdChallengeFormDraft, string>>>({})

watch(
  () => props.draft,
  (draft) => {
    Object.assign(localDraft, draft)
    resetErrors()
  },
  { immediate: true, deep: true }
)

function resetErrors() {
  fieldErrors.name = ''
  fieldErrors.slug = ''
  fieldErrors.category = ''
  fieldErrors.description = ''
}

function closeDialog() {
  emit('update:open', false)
}

function validate(): boolean {
  resetErrors()

  if (!localDraft.name.trim()) {
    fieldErrors.name = '请填写题目名称'
  }

  if (!localDraft.slug.trim()) {
    fieldErrors.slug = '请填写题目 slug'
  } else if (!/^[a-z0-9-]+$/.test(localDraft.slug.trim())) {
    fieldErrors.slug = 'slug 仅支持小写字母、数字和中划线'
  }

  if (!localDraft.category.trim()) {
    fieldErrors.category = '请填写分类'
  }

  if (localDraft.description.trim().length > 5000) {
    fieldErrors.description = '描述不能超过 5000 个字符'
  }

  return !fieldErrors.name && !fieldErrors.slug && !fieldErrors.category && !fieldErrors.description
}

function handleSubmit() {
  if (props.saving) {
    return
  }

  if (!validate()) {
    return
  }

  emit('save', {
    name: localDraft.name,
    slug: localDraft.slug,
    category: localDraft.category,
    difficulty: localDraft.difficulty,
    description: localDraft.description,
    service_type: localDraft.service_type,
    deployment_mode: localDraft.deployment_mode,
    status: localDraft.status,
  })
}
</script>

<template>
  <SlideOverDrawer
    :open="open"
    :title="mode === 'create' ? '创建 AWD 题目' : '编辑 AWD 题目'"
    :subtitle="
      mode === 'create'
        ? '先登记 AWD 题目的基础服务属性，后续再继续补 checker、flag 和运行配置。'
        : '更新 AWD 题目的基础信息、部署方式和发布状态，保持 AWD 题库与比赛配置分离。'
    "
    eyebrow="AWD Service Library"
    width="42rem"
    @close="closeDialog"
    @update:open="emit('update:open', $event)"
  >
    <form
      class="awd-template-dialog"
      @submit.prevent="handleSubmit"
    >
      <div class="awd-template-dialog__grid">
        <label
          class="ui-field awd-template-dialog__field"
          for="awd-template-name"
        >
          <span class="ui-field__label">题目名称</span>
          <span
            class="ui-control-wrap"
            :class="{ 'is-error': !!fieldErrors.name }"
          >
            <input
              id="awd-template-name"
              v-model="localDraft.name"
              type="text"
              class="ui-control"
              placeholder="例如：Bank Portal AWD"
            >
          </span>
          <p
            v-if="fieldErrors.name"
            class="ui-field__error"
          >{{ fieldErrors.name }}</p>
        </label>

        <label
          class="ui-field awd-template-dialog__field"
          for="awd-template-slug"
        >
          <span class="ui-field__label">题目 slug</span>
          <span
            class="ui-control-wrap"
            :class="{ 'is-error': !!fieldErrors.slug }"
          >
            <input
              id="awd-template-slug"
              v-model="localDraft.slug"
              type="text"
              class="ui-control awd-template-dialog__mono"
              placeholder="bank-portal-awd"
            >
          </span>
          <p
            v-if="fieldErrors.slug"
            class="ui-field__error"
          >{{ fieldErrors.slug }}</p>
        </label>
      </div>

      <div class="awd-template-dialog__grid">
        <label
          class="ui-field awd-template-dialog__field"
          for="awd-template-category"
        >
          <span class="ui-field__label">分类</span>
          <span
            class="ui-control-wrap"
            :class="{ 'is-error': !!fieldErrors.category }"
          >
            <select
              id="awd-template-category"
              v-model="localDraft.category"
              class="ui-control"
            >
              <option
                v-for="option in categoryOptions"
                :key="option.value"
                :value="option.value"
              >
                {{ option.label }}
              </option>
            </select>
          </span>
          <p
            v-if="fieldErrors.category"
            class="ui-field__error"
          >{{ fieldErrors.category }}</p>
        </label>

        <label
          class="ui-field awd-template-dialog__field"
          for="awd-template-difficulty"
        >
          <span class="ui-field__label">难度</span>
          <span class="ui-control-wrap">
            <select
              id="awd-template-difficulty"
              v-model="localDraft.difficulty"
              class="ui-control"
            >
              <option value="beginner">beginner</option>
              <option value="easy">easy</option>
              <option value="medium">medium</option>
              <option value="hard">hard</option>
              <option value="insane">insane</option>
            </select>
          </span>
        </label>
      </div>

      <div class="awd-template-dialog__grid">
        <label
          class="ui-field awd-template-dialog__field"
          for="awd-template-service-type"
        >
          <span class="ui-field__label">服务类型</span>
          <span class="ui-control-wrap">
            <select
              id="awd-template-service-type"
              v-model="localDraft.service_type"
              class="ui-control"
            >
              <option value="web_http">web_http</option>
              <option value="binary_tcp">binary_tcp</option>
              <option value="multi_container">multi_container</option>
            </select>
          </span>
        </label>

        <label
          class="ui-field awd-template-dialog__field"
          for="awd-template-deployment"
        >
          <span class="ui-field__label">部署方式</span>
          <span class="ui-control-wrap">
            <select
              id="awd-template-deployment"
              v-model="localDraft.deployment_mode"
              class="ui-control"
            >
              <option value="single_container">single_container</option>
              <option value="topology">topology</option>
            </select>
          </span>
        </label>
      </div>

      <label
        class="ui-field awd-template-dialog__field awd-template-dialog__field--wide"
        for="awd-template-description"
      >
        <span class="ui-field__label">描述</span>
        <span
          class="ui-control-wrap"
          :class="{ 'is-error': !!fieldErrors.description }"
        >
          <textarea
            id="awd-template-description"
            v-model="localDraft.description"
            rows="5"
            class="ui-control awd-template-dialog__textarea"
            placeholder="概述服务的核心攻击面、运行方式和目标业务场景。"
          />
        </span>
        <p
          v-if="fieldErrors.description"
          class="ui-field__error"
        >{{ fieldErrors.description }}</p>
      </label>

      <label
        v-if="mode === 'edit'"
        class="ui-field awd-template-dialog__field awd-template-dialog__field--wide"
        for="awd-template-status"
      >
        <span class="ui-field__label">发布状态</span>
        <span class="ui-control-wrap">
          <select
            id="awd-template-status"
            v-model="localDraft.status"
            class="ui-control"
          >
            <option value="draft">draft</option>
            <option value="published">published</option>
            <option value="archived">archived</option>
          </select>
        </span>
      </label>
    </form>

    <template #footer>
      <div class="awd-template-drawer-footer">
        <button
          type="button"
          class="ui-btn ui-btn--secondary"
          @click="closeDialog"
        >
          取消
        </button>
        <button
          id="awd-template-dialog-submit"
          type="button"
          class="ui-btn ui-btn--primary"
          :disabled="saving"
          @click="handleSubmit"
        >
          {{ saving ? '保存中...' : mode === 'create' ? '创建题目' : '保存修改' }}
        </button>
      </div>
    </template>
  </SlideOverDrawer>
</template>

<style scoped>
.awd-template-dialog {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.awd-template-dialog__grid {
  display: grid;
  gap: 1.25rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.awd-template-dialog__field {
  --ui-field-gap: var(--space-2);
}

.awd-template-dialog__field--wide {
  width: 100%;
}

.awd-template-dialog__textarea {
  min-height: 10rem;
}

.awd-template-dialog__mono {
  font-family: var(--font-family-mono);
}

.awd-template-drawer-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  width: 100%;
}

@media (max-width: 768px) {
  .awd-template-dialog__grid {
    grid-template-columns: 1fr;
  }
}
</style>
