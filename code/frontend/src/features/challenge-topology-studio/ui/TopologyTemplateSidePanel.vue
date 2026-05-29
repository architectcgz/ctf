<script setup lang="ts">
import { computed } from 'vue'
import { Layout, Network, Plus, RefreshCw, Server, Trash2 } from 'lucide-vue-next'

import type { EnvironmentTemplateData } from '@/api/contracts'
import AppCard from '@/components/common/AppCard.vue'
import SectionCard from '@/components/common/SectionCard.vue'

const props = defineProps<{
  isTemplateLibraryMode: boolean
  selectedTemplateSummary: string
  selectedTemplateId: string | null
  templates: EnvironmentTemplateData[]
  templateKeyword: string
  templateName: string
  templateDescription: string
  templateBusy: boolean
}>()

const emit = defineEmits<{
  'update:templateKeyword': [value: string]
  'update:templateName': [value: string]
  'update:templateDescription': [value: string]
  loadTemplate: [template: EnvironmentTemplateData]
  clearTemplateSelection: []
  searchTemplates: []
  resetTemplateForm: [template: EnvironmentTemplateData]
  applyTemplate: [template: EnvironmentTemplateData]
  deleteTemplate: [templateId: string]
  resetTemplateEditor: []
  createTemplate: []
  updateTemplate: []
}>()

const rootClasses = computed(() =>
  props.isTemplateLibraryMode
    ? 'topology-side-stack topology-side-stack--library topology-template-side-panel'
    : 'topology-side-stack topology-template-side-panel space-y-6'
)
const selectedTemplate = computed(
  () => props.templates.find((template) => template.id === props.selectedTemplateId) || null
)

const focusCardClasses = computed(() =>
  props.isTemplateLibraryMode
    ? 'template-focus-card'
    : 'rounded-2xl border border-border bg-elevated px-4 py-4'
)

const searchRowClasses = computed(() =>
  props.isTemplateLibraryMode ? 'template-search-row' : 'grid gap-3 md:grid-cols-[1fr_auto]'
)

const emptyStateClasses = computed(() =>
  props.isTemplateLibraryMode
    ? 'template-empty-state'
    : 'rounded-xl border border-dashed border-border px-4 py-6 text-sm text-text-muted'
)

const templateListClasses = computed(() =>
  props.isTemplateLibraryMode ? 'template-library-list' : 'space-y-3'
)

const writebackFormClasses = computed(() =>
  props.isTemplateLibraryMode ? 'template-writeback-form' : 'space-y-4'
)

const boundaryListClasses = computed(() =>
  props.isTemplateLibraryMode ? 'template-boundary-list' : 'space-y-4'
)

function templateItemClasses(template: EnvironmentTemplateData) {
  if (props.isTemplateLibraryMode) {
    return [
      'template-library-item',
      props.selectedTemplateId === template.id
        ? 'template-library-item--active'
        : 'template-library-item--idle',
    ]
  }

  return [
    'rounded-2xl border p-4 transition',
    props.selectedTemplateId === template.id
      ? 'border-primary bg-primary/8'
      : 'border-border bg-elevated',
  ]
}

function templateActionClass(variant: 'default' | 'primary' | 'danger' = 'default') {
  if (!props.isTemplateLibraryMode) {
    const variantClass =
      variant === 'primary'
        ? 'ui-btn--primary'
        : variant === 'danger'
          ? 'ui-btn--danger'
          : 'ui-btn--secondary'
    return `ui-btn ui-btn--sm ${variantClass}`
  }

  if (variant === 'primary') return 'template-action-btn template-action-btn--primary'
  if (variant === 'danger') return 'template-action-btn template-action-btn--danger'
  return 'template-action-btn'
}

function reloadSelectedTemplate() {
  if (!selectedTemplate.value) return
  emit('loadTemplate', selectedTemplate.value)
}
</script>

<template>
  <div :class="rootClasses">
    <SectionCard
      title="模板库"
      :subtitle="
        isTemplateLibraryMode
          ? '从模板库载入后可直接编辑并覆盖模板，或另存为新模板。'
          : '可按模板快速回填编辑器，或直接应用到题目。'
      "
    >
      <div class="space-y-3">
        <div :class="focusCardClasses">
          <div class="text-xs font-semibold uppercase tracking-[0.22em] text-text-muted">
            当前模板
          </div>
          <div class="mt-2 text-sm text-text-primary">
            {{ selectedTemplateSummary }}
          </div>
          <div class="mt-3 flex flex-wrap gap-2">
            <button
              v-if="selectedTemplate"
              type="button"
              :class="
                isTemplateLibraryMode
                  ? 'ui-btn ui-btn--secondary topology-action-btn'
                  : templateActionClass()
              "
              @click="reloadSelectedTemplate"
            >
              重新载入当前模板
            </button>
            <button
              v-if="selectedTemplateId"
              type="button"
              :class="
                isTemplateLibraryMode
                  ? 'ui-btn ui-btn--secondary topology-action-btn'
                  : templateActionClass()
              "
              @click="emit('clearTemplateSelection')"
            >
              清空模板选择
            </button>
          </div>
        </div>

        <div :class="searchRowClasses">
          <input
            :value="templateKeyword"
            type="text"
            class="w-full rounded-xl border border-border bg-elevated px-3 py-3 text-sm text-text-primary outline-none transition focus:border-primary"
            placeholder="按模板名称搜索"
            @input="emit('update:templateKeyword', ($event.target as HTMLInputElement).value)"
          />
          <button
            type="button"
            :class="
              isTemplateLibraryMode
                ? 'ui-btn ui-btn--secondary topology-action-btn'
                : 'ui-btn ui-btn--secondary'
            "
            @click="emit('searchTemplates')"
          >
            搜索
          </button>
        </div>

        <div v-if="templates.length === 0" :class="emptyStateClasses">当前没有模板数据</div>

        <div v-else :class="templateListClasses">
          <div v-if="isTemplateLibraryMode" class="template-directory-head" aria-hidden="true">
            <span>模板</span>
            <span>概况</span>
            <span>操作</span>
          </div>

          <article
            v-for="template in templates"
            :key="template.id"
            :class="templateItemClasses(template)"
          >
            <div :class="isTemplateLibraryMode ? 'template-library-item__main' : 'min-w-0'">
              <div
                :class="
                  isTemplateLibraryMode
                    ? 'truncate text-base font-bold text-text-primary'
                    : 'truncate text-base font-semibold text-text-primary'
                "
              >
                {{ template.name }}
              </div>
              <div
                :class="
                  isTemplateLibraryMode
                    ? 'mt-1 line-clamp-2 text-xs leading-relaxed text-text-secondary'
                    : 'mt-1 text-sm text-text-secondary'
                "
              >
                {{ template.description || '无描述' }}
              </div>
              <div
                :class="
                  isTemplateLibraryMode
                    ? 'mt-3 flex flex-wrap gap-x-3 gap-y-1.5 text-[10px] font-bold uppercase tracking-wider text-text-muted'
                    : 'mt-2 flex flex-wrap gap-2 text-xs text-text-muted'
                "
              >
                <template v-if="isTemplateLibraryMode">
                  <span class="flex items-center gap-1">
                    <Layout class="h-3 w-3" />
                    {{ template.entry_node_key }}
                  </span>
                  <span class="flex items-center gap-1">
                    <Server class="h-3 w-3" />
                    {{ template.nodes.length }}
                  </span>
                  <span class="flex items-center gap-1">
                    <Network class="h-3 w-3" />
                    {{ template.networks?.length || 0 }}
                  </span>
                  <span class="flex items-center gap-1">
                    <RefreshCw class="h-3 w-3" />
                    {{ template.usage_count }}
                  </span>
                </template>
                <template v-else>
                  <span>入口：{{ template.entry_node_key }}</span>
                  <span>节点：{{ template.nodes.length }}</span>
                  <span>网络：{{ template.networks?.length || 0 }}</span>
                  <span>使用：{{ template.usage_count }}</span>
                </template>
              </div>
            </div>

            <div v-if="isTemplateLibraryMode" class="template-library-item__meta">
              <span>入口 {{ template.entry_node_key }}</span>
              <span>{{ template.nodes.length }} 节点</span>
              <span>{{ template.networks?.length || 0 }} 网络</span>
              <span>使用 {{ template.usage_count }}</span>
            </div>

            <div
              :class="
                isTemplateLibraryMode
                  ? 'template-library-item__actions'
                  : 'mt-4 flex flex-wrap gap-2'
              "
            >
              <button
                type="button"
                :class="
                  isTemplateLibraryMode
                    ? 'ui-btn ui-btn--secondary topology-action-btn'
                    : templateActionClass()
                "
                @click="emit('loadTemplate', template)"
              >
                {{ isTemplateLibraryMode ? '载入编辑' : '载入草稿' }}
              </button>
              <button
                v-if="!isTemplateLibraryMode"
                type="button"
                :class="templateActionClass()"
                @click="emit('resetTemplateForm', template)"
              >
                选中
              </button>
              <button
                v-if="!isTemplateLibraryMode"
                type="button"
                :class="templateActionClass('primary')"
                :disabled="templateBusy"
                @click="emit('applyTemplate', template)"
              >
                应用到题目
              </button>
              <button
                type="button"
                :class="
                  isTemplateLibraryMode
                    ? 'ui-btn ui-btn--danger topology-action-btn'
                    : templateActionClass('danger')
                "
                :disabled="templateBusy"
                @click="emit('deleteTemplate', template.id)"
              >
                <Trash2 v-if="isTemplateLibraryMode" class="h-3 w-3" />
                <span v-else>删除模板</span>
              </button>
            </div>
          </article>
        </div>
      </div>
    </SectionCard>

    <SectionCard
      title="模板写回"
      :subtitle="
        isTemplateLibraryMode
          ? '在独立模板库中可新建空白草稿、载入现有模板后覆盖，或另存为新模板。'
          : '把当前编辑器草稿保存为新模板，或覆盖已选中的模板。'
      "
    >
      <div :class="writebackFormClasses">
        <label class="space-y-2">
          <span class="text-sm text-text-secondary">模板名称</span>
          <input
            :value="templateName"
            type="text"
            class="w-full rounded-xl border border-border bg-elevated px-3 py-3 text-sm text-text-primary outline-none transition focus:border-primary"
            placeholder="例如 双节点 Web + DB"
            @input="emit('update:templateName', ($event.target as HTMLInputElement).value)"
          />
        </label>

        <label class="space-y-2">
          <span class="text-sm text-text-secondary">模板描述</span>
          <textarea
            :value="templateDescription"
            rows="4"
            class="w-full rounded-xl border border-border bg-elevated px-3 py-3 text-sm text-text-primary outline-none transition focus:border-primary"
            placeholder="说明这个模板的适用场景"
            @input="
              emit('update:templateDescription', ($event.target as HTMLTextAreaElement).value)
            "
          />
        </label>

        <div class="flex flex-wrap gap-2">
          <button
            v-if="isTemplateLibraryMode"
            type="button"
            :class="templateActionClass()"
            @click="emit('resetTemplateEditor')"
          >
            新建空白模板
          </button>
          <button
            type="button"
            :class="
              isTemplateLibraryMode
                ? 'template-action-btn template-action-btn--primary'
                : 'ui-btn ui-btn--primary topology-action-btn'
            "
            :disabled="templateBusy"
            @click="emit('createTemplate')"
          >
            <Plus class="h-4 w-4" />
            {{ isTemplateLibraryMode ? '另存为' : '保存为新模板' }}
          </button>
          <button
            type="button"
            :class="
              isTemplateLibraryMode
                ? 'template-action-btn'
                : 'ui-btn ui-btn--ghost topology-action-btn'
            "
            :disabled="templateBusy || !selectedTemplateId"
            @click="emit('updateTemplate')"
          >
            {{ isTemplateLibraryMode ? '覆盖' : '覆盖已选模板' }}
          </button>
        </div>
      </div>
    </SectionCard>

    <SectionCard title="当前边界" subtitle="避免把未生效能力继续暴露成可用配置。">
      <div :class="boundaryListClasses">
        <template v-if="isTemplateLibraryMode">
          <article class="template-boundary-item template-boundary-item--warning">
            <div class="template-boundary-item__label">已开放</div>
            <div class="template-boundary-item__copy">
              多网络、节点、逻辑连线、粗粒度 allow/deny 策略、模板复用。
            </div>
          </article>
          <article class="template-boundary-item template-boundary-item--danger">
            <div class="template-boundary-item__label">暂未开放</div>
            <div class="template-boundary-item__copy">
              protocol / ports 级细粒度 ACL 前端字段、模板版本化与批量比对能力。
            </div>
          </article>
          <article class="template-boundary-item template-boundary-item--neutral">
            <div class="template-boundary-item__label">建议</div>
            <div class="template-boundary-item__copy">
              继续开放高级能力前，先补参数校验、可视化提示和误操作保护。
            </div>
          </article>
        </template>

        <template v-else>
          <AppCard
            variant="action"
            accent="warning"
            eyebrow="已开放"
            subtitle="多网络、节点、逻辑连线、粗粒度 allow/deny 策略、模板复用。"
          >
            <template #default />
          </AppCard>
          <AppCard
            variant="action"
            accent="danger"
            eyebrow="暂未开放"
            subtitle="protocol / ports 级细粒度 ACL 前端字段、模板版本化与批量比对能力。"
          >
            <template #default />
          </AppCard>
          <AppCard
            variant="action"
            accent="neutral"
            eyebrow="建议"
            subtitle="继续开放高级能力前，先补参数校验、可视化提示和误操作保护。"
          >
            <template #default />
          </AppCard>
        </template>
      </div>
    </SectionCard>
  </div>
</template>

<style scoped>
.topology-side-stack--library {
  border: 0;
  background: transparent;
  box-shadow: none;
  padding: 0;
}

.topology-side-stack--library :deep(.section-card:first-child) {
  padding-top: 0;
  border-top: 0;
}

.template-focus-card,
.template-empty-state {
  padding: 0 0 0 var(--space-4);
  border: 0;
  border-left: 2px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

.template-search-row {
  display: grid;
  gap: var(--space-3);
  grid-template-columns: minmax(0, 1fr) auto;
}

.template-library-list,
.template-writeback-form,
.template-boundary-list {
  display: grid;
  gap: 0;
}

.template-directory-head {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(0, 0.95fr) auto;
  gap: var(--space-4);
  padding: 0 0 var(--space-3);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.template-library-item {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(0, 0.95fr) auto;
  align-items: start;
  gap: var(--space-4);
  border-radius: 0;
  padding: var(--space-4) 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  box-shadow: none;
  transition:
    border-color 150ms ease,
    background 150ms ease,
    color 150ms ease;
}

.template-library-item--idle {
  border-left: 0;
  background: transparent;
}

.template-library-item--idle:hover {
  background: color-mix(in srgb, var(--journal-accent) 4%, transparent);
}

.template-library-item--active {
  border-left: 2px solid color-mix(in srgb, var(--journal-accent) 58%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 6%, transparent);
}

.template-library-item__main {
  min-width: 0;
}

.template-library-item__meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2) var(--space-3);
  align-content: start;
  padding-top: var(--space-0-5);
  font-size: var(--font-size-0-76);
  line-height: 1.6;
  color: var(--journal-muted);
}

.template-library-item__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--space-2);
}

.template-action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: 2.45rem;
  border-radius: 0.75rem;
  border: 1px solid var(--journal-border);
  padding: var(--space-2) var(--space-4);
  background: color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base));
  color: var(--journal-ink);
  font-size: var(--font-size-0-84);
  font-weight: 600;
  transition:
    border-color 150ms ease,
    background-color 150ms ease,
    color 150ms ease;
}

.template-action-btn:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
  color: var(--journal-accent);
}

.template-action-btn--primary {
  border-color: transparent;
  background: var(--journal-accent);
  color: var(--color-bg-base);
}

.template-action-btn--primary:hover {
  background: color-mix(in srgb, var(--journal-accent) 88%, var(--color-bg-base));
}

.template-action-btn--danger {
  border-color: color-mix(in srgb, var(--color-danger) 28%, transparent);
  background: color-mix(in srgb, var(--color-danger) 10%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
}

.template-action-btn--danger:hover {
  border-color: color-mix(in srgb, var(--color-danger) 34%, transparent);
  background: color-mix(in srgb, var(--color-danger) 14%, var(--journal-surface));
}

.template-boundary-item {
  display: grid;
  gap: var(--space-1-5);
  padding: var(--space-4) 0 var(--space-4) var(--space-4);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-left: 2px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.template-boundary-item__label {
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.template-boundary-item__copy {
  font-size: var(--font-size-0-88);
  line-height: 1.65;
  color: var(--journal-muted);
}

.template-boundary-item--warning .template-boundary-item__label {
  color: var(--color-warning);
}

.template-boundary-item--danger .template-boundary-item__label {
  color: var(--color-danger);
}

.template-boundary-item--neutral .template-boundary-item__label {
  color: var(--journal-accent);
}

@media (max-width: 1023px) {
  .template-directory-head {
    display: none;
  }

  .template-library-item {
    grid-template-columns: minmax(0, 1fr);
  }

  .template-library-item__actions {
    justify-content: flex-start;
  }
}

@media (max-width: 767px) {
  .template-search-row {
    grid-template-columns: 1fr;
  }
}
</style>
